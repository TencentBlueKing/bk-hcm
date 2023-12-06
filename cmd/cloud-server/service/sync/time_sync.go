/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

package sync

import (
	"fmt"
	"sync"
	"time"

	"hcm/cmd/cloud-server/service/sync/aws"
	"hcm/cmd/cloud-server/service/sync/azure"
	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/cmd/cloud-server/service/sync/gcp"
	"hcm/cmd/cloud-server/service/sync/huawei"
	"hcm/cmd/cloud-server/service/sync/tcloud"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/retry"
)

// CloudResourceSync 定时同步云资源
func CloudResourceSync(intervalMin time.Duration, sd serviced.ServiceDiscover, cliSet *client.ClientSet) {
	logs.Infof("cloud resource sync enable, syncIntervalMin: %v", intervalMin)

	for {
		time.Sleep(intervalMin)

		if !sd.IsMaster() {
			continue
		}

		start := time.Now()
		logs.Infof("cloud resource all sync start, time: %v", start)

		waitGroup := new(sync.WaitGroup)

		vendors := []enumor.Vendor{enumor.TCloud, enumor.Aws, enumor.HuaWei, enumor.Azure, enumor.Gcp}
		waitGroup.Add(len(vendors))
		for _, vendor := range vendors {
			go func(vendor enumor.Vendor) {
				allAccountSync(core.NewBackendKit(), cliSet, vendor)
				waitGroup.Done()
			}(vendor)
		}

		waitGroup.Wait()

		logs.Infof("cloud resource all sync end, time: %v", start)
	}
}

// allAccountSync all account sync.
func allAccountSync(kt *kit.Kit, cliSet *client.ClientSet, vendor enumor.Vendor) {

	startTime := time.Now()
	logs.Infof("%s start sync all cloud resource, time: %v, rid: %s", vendor, startTime, kt.Rid)

	defer func() {
		logs.Infof("%s sync all cloud resource end, cost: %v, rid: %s", vendor, time.Since(startTime), kt.Rid)
	}()

	listReq := &protocloud.AccountListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: []filter.RuleFactory{
			&filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: vendor},
			&filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: enumor.ResourceAccount}}},
		Page: &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	start := uint32(0)
	syncPublicResource := true
	for {
		listReq.Page.Start = start
		accounts, err := listAccountWithRetry(kt, cliSet.DataService(), listReq)
		if err != nil {
			logs.Errorf("list account failed, err: %v, rid: %s", err, kt.Rid)
			break
		}

		for _, one := range accounts {
			resName := enumor.CloudResourceType("")
			sd := &detail.SyncDetail{
				Kt:        kt,
				DataCli:   cliSet.DataService(),
				AccountID: one.ID,
				Vendor:    string(one.Vendor),
			}

			switch one.Vendor {
			case enumor.TCloud:
				opt := &tcloud.SyncAllResourceOption{AccountID: one.ID, SyncPublicResource: syncPublicResource}
				resName, err = tcloud.SyncAllResource(kt, cliSet, opt)

			case enumor.Aws:
				opt := &aws.SyncAllResourceOption{AccountID: one.ID, SyncPublicResource: syncPublicResource}
				resName, err = aws.SyncAllResource(kt, cliSet, opt)

			case enumor.HuaWei:
				opt := &huawei.SyncAllResourceOption{AccountID: one.ID, SyncPublicResource: syncPublicResource}
				resName, err = huawei.SyncAllResource(kt, cliSet, opt)

			case enumor.Azure:
				opt := &azure.SyncAllResourceOption{AccountID: one.ID, SyncPublicResource: syncPublicResource}
				resName, err = azure.SyncAllResource(kt, cliSet, opt)

			case enumor.Gcp:
				opt := &gcp.SyncAllResourceOption{AccountID: one.ID, SyncPublicResource: syncPublicResource}
				resName, err = gcp.SyncAllResource(kt, cliSet, opt)

			default:
				logs.Errorf("unknown %s vendor type", one.Vendor)
				continue
			}
			if err != nil {
				if resName != "" {
					if err := sd.ResSyncStatusFailed(resName, err); err != nil {
						logs.Errorf("%s sync %s res detail failed, err: %v, accountID: %s, rid: %s", vendor,
							resName, err, one.ID, kt.Rid)
						return
					}
				}
				logs.Errorf("sync %s all resource failed, err: %v, accountID: %s, rid: %s", vendor, err, one.ID, kt.Rid)
				// 跳过当前账号
				continue
			}

			// 公共资源仅需要同步一次即可
			syncPublicResource = false
		}
		if len(accounts) < int(core.DefaultMaxPageLimit) {
			break
		}
		start += uint32(core.DefaultMaxPageLimit)
	}
}

const maxRetryCount = 3

// listAccountWithRetry 查询账号列表，最多重试3次，每次等待
func listAccountWithRetry(kt *kit.Kit, cli *dataservice.Client, req *protocloud.AccountListReq) (
	[]*corecloud.BaseAccount, error) {
	rty := retry.NewRetryPolicy(maxRetryCount, [2]uint{500, 15000})

	for {
		if rty.RetryCount() == maxRetryCount {
			return nil, fmt.Errorf("list account failed count over %d", maxRetryCount)
		}

		list, err := cli.Global.Account.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list account failed, err: %v, rid: %s", err, kt.Rid)
			rty.Sleep()
			continue
		}

		return list.Details, nil
	}
}
