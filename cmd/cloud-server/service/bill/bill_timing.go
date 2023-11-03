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

package bill

import (
	"fmt"
	"sync"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
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

// CloudBillConfigCreate 定时生成云账单配置
func CloudBillConfigCreate(intervalMin time.Duration, sd serviced.ServiceDiscover, cliSet *client.ClientSet) {
	logs.Infof("account cloud bill config pipeline enable && start, syncIntervalMin: %v", intervalMin)

	for {
		time.Sleep(intervalMin)

		if !sd.IsMaster() {
			continue
		}

		kt := core.NewBackendKit()

		start := time.Now()
		logs.Infof("account cloud bill config pipeline start, time: %v, rid: %s", start, kt.Rid)

		waitGroup := new(sync.WaitGroup)

		vendors := []enumor.Vendor{enumor.Aws}
		waitGroup.Add(len(vendors))
		for _, vendor := range vendors {
			go func(vendor enumor.Vendor) {
				allAccountBillConfig(kt, cliSet, vendor)
				waitGroup.Done()
			}(vendor)
		}

		waitGroup.Wait()

		logs.Infof("cloud resource all sync end, time: %v, rid: %s", start, kt.Rid)
	}
}

// allAccountBillConfig all account bill config.
func allAccountBillConfig(kt *kit.Kit, cliSet *client.ClientSet, vendor enumor.Vendor) {
	startTime := time.Now()
	logs.Infof("%s all account bill config start, time: %v, rid: %s", vendor, startTime, kt.Rid)

	defer func() {
		logs.Infof("%s all account bill config end, cost: %v, rid: %s", vendor, time.Since(startTime), kt.Rid)
	}()

	listReq := &protocloud.AccountListReq{
		Filter: &filter.Expression{
			Op: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "vendor",
					Op:    filter.Equal.Factory(),
					Value: vendor,
				},
				&filter.AtomRule{
					Field: "type",
					Op:    filter.Equal.Factory(),
					Value: enumor.ResourceAccount,
				},
			},
		},
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
		},
	}

	start := uint32(0)
	for {
		listReq.Page.Start = start
		accounts, err := listAccountWithRetry(kt, cliSet.DataService(), listReq)
		if err != nil {
			logs.Errorf("%s account bill config get list account failed, err: %v, rid: %s", vendor, err, kt.Rid)
			break
		}

		for _, one := range accounts {
			switch one.Vendor {
			case enumor.Aws:
				err = AccountBillConfig(kt, cliSet, &BillConfigOption{AccountID: one.ID})
			default:
				logs.Errorf("unknown %s vendor type", one.Vendor)
				continue
			}

			if err != nil {
				logs.Errorf("%s account bill config failed, accountID: %s, err: %+v, rid: %s",
					vendor, one.ID, err, kt.Rid)
				continue
			}
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
	[]*cloud.BaseAccount, error) {

	rty := retry.NewRetryPolicy(maxRetryCount, [2]uint{500, 15000})

	for {
		if rty.RetryCount() == maxRetryCount {
			return nil, fmt.Errorf("list account with retry failed count over %d", maxRetryCount)
		}

		list, err := cli.Global.Account.List(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("list account with retry failed, err: %v, rid: %s", err, kt.Rid)
			rty.Sleep()
			continue
		}

		return list.Details, nil
	}
}
