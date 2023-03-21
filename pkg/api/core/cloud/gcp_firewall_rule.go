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
	"time"
)

// GcpFirewallRule define gcp firewall rule.
type GcpFirewallRule struct {
	ID                    string           `json:"id"`
	CloudID               string           `json:"cloud_id"`
	Name                  string           `json:"name"`
	Priority              int64            `json:"priority"`
	Memo                  string           `json:"memo"`
	CloudVpcID            string           `json:"cloud_vpc_id"`
	VpcSelfLink           string           `json:"vpc_self_link"`
	SourceRanges          []string         `json:"source_ranges"`
	BkBizID               int64            `json:"bk_biz_id"`
	VpcId                 string           `json:"vpc_id"`
	DestinationRanges     []string         `json:"destination_ranges"`
	SourceTags            []string         `json:"source_tags"`
	TargetTags            []string         `json:"target_tags"`
	SourceServiceAccounts []string         `json:"source_service_accounts"`
	TargetServiceAccounts []string         `json:"target_service_accounts"`
	Denied                []GcpProtocolSet `json:"denied"`
	Allowed               []GcpProtocolSet `json:"allowed"`
	Type                  string           `json:"type"`
	LogEnable             bool             `json:"log_enable"`
	Disabled              bool             `json:"disabled"`
	AccountID             string           `json:"account_id"`
	SelfLink              string           `json:"self_link"`
	Creator               string           `json:"creator"`
	Reviser               string           `json:"reviser"`
	CreatedAt             *time.Time       `json:"created_at"`
	UpdatedAt             *time.Time       `json:"updated_at"`
}

// GcpProtocolSet define gcp protocol set.
type GcpProtocolSet struct {
	Protocol string   `json:"protocol"`
	Port     []string `json:"port"`
}
