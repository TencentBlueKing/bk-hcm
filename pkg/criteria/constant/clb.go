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

package constant

import "time"

// 负载均衡相关的常量
const (
	// LoadBalancerBindSecurityGroupMaxLimit 一个负载均衡实例最多可绑定安全组的最大数量限制
	LoadBalancerBindSecurityGroupMaxLimit = 5
	// BatchListenerMaxLimit 单次操作监听器的最大数量
	BatchListenerMaxLimit = 20
	// ListenerMinSessionExpire 监听器最短的会话过期时间，单位秒
	ListenerMinSessionExpire = 30
	// ResFlowLockExpireDays 锁定资源与Flow的最大超时时间，默认7天
	ResFlowLockExpireDays = 7
	// FlowRetryTimeout Flow重试的最大重试超时
	FlowRetryTimeout = 7 * 24 * time.Hour
)

// 腾讯云CLB相关常量
const (
	// TCLBDescribeMax 腾讯云CLB默认查询大小
	TCLBDescribeMax = 20
	// TCLBDeleteProtect 腾讯云负载均衡删除保护
	TCLBDeleteProtect = "DeleteProtect"
)

const (
	// ExportLayer7ListenerLimit 导出七层监听器数量限制
	ExportLayer7ListenerLimit = 5000
	// ExportLayer4ListenerLimit 导出四层监听器数量限制
	ExportLayer4ListenerLimit = 5000
	// ExportRuleLimit 导出规则数量限制
	ExportRuleLimit = 5000
	// ExportLayer7RsLimit 导出七层RS数量限制
	ExportLayer7RsLimit = 5000
	// ExportLayer4RsLimit 导出四层RS数量限制
	ExportLayer4RsLimit = 5000
	// ExportClbOneFileRowLimit 导出文件行数限制
	ExportClbOneFileRowLimit = 5000
	// ExportListenerParamLimit 导出监听器参数限制
	ExportListenerParamLimit = 5000
)

const (
	// CLBFilePrefix 负载均衡文件名前缀
	CLBFilePrefix = "hcm-clb"
	// Layer4ListenerFilePrefix 四层监听器文件名前缀
	Layer4ListenerFilePrefix = "tcp_udp监听器"
	// Layer7ListenerFilePrefix 七层监听器文件名前缀
	Layer7ListenerFilePrefix = "http_https监听器"
	// RuleFilePrefix 规则文件名前缀
	RuleFilePrefix = "http_https规则URL"
	// Layer4RsFilePrefix 四层RS文件名前缀
	Layer4RsFilePrefix = "tcp_udp绑定的RS"
	// Layer7RsFilePrefix 七层RS文件名前缀
	Layer7RsFilePrefix = "http_https绑定的RS"
	// Layer4ListenerSheetName 四层监听器sheet名
	Layer4ListenerSheetName = "批量创建监听器-TCP-UDP"
	// Layer7ListenerSheetName 七层监听器sheet名
	Layer7ListenerSheetName = "批量创建监听器-HTTP-HTTPS"
	// RuleSheetName 规则sheet名
	RuleSheetName = "批量创建URL规则-HTTP-HTTPS"
	// Layer4RsSheetName 四层RS sheet名
	Layer4RsSheetName = "绑定RS-TCP-UDP"
	// Layer7RsSheetName 七层RS sheet名
	Layer7RsSheetName = "绑定RS-HTTP-HTTPS"

	// CLBExcelHeaderVendor excel表头云厂商字段值
	CLBExcelHeaderVendor = "vendor(云厂商)"
	// CLBExcelHeaderTCloud excel表头腾讯云字段值
	CLBExcelHeaderTCloud = "tencent_cloud_public(腾讯云-公有云)"

	// CLBTopoFindInLimit 负载均衡各级拓扑in查询数量限制
	CLBTopoFindInLimit = 10000
)
