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

package azure

import (
	"fmt"
	gosync "sync"
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/api/core"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SyncSubnet ...
func SyncSubnet(kt *kit.Kit, cliSet *client.ClientSet, accountID string, resourceGroupNames []string,
	sd *detail.SyncDetail) error {

	// 重新设置rid方便定位
	prefix := fmt.Sprintf("%s", enumor.SubnetCloudResType)
	kt = kt.NewSubKit(prefix)

	start := time.Now()
	logs.V(3).Infof("azure account[%s] sync subnet start, time: %v, rid: %s", accountID, start, kt.Rid)

	// 同步中
	if err := sd.ResSyncStatusSyncing(enumor.SubnetCloudResType); err != nil {
		return err
	}

	defer func() {
		logs.V(3).Infof("azure account[%s] sync subnet end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for _, name := range resourceGroupNames {
		filterRules := []filter.RuleFactory{
			&filter.AtomRule{Field: "account_id", Op: filter.Equal.Factory(), Value: accountID},
			&filter.AtomRule{Field: "extension.resource_group_name", Op: filter.JSONEqual.Factory(), Value: name},
		}
		listReq := &core.ListReq{
			Filter: &filter.Expression{Op: filter.And, Rules: filterRules},
			Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
			Fields: []string{"cloud_id"},
		}
		startIndex := uint32(0)
		for {
			listReq.Page.Start = startIndex

			vpcResult, err := cliSet.DataService().Global.Vpc.List(kt.Ctx, kt.Header(), listReq)
			if err != nil {
				logs.Errorf("list huawei vpc failed, err: %v, rid: %s", err, kt.Rid)
				return err
			}

			for _, vpc := range vpcResult.Details {
				pipeline <- true
				wg.Add(1)

				go func(vpcID, name string) {
					defer func() {
						wg.Done()
						<-pipeline
					}()

					req := &sync.AzureSubnetSyncReq{AccountID: accountID, ResourceGroupName: name, CloudVpcID: vpcID}
					err := cliSet.HCService().Azure.Subnet.SyncSubnet(kt.Ctx, kt.Header(), req)
					if firstErr == nil && err != nil {
						logs.Errorf("sync azure subnet failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
						firstErr = err
						return
					}
				}(vpc.CloudID, name)
			}

			if len(vpcResult.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			startIndex += uint32(core.DefaultMaxPageLimit)
		}
	}

	wg.Wait()
	if firstErr != nil {
		return firstErr
	}
	// 同步成功
	if err := sd.ResSyncStatusSuccess(enumor.SubnetCloudResType); err != nil {
		return err
	}
	return nil
}
