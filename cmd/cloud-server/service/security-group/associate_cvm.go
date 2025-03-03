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
	proto "hcm/pkg/api/cloud-server"
	protoaudit "hcm/pkg/api/data-service/audit"
	"hcm/pkg/api/data-service/cloud"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AssociateCvm associate cvm.
func (svc *securityGroupSvc) AssociateCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.associateCvm(cts, handler.ResOperateAuth)
}

// AssociateBizCvm associate biz cvm.
func (svc *securityGroupSvc) AssociateBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.associateCvm(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) associateCvm(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req, vendor, err := svc.decodeAndValidateAssocCvmReq(cts, meta.Associate, validHandler)
	if err != nil {
		return nil, err
	}

	// create operation audit.
	audit := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.SecurityGroupRuleAuditResType,
		ResID:             req.SecurityGroupID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   req.CvmID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		associateReq := &hcproto.SecurityGroupAssociateCvmReq{
			SecurityGroupID: req.SecurityGroupID,
			CvmID:           req.CvmID,
		}
		err = svc.client.HCService().TCloud.SecurityGroup.AssociateCvm(cts.Kit.Ctx, cts.Kit.Header(),
			associateReq)

	case enumor.HuaWei:
		associateReq := &hcproto.SecurityGroupAssociateCvmReq{
			SecurityGroupID: req.SecurityGroupID,
			CvmID:           req.CvmID,
		}
		err = svc.client.HCService().HuaWei.SecurityGroup.AssociateCvm(cts.Kit.Ctx, cts.Kit.Header(),
			associateReq)

	case enumor.Aws:
		associateReq := &hcproto.SecurityGroupAssociateCvmReq{
			SecurityGroupID: req.SecurityGroupID,
			CvmID:           req.CvmID,
		}
		err = svc.client.HCService().Aws.SecurityGroup.AssociateCvm(cts.Kit.Ctx, cts.Kit.Header(),
			associateReq)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	if err != nil {
		logs.Errorf("security group associate cvm failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisassociateCvm disassociate cvm.
func (svc *securityGroupSvc) DisassociateCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateCvm(cts, handler.ResOperateAuth)
}

// DisassociateBizCvm disassociate biz cvm.
func (svc *securityGroupSvc) DisassociateBizCvm(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateCvm(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) disassociateCvm(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req, vendor, err := svc.decodeAndValidateAssocCvmReq(cts, meta.Disassociate, validHandler)
	if err != nil {
		return nil, err
	}

	// create operation audit.
	audit := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.SecurityGroupRuleAuditResType,
		ResID:             req.SecurityGroupID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.CvmAuditResType,
		AssociatedResID:   req.CvmID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		associateReq := &hcproto.SecurityGroupAssociateCvmReq{
			SecurityGroupID: req.SecurityGroupID,
			CvmID:           req.CvmID,
		}
		err = svc.client.HCService().TCloud.SecurityGroup.DisassociateCvm(cts.Kit.Ctx, cts.Kit.Header(),
			associateReq)

	case enumor.HuaWei:
		associateReq := &hcproto.SecurityGroupAssociateCvmReq{
			SecurityGroupID: req.SecurityGroupID,
			CvmID:           req.CvmID,
		}
		err = svc.client.HCService().HuaWei.SecurityGroup.DisassociateCvm(cts.Kit.Ctx, cts.Kit.Header(),
			associateReq)

	case enumor.Aws:
		associateReq := &hcproto.SecurityGroupAssociateCvmReq{
			SecurityGroupID: req.SecurityGroupID,
			CvmID:           req.CvmID,
		}
		err = svc.client.HCService().Aws.SecurityGroup.DisassociateCvm(cts.Kit.Ctx, cts.Kit.Header(),
			associateReq)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}

	if err != nil {
		logs.Errorf("security group disassociate cvm failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) decodeAndValidateAssocCvmReq(cts *rest.Contexts, action meta.Action,
	validHandler handler.ValidWithAuthHandler) (*proto.SecurityGroupAssociateCvmReq, enumor.Vendor, error) {

	req := new(proto.SecurityGroupAssociateCvmReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, "", errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, "", errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 鉴权和校验资源分配状态和回收状态
	basicReq := &cloud.BatchListResourceBasicInfoReq{
		Items: []cloud.ListResourceBasicInfoReq{
			{ResourceType: enumor.SecurityGroupCloudResType, IDs: []string{req.SecurityGroupID},
				Fields: types.CommonBasicInfoFields},
			{ResourceType: enumor.CvmCloudResType, IDs: []string{req.CvmID}, Fields: types.ResWithRecycleBasicFields},
		},
	}

	basicInfos, err := svc.client.DataService().Global.Cloud.BatchListResBasicInfo(cts.Kit, basicReq)
	if err != nil {
		logs.Errorf("batch list resource basic info failed, err: %v, req: %+v, rid: %s", err, basicReq, cts.Kit.Rid)
		return nil, "", err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: action, BasicInfos: basicInfos})
	if err != nil {
		return nil, "", err
	}

	var vendor enumor.Vendor
	for _, info := range basicInfos {
		vendor = info.Vendor
		break
	}

	return req, vendor, nil
}
