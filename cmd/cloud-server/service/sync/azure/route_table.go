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
	gosync "sync"
	"time"

	routetable "hcm/pkg/api/hc-service/route-table"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncRouteTable 同步路由表
func SyncRouteTable(kt *kit.Kit, service *hcservice.Client, accountID string, resourceGroupNames []string) error {
	start := time.Now()
	logs.V(3).Infof("cloud-server-sync-%s account[%s] sync route table start, time: %v, rid: %s",
		enumor.Azure, accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("cloud-server-sync-%s account[%s] sync route table end, cost: %v, rid: %s",
			enumor.Azure, accountID, time.Since(start), kt.Rid)
	}()

	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for _, name := range resourceGroupNames {
		pipeline <- true
		wg.Add(1)

		go func(name string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			req := &routetable.AzureRouteTableSyncReq{
				AccountID:         accountID,
				ResourceGroupName: name,
			}
			err := service.Azure.RouteTable.SyncRouteTable(kt.Ctx, kt.Header(), req)
			if firstErr == nil && err != nil {
				logs.Errorf("sync azure route table failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
				firstErr = err
				return
			}
		}(name)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	return nil
}
