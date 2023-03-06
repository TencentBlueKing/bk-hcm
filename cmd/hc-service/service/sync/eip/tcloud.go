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

package eip

import (
	eip "hcm/cmd/hc-service/logics/sync/eip"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncTCloudEip ...
func (svc *syncEipSvc) SyncTCloudEip(cts *rest.Contexts) (interface{}, error) {
	req, err := decodeEipSyncReq(cts)
	if err != nil {
		logs.Errorf("request decodeEipSyncReq failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	_, err = eip.SyncTCloudEip(cts.Kit, req, svc.adaptor, svc.dataCli)
	if err != nil {
		logs.Errorf("request to sync tcloud eip failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
