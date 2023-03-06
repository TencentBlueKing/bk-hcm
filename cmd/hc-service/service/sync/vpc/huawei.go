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

// Package vpc defines vpc service.
package vpc

import (
	vpclogic "hcm/cmd/hc-service/logics/sync/vpc"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncHuaWeiVpc sync huawei vpc to hcm.
func (v syncVpcSvc) SyncHuaWeiVpc(cts *rest.Contexts) (interface{}, error) {
	req, err := decodeVpcSyncReq(cts)
	if err != nil {
		return nil, err
	}

	resp, err := vpclogic.HuaWeiVpcSync(cts.Kit, req, v.ad, v.dataCli)
	if err != nil {
		logs.Errorf("request to sync huawei vpc logic failed, req: %+v, err: %v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	return resp, nil
}
