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
	SlaType *string `json:"sla_type,omitempty"`
	/*
		ChargeType 负载均衡实例的计费类型，PREPAID：包年包月，POSTPAID_BY_HOUR：按量计费
	*/
	ChargeType *string `json:"charge_type,omitempty"`
	/*
		VipIsp 仅适用于公网负载均衡。
		目前仅广州、上海、南京、济南、杭州、福州、北京、石家庄、武汉、长沙、成都、重庆地域支持静态单线 IP 线路类型，
		可选择中国移动（CMCC）、中国联通（CUCC）或中国电信（CTCC）的运营商类型，网络计费模式只能使用按带宽包计费(BANDWIDTH_PACKAGE)。
		如果不指定本参数，则默认使用BGP。可通过 DescribeResources 接口查询一个地域所支持的Isp。
	*/
	VipIsp *string `json:"vip_isp,omitempty"`

	/*
		LoadBalancerPassToTarget Target是否放通来自CLB的流量。
		开启放通（true）：只验证CLB上的安全组；
		不开启放通（false）：需同时验证CLB和后端实例上的安全组。
	*/
	LoadBalancerPassToTarget *bool `json:"load_balancer_pass_to_target,omitempty"`

	/*
		InternetMaxBandwidthOut 最大出带宽，单位Mbps，仅对公网属性的共享型、性能容量型和独占型 CLB 实例、以及内网属性的性能容量型 CLB 实例生效。
		- 对于公网属性的共享型和独占型 CLB 实例，最大出带宽的范围为1Mbps-2048Mbps。
		- 对于公网属性和内网属性的性能容量型 CLB实例，最大出带宽的范围为1Mbps-61440Mbps。
		（调用CreateLoadBalancer创建LB时不指定此参数则设置为默认值10Mbps。此上限可调整）
		注意：此字段可能返回 null，表示取不到有效值。
		示例值：1
	*/
	InternetMaxBandwidthOut *int64 `json:"internet_max_bandwidth_out,omitempty"`

	/*
		InternetChargeType
		TRAFFIC_POSTPAID_BY_HOUR 按流量按小时后计费 ;
		BANDWIDTH_POSTPAID_BY_HOUR 按带宽按小时后计费;
		BANDWIDTH_PACKAGE 按带宽包计费;BANDWIDTH_PREPAID按带宽预付费。
		注意：此字段可能返回 null，表示取不到有效值。
		示例值：BANDWIDTH_POSTPAID_BY_HOUR
	*/
	InternetChargeType *string `json:"internet_charge_type,omitempty"`
	/*
		BandwidthpkgSubType 带宽包的类型，如SINGLEISP（单线）、BGP（多线）。
		注意：此字段可能返回 null，表示取不到有效值。
	*/
	BandwidthpkgSubType *string `json:"bandwidthpkg_sub_type,omitempty"`

	/*
		BandwidthPackageId 带宽包ID，
		指定此参数时，网络计费方式（InternetAccessible.InternetChargeType）只支持按带宽包计费（BANDWIDTH_PACKAGE）。
		非上移用户购买的 IPv6 负载均衡实例，且运营商类型非 BGP 时 ，不支持指定具体带宽包id。
		示例值：bwp-pnbe****
	*/
	BandwidthPackageId *string `json:"bandwidth_package_id,omitempty"`

	// IP地址版本为ipv6时此字段有意义， IPv6Nat64 | IPv6FullChain
	IPv6Mode *string `json:"ipv6_mode,omitempty"`

	// 在 2016 年 12 月份之前的传统型内网负载均衡都是开启了 snat 的。
	Snat *bool `json:"snat,omitempty" `

	// 是否开启SnatPro。
	SnatPro *bool `json:"snat_pro,omitempty"`

	// 开启SnatPro负载均衡后，SnatIp列表。
	SnatIps []SnatIp `json:"snat_ips,omitempty"`

	// 删除保护
	DeleteProtect *bool `json:"delete_protect,omitempty"`

	// 网络出口
	Egress *string `json:"egress,omitempty"`

	// 双栈混绑 开启IPv6FullChain负载均衡7层监听器支持混绑IPv4/IPv6目标功能。
	MixIpTarget *bool `json:"mix_ip_target,omitempty"`
}

// SnatIp ...
type SnatIp struct {
	// 私有网络子网的唯一性id，如subnet-12345678
	SubnetId *string `json:"subnet_id"`

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

// TCloudHealthCheckInfo define health check.
type TCloudHealthCheckInfo struct {
	// 是否开启健康检查：1（开启）、0（关闭）
	HealthSwitch *int64 `json:"health_switch"`
	// 健康检查的响应超时时间（仅适用于四层监听器），可选值：2~60，默认值：2，单位：秒。响应超时时间要小于检查间隔时间
	TimeOut *int64 `json:"time_out,omitempty" validate:"omitempty,min=2,max=60"`
	// 健康检查探测间隔时间，默认值：5，IPv4 CLB实例的取值范围为：2-300，IPv6 CLB 实例的取值范围为：5-300。单位：秒
	// 说明：部分老旧 IPv4 CLB实例的取值范围为：5-300
	IntervalTime *int64 `json:"interval_time,omitempty" validate:"omitempty,min=5,max=300"`
	// 健康阈值，默认值：3，表示当连续探测三次健康则表示该转发正常，可选值：2~10，单位：次
	HealthNum *int64 `json:"health_num,omitempty" validate:"omitempty,min=2,max=10"`
	// 不健康阈值，默认值：3，表示当连续探测三次不健康则表示该转发异常，可选值：2~10，单位：次。
	UnHealthNum *int64 `json:"un_health_num,omitempty" validate:"omitempty,min=2,max=10"`
	// 健康检查状态码（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）。可选值：1~31，默认 31。
	// 1 表示探测后返回值 1xx 代表健康，2 表示返回 2xx 代表健康，4 表示返回 3xx 代表健康，8 表示返回 4xx 代表健康，
	// 16 表示返回 5xx 代表健康。若希望多种返回码都可代表健康，则将相应的值相加。
	HttpCode *int64 `json:"http_code" validate:"omitempty,min=1,max=31"`
	// 自定义探测相关参数。健康检查端口，默认为后端服务的端口，除非您希望指定特定端口，否则建议留空。（仅适用于TCP/UDP监听器）
	CheckPort *int64 `json:"check_port,omitempty"`
	// 健康检查使用的协议。取值 TCP | HTTP | HTTPS | GRPC | PING | CUSTOM，UDP监听器支持PING/CUSTOM，
	// TCP监听器支持TCP/HTTP/CUSTOM，TCP_SSL/QUIC监听器支持TCP/HTTP，HTTP规则支持HTTP/GRPC，HTTPS规则支持HTTP/HTTPS/GRPC
	CheckType *string `json:"check_type,omitempty"`
	// HTTP版本。健康检查协议CheckType的值取HTTP时，必传此字段，代表后端服务的HTTP版本：HTTP/1.0、HTTP/1.1；（仅适用于TCP监听器）
	HttpVersion *string `json:"http_version,omitempty"`
	// 健康检查路径（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式）
	HttpCheckPath *string `json:"http_check_path,omitempty"`
	// 健康检查域名（仅适用于HTTP/HTTPS监听器和TCP监听器的HTTP健康检查方式。针对TCP监听器，当使用HTTP健康检查方式时，该参数为必填项）
	HttpCheckDomain *string `json:"http_check_domain,omitempty"`
	// 健康检查方法（仅适用于HTTP/HTTPS转发规则、TCP监听器的HTTP健康检查方式），默认值：HEAD，可选值HEAD或GET
	HttpCheckMethod *string `json:"http_check_method,omitempty"`
	// 健康检查源IP类型：0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP），默认值：0
	SourceIpType *int64 `json:"source_ip_type,omitempty"`
	// 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，代表健康检查的输入格式，可取值：HEX或TEXT；
	// 取值为HEX时，SendContext和RecvContext的字符只能在0123456789ABCDEF中选取且长度必须是偶数位。（仅适用于TCP/UDP监听器）
	ContextType *string `json:"context_type,omitempty"`
	// （仅适用于TCP/UDP监听器）。自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，
	// 代表健康检查发送的请求内容，只允许ASCII可见字符，最大长度限制500。
	SendContext *string `json:"send_context,omitempty"`
	// （仅适用于TCP/UDP监听器）。 自定义探测相关参数。健康检查协议CheckType的值取CUSTOM时，必填此字段，
	// 代表健康检查返回的结果，只允许ASCII可见字符，最大长度限制500。
	RecvContext *string `json:"recv_context,omitempty"`
	// GRPC健康检查状态码（仅适用于后端转发协议为GRPC的规则）。
	// 默认值为 12，可输入值为数值、多个数值, 或者范围，例如 20 或 20,25 或 0-99
	ExtendedCode *string `json:"extended_code,omitempty"`
}

// TCloudCertificateInfo 证书信息，不存储具体证书信息，只保存对应id
type TCloudCertificateInfo struct {
	// 认证类型，UNIDIRECTIONAL：单向认证，MUTUAL：双向认证
	SSLMode *string `json:"ssl_mode"`
	// CA证书，认证RS侧用的证书，只需要公钥
	CaCloudID *string `json:"ca_cloud_id,omitempty"`
	// 客户端证书，客户端向CLB发起请求时认证CLB来源是否可靠的证书。可以配置两个不同的加密类型的证书，
	CertCloudIDs []string `json:"cert_cloud_ids,omitempty"`
}

// TCloudListenerExtension 腾讯云监听器拓展
type TCloudListenerExtension struct {
	Certificate *TCloudCertificateInfo `json:"certificate,omitempty"`
}

// TCloudListener ...
type TCloudListener = Listener[TCloudListenerExtension]
