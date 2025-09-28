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
	"hcm/cmd/cloud-server/logics/async"
	actionsg "hcm/cmd/task-server/logics/action/security-group"
	proto "hcm/pkg/api/cloud-server"
	hcproto "hcm/pkg/api/hc-service"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/assert"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/hooks/handler"
	"hcm/pkg/tools/slice"
)

// CreateSecurityGroupRule create security group rule.
func (svc *securityGroupSvc) CreateSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	return svc.createSGRule(cts, handler.ResOperateAuth)
}

// CreateBizSGRule create biz security group rule.
func (svc *securityGroupSvc) CreateBizSGRule(cts *rest.Contexts) (interface{}, error) {
	return svc.createSGRule(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) createSGRule(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	sgBaseInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroupRule,
		Action: meta.Create, BasicInfo: sgBaseInfo})
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

func (svc *securityGroupSvc) createTCloudSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
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
				CloudServiceID:             one.CloudServiceID,
				CloudServiceGroupID:        one.CloudServiceGroupID,
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				CloudAddressID:             one.CloudAddressID,
				CloudAddressGroupID:        one.CloudAddressGroupID,
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
				CloudServiceID:             one.CloudServiceID,
				CloudServiceGroupID:        one.CloudServiceGroupID,
				IPv4Cidr:                   one.IPv4Cidr,
				IPv6Cidr:                   one.IPv6Cidr,
				CloudAddressID:             one.CloudAddressID,
				CloudAddressGroupID:        one.CloudAddressGroupID,
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

func (svc *securityGroupSvc) createAwsSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
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

func (svc *securityGroupSvc) createHuaWeiSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
	interface{}, error) {

	req := new(proto.SecurityGroupRuleCreateReq[proto.HuaWeiSecurityGroupRule])
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	getTaskID := counter.NewNumStringCounter(1, 10)
	tasks := slice.Map(req.EgressRuleSet, func(r proto.HuaWeiSecurityGroupRule) ts.CustomFlowTask {
		return ts.CustomFlowTask{
			ActionID:   action.ActIDType(getTaskID()),
			ActionName: enumor.ActionCreateHuaweiSGRule,
			Params:     convSGEgressRuleReq(sgBaseInfo, r),
		}
	})
	tasks = append(tasks, slice.Map(req.IngressRuleSet, func(r proto.HuaWeiSecurityGroupRule) ts.CustomFlowTask {
		return ts.CustomFlowTask{
			ActionID:   action.ActIDType(getTaskID()),
			ActionName: enumor.ActionCreateHuaweiSGRule,
			Params:     convSGIngressRuleReq(sgBaseInfo, r),
		}
	})...)

	flowReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowCreateHuaweiSGRule,
		Tasks: tasks,
	}

	result, err := svc.client.TaskServer().CreateCustomFlow(cts.Kit, flowReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err := async.WaitTaskToEnd(cts.Kit, svc.client.TaskServer(), result.ID); err != nil {
		return nil, err
	}
	return result, nil
}

func convSGEgressRuleReq(sgBaseInfo *types.CloudResourceBasicInfo,
	rule proto.HuaWeiSecurityGroupRule) *actionsg.CreateHuaweiSGRuleOption {
	return convSGRuleReq(sgBaseInfo, rule, true)
}
func convSGIngressRuleReq(sgBaseInfo *types.CloudResourceBasicInfo,
	rule proto.HuaWeiSecurityGroupRule) *actionsg.CreateHuaweiSGRuleOption {
	return convSGRuleReq(sgBaseInfo, rule, false)
}

func convSGRuleReq(sgBaseInfo *types.CloudResourceBasicInfo, rule proto.HuaWeiSecurityGroupRule,
	isEgress bool) *actionsg.CreateHuaweiSGRuleOption {

	actionOpt := &actionsg.CreateHuaweiSGRuleOption{
		SGID: sgBaseInfo.ID,
		RuleReq: &hcproto.HuaWeiSGRuleCreateReq{
			AccountID: sgBaseInfo.AccountID,
		},
	}
	r := &hcproto.HuaWeiSGRuleCreate{
		Memo:               rule.Memo,
		Ethertype:          rule.Ethertype,
		Protocol:           rule.Protocol,
		RemoteIPPrefix:     rule.RemoteIPPrefix,
		CloudRemoteGroupID: rule.CloudRemoteGroupID,
		Port:               rule.Port,
		Action:             rule.Action,
		Priority:           rule.Priority,
	}
	if isEgress {
		actionOpt.RuleReq.EgressRule = r
	} else {
		actionOpt.RuleReq.IngressRule = r
	}

	return actionOpt
}

func (svc *securityGroupSvc) createAzureSGRule(cts *rest.Contexts, sgBaseInfo *types.CloudResourceBasicInfo) (
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
			tmpEgressRule := hcproto.AzureSGRuleCreate{
				Name:                       one.Name,
				Memo:                       one.Memo,
				DestinationAddressPrefix:   one.DestinationAddressPrefix,
				DestinationAddressPrefixes: one.DestinationAddressPrefixes,
				DestinationPortRange:       one.DestinationPortRange,
				DestinationPortRanges:      one.DestinationPortRanges,
				Protocol:                   one.Protocol,
				SourceAddressPrefix:        one.SourceAddressPrefix,
				SourceAddressPrefixes:      one.SourceAddressPrefixes,
				SourcePortRange:            one.SourcePortRange,
				SourcePortRanges:           one.SourcePortRanges,
				Priority:                   one.Priority,
				Type:                       enumor.Egress,
				Access:                     one.Access,
			}

			if err := svc.checkCreateAzureSGRuleParams(tmpEgressRule); err != nil {
				return nil, err
			}

			createReq.EgressRuleSet = append(createReq.EgressRuleSet, tmpEgressRule)
		}
	}

	if len(req.IngressRuleSet) != 0 {
		createReq.IngressRuleSet = make([]hcproto.AzureSGRuleCreate, 0, len(req.IngressRuleSet))
		for _, one := range req.IngressRuleSet {
			tmpIngressRule := hcproto.AzureSGRuleCreate{
				Name:                       one.Name,
				Memo:                       one.Memo,
				DestinationAddressPrefix:   one.DestinationAddressPrefix,
				DestinationAddressPrefixes: one.DestinationAddressPrefixes,
				DestinationPortRange:       one.DestinationPortRange,
				DestinationPortRanges:      one.DestinationPortRanges,
				Protocol:                   one.Protocol,
				SourceAddressPrefix:        one.SourceAddressPrefix,
				SourceAddressPrefixes:      one.SourceAddressPrefixes,
				SourcePortRange:            one.SourcePortRange,
				SourcePortRanges:           one.SourcePortRanges,
				Priority:                   one.Priority,
				Type:                       enumor.Ingress,
				Access:                     one.Access,
			}

			if err := svc.checkCreateAzureSGRuleParams(tmpIngressRule); err != nil {
				return nil, err
			}

			createReq.IngressRuleSet = append(createReq.IngressRuleSet, tmpIngressRule)
		}
	}

	result, err := svc.client.HCService().Azure.SecurityGroup.BatchCreateSecurityGroupRule(cts.Kit.Ctx,
		cts.Kit.Header(), sgBaseInfo.ID, createReq)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// checkCreateAzureSGRuleParams check create azure security group rule params
func (svc *securityGroupSvc) checkCreateAzureSGRuleParams(req hcproto.AzureSGRuleCreate) error {
	if !assert.IsSameCaseString(req.Name) {
		return errf.New(errf.InvalidParameter, "name can only be lowercase")
	}

	return nil
}
