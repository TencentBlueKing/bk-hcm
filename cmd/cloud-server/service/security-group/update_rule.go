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
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
)

// UpdateSecurityGroupRule update security group rule.
func (svc *securityGroupSvc) UpdateSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	sgBaseInfo, err := svc.client.DataService().Global.Cloud.GetResourceBasicInfo(cts.Kit.Ctx, cts.Kit.Header(),
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		return nil, err
	}

	// authorize
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.SecurityGroupRule, Action: meta.Update,
		ResourceID: sgBaseInfo.AccountID}}
	err = svc.authorizer.AuthorizeWithPerm(cts.Kit, authRes)
	if err != nil {
		return nil, err
	}

	// 已分配业务的资源，不允许操作
	flt := &filter.AtomRule{Field: "id", Op: filter.In.Factory(), Value: sgID}
	err = svc.checkSecurityGroupsInBiz(cts.Kit, flt, constant.UnassignedBiz)
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		return svc.updateTCloudSGRule(cts, sgBaseInfo, id)

	case enumor.Aws:
		return svc.updateAwsSGRule(cts, sgBaseInfo, id)

	case enumor.Azure:
		return svc.updateAzureSGRule(cts, sgBaseInfo, id)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

func (svc *securityGroupSvc) updateTCloudSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo,
	id string) (interface{}, error) {

	req := new(proto.TCloudSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ChildResUpdateAudit(cts.Kit, enumor.SecurityGroupRuleAuditResType, sgBaseInfo.ID, id, updateFields)
	if err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &hcproto.TCloudSGRuleUpdateReq{
		Protocol:                   req.Protocol,
		Port:                       req.Port,
		IPv4Cidr:                   req.IPv4Cidr,
		IPv6Cidr:                   req.IPv6Cidr,
		CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
		Action:                     req.Action,
		Memo:                       req.Memo,
	}
	if err := svc.client.HCService().TCloud.SecurityGroup.UpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		sgBaseInfo.ID, id, updateReq); err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) updateAwsSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo,
	id string) (interface{}, error) {

	req := new(proto.AwsSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ChildResUpdateAudit(cts.Kit, enumor.SecurityGroupRuleAuditResType, sgBaseInfo.ID, id, updateFields)
	if err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &hcproto.AwsSGRuleUpdateReq{
		IPv4Cidr:                   req.IPv4Cidr,
		IPv6Cidr:                   req.IPv6Cidr,
		Memo:                       req.Memo,
		FromPort:                   req.FromPort,
		ToPort:                     req.ToPort,
		Protocol:                   req.Protocol,
		CloudTargetSecurityGroupID: req.CloudTargetSecurityGroupID,
	}
	if err := svc.client.HCService().Aws.SecurityGroup.UpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		sgBaseInfo.ID, id, updateReq); err != nil {
		return nil, err
	}

	return nil, nil
}

func (svc *securityGroupSvc) updateAzureSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo,
	id string) (interface{}, error) {

	req := new(proto.AzureSGRuleUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// create update audit.
	updateFields, err := converter.StructToMap(req)
	if err != nil {
		logs.Errorf("convert request to map failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	err = svc.audit.ChildResUpdateAudit(cts.Kit, enumor.SecurityGroupRuleAuditResType, sgBaseInfo.ID, id, updateFields)
	if err != nil {
		logs.Errorf("create update audit failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	updateReq := &hcproto.AzureSGRuleUpdateReq{
		Name:                             req.Name,
		Memo:                             req.Memo,
		DestinationAddressPrefix:         req.DestinationAddressPrefix,
		DestinationAddressPrefixes:       req.DestinationAddressPrefixes,
		CloudDestinationSecurityGroupIDs: req.CloudDestinationSecurityGroupIDs,
		DestinationPortRange:             req.DestinationPortRange,
		DestinationPortRanges:            req.DestinationPortRanges,
		Protocol:                         req.Protocol,
		SourceAddressPrefix:              req.SourceAddressPrefix,
		SourceAddressPrefixes:            req.SourceAddressPrefixes,
		CloudSourceSecurityGroupIDs:      req.CloudSourceSecurityGroupIDs,
		SourcePortRange:                  req.SourcePortRange,
		SourcePortRanges:                 req.SourcePortRanges,
		Priority:                         req.Priority,
		Access:                           req.Access,
	}
	if err := svc.client.HCService().Azure.SecurityGroup.UpdateSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
		sgBaseInfo.ID, id, updateReq); err != nil {
		return nil, err
	}

	return nil, nil
}
