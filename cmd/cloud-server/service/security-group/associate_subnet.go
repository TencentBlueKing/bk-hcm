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
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// AssociateSubnet associate subnet.
func (svc *securityGroupSvc) AssociateSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.associateSubnet(cts, handler.ResValidWithAuth)
}

// AssociateBizSubnet associate biz subnet.
func (svc *securityGroupSvc) AssociateBizSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.associateSubnet(cts, handler.BizValidWithAuth)
}

func (svc *securityGroupSvc) associateSubnet(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req, err := svc.decodeAndValidateAssocSubnetReq(cts, meta.Associate, validHandler)
	if err != nil {
		return nil, err
	}

	// create operation audit.
	audit := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.SecurityGroupRuleAuditResType,
		ResID:             req.SecurityGroupID,
		Action:            protoaudit.Associate,
		AssociatedResType: enumor.SubnetAuditResType,
		AssociatedResID:   req.SubnetID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	associateReq := &hcproto.AzureSecurityGroupAssociateSubnetReq{
		SecurityGroupID: req.SecurityGroupID,
		SubnetID:        req.SubnetID,
	}
	err = svc.client.HCService().Azure.SecurityGroup.AssociateSubnet(cts.Kit.Ctx,
		cts.Kit.Header(), associateReq)
	if err != nil {
		logs.Errorf("security group associate subnet failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// DisAssociateSubnet disassociate subnet.
func (svc *securityGroupSvc) DisAssociateSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateSubnet(cts, handler.ResValidWithAuth)
}

// DisAssociateBizSubnet disassociate biz subnet.
func (svc *securityGroupSvc) DisAssociateBizSubnet(cts *rest.Contexts) (interface{}, error) {
	return svc.disassociateSubnet(cts, handler.BizValidWithAuth)
}

func (svc *securityGroupSvc) disassociateSubnet(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req, err := svc.decodeAndValidateAssocSubnetReq(cts, meta.Disassociate, validHandler)
	if err != nil {
		return nil, err
	}

	// create operation audit.
	audit := protoaudit.CloudResourceOperationInfo{
		ResType:           enumor.SecurityGroupRuleAuditResType,
		ResID:             req.SecurityGroupID,
		Action:            protoaudit.Disassociate,
		AssociatedResType: enumor.SubnetAuditResType,
		AssociatedResID:   req.SubnetID,
	}
	if err = svc.audit.ResOperationAudit(cts.Kit, audit); err != nil {
		logs.Errorf("create operation audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	associateReq := &hcproto.AzureSecurityGroupAssociateSubnetReq{
		SecurityGroupID: req.SecurityGroupID,
		SubnetID:        req.SubnetID,
	}
	err = svc.client.HCService().Azure.SecurityGroup.DisassociateSubnet(cts.Kit.Ctx,
		cts.Kit.Header(), associateReq)
	if err != nil {
		logs.Errorf("security group disassociate subnet failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) decodeAndValidateAssocSubnetReq(cts *rest.Contexts, action meta.Action,
	validHandler handler.ValidWithAuthHandler) (*proto.SecurityGroupAssociateSubnetReq, error) {

	req := new(proto.SecurityGroupAssociateSubnetReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, req.SecurityGroupID)
	if err != nil {
		logs.Errorf("get resource vendor failed, id: %s, err: %s, rid: %s", basicInfo, err, cts.Kit.Rid)
		return nil, err
	}

	if basicInfo.Vendor != enumor.Azure {
		return nil, errf.Newf(errf.InvalidParameter, "associate subnet only support azure")
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroup,
		Action: action, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	return req, nil
}
