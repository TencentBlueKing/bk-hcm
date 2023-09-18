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

package aws

import (
	"fmt"
	gosync "sync"
	"time"

	"hcm/pkg/api/hc-service/zone"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncZone sync zone
func SyncZone(kt *kit.Kit, hcCli *hcservice.Client, accountID string, regions []string) error {

	// 重新设置rid方便定位
	prefix := fmt.Sprintf("%s", enumor.ZoneCloudResType)
	kt = kt.NewSubKit(prefix)

	start := time.Now()
	logs.V(3).Infof("aws account[%s] sync zone start, time: %v, rid: %s", accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("aws account[%s] sync zone end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
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

			syncReq := &zone.AwsZoneSyncReq{
				AccountID: accountID,
				Region:    region,
			}
			err := hcCli.Aws.Zone.SyncZone(kt.Ctx, kt.Header(), syncReq)
			if firstErr == nil && err != nil {
				logs.Errorf("sync aws zone failed, err: %v, req: %v, rid: %s", err, syncReq, kt.Rid)
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
