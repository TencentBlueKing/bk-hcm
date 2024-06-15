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

package tcloud

import (
	typeslb "hcm/pkg/adaptor/types/load-balancer"
)

var (
	// LoadBalancerTypeMap 负载均衡类型翻译
	LoadBalancerTypeMap = map[typeslb.TCloudLoadBalancerType]string{
		typeslb.OpenLoadBalancerType:     "公网",
		typeslb.InternalLoadBalancerType: "内网",
	}
	// LoadBalancerChargeTypeNameMap 负载均衡实例计费类型
	LoadBalancerChargeTypeNameMap = map[typeslb.TCloudLoadBalancerChargeType]string{
		typeslb.Prepaid:  "包年包月",
		typeslb.Postpaid: "按量计费",
	}
	// LoadBalancerNetworkChargeTypeNameMap 负载均衡网络计费类型
	LoadBalancerNetworkChargeTypeNameMap = map[typeslb.TCloudLoadBalancerNetworkChargeType]string{
		typeslb.TrafficPostPaidByHour:   "按流量按小时后计费",
		typeslb.BandwidthPostpaidByHour: "按带宽按小时后计费",
		typeslb.BandwidthPackage:        "带宽包计费",
	}
	// IPVersionNameMap ...
	IPVersionNameMap = map[typeslb.TCloudIPVersionForCreate]string{
		typeslb.IPV4IPVersion:          "IPV4",
		typeslb.IPV6NAT64IPVersion:     "IPV6Nat64",
		typeslb.IPV6FullChainIPVersion: "IPV6FullChain",
	}
)
