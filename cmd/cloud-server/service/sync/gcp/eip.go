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

	"hcm/pkg/api/hc-service/sync"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncEip ...
func SyncEip(kt *kit.Kit, service *hcservice.Client, accountID string, regions []string) error {

	start := time.Now()
	logs.V(3).Infof("gcp account[%s] sync eip start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("gcp account[%s] sync eip end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	pipeline := make(chan bool, syncConcurrencyCount)
	var firstErr error
	var wg gosync.WaitGroup
	for _, region := range regions {
		pipeline <- true
		wg.Add(1)

		go func(region string) {
			defer func() {
				wg.Done()
				<-pipeline
			}()

			req := &sync.GcpSyncReq{
				AccountID: accountID,
				Region:    region,
			}
			err := service.Gcp.Eip.SyncEip(kt.Ctx, kt.Header(), req)
			if firstErr == nil && err != nil {
				logs.Errorf("sync gcp eip failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
				firstErr = err
				return
			}
		}(region)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	return nil
}
