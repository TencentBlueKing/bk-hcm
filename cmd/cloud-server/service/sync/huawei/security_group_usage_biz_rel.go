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

package huawei

import (
	gosync "sync"
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/adaptor/huawei"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncSGUsageBizRel ...
func SyncSGUsageBizRel(kt *kit.Kit, cliSet *client.ClientSet, accountID string, sd *detail.SyncDetail) error {

	// 重新设置rid方便定位
	kt = kt.NewSubKit()

	start := time.Now()
	logs.V(3).Infof("huawei account[%s] sync sg usage biz rel start, time: %v, rid: %s", accountID, start, kt.Rid)

	// 同步中
	if err := sd.ResSyncStatusSyncing(enumor.SecurityGroupUsageBizRelResType); err != nil {
		return err
	}

	defer func() {
		logs.V(3).Infof("huawei account[%s] sync sg usage biz rel  end, cost: %v, rid: %s",
			accountID, time.Since(start), kt.Rid)
	}()

	regions, err := ListRegionByService(kt, cliSet.DataService(), huawei.Vpc)
	if err != nil {
		logs.Errorf("sync huawei list region failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

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

			req := &sync.HuaWeiSyncReq{
				AccountID: accountID,
				Region:    region,
			}
			err = cliSet.HCService().HuaWei.SecurityGroup.SyncSecurityGroupUsageBizRel(kt, req)
			if firstErr == nil && Error(err) != nil {
				logs.Errorf("sync huawei security group usage biz rel failed, err: %v, req: %v, rid: %s",
					err, req, kt.Rid)
				firstErr = err
				return
			}
		}(region)
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	// 同步成功
	if err := sd.ResSyncStatusSuccess(enumor.SecurityGroupUsageBizRelResType); err != nil {
		return err
	}

	return nil
}
