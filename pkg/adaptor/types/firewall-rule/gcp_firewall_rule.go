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

package firewallrule

import (
	"errors"
	"fmt"

	"hcm/pkg/adaptor/types/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"

	"google.golang.org/api/compute/v1"
)

// -------------------------- Create --------------------------

// CreateOption define gcp firewall rule create option.
type CreateOption struct {
	Type              string                     `json:"type" validate:"required"`
	Name              string                     `json:"name" validate:"required"`
	Description       string                     `json:"description"`
	Priority          int64                      `json:"priority"`
	VpcSelfLink       string                     `json:"vpc_self_link" validate:"required"`
	SourceTags        []string                   `json:"source_tags"`
	TargetTags        []string                   `json:"target_tags"`
	Denied            []corecloud.GcpProtocolSet `json:"denied"`
	Allowed           []corecloud.GcpProtocolSet `json:"allowed"`
	SourceRanges      []string                   `json:"source_ranges"`
	DestinationRanges []string                   `json:"destination_ranges"`
	Disabled          bool                       `json:"disabled"`
}

// Validate gcp firewall rule create option.
func (opt CreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update --------------------------

// UpdateOption define gcp firewall rule update option.
type UpdateOption struct {
	CloudID         string  `json:"cloud_id" validate:"required"`
	GcpFirewallRule *Update `json:"gcp_firewall_rule" validate:"required"`
}

// Validate gcp firewall rule update option.
func (opt UpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// Update define gcp firewall rule when update.
type Update struct {
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

// ListOption define gcp firewall rule list option.
type ListOption struct {
	CloudIDs []uint64      `json:"cloud_ids,omitempty"`
	Page     *core.GcpPage `json:"page" validate:"omitempty"`
}

// Validate gcp firewall rule list option.
func (opt ListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if len(opt.CloudIDs) != 0 && opt.Page != nil {
		return errors.New("list firewall by ids, that not support page")
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Delete --------------------------

// DeleteOption gcp firewall rule delete option.
type DeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate gcp firewall rule delete option.
func (opt DeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpFirewall for compute Firewall
type GcpFirewall struct {
	*compute.Firewall
}

// GetCloudID ...
func (firewall GcpFirewall) GetCloudID() string {
	return fmt.Sprint(firewall.Id)
}

// ConvGcpAllowedProtoSets ...
func ConvGcpAllowedProtoSets(allows []*compute.FirewallAllowed) []corecloud.GcpProtocolSet {
	var result []corecloud.GcpProtocolSet
	for _, allowed := range allows {
		result = append(result, corecloud.GcpProtocolSet{
			Protocol: allowed.IPProtocol,
			Port:     allowed.Ports,
		})
	}
	return result
}

// ConvGcpDeniedProtoSets ...
func ConvGcpDeniedProtoSets(denies []*compute.FirewallDenied) []corecloud.GcpProtocolSet {
	var result []corecloud.GcpProtocolSet
	for _, denied := range denies {
		result = append(result, corecloud.GcpProtocolSet{
			Protocol: denied.IPProtocol,
			Port:     denied.Ports,
		})
	}
	return result
}
