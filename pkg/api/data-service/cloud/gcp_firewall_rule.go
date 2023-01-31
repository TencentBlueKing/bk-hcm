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

package cloud

import (
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// GcpFirewallRuleBatchCreateReq gcp firewall rule batch create request.
type GcpFirewallRuleBatchCreateReq struct {
	FirewallRules []GcpFirewallRuleBatchCreate `json:"gcp_firewall_rules" validate:"required"`
}

// Validate gcp firewall rule create request.
func (req *GcpFirewallRuleBatchCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpFirewallRuleBatchCreate define gcp firewall rule when create.
type GcpFirewallRuleBatchCreate struct {
	CloudID               string                     `json:"cloud_id" validate:"required"`
	AccountID             string                     `json:"account" validate:"required"`
	Name                  string                     `json:"name" validate:"required"`
	Priority              int64                      `json:"priority"`
	Memo                  string                     `json:"memo"`
	CloudVpcID            string                     `json:"cloud_vpc_id" validate:"required"`
	VpcId                 string                     `json:"vpc_id" validate:"required"`
	SourceRanges          []string                   `json:"source_ranges"`
	BkBizID               int64                      `json:"bk_biz_id" validate:"required"`
	DestinationRanges     []string                   `json:"destination_ranges"`
	SourceTags            []string                   `json:"source_tags"`
	TargetTags            []string                   `json:"target_tags"`
	SourceServiceAccounts []string                   `json:"source_service_accounts"`
	TargetServiceAccounts []string                   `json:"target_service_accounts"`
	Denied                []corecloud.GcpProtocolSet `json:"denied"`
	Allowed               []corecloud.GcpProtocolSet `json:"allowed"`
	Type                  string                     `json:"type" validate:"required"`
	LogEnable             bool                       `json:"log_enable"`
	Disabled              bool                       `json:"disabled"`
	SelfLink              string                     `json:"self_link"`
}

// -------------------------- Update --------------------------

// GcpFirewallRuleBatchUpdateReq gcp firewall rule batch update request.
type GcpFirewallRuleBatchUpdateReq struct {
	FirewallRules []GcpFirewallRuleBatchUpdate `json:"gcp_firewall_rules" validate:"required"`
}

// Validate gcp firewall rule update request.
func (req *GcpFirewallRuleBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpFirewallRuleBatchUpdate define gcp firewall rule when update.
type GcpFirewallRuleBatchUpdate struct {
	ID                    string                     `json:"id" validate:"required"`
	CloudID               string                     `json:"cloud_id"`
	AccountID             string                     `json:"account_id"`
	Name                  string                     `json:"name"`
	Priority              int64                      `json:"priority"`
	Memo                  string                     `json:"memo"`
	CloudVpcID            string                     `json:"cloud_vpc_id"`
	VpcId                 string                     `json:"vpc_id"`
	SourceRanges          []string                   `json:"source_ranges"`
	BkBizID               int64                      `json:"bk_biz_id"`
	DestinationRanges     []string                   `json:"destination_ranges"`
	SourceTags            []string                   `json:"source_tags"`
	TargetTags            []string                   `json:"target_tags"`
	SourceServiceAccounts []string                   `json:"source_service_accounts"`
	TargetServiceAccounts []string                   `json:"target_service_accounts"`
	Denied                []corecloud.GcpProtocolSet `json:"denied"`
	Allowed               []corecloud.GcpProtocolSet `json:"allowed"`
	Type                  string                     `json:"type" validate:"required"`
	LogEnable             bool                       `json:"log_enable"`
	Disabled              bool                       `json:"disabled"`
	SelfLink              string                     `json:"self_link"`
}

// -------------------------- List --------------------------

// GcpFirewallRuleListReq gcp firewall rule list request.
type GcpFirewallRuleListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *types.BasePage    `json:"page" validate:"required"`
}

// Validate gcp firewall rule list request.
func (req *GcpFirewallRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpFirewallRuleListResult define gcp firewall rule list result.
type GcpFirewallRuleListResult struct {
	Count   uint64                      `json:"count,omitempty"`
	Details []corecloud.GcpFirewallRule `json:"details,omitempty"`
}

// GcpFirewallRuleListResp define gcp firewall rule list resp.
type GcpFirewallRuleListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *GcpFirewallRuleListResult `json:"data"`
}

// -------------------------- Delete --------------------------

// GcpFirewallRuleBatchDeleteReq gcp firewall rule delete request.
type GcpFirewallRuleBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate gcp firewall rule delete request.
func (req *GcpFirewallRuleBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
