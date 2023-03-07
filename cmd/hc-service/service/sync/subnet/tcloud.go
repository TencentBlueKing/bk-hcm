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

// Package subnet defines subnet service.
package subnet

import (
	subnetlogic "hcm/cmd/hc-service/logics/sync/subnet"
	"hcm/pkg/api/core"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// SyncTCloudSubnet sync tcloud subnet to hcm.
func (v syncSubnetSvc) SyncTCloudSubnet(cts *rest.Contexts) (interface{}, error) {
	req := new(hcproto.TCloudResourceSyncReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(req.CloudIDs) > 0 && len(req.CloudIDs) > int(core.DefaultMaxPageLimit) {
		return nil, errf.New(errf.TooManyRequest, "cloud_ids length should <= 500")
	}

	resp, err := subnetlogic.TCloudSubnetSync(cts.Kit, req, v.ad, v.dataCli)
	if err != nil {
		logs.Errorf("request to sync tcloud subnet logic failed, req: %+v, err: %v, rid: %s",
			req, err, cts.Kit.Rid)
		return nil, err
	}

	return resp, nil
}
