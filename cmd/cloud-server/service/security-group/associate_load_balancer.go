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

package securitygroup

import (
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AssociateLb associate lb.
func (svc *securityGroupSvc) AssociateLb(cts *rest.Contexts) (interface{}, error) {
	return svc.associateLb(cts, handler.ResOperateAuth)
}

// AssociateBizLb associate biz lb.
func (svc *securityGroupSvc) AssociateBizLb(cts *rest.Contexts) (interface{}, error) {
	return svc.associateLb(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) associateLb(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(hclb.TCloudSetLbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 鉴权和校验资源分配状态和回收状态
	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.SecurityGroupCloudResType, IDs: req.SecurityGroupIDs,
				Fields: types.CommonBasicInfoFields},
			{ResourceType: enumor.LoadBalancerCloudResType, IDs: []string{req.LbID},
				Fields: types.CommonBasicInfoFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list lb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Associate, BasicInfos: basicInfos})
	if err != nil {
		logs.Errorf("batch list lb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	var vendor enumor.Vendor
	for _, info := range basicInfos {
		vendor = info.Vendor
		break
	}

	for _, sgID := range req.SecurityGroupIDs {
		// create operation audit.
		audit := protoaudit.CloudResourceOperationInfo{
			ResType:           enumor.SecurityGroupRuleAuditResType,
			ResID:             sgID,
			Action:            protoaudit.Associate,
			AssociatedResType: enumor.LoadBalancerAuditResType,
			AssociatedResID:   req.LbID,
		}
		if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
			logs.Errorf("create lb operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	switch vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.SecurityGroup.AssociateLb(cts.Kit.Ctx, cts.Kit.Header(), req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	if err != nil {
		logs.Errorf("security group associate lb failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateLb disassociate lb.
func (svc *securityGroupSvc) DisassociateLb(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateLb(cts, handler.ResOperateAuth)
}

// DisassociateBizLb disassociate biz lb.
func (svc *securityGroupSvc) DisassociateBizLb(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateLb(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) disassociateLb(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(hclb.TCloudDisAssociateLbSecurityGroupReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 鉴权和校验资源分配状态和回收状态
	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.SecurityGroupCloudResType, IDs: []string{req.SecurityGroupID},
				Fields: types.CommonBasicInfoFields},
			{ResourceType: enumor.LoadBalancerCloudResType,
				IDs: []string{req.LbID}, Fields: types.CommonBasicInfoFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list lb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Disassociate, BasicInfos: basicInfos})
	if err != nil {
		logs.Errorf("batch list lb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	var vendor enumor.Vendor
	for _, info := range basicInfos {
		vendor = info.Vendor
		break
	}

	// create operation audit.
	audit := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.SecurityGroupRuleAuditResType,
		ResID:             req.SecurityGroupID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.LoadBalancerAuditResType,
		AssociatedResID:   req.LbID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.SecurityGroup.DisassociateLb(cts.Kit.Ctx, cts.Kit.Header(), req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	if err != nil {
		logs.Errorf("security group disassociate lb failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
