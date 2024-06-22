/*
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
	"errors"

	cslb "hcm/pkg/api/cloud-server/load-balancer"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// TCloudCreateSnatIps ...
func (svc *lbSvc) TCloudCreateSnatIps(cts *rest.Contexts) (any, error) {
	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "load balancer id is required")
	}

	req := new(cslb.TCloudCreateSnatIpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lbInfo, err := svc.client.DataService().TCloud.LoadBalancer.Get(cts.Kit, lbID)
	if err != nil {
		logs.Errorf("getLoadBalancer resource vendor failed, id: %s, err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	basicInfo := &types.CloudResourceBasicInfo{
		ResType:   enumor.LoadBalancerCloudResType,
		ID:        lbID,
		Vendor:    enumor.TCloud,
		AccountID: lbInfo.AccountID,
		BkBizID:   lbInfo.BkBizID,
		Region:    lbInfo.Region,
	}

	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  basicInfo})
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
		logs.Errorf("create lb create snat ip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	hcReq := &hclb.TCloudCreateSnatIpReq{
		AccountID:           lbInfo.AccountID,
		Region:              lbInfo.Region,
		LoadBalancerCloudId: lbInfo.CloudID,
		SnatIPs:             req.SnatIps,
	}

	if err := svc.client.HCService().TCloud.Clb.CreateSnatIp(cts.Kit, hcReq); err != nil {
		logs.Errorf("fail to call hc service to create snat ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

// TCloudDeleteSnatIps ...
func (svc *lbSvc) TCloudDeleteSnatIps(cts *rest.Contexts) (any, error) {

	lbID := cts.PathParameter("lb_id").String()
	if len(lbID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "load balancer id is required")
	}

	req := new(cslb.TCloudDeleteSnatIpReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	lbInfo, err := svc.client.DataService().TCloud.LoadBalancer.Get(cts.Kit, lbID)
	if err != nil {
		logs.Errorf("getLoadBalancer resource vendor failed, id: %s, err: %v, rid: %s", lbID, err, cts.Kit.Rid)
		return nil, err
	}

	basicInfo := &types.CloudResourceBasicInfo{
		ResType:   enumor.LoadBalancerCloudResType,
		ID:        lbID,
		Vendor:    enumor.TCloud,
		AccountID: lbInfo.AccountID,
		BkBizID:   lbInfo.BkBizID,
		Region:    lbInfo.Region,
	}

	// validate biz and authorize
	err = handler.BizOperateAuth(cts, &handler.ValidWithAuthOption{
		Authorizer: svc.authorizer,
		ResType:    meta.LoadBalancer,
		Action:     meta.Update,
		BasicInfo:  basicInfo})
	if err != nil {
		return nil, err
	}

	if converter.PtrToVal(lbInfo.Extension.SnatPro) == false {
		return nil, errors.New("deleting snat ip is not allowed on load balancer whose snat pro flag is false")
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert lb delete snat ip to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if err = svc.audit.ResUpdateAudit(cts.Kit, enumor.LoadBalancerAuditResType, lbID, updateFields); err != nil {
		logs.Errorf("create lb delete snat ip audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	hcReq := &hclb.TCloudDeleteSnatIpReq{
		AccountID:           lbInfo.AccountID,
		Region:              lbInfo.Region,
		LoadBalancerCloudId: lbInfo.CloudID,
		Ips:                 req.DeleteIps,
	}

	if err := svc.client.HCService().TCloud.Clb.DeleteSnatIp(cts.Kit, hcReq); err != nil {
		logs.Errorf("fail to call hc service to delete snat ip, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}
