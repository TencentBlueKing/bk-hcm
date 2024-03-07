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
	hcclb "hcm/pkg/api/hc-service/clb"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AssociateClb associate clb.
func (svc *securityGroupSvc) AssociateClb(cts *rest.Contexts) (interface{}, error) {
	return svc.associateClb(cts, handler.ResOperateAuth)
}

// AssociateBizClb associate biz clb.
func (svc *securityGroupSvc) AssociateBizClb(cts *rest.Contexts) (interface{}, error) {
	return svc.associateClb(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) associateClb(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(hcclb.TCloudSetClbSecurityGroupReq)
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
			{ResourceType: enumor.ClbCloudResType, IDs: []string{req.ClbID}, Fields: types.CommonBasicInfoFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list clb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Associate, BasicInfos: basicInfos})
	if err != nil {
		logs.Errorf("batch list clb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
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
			AssociatedResType: enumor.ClbAuditResType,
			AssociatedResID:   req.ClbID,
		}
		if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
			logs.Errorf("create clb operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	switch vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.SecurityGroup.AssociateClb(cts.Kit.Ctx, cts.Kit.Header(), req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	if err != nil {
		logs.Errorf("security group associate clb failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateClb disassociate clb.
func (svc *securityGroupSvc) DisassociateClb(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateClb(cts, handler.ResOperateAuth)
}

// DisassociateBizClb disassociate biz clb.
func (svc *securityGroupSvc) DisassociateBizClb(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateClb(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) disassociateClb(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(hcclb.TCloudDisAssociateClbSecurityGroupReq)
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
			{ResourceType: enumor.ClbCloudResType, IDs: []string{req.ClbID}, Fields: types.CommonBasicInfoFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list clb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: meta.Disassociate, BasicInfos: basicInfos})
	if err != nil {
		logs.Errorf("batch list clb resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
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
		AssociatedResType: enumor.ClbAuditResType,
		AssociatedResID:   req.ClbID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		err = svc.client.HCService().TCloud.SecurityGroup.DisassociateClb(cts.Kit.Ctx, cts.Kit.Header(), req)
	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	if err != nil {
		logs.Errorf("security group disassociate clb failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
