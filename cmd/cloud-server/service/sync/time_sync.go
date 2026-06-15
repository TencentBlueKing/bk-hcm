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
	"encoding/json"
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

		tenantSyncBizIDs, err := fetchSyncBizIDs(syncKit, cliSet)
		if err != nil {
			logs.Errorf("fetch sync biz ids failed, err: %v, rid: %s", err, syncKit.Rid)
			continue
		}
		logs.Infof("cloud resource sync tenant biz ids config: %v, rid: %s", tenantSyncBizIDs, syncKit.Rid)

		tenantIDs, err := listAllTenantIDs(syncKit, cliSet)
		if err != nil {
			logs.Errorf("list tenant failed, err: %v, rid: %s", err, syncKit.Rid)
			continue
		}
		logs.Infof("cloud resource sync tenant total count: %d, rid: %s", len(tenantIDs), syncKit.Rid)

		for _, tenantID := range tenantIDs {
			waitGroup := new(sync.WaitGroup)
			syncers := account.GetAvailableVendorSyncers()

			syncBizIDs := tenantSyncBizIDs[tenantID]
			waitGroup.Add(len(syncers))
			for _, vendorSyncer := range syncers {
				go func(vendor account.VendorSyncer) {
					kt := core.NewTenantBackendKit(tenantID)
					// for retry
					kt.RequestSource = enumor.AsynchronousTasks
					allAccountSync(kt, cliSet, vendor, syncBizIDs)
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
func allAccountSync(kt *kit.Kit, cliSet *client.ClientSet, syncer account.VendorSyncer, syncBizIDs []int64) {
	startTime := time.Now()
	logs.Infof("%s start sync all cloud resource, tenant: %s, time: %v, rid: %s",
		syncer.Vendor(), kt.TenantID, startTime, kt.Rid)

	defer func() {
		logs.Infof("%s sync all cloud resource end, tenant: %s, cost: %v, rid: %s",
			syncer.Vendor(), kt.TenantID, time.Since(startTime), kt.Rid)
	}()

	filterRules := []filter.RuleFactory{
		&filter.AtomRule{Field: "vendor", Op: filter.Equal.Factory(), Value: syncer.Vendor()},
		&filter.AtomRule{Field: "type", Op: filter.Equal.Factory(), Value: enumor.ResourceAccount},
	}
	if len(syncBizIDs) > 0 {
		filterRules = append(filterRules, buildBizFilter(syncBizIDs))
	}
	listReq := &protocloud.AccountListReq{
		Filter: &filter.Expression{Op: filter.And, Rules: filterRules},
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
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

// fetchSyncBizIDs 从 global_config 读取云资源同步业务白名单。
// config_value 格式为 JSON 对象：{"tenantID": [bizID1, bizID2], ...}。
// 若配置不存在或为空对象，返回 nil（各租户全量同步）；
// 若租户不在 map 中，该租户也执行全量同步。
func fetchSyncBizIDs(kt *kit.Kit, cliSet *client.ClientSet) (map[string][]int64, error) {
	req := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", string(enumor.GlobalConfigTypeCloudSync)),
			tools.RuleEqual("config_key", string(enumor.GlobalConfigKeyCloudSyncBizIDs)),
		),
		Page: &core.BasePage{Limit: 1},
	}
	result, err := cliSet.DataService().Global.GlobalConfig.List(kt, req)
	if err != nil {
		return nil, fmt.Errorf("list global config sync_biz_ids failed, err: %v", err)
	}
	if len(result.Details) == 0 {
		return nil, nil
	}
	tenantBizIDs := make(map[string][]int64)
	if err = json.Unmarshal([]byte(result.Details[0].ConfigValue), &tenantBizIDs); err != nil {
		return nil, fmt.Errorf("unmarshal sync_biz_ids failed, err: %v", err)
	}
	return tenantBizIDs, nil
}

func buildBizFilter(syncBizIDs []int64) filter.RuleFactory {
	if len(syncBizIDs) <= int(filter.DefaultMaxInLimit) {
		return tools.RuleIn("bk_biz_id", syncBizIDs)
	}
	rules := make([]*filter.AtomRule, 0)
	for i := 0; i < len(syncBizIDs); i += int(filter.DefaultMaxInLimit) {
		end := i + int(filter.DefaultMaxInLimit)
		if end > len(syncBizIDs) {
			end = len(syncBizIDs)
		}
		rules = append(rules, tools.RuleIn("bk_biz_id", syncBizIDs[i:end]))
	}
	return tools.ExpressionOr(rules...)
}
