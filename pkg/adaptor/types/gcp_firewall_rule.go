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

package types

import (
	"hcm/pkg/adaptor/types/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Update --------------------------

// GcpFirewallRuleUpdateOption define gcp firewall rule update option.
type GcpFirewallRuleUpdateOption struct {
	CloudID         string                 `json:"cloud_id" validate:"required"`
	GcpFirewallRule *GcpFirewallRuleUpdate `json:"gcp_firewall_rule" validate:"required"`
}

// Validate gcp firewall rule update option.
func (opt GcpFirewallRuleUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpFirewallRuleUpdate define gcp firewall rule when update.
type GcpFirewallRuleUpdate struct {
	Description       string                     `json:"description"`
	Priority          int64                      `json:"priority"`
	SourceTags        []string                   `json:"source_tags"`
	TargetTags        []string                   `json:"target_tags"`
	Denied            []corecloud.GcpProtocolSet `json:"denied"`
	Allowed           []corecloud.GcpProtocolSet `json:"allowed"`
	SourceRanges      []string                   `json:"source_ranges"`
	DestinationRanges []string                   `json:"destination_ranges"`
	Disabled          bool                       `json:"disabled"`
	// SourceServiceAccounts 因为产品侧未引入该概念，所以只能支持将该字段更新为空。
	SourceServiceAccounts []string `json:"source_service_accounts"`
	// SourceServiceAccounts 因为产品侧未引入该概念，所以只能支持将该字段更新为空。
	TargetServiceAccounts []string `json:"target_service_accounts"`
}

// -------------------------- List --------------------------

// GcpFirewallRuleListOption define gcp firewall rule list option.
type GcpFirewallRuleListOption struct {
	Page *core.GcpPage `json:"page" validate:"omitempty"`
}

// Validate gcp firewall rule list option.
func (opt GcpFirewallRuleListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Delete --------------------------

// GcpFirewallRuleDeleteOption gcp firewall rule delete option.
type GcpFirewallRuleDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate gcp firewall rule delete option.
func (opt GcpFirewallRuleDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}
