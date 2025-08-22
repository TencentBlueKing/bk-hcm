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

package cslb

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// ListUrlRulesByTopologyReq list url rules by topology.
type ListUrlRulesByTopologyReq struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`

	AccountID string `json:"account_id" validate:"required"`

	LbRegions []string `json:"lb_regions" validate:"omitempty,max=500"`

	LbNetworkTypes []string `json:"lb_network_types" validate:"omitempty"`

	LbIpVersions []string `json:"lb_ip_versions" validate:"omitempty"`

	CloudLbIds []string `json:"cloud_lb_ids" validate:"omitempty,max=500"`

	LbVips []string `json:"lb_vips" validate:"omitempty,max=500"`

	LbDomains []string `json:"lb_domains" validate:"omitempty,max=500"`

	LblProtocols []string `json:"lbl_protocols" validate:"omitempty"`

	LblPorts []int `json:"lbl_ports" validate:"omitempty,max=1000"`

	RuleDomains []string `json:"rule_domains" validate:"omitempty,max=500"`

	RuleUrls []string `json:"rule_urls" validate:"omitempty,max=500"`

	TargetIps []string `json:"target_ips" validate:"omitempty,max=5000"`

	TargetPorts []int `json:"target_ports" validate:"omitempty,max=500"`

	Page *core.BasePage `json:"page" validate:"required"`
}

// ListUrlRulesByTopologyResp list url rules by topology resp.
type ListUrlRulesByTopologyResp struct {
	Count   int             `json:"count"`
	Details []UrlRuleDetail `json:"details"`
}

// UrlRuleDetail url rule detail.
type UrlRuleDetail struct {
	ID           string `json:"id"`
	Ip           string `json:"ip"`
	LblProtocols string `json:"lbl_protocols"`
	LblPort      int    `json:"lbl_port"`
	RuleUrl      string `json:"rule_url"`
	RuleDomain   string `json:"rule_domain"`
	TargetCount  int    `json:"target_count"`
	ListenerID   string `json:"listener_id"`
}

// TargetDetail target detail.
type TargetDetail struct {
	ID      string `json:"id"`
	CloudID string `json:"cloud_id"`
	IP      string `json:"ip"`
	Port    int    `json:"port"`
	Weight  int    `json:"weight"`
	Status  string `json:"status"`
}
