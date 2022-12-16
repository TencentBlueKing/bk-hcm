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

package huawei

import (
	"fmt"

	"hcm/pkg/adaptor/types"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/region"
)

// CreateSecurityGroupRule create security group rule.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0016.html
func (h *Huawei) CreateSecurityGroupRule(kt *kit.Kit, opt *types.HuaWeiSGRuleCreateOption) (*model.SecurityGroupRule,
		error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule create option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(region.ValueOf(opt.Region))
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	rule := &model.CreateSecurityGroupRuleOption{
		SecurityGroupId: opt.CloudSecurityGroupID,
		Description:     opt.Rule.Description,
		Ethertype:       opt.Rule.Ethertype,
		Protocol:        opt.Rule.Protocol,
		Multiport:       opt.Rule.Port,
		RemoteIpPrefix:  opt.Rule.RemoteIPPrefix,
		RemoteGroupId:   opt.Rule.CloudRemoteGroupID,
		Action:          opt.Rule.Action,
		Priority:        opt.Rule.Priority,
	}
	switch opt.Rule.Type {
	case enumor.Egress:
		rule.Direction = "egress"
	case enumor.Ingress:
		rule.Direction = "ingress"
	default:
		return nil, fmt.Errorf("unknown security group rule type: %s", opt.Rule.Type)
	}

	req := &model.CreateSecurityGroupRuleRequest{
		Body: &model.CreateSecurityGroupRuleRequestBody{
			SecurityGroupRule: rule,
		},
	}
	resp, err := client.CreateSecurityGroupRule(req)
	if err != nil {
		logs.Errorf("create huawei security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp.SecurityGroupRule, nil
}

// DeleteSecurityGroupRule delete security group rule.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0019.html
func (h *Huawei) DeleteSecurityGroupRule(kt *kit.Kit, opt *types.HuaWeiSGRuleDeleteOption) error {

	if opt == nil {
		return errf.New(errf.InvalidParameter, "security group rule delete option is required")
	}

	if err := opt.Validate(); err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(region.ValueOf(opt.Region))
	if err != nil {
		return fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &model.DeleteSecurityGroupRuleRequest{
		SecurityGroupRuleId: opt.CloudRuleID,
	}
	_, err = client.DeleteSecurityGroupRule(req)
	if err != nil {
		logs.Errorf("delete huawei security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroupRule list security group rule.
// reference: https://support.huaweicloud.com/api-vpc/vpc_apiv3_0019.html
func (h *Huawei) ListSecurityGroupRule(kt *kit.Kit, opt *types.HuaWeiSGRuleListOption) (*model.
ListSecurityGroupRulesResponse, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "security group rule list option is required")
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	client, err := h.clientSet.vpcClient(region.ValueOf(opt.Region))
	if err != nil {
		return nil, fmt.Errorf("new vpc client failed, err: %v", err)
	}

	req := &model.ListSecurityGroupRulesRequest{
		SecurityGroupId: sliceToPtr[string]([]string{opt.CloudSecurityGroupID}),
	}

	if opt.Page != nil {
		req.Marker = opt.Page.Marker
		req.Limit = opt.Page.Limit
	}

	resp, err := client.ListSecurityGroupRules(req)
	if err != nil {
		logs.Errorf("list huawei security group rule failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return resp, nil
}
