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

package cloudserver

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- List --------------------------

// GcpFirewallRuleListReq define gcp firewall rule list req.
type GcpFirewallRuleListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate gcp firewall rule list req.
func (req *GcpFirewallRuleListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// GcpFirewallRuleUpdateReq define gcp firewall rule update req.
type GcpFirewallRuleUpdateReq struct {
	Memo              string                     `json:"memo"`
	Priority          int64                      `json:"priority"`
	SourceTags        []string                   `json:"source_tags"`
	TargetTags        []string                   `json:"target_tags"`
	Denied            []corecloud.GcpProtocolSet `json:"denied"`
	Allowed           []corecloud.GcpProtocolSet `json:"allowed"`
	SourceRanges      []string                   `json:"source_ranges"`
	DestinationRanges []string                   `json:"destination_ranges"`
	Disabled          bool                       `json:"disabled"`
}

// Validate gcp firewall rule update req.
func (req *GcpFirewallRuleUpdateReq) Validate() error {

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	for _, tag := range req.SourceTags {
		if err := validator.ValidateGcpName(tag); err != nil {
			return fmt.Errorf("source tags validate failed, err: %v", err)
		}
	}

	for _, tag := range req.TargetTags {
		if err := validator.ValidateGcpName(tag); err != nil {
			return fmt.Errorf("target tags validate failed, err: %v", err)
		}
	}

	return nil
}

// AssignGcpFirewallRuleToBizReq define assign gcp firewall rule to biz req.
type AssignGcpFirewallRuleToBizReq struct {
	BkBizID         int64    `json:"bk_biz_id" validate:"required"`
	FirewallRuleIDs []string `json:"firewall_rule_ids" validate:"required"`
}

// Validate assign security group to biz request.
func (req *AssignGcpFirewallRuleToBizReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.FirewallRuleIDs) == 0 {
		return errors.New("firewall rule ids is required")
	}

	if len(req.FirewallRuleIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("firewall rule ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Delete --------------------------

// GcpFirewallRuleBatchDeleteReq gcp firewall rule batch delete request.
type GcpFirewallRuleBatchDeleteReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate gcp firewall rule batch delete request.
func (req *GcpFirewallRuleBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Create --------------------------

// GcpFirewallRuleCreateReq ...
type GcpFirewallRuleCreateReq struct {
	AccountID         string                     `json:"account_id" validate:"required"`
	CloudVpcID        string                     `json:"cloud_vpc_id" validate:"required"`
	Name              string                     `json:"name" validate:"required"`
	Memo              string                     `json:"memo"`
	Priority          int64                      `json:"priority" validate:"omitempty"`
	Type              string                     `json:"type" validate:"required"`
	SourceTags        []string                   `json:"source_tags"`
	TargetTags        []string                   `json:"target_tags"`
	Denied            []corecloud.GcpProtocolSet `json:"denied"`
	Allowed           []corecloud.GcpProtocolSet `json:"allowed"`
	SourceRanges      []string                   `json:"source_ranges"`
	DestinationRanges []string                   `json:"destination_ranges"`
	Disabled          bool                       `json:"disabled"`
}

// Validate gcp firewall rule create req.
func (req *GcpFirewallRuleCreateReq) Validate() error {

	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := validator.ValidateGcpName(req.Name); err != nil {
		return err
	}

	for _, tag := range req.SourceTags {
		if err := validator.ValidateGcpName(tag); err != nil {
			return fmt.Errorf("source tags validate failed, err: %v", err)
		}
	}

	for _, tag := range req.TargetTags {
		if err := validator.ValidateGcpName(tag); err != nil {
			return fmt.Errorf("target tags validate failed, err: %v", err)
		}
	}

	return nil
}
