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

// Package networkinterface defines networkinterface service.
package networkinterface

import (
	netinterfacelogic "hcm/cmd/hc-service/logics/sync/network-interface"
	hcservice "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncHuaWeiNetworkInterface sync huawei network interface to hcm.
func (v syncNetworkInterfaceSvc) SyncHuaWeiNetworkInterface(cts *rest.Contexts) (interface{}, error) {
	req := new(hcservice.HuaWeiNetworkInterfaceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	err := req.Validate()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	resp, err := netinterfacelogic.HuaWeiNetworkInterfaceSync(cts.Kit, req, v.ad, v.dataCli)
	if err != nil {
		logs.Errorf("request to sync huawei network interface logic failed, req: %+v, err: %v, rid: %s",
			req, err, cts.Kit.Rid)
		return nil, err
	}

	return resp, nil
}
