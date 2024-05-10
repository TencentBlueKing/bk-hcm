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
	// FlowRetryMaxLimit Flow重试的最大次数
	FlowRetryMaxLimit = 864000
)

// 腾讯云CLB相关常量
const (
	// TCLBDescribeMax 腾讯云CLB默认查询大小
	TCLBDescribeMax = 20
	// TCLBDeleteProtect 腾讯云负载均衡删除保护
	TCLBDeleteProtect = "DeleteProtect"
)
