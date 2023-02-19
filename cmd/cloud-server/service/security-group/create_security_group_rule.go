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
	"hcm/pkg/api/core"
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CreateSecurityGroupRule create security group rule.
func (svc securityGroupSvc) CreateSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
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
	authRes := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.SecurityGroupRule, Action: meta.Create,
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
		return svc.createTCloudSGRule(cts, sgBaseInfo)

	case enumor.Aws:
		return svc.createAwsSGRule(cts, sgBaseInfo)

	case enumor.HuaWei:
		return svc.createHuaWeiSGRule(cts, sgBaseInfo)

	case enumor.Azure:
		return svc.createAzureSGRule(cts, sgBaseInfo)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support", vendor)
	}
}

func (svc securityGroupSvc) createTCloudSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
	interface{}, error) {

	req := new(proto.SecurityGroupRuleCreateReq[proto.TCloudSecurityGroupRule])
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createReq := &hcproto.TCloudSGRuleCreateReq{
		AccountID: sgBaseInfo.AccountID,
	}
	if len(req.EgressRuleSet) != 0 {
		createReq.EgressRuleSet = make([]hcproto.TCloudSGRuleCreate, 0, len(req.EgressRuleSet))
		for _, one := range req.EgressRuleSet {
			createReq.EgressRuleSet = append(createReq.EgressRuleSet, hcproto.TCloudSGRuleCreate{
				Protocol:                   one.Protocol,
				Port:                       one.Port,
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
				Action:                     one.Action,
				Memo:                       one.Memo,
			})
		}
	}

	if len(req.IngressRuleSet) != 0 {
		createReq.IngressRuleSet = make([]hcproto.TCloudSGRuleCreate, 0, len(req.IngressRuleSet))
		for _, one := range req.IngressRuleSet {
			createReq.IngressRuleSet = append(createReq.IngressRuleSet, hcproto.TCloudSGRuleCreate{
				Protocol:                   one.Protocol,
				Port:                       one.Port,
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
				Action:                     one.Action,
				Memo:                       one.Memo,
			})
		}
	}

	result, err := svc.client.HCService().TCloud.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx,
		cts.Kit.Header(), sgBaseInfo.ID, createReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (svc securityGroupSvc) createAwsSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
	interface{}, error) {

	req := new(proto.SecurityGroupRuleCreateReq[proto.AwsSecurityGroupRule])
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createReq := &hcproto.AwsSGRuleCreateReq{
		AccountID: sgBaseInfo.AccountID,
	}
	if len(req.EgressRuleSet) != 0 {
		createReq.EgressRuleSet = make([]hcproto.AwsSGRuleCreate, 0, len(req.EgressRuleSet))
		for _, one := range req.EgressRuleSet {
			createReq.EgressRuleSet = append(createReq.EgressRuleSet, hcproto.AwsSGRuleCreate{
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				Memo:                       one.Memo,
				FromPort:                   one.FromPort,
				ToPort:                     one.ToPort,
				Protocol:                   one.Protocol,
				CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
			})
		}
	}

	if len(req.IngressRuleSet) != 0 {
		createReq.IngressRuleSet = make([]hcproto.AwsSGRuleCreate, 0, len(req.IngressRuleSet))
		for _, one := range req.IngressRuleSet {
			createReq.IngressRuleSet = append(createReq.IngressRuleSet, hcproto.AwsSGRuleCreate{
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				Memo:                       one.Memo,
				FromPort:                   one.FromPort,
				ToPort:                     one.ToPort,
				Protocol:                   one.Protocol,
				CloudTargetSecurityGroupID: one.CloudTargetSecurityGroupID,
			})
		}
	}

	result, err := svc.client.HCService().Aws.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx,
		cts.Kit.Header(), sgBaseInfo.ID, createReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (svc securityGroupSvc) createHuaWeiSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
	interface{}, error) {

	req := new(proto.SecurityGroupRuleCreateReq[proto.HuaWeiSecurityGroupRule])
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(req.EgressRuleSet) != 0 {
		successIDs := make([]string, 0, len(req.EgressRuleSet))
		for _, one := range req.EgressRuleSet {
			createReq := &hcproto.HuaWeiSGRuleCreateReq{
				AccountID: sgBaseInfo.AccountID,
				EgressRule: &hcproto.HuaWeiSGRuleCreate{
					Memo:               one.Memo,
					Ethertype:          one.Ethertype,
					Protocol:           one.Protocol,
					RemoteIPPrefix:     one.RemoteIPPrefix,
					CloudRemoteGroupID: one.CloudRemoteGroupID,
					Port:               one.Port,
					Action:             one.Action,
					Priority:           one.Priority,
				},
			}
			result, err := svc.client.HCService().HuaWei.SecurityGroup.CreateSecurityGroupRule(cts.Kit.Ctx,
				cts.Kit.Header(), sgBaseInfo.ID, createReq)
			if err != nil {
				return &core.BatchCreateResult{IDs: successIDs}, errf.NewFromErr(errf.PartialFailed, err)
			}

			successIDs = append(successIDs, result.ID)
		}
		return &core.BatchCreateResult{IDs: successIDs}, nil
	}

	if len(req.IngressRuleSet) != 0 {
		successIDs := make([]string, 0)
		for _, one := range req.IngressRuleSet {
			createReq := &hcproto.HuaWeiSGRuleCreateReq{
				AccountID: sgBaseInfo.AccountID,
				IngressRule: &hcproto.HuaWeiSGRuleCreate{
					Memo:               one.Memo,
					Ethertype:          one.Ethertype,
					Protocol:           one.Protocol,
					RemoteIPPrefix:     one.RemoteIPPrefix,
					CloudRemoteGroupID: one.CloudRemoteGroupID,
					Port:               one.Port,
					Action:             one.Action,
					Priority:           one.Priority,
				},
			}
			result, err := svc.client.HCService().HuaWei.SecurityGroup.CreateSecurityGroupRule(cts.Kit.Ctx,
				cts.Kit.Header(), sgBaseInfo.ID, createReq)
			if err != nil {
				return &core.BatchCreateResult{IDs: successIDs}, errf.NewFromErr(errf.PartialFailed, err)
			}

			successIDs = append(successIDs, result.ID)
		}
		return &core.BatchCreateResult{IDs: successIDs}, nil
	}

	return nil, nil
}

func (svc securityGroupSvc) createAzureSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
	interface{}, error) {

	req := new(proto.SecurityGroupRuleCreateReq[proto.AzureSecurityGroupRule])
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	createReq := &hcproto.AzureSGRuleCreateReq{
		AccountID: sgBaseInfo.AccountID,
	}
	if len(req.EgressRuleSet) != 0 {
		createReq.EgressRuleSet = make([]hcproto.AzureSGRuleCreate, 0, len(req.EgressRuleSet))
		for _, one := range req.EgressRuleSet {
			createReq.EgressRuleSet = append(createReq.EgressRuleSet, hcproto.AzureSGRuleCreate{
				Name:                             one.Name,
				Memo:                             one.Memo,
				DestinationAddressPrefix:         one.DestinationAddressPrefix,
				DestinationAddressPrefixes:       one.DestinationAddressPrefixes,
				CloudDestinationSecurityGroupIDs: one.CloudDestinationSecurityGroupIDs,
				DestinationPortRange:             one.DestinationPortRange,
				DestinationPortRanges:            one.DestinationPortRanges,
				Protocol:                         one.Protocol,
				SourceAddressPrefix:              one.SourceAddressPrefix,
				SourceAddressPrefixes:            one.SourceAddressPrefixes,
				CloudSourceSecurityGroupIDs:      one.CloudSourceSecurityGroupIDs,
				SourcePortRange:                  one.SourcePortRange,
				SourcePortRanges:                 one.SourcePortRanges,
				Priority:                         one.Priority,
				Type:                             one.Type,
				Access:                           one.Access,
			})
		}
	}

	if len(req.IngressRuleSet) != 0 {
		createReq.IngressRuleSet = make([]hcproto.AzureSGRuleCreate, 0, len(req.IngressRuleSet))
		for _, one := range req.IngressRuleSet {
			createReq.IngressRuleSet = append(createReq.IngressRuleSet, hcproto.AzureSGRuleCreate{
				Name:                             one.Name,
				Memo:                             one.Memo,
				DestinationAddressPrefix:         one.DestinationAddressPrefix,
				DestinationAddressPrefixes:       one.DestinationAddressPrefixes,
				CloudDestinationSecurityGroupIDs: one.CloudDestinationSecurityGroupIDs,
				DestinationPortRange:             one.DestinationPortRange,
				DestinationPortRanges:            one.DestinationPortRanges,
				Protocol:                         one.Protocol,
				SourceAddressPrefix:              one.SourceAddressPrefix,
				SourceAddressPrefixes:            one.SourceAddressPrefixes,
				CloudSourceSecurityGroupIDs:      one.CloudSourceSecurityGroupIDs,
				SourcePortRange:                  one.SourcePortRange,
				SourcePortRanges:                 one.SourcePortRanges,
				Priority:                         one.Priority,
				Type:                             one.Type,
				Access:                           one.Access,
			})
		}
	}

	result, err := svc.client.HCService().Azure.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx,
		cts.Kit.Header(), sgBaseInfo.ID, createReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}
