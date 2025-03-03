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
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/hooks/handler"
)

// ListSecurityGroupRule list security group rule.
func (svc *securityGroupSvc) ListSecurityGroupRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRule(cts, handler.ResOperateAuth)
}

// ListBizSGRule list biz security group rule.
func (svc *securityGroupSvc) ListBizSGRule(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRule(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSGRule(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	sgID := cts.PathParameter("security_group_id").String()
	if len(sgID) == 0 {
		return nil, errf.New(errf.InvalidParameter, "security group id is required")
	}

	req := new(proto.SecurityGroupRuleListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	basicInfo, err := svc.client.DataService().Global.Cloud.GetResBasicInfo(cts.Kit,
		enumor.SecurityGroupCloudResType, sgID)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroupRule,
		Action: meta.Find, BasicInfo: basicInfo})
	if err != nil {
		return nil, err
	}

	switch vendor {
	case enumor.TCloud:
		listReq := &dataproto.TCloudSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().TCloud.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx,
			cts.Kit.Header(), listReq, sgID)

	case enumor.Aws:
		listReq := &dataproto.AwsSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().Aws.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			listReq, sgID)

	case enumor.HuaWei:
		listReq := &dataproto.HuaWeiSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().HuaWei.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			listReq, sgID)

	case enumor.Azure:
		listReq := &dataproto.AzureSGRuleListReq{
			Filter: req.Filter,
			Page:   req.Page,
		}
		return svc.client.DataService().Azure.SecurityGroup.ListSecurityGroupRule(cts.Kit.Ctx, cts.Kit.Header(),
			listReq, sgID)

	default:
		return nil, errf.Newf(errf.Unknown, "vendor: %s not support for list security group rule", vendor)
	}
}

// GetAzureDefaultSGRule get azure default security group rule.
func (svc *securityGroupSvc) GetAzureDefaultSGRule(cts *rest.Contexts) (interface{}, error) {
	ruleType := enumor.SecurityGroupRuleType(cts.PathParameter("type").String())

	rules, exist := azureDefaultSGRuleMap[ruleType]
	if !exist {
		return nil, errf.Newf(errf.InvalidParameter, "rule type: %s not support", ruleType)
	}

	return rules, nil
}

// AzureDefaultSGRule define azure default security group rule.
type AzureDefaultSGRule struct {
	Name                                string                       `json:"name"`
	Memo                                *string                      `json:"memo"`
	DestinationAddressPrefix            *string                      `json:"destination_address_prefix"`
	DestinationAddressPrefixes          []*string                    `json:"destination_address_prefixes"`
	CloudDestinationAppSecurityGroupIDs []*string                    `json:"cloud_destination_app_security_group_ids"`
	DestinationPortRange                *string                      `json:"destination_port_range"`
	DestinationPortRanges               []*string                    `json:"destination_port_ranges"`
	Protocol                            string                       `json:"protocol"`
	ProvisioningState                   string                       `json:"provisioning_state"`
	SourceAddressPrefix                 *string                      `json:"source_address_prefix"`
	SourceAddressPrefixes               []*string                    `json:"source_address_prefixes"`
	CloudSourceAppSecurityGroupIDs      []*string                    `json:"cloud_source_app_security_group_ids"`
	SourcePortRange                     *string                      `json:"source_port_range"`
	SourcePortRanges                    []*string                    `json:"source_port_ranges"`
	Priority                            int32                        `json:"priority"`
	Type                                enumor.SecurityGroupRuleType `json:"type"`
	Access                              string                       `json:"access"`
}

// TODO: 之后考虑是否通过同步的方式将这几条默认安全组规则同步进来，而不是写死。
// reference:
// https://learn.microsoft.com/zh-cn/azure/virtual-network/network-security-groups-overview#default-security-rules
var azureDefaultSGRuleMap = map[enumor.SecurityGroupRuleType][]AzureDefaultSGRule{
	enumor.Egress: {
		{
			Name:                     "AllowVnetOutBound",
			Memo:                     converter.ValToPtr("Allow outbound traffic from all VMs to all VMs in VNET"),
			Protocol:                 "*",
			SourcePortRange:          converter.ValToPtr("*"),
			DestinationPortRange:     converter.ValToPtr("*"),
			SourceAddressPrefix:      converter.ValToPtr("VirtualNetwork"),
			DestinationAddressPrefix: converter.ValToPtr("VirtualNetwork"),
			Access:                   "Allow",
			Priority:                 65000,
			Type:                     enumor.Egress,
		},
		{
			Name:                     "AllowInternetOutBound",
			Memo:                     converter.ValToPtr("Allow outbound traffic from all VMs to Internet"),
			Protocol:                 "*",
			SourcePortRange:          converter.ValToPtr("*"),
			DestinationPortRange:     converter.ValToPtr("*"),
			SourceAddressPrefix:      converter.ValToPtr("*"),
			DestinationAddressPrefix: converter.ValToPtr("Internet"),
			Access:                   "Allow",
			Priority:                 65001,
			Type:                     enumor.Egress,
		},
		{
			Name:                     "DenyAllOutBound",
			Memo:                     converter.ValToPtr("Deny all outbound traffic"),
			Protocol:                 "*",
			SourcePortRange:          converter.ValToPtr("*"),
			DestinationPortRange:     converter.ValToPtr("*"),
			SourceAddressPrefix:      converter.ValToPtr("*"),
			DestinationAddressPrefix: converter.ValToPtr("*"),
			Access:                   "Deny",
			Priority:                 65500,
			Type:                     enumor.Egress,
		},
	},
	enumor.Ingress: {
		{
			Name:                     "AllowVnetInBound",
			Memo:                     converter.ValToPtr("Allow inbound traffic from all VMs in VNET"),
			Protocol:                 "*",
			SourcePortRange:          converter.ValToPtr("*"),
			DestinationPortRange:     converter.ValToPtr("*"),
			SourceAddressPrefix:      converter.ValToPtr("VirtualNetwork"),
			DestinationAddressPrefix: converter.ValToPtr("VirtualNetwork"),
			Access:                   "Allow",
			Priority:                 65000,
			Type:                     enumor.Ingress,
		},
		{
			Name:                     "AllowAzureLoadBalancerInBound",
			Memo:                     converter.ValToPtr("Allow inbound traffic from azure load balancer"),
			Protocol:                 "*",
			SourcePortRange:          converter.ValToPtr("*"),
			DestinationPortRange:     converter.ValToPtr("*"),
			SourceAddressPrefix:      converter.ValToPtr("AzureLoadBalancer"),
			DestinationAddressPrefix: converter.ValToPtr("*"),
			Access:                   "Allow",
			Priority:                 65001,
			Type:                     enumor.Ingress,
		},
		{
			Name:                     "DenyAllInBound",
			Memo:                     converter.ValToPtr("Deny all inbound traffic"),
			Protocol:                 "*",
			SourcePortRange:          converter.ValToPtr("*"),
			DestinationPortRange:     converter.ValToPtr("*"),
			SourceAddressPrefix:      converter.ValToPtr("*"),
			DestinationAddressPrefix: converter.ValToPtr("*"),
			Access:                   "Deny",
			Priority:                 65500,
			Type:                     enumor.Ingress,
		},
	},
}

// CountSecurityGroupRules list security group rules count.
func (svc *securityGroupSvc) CountSecurityGroupRules(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRulesCount(cts, handler.ResOperateAuth)
}

// CountBizSecurityGroupRules list biz security group rules count.
func (svc *securityGroupSvc) CountBizSecurityGroupRules(cts *rest.Contexts) (interface{}, error) {
	return svc.listSGRulesCount(cts, handler.BizOperateAuth)
}

func (svc *securityGroupSvc) listSGRulesCount(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (
	interface{}, error) {

	req := new(proto.ListSecurityGroupRuleCountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listBasicInfoReq := dataproto.ListResourceBasicInfoReq{
		ResourceType: enumor.SecurityGroupCloudResType,
		IDs:          req.SecurityGroupIDs,
	}
	basicInfoMap, err := svc.client.DataService().Global.Cloud.ListResBasicInfo(cts.Kit, listBasicInfoReq)
	if err != nil {
		return nil, err
	}

	// validate biz and authorize
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.SecurityGroupRule,
		Action: meta.Find, BasicInfos: basicInfoMap})
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	vendorToSGIDMap := make(map[enumor.Vendor][]string)
	for _, info := range basicInfoMap {
		result[info.ID] = 0
		vendorToSGIDMap[info.Vendor] = append(vendorToSGIDMap[info.Vendor], info.ID)
	}

	for vendor, ids := range vendorToSGIDMap {
		resp, err := svc.client.DataService().Global.SecurityGroup.CountSecurityGroupRules(cts.Kit, vendor, ids)
		if err != nil {
			logs.Errorf("list security group rules count from data service failed, err: %v, vendor: %s, ids: %v, rid: %s",
				err, vendor, ids, cts.Kit.Rid)
			return nil, err
		}

		for sgID, count := range resp {
			result[sgID] = count
		}
	}
	return result, nil
}
