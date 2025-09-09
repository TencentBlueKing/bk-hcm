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
 * to the current version of the project delivered to anyone in the future.
 */

package cslb

import (
	"hcm/pkg/api/core"
)

// ListUrlRulesByTopologyReq 查询URL规则请求
type ListUrlRulesByTopologyReq struct {
	AccountID string `json:"account_id"`
	Vendor    string `json:"vendor"`

	// 负载均衡器相关条件
	LbRegions      []string `json:"lb_regions"`
	LbNetworkTypes []string `json:"lb_network_types"`
	LbIpVersions   []string `json:"lb_ip_versions"`
	CloudLbIds     []string `json:"cloud_lb_ids"`
	LbVips         []string `json:"lb_vips"`
	LbDomains      []string `json:"lb_domains"`

	// 监听器相关条件
	LblProtocols []string `json:"lbl_protocols"`
	LblPorts     []int    `json:"lbl_ports"`

	// 目标相关条件
	TargetIps   []string `json:"target_ips"`
	TargetPorts []int    `json:"target_ports"`

	// 规则相关条件
	RuleUrls    []string `json:"rule_urls"`
	RuleDomains []string `json:"rule_domains"`

	Page *core.BasePage `json:"page"`
}

// ListUrlRulesByTopologyResp 查询URL规则响应
type ListUrlRulesByTopologyResp struct {
	Count   int             `json:"count"`
	Details []UrlRuleDetail `json:"details"`
}

// UrlRuleDetail URL规则详情
type UrlRuleDetail struct {
	ID           string `json:"id"`
	Ip           string `json:"ip"`
	LbID         string `json:"lb_id"`
	LblProtocols string `json:"lbl_protocols"`
	LblPort      int    `json:"lbl_port"`
	RuleUrl      string `json:"rule_url"`
	RuleDomain   string `json:"rule_domain"`
	TargetCount  int    `json:"target_count"`
	CloudLblID   string `json:"cloud_lbl_id"`
}

// HasLbConditions 是否有负载均衡器相关条件
func (req *ListUrlRulesByTopologyReq) HasLbConditions() bool {
	return len(req.LbRegions) > 0 || len(req.LbNetworkTypes) > 0 || len(req.LbIpVersions) > 0 ||
		len(req.CloudLbIds) > 0 || len(req.LbVips) > 0 || len(req.LbDomains) > 0
}

// HasListenerConditions 是否有监听器相关条件
func (req *ListUrlRulesByTopologyReq) HasListenerConditions() bool {
	return len(req.LblProtocols) > 0 || len(req.LblPorts) > 0
}

// HasRuleConditions 是否有规则相关条件
func (req *ListUrlRulesByTopologyReq) HasRuleConditions() bool {
	return len(req.RuleUrls) > 0 || len(req.RuleDomains) > 0
}

// HasTargetConditions 是否有目标相关条件
func (req *ListUrlRulesByTopologyReq) HasTargetConditions() bool {
	return len(req.TargetIps) > 0 || len(req.TargetPorts) > 0
}

// Validate 验证请求参数
func (req *ListUrlRulesByTopologyReq) Validate() error {
	return nil
}
