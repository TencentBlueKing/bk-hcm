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

	"hcm/cmd/cloud-server/logics/account"
	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/retry"
)

// CloudResourceSync 定时同步云资源
func CloudResourceSync(intervalMin time.Duration, sd serviced.ServiceDiscover, cliSet *client.ClientSet) {

	syncKit := core.NewBackendKit()
	logs.Infof("cloud resource sync enabled, syncIntervalMin: %s, rid: %s", intervalMin, syncKit.Rid)

	for {
		time.Sleep(intervalMin)

		if !sd.IsMaster() {
			continue
		}

		start := time.Now()
		logs.Infof("cloud resource all sync start, time: %v, rid: %s", start, syncKit.Rid)

		tenantIDs, err := listAllTenantIDs(syncKit, cliSet)
		if err != nil {
			logs.Errorf("list tenant failed, err: %v, rid: %s", err, syncKit.Rid)
			continue
		}
		logs.Infof("cloud resource sync tenant total count: %d, rid: %s", len(tenantIDs), syncKit.Rid)

		for _, tenantID := range tenantIDs {
			waitGroup := new(sync.WaitGroup)
			syncers := account.GetAvailableVendorSyncers()

			waitGroup.Add(len(syncers))
			for _, vendorSyncer := range syncers {
				go func(vendor account.VendorSyncer) {
					kt := core.NewTenantBackendKit(tenantID)
					// for retry
					kt.RequestSource = enumor.AsynchronousTasks
					allAccountSync(kt, cliSet, vendor)
					waitGroup.Done()
				}(vendorSyncer)
			}

			waitGroup.Wait()
		}

		logs.Infof("cloud resource all sync end, time: %s, rid: %s", time.Since(start), syncKit.Rid)
	}
}

func listAllTenantIDs(kt *kit.Kit, cliSet *client.ClientSet) ([]string, error) {

	tenantIDs := make([]string, 0)

	tenantListReq := &core.ListReq{
		Filter: tools.EqualExpression("status", enumor.TenantEnable),
		Page:   core.NewDefaultBasePage(),
	}
	for {
		tenantResp, err := cliSet.DataService().Global.Tenant.List(kt, tenantListReq)
		if err != nil {
			logs.Errorf("list tenant failed, err: %v, req: %v, rid: %s", err, tenantListReq, kt.Rid)
			return nil, err
		}

		for _, tenant := range tenantResp.Details {
			tenantIDs = append(tenantIDs, tenant.TenantID)
		}
		if len(tenantResp.Details) < int(tenantListReq.Page.Limit) {
			break
		}
		tenantListReq.Page.Start += uint32(tenantListReq.Page.Limit)
	}
	return tenantIDs, nil
}

// allAccountSync all account sync.
func allAccountSync(kt *kit.Kit, cliSet *client.ClientSet, syncer account.VendorSyncer) {
	startTime := time.Now()
	logs.Infof("%s start sync all cloud resource, tenant: %s, time: %v, rid: %s",
		syncer.Vendor(), kt.TenantID, startTime, kt.Rid)

	defer func() {
		logs.Infof("%s sync all cloud resource end, tenant: %s, cost: %v, rid: %s",
			syncer.Vendor(), kt.TenantID, time.Since(startTime), kt.Rid)
	}()

	listReq := &protocloud.AccountListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: []filter.RuleFactory{
			&filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: syncer.Vendor()},
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

		for _, acc := range accounts {
			sd := &detail.SyncDetail{
				Kt:        kt,
				DataCli:   cliSet.DataService(),
				AccountID: acc.ID,
				Vendor:    string(acc.Vendor),
			}
			resName, err := syncer.SyncAllResource(kt, cliSet, acc.ID, syncPublicResource)
			if err != nil {
				if resName != "" {
					if err := sd.ResSyncStatusFailed(resName, err); err != nil {
						logs.Errorf("%s sync %s res detail failed, err: %v, accountID: %s, rid: %s",
							syncer.Vendor(), resName, err, acc.ID, kt.Rid)
						return
					}
				}
				logs.Errorf("sync %s all resource failed, err: %v, accountID: %s, rid: %s",
					syncer.Vendor(), err, acc.ID, kt.Rid)
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
