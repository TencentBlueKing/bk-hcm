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
	"fmt"
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncFireWall ...
func SyncFireWall(kt *kit.Kit, cliSet *client.ClientSet, accountID string,
	sd *detail.SyncDetail) error {

	// 重新设置rid方便定位
	prefix := fmt.Sprintf("%s", enumor.GcpFirewallRuleCloudResType)
	kt = kt.NewSubKit(prefix)

	start := time.Now()
	logs.V(3).Infof("gcp account[%s] sync firewall start, time: %v, rid: %s", accountID, start, kt.Rid)

	// 同步中
	if err := sd.ResSyncStatusSyncing(enumor.GcpFirewallRuleCloudResType); err != nil {
		return err
	}

	defer func() {
		logs.V(3).Infof("gcp account[%s] sync firewall end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	req := &sync.GcpGlobalSyncReq{
		AccountID: accountID,
	}
	if err := cliSet.HCService().Gcp.Firewall.SyncFirewall(kt.Ctx, kt.Header(), req); err != nil {
		logs.Errorf("sync gcp firewall failed, err: %v, req: %v, rid: %s", err, req, kt.Rid)
		return err
	}

	// 同步成功
	if err := sd.ResSyncStatusSuccess(enumor.GcpFirewallRuleCloudResType); err != nil {
		return err
	}

	return nil
}
