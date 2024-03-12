/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package loadbalancer

import (
	"fmt"

	hcclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// UpdateBizTCloudLoadBalancer  业务下更新clb
func (svc *clbSvc) UpdateBizTCloudLoadBalancer(cts *rest.Contexts) (any, error) {

	lbID := cts.PathParameter("id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	req := new(hcclb.TCloudUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	baseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit, enumor.LoadBalancerCloudResType, lbID)
	if err != nil {
		logs.Errorf("getLoadBalancer resource vendor failed, id: %s, err: %s, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  baseInfo})
	if err != nil {
		return nil, err
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.LoadBalancerAuditResType, lbID, updateFields); err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	switch baseInfo.Vendor {
	case enumor.TCloud:
		return nil, svc.client.HCService().TCloud.Clb.Update(cts.Kit, lbID, req)
	default:
		return nil, fmt.Errorf("vendor: %s not support", baseInfo.Vendor)
	}

}
