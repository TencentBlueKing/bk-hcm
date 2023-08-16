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

package hcservice

import (
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Update --------------------------

// GcpFirewallRuleUpdateReq define gcp firewall rule update req.
type GcpFirewallRuleUpdateReq struct {
	Memo              string                     `json:"memo"`
	Priority          int64                      `json:"priority" validate:"required"`
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
	return validator.Validate.Struct(req)
}

// -------------------------- Create --------------------------

// GcpFirewallRuleCreateReq define gcp firewall rule create req.
type GcpFirewallRuleCreateReq struct {
	BkBizID           int64                      `json:"bk_biz_id" validate:"required"`
	AccountID         string                     `json:"account_id" validate:"required"`
	CloudVpcID        string                     `json:"cloud_vpc_id" validate:"required"`
	Type              string                     `json:"type" validate:"required"`
	Name              string                     `json:"name" validate:"required"`
	Priority          int64                      `json:"priority" validate:"omitempty"`
	Memo              string                     `json:"memo"`
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
	return validator.Validate.Struct(req)
}
