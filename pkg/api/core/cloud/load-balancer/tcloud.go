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

package loadbalancer

import (
	cvt "hcm/pkg/tools/converter"
)

// TCloudLoadBalancer ...
type TCloudLoadBalancer = LoadBalancer[TCloudClbExtension]

// TCloudClbExtension tcloud clb extension.
type TCloudClbExtension struct {
	/*
		SlaType 性能容量型规格。
		若需要创建性能容量型实例，则此参数必填，取值范围：
		clb.c2.medium：标准型规格
		clb.c3.small：高阶型1规格
		clb.c3.medium：高阶型2规格
		clb.c4.small：超强型1规格
		clb.c4.medium：超强型2规格
		clb.c4.large：超强型3规格
		clb.c4.xlarge：超强型4规格
		若需要创建共享型实例，则无需填写此参数。
	*/
	SlaType string `json:"sla_type,omitempty"`

	/*
		VipIsp 仅适用于公网负载均衡。
		目前仅广州、上海、南京、济南、杭州、福州、北京、石家庄、武汉、长沙、成都、重庆地域支持静态单线 IP 线路类型，
		可选择中国移动（CMCC）、中国联通（CUCC）或中国电信（CTCC）的运营商类型，网络计费模式只能使用按带宽包计费(BANDWIDTH_PACKAGE)。
		如果不指定本参数，则默认使用BGP。可通过 DescribeResources 接口查询一个地域所支持的Isp。
	*/
	VipIsp string `json:"vip_isp,omitempty"`

	/*
		LoadBalancerPassToTarget Target是否放通来自CLB的流量。
		开启放通（true）：只验证CLB上的安全组；
		不开启放通（false）：需同时验证CLB和后端实例上的安全组。
	*/
	LoadBalancerPassToTarget bool `json:"load_balancer_pass_to_target,omitempty"`

	/*
		InternetMaxBandwidthOut 最大出带宽，单位Mbps，仅对公网属性的共享型、性能容量型和独占型 CLB 实例、以及内网属性的性能容量型 CLB 实例生效。
		- 对于公网属性的共享型和独占型 CLB 实例，最大出带宽的范围为1Mbps-2048Mbps。
		- 对于公网属性和内网属性的性能容量型 CLB实例，最大出带宽的范围为1Mbps-61440Mbps。
		（调用CreateLoadBalancer创建LB时不指定此参数则设置为默认值10Mbps。此上限可调整）
		注意：此字段可能返回 null，表示取不到有效值。
		示例值：1
	*/
	InternetMaxBandwidthOut int64 `json:"internet_max_bandwidth_out,omitempty"`
	/*
		InternetChargeType
		TRAFFIC_POSTPAID_BY_HOUR 按流量按小时后计费 ;
		BANDWIDTH_POSTPAID_BY_HOUR 按带宽按小时后计费;
		BANDWIDTH_PACKAGE 按带宽包计费;BANDWIDTH_PREPAID按带宽预付费。
		注意：此字段可能返回 null，表示取不到有效值。
		示例值：BANDWIDTH_POSTPAID_BY_HOUR
	*/
	InternetChargeType string `json:"internet_charge_type,omitempty"`
	/*
		BandwidthpkgSubType 带宽包的类型，如SINGLEISP（单线）、BGP（多线）。
		注意：此字段可能返回 null，表示取不到有效值。
	*/
	BandwidthpkgSubType string `json:"bandwidthpkg_sub_type,omitempty"`

	/*
		BandwidthPackageId 带宽包ID，
		指定此参数时，网络计费方式（InternetAccessible.InternetChargeType）只支持按带宽包计费（BANDWIDTH_PACKAGE）。
		非上移用户购买的 IPv6 负载均衡实例，且运营商类型非 BGP 时 ，不支持指定具体带宽包id。
		示例值：bwp-pnbe****
	*/
	BandwidthPackageId *string `json:"bandwidth_package_id,omitempty"`

	// IP地址版本为ipv6时此字段有意义， IPv6Nat64 | IPv6FullChain
	IPv6Mode string `json:"ipv6_mode,omitempty"`

	// 在 2016 年 12 月份之前的传统型内网负载均衡都是开启了 snat 的。
	Snat bool `json:"snat,omitempty" `

	// 是否开启SnatPro。
	SnatPro bool `json:"snat_pro,omitempty"`

	// 开启SnatPro负载均衡后，SnatIp列表。
	SnatIps []SnatIp `json:"snat_ips,omitempty"`

	// 删除保护
	DeleteProtect bool `json:"delete_protect,omitempty"`

	// 网络出口
	Egress string `json:"egress,omitempty"`

	// 双栈混绑 开启IPv6FullChain负载均衡7层监听器支持混绑IPv4/IPv6目标功能。
	MixIpTarget bool `json:"mix_ip_target,omitempty"`
}

// SnatIp ...
type SnatIp struct {
	// 私有网络子网的唯一性id，如subnet-12345678
	SubnetId *string `json:"subnet_id" `

	// IP地址，如192.168.0.1
	Ip *string `json:"ip"`
}

// Hash use to compare: {SubnetId},{Ip}
func (ip *SnatIp) Hash() string {
	return ip.String()
}

// String()
func (ip *SnatIp) String() string {
	return cvt.PtrToVal(ip.SubnetId) + "," + cvt.PtrToVal(ip.Ip)
}

// TCloudTargetGroupExtension tcloud target group extension.
type TCloudTargetGroupExtension struct {
}
