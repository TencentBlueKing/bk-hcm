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

package tcloud

import (
	"time"

	"hcm/cmd/cloud-server/service/sync/detail"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncCert ...
func SyncCert(kt *kit.Kit, cliSet *client.ClientSet, accountID string, regions []string, sd *detail.SyncDetail) error {
	// 重新设置rid方便定位
	kt = kt.NewSubKit()

	start := time.Now()
	logs.V(3).Infof("tcloud account[%s] sync cert start, time: %v, rid: %s", accountID, start, kt.Rid)

	// 同步中
	if err := sd.ResSyncStatusSyncing(enumor.CertCloudResType); err != nil {
		return err
	}

	defer func() {
		logs.V(3).Infof("tcloud account[%s] sync cert end, cost: %v, rid: %s", accountID, time.Since(start), kt.Rid)
	}()

	if len(regions) > 0 {
		req := &sync.TCloudSyncReq{
			AccountID: accountID,
			Region:    regions[0],
		}
		if err := cliSet.HCService().TCloud.Cert.SyncCert(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("sync tcloud cert failed, req: %+v, err: %v, rid: %s", req, err, kt.Rid)
			return err
		}
	}

	// 同步成功
	if err := sd.ResSyncStatusSuccess(enumor.CertCloudResType); err != nil {
		return err
	}

	return nil
}
