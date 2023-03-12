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
	"time"

	hcapiproto "hcm/pkg/api/hc-service"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// SyncNetworkInterface 网络接口同步
func SyncNetworkInterface(kt *kit.Kit, service *hcservice.Client, accountID string, zones []string) error {

	start := time.Now()
	logs.V(3).Infof("cloud-server-sync-%s account[%s] sync network interface start, time: %v, rid: %s",
		enumor.Gcp, accountID, start, kt.Rid)

	defer func() {
		logs.V(3).Infof("cloud-server-sync-%s account[%s] sync network interface end, cost: %v, rid: %s",
			enumor.Gcp, accountID, time.Since(start), kt.Rid)
	}()

	for _, zone := range zones {
		req := &hcapiproto.GcpNetworkInterfaceSyncReq{
			AccountID: accountID,
			Zone:      zone,
		}
		if err := service.Gcp.NetworkInterface.SyncNetworkInterface(kt.Ctx, kt.Header(), req); err != nil {
			logs.Errorf("cloud-server-sync-%s network interface failed, req: %v, err: %v, rid: %s",
				enumor.Gcp, req, err, kt.Rid)
			return err
		}
	}

	return nil
}
