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

package gcp

import (
	gosync "sync"
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncRoute 同步路由表
func SyncRoute(kt *kit.Kit, cliSet *client.ClientSet, accountID string, regionZoneMap map[string][]string,
	sd *detail.SyncDetail) error {

	start := time.Now()
	logs.V(3).Infof("[%s] account[%s] sync route table start, time: %v, rid: %s",
		enumor.Gcp, accountID, start, kt.Rid)

	// 同步中
	if err := sd.ResSyncStatusSyncing(enumor.RouteTableCloudResType); err != nil {
		return err
	}

	defer func() {
		logs.V(3).Infof("[%s] account[%s] sync route table end, cost: %v, rid: %s",
			enumor.Gcp, accountID, time.Since(start), kt.Rid)
	}()

	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for _, zones := range regionZoneMap {
		for _, zone := range zones {
			pipeline <- true
			wg.Add(1)

			go func(zone string) {
				defer func() {
					wg.Done()
					<-pipeline
				}()

				req := &sync.GcpRouteSyncReq{
					AccountID: accountID,
					Zone:      zone,
				}
				err := cliSet.HCService().Gcp.RouteTable.SyncRoute(kt.Ctx, kt.Header(), req)
				if firstErr == nil && err != nil {
					logs.Errorf("[%s] account[%s] sync route failed, req: %v, err: %v, rid: %s",
						enumor.Gcp, accountID, req, err, kt.Rid)
					firstErr = err
					return
				}
			}(zone)
		}
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	// 同步成功
	if err := sd.ResSyncStatusSuccess(enumor.RouteTableCloudResType); err != nil {
		return err
	}

	return nil
}
