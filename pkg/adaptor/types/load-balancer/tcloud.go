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
	"errors"
	"fmt"
	"strings"

	"hcm/pkg/adaptor/types/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	cvt "hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// TCloudLoadBalancerType 负载均衡实例的网络类型
type TCloudLoadBalancerType string

const (
	// OpenLoadBalancerType 公网属性
	OpenLoadBalancerType TCloudLoadBalancerType = "OPEN"
	// InternalLoadBalancerType 内网属性
	InternalLoadBalancerType TCloudLoadBalancerType = "INTERNAL"
)

// TCloudLoadBalancerInstType 负载均衡实例的类型
type TCloudLoadBalancerInstType int64

const (
	// DefaultLoadBalancerInstType 1:通用的负载均衡实例，目前只支持传入1
	DefaultLoadBalancerInstType TCloudLoadBalancerInstType = 1
)

// TCloudDefaultISP 默认ISP
const TCloudDefaultISP = "BGP"

// TCloudSslMode 腾讯云clb SSL认证类型
type TCloudSslMode string

const (
	// TCloudSslUniDirect 单向认证
	TCloudSslUniDirect TCloudSslMode = "UNIDIRECTIONAL"
	// TCloudSslMutual 双向认证
	TCloudSslMutual TCloudSslMode = "MUTUAL"
)

// TCloudIPVersionForCreate 仅适用于公网负载均衡。IP版本，可取值：IPV4、IPV6、IPv6FullChain，不区分大小写，默认值 IPV4。
// 说明：取值为IPV6表示为IPV6 NAT64版本；取值为IPv6FullChain，表示为IPv6版本。
type TCloudIPVersionForCreate string

// Convert 转换为中的含义 `enumor.IPAddressType`
func (c TCloudIPVersionForCreate) Convert() enumor.IPAddressType {
	switch c {
	case IPV4IPVersion:
		return enumor.Ipv4
	case IPV6NAT64IPVersion:
		return enumor.Ipv6Nat64
	case IPV6FullChainIPVersion:
		return enumor.Ipv6
	}
	return enumor.IPAddressType("UNKNOWN_" + string(c))
}

const (
	// IPV4IPVersion IPV4版本
	IPV4IPVersion TCloudIPVersionForCreate = "IPV4"
	// IPV6NAT64IPVersion IPV6版本
	IPV6NAT64IPVersion TCloudIPVersionForCreate = "IPV6"
	// IPV6FullChainIPVersion IPv6FullChain版本
	IPV6FullChainIPVersion TCloudIPVersionForCreate = "IPv6FullChain"
)

// TCloudLoadBalancerStatus 负载均衡实例的状态
type TCloudLoadBalancerStatus uint64

const (
	// IngStatus 负载均衡实例的状态-创建中
	IngStatus TCloudLoadBalancerStatus = 0
	// SuccessStatus 负载均衡实例的状态-正常运行
	SuccessStatus TCloudLoadBalancerStatus = 1
)

// TCloudLoadBalancerChargeType 实例计费类型。
type TCloudLoadBalancerChargeType string

const (
	// Prepaid 预付费，即包年包月
	Prepaid TCloudLoadBalancerChargeType = "PREPAID"
	// Postpaid 按量计费
	Postpaid TCloudLoadBalancerChargeType = "POSTPAID"
)

// TCloudLoadBalancerNetworkChargeType 腾讯云网络计费类型
type TCloudLoadBalancerNetworkChargeType string

const (
	// TrafficPostPaidByHour  按流量按小时后计费
	TrafficPostPaidByHour TCloudLoadBalancerNetworkChargeType = `TRAFFIC_POSTPAID_BY_HOUR`
	// BandwidthPostpaidByHour 按带宽按小时后计费
	BandwidthPostpaidByHour TCloudLoadBalancerNetworkChargeType = `BANDWIDTH_POSTPAID_BY_HOUR`
	// BandwidthPackage 带宽包计费
	BandwidthPackage TCloudLoadBalancerNetworkChargeType = `BANDWIDTH_PACKAGE`
)

// -------------------------- List Clb--------------------------

// TCloudListOption defines options to list tcloud clb instances.
type TCloudListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty,max=20"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud clb list option.
func (opt TCloudListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// TCloudClb for clb Instance
type TCloudClb struct {
	*tclb.LoadBalancer
}

// GetCloudID get cloud id
func (clb TCloudClb) GetCloudID() string {
	return cvt.PtrToVal(clb.LoadBalancerId)
}

// GetIPVersion 返回ip版本信息
func (clb TCloudClb) GetIPVersion() enumor.IPAddressType {

	ver := strings.ToLower(cvt.PtrToVal(clb.AddressIPVersion))
	if ver == "ipv4" {
		return enumor.Ipv4
	}
	mode := strings.ToLower(cvt.PtrToVal(clb.IPv6Mode))
	if ver == "ipv6" && mode == "ipv6fullchain" {
		return enumor.Ipv6
	}
	if ver == "ipv6" && mode == "ipv6nat64" {
		return enumor.Ipv6Nat64
	}
	// fall back to unknown
	return enumor.IPAddressType(ver)
}

// -------------------------- List Listeners--------------------------

// TCloudListListenersOption defines options to list tcloud listeners instances.
type TCloudListListenersOption struct {
	Region         string              `json:"region" validate:"required"`
	LoadBalancerId string              `json:"load_balancer_id" validate:"required"`
	CloudIDs       []string            `json:"cloud_ids" validate:"omitempty"`
	Protocol       enumor.ProtocolType `json:"protocol" validate:"omitempty"`
	Port           int64               `json:"port" validate:"omitempty"`
}

// Validate tcloud listeners list option.
func (opt TCloudListListenersOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}

// TCloudListener for listener Instance
type TCloudListener struct {
	*tclb.Listener
}

// GetCloudID get cloud id
func (clb TCloudListener) GetCloudID() string {
	return cvt.PtrToVal(clb.ListenerId)
}

// GetProtocol ...
func (clb TCloudListener) GetProtocol() enumor.ProtocolType {
	return cvt.PtrToVal((*enumor.ProtocolType)(clb.Protocol))
}

// -------------------------- List Targets--------------------------

// TCloudListTargetsOption defines options to list tcloud targets instances.
type TCloudListTargetsOption struct {
	Region         string              `json:"region" validate:"required"`
	LoadBalancerId string              `json:"load_balancer_id" validate:"required"`
	ListenerIds    []string            `json:"cloud_ids" validate:"omitempty"`
	Protocol       enumor.ProtocolType `json:"protocol" validate:"omitempty"`
	Port           int64               `json:"port" validate:"omitempty"`
}

// Validate tcloud targets list option.
func (opt TCloudListTargetsOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}

// TCloudListenerTarget for listener target Instance
type TCloudListenerTarget struct {
	*tclb.ListenerBackend
}

// GetProtocol ...
func (t *TCloudListenerTarget) GetProtocol() enumor.ProtocolType {
	return enumor.ProtocolType(cvt.PtrToVal(t.Protocol))
}

// -------------------------- Create Clb--------------------------

// TCloudCreateClbOption defines options to create tcloud clb instances.
type TCloudCreateClbOption struct {
	Region                   string                     `json:"region" validate:"required"`
	LoadBalancerType         TCloudLoadBalancerType     `json:"load_balancer_type" validate:"required"`
	Forward                  TCloudLoadBalancerInstType `json:"forward" validate:"omitempty"`
	LoadBalancerName         *string                    `json:"load_balancer_name" validate:"omitempty,max=60"`
	VpcID                    *string                    `json:"vpc_id" validate:"omitempty"`
	SubnetID                 *string                    `json:"subnet_id" validate:"omitempty"`
	ProjectID                *int64                     `json:"project_id" validate:"omitempty"`
	AddressIPVersion         TCloudIPVersionForCreate   `json:"address_ip_version" validate:"omitempty"`
	Number                   *uint64                    `json:"number" validate:"omitempty,min=1"`
	MasterZoneID             *string                    `json:"master_zone_id" validate:"omitempty"`
	ZoneID                   *string                    `json:"zone_id" validate:"omitempty"`
	VipIsp                   *string                    `json:"vip_isp" validate:"omitempty"`
	Tags                     []*tclb.TagInfo            `json:"tags" validate:"omitempty"`
	Vip                      *string                    `json:"vip" validate:"omitempty"`
	BandwidthPackageID       *string                    `json:"bandwidth_package_id" validate:"omitempty"`
	ExclusiveCluster         *tclb.ExclusiveCluster     `json:"exclusive_cluster" validate:"omitempty"`
	SlaType                  *string                    `json:"sla_type" validate:"omitempty"`
	ClusterIds               []*string                  `json:"cluster_ids" validate:"omitempty"`
	ClientToken              *string                    `json:"client_token" validate:"omitempty"`
	SnatPro                  *bool                      `json:"snat_pro" validate:"omitempty"`
	SnatIps                  []*tclb.SnatIp             `json:"snat_ips" validate:"omitempty"`
	ClusterTag               *string                    `json:"cluster_tag" validate:"omitempty"`
	SlaveZoneID              *string                    `json:"slave_zone_id" validate:"omitempty"`
	EipAddressID             *string                    `json:"eip_address_id" validate:"omitempty"`
	LoadBalancerPassToTarget *bool                      `json:"load_balancer_pass_to_target" validate:"omitempty"`
	DynamicVip               *bool                      `json:"dynamic_vip" validate:"omitempty"`
	Egress                   *string                    `json:"egress" validate:"omitempty"`

	InternetChargeType      *TCloudLoadBalancerNetworkChargeType `json:"internet_charge_type"`
	InternetMaxBandwidthOut *int64                               `json:"internet_max_bandwidth_out" `
	BandwidthpkgSubType     *string                              `json:"bandwidthpkg_sub_type" validate:"omitempty"`

	// 不填默认按量付费
	LoadBalancerChargeType TCloudLoadBalancerChargeType `json:"load_balancer_charge_type"`
}

// Validate tcloud clb create option.
func (opt TCloudCreateClbOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudDescribeResourcesOption defines options to list tcloud listeners instances.
type TCloudDescribeResourcesOption struct {
	Region      string   `json:"region" validate:"required"`
	MasterZones []string `json:"master_zone" validate:"omitempty"`
	IPVersion   []string `json:"ip_version" validate:"omitempty"`
	ISP         []string `json:"isp" validate:"omitempty"`
	Limit       *uint64  `json:"limit"  validate:"omitempty"`
	Offset      *uint64  `json:"offset" validate:"omitempty"`
}

// Validate tcloud listeners list option.
func (opt TCloudDescribeResourcesOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// --------------------------[设置负载均衡实例的安全组]--------------------------

// TCloudSetClbSecurityGroupOption defines options to set tcloud clb security-group option.
type TCloudSetClbSecurityGroupOption struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerID string   `json:"load_balancer_id" validate:"required"`
	SecurityGroups []string `json:"security_groups" validate:"omitempty,max=50"`
}

// Validate tcloud clb security-group option.
func (opt TCloudSetClbSecurityGroupOption) Validate() error {
	if len(opt.SecurityGroups) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("invalid page.limit max value: %d", constant.LoadBalancerBindSecurityGroupMaxLimit)
	}

	return validator.Validate.Struct(opt)
}

// TCloudDeleteOption 批量删除
type TCloudDeleteOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required,min=1"`
}

// Validate ...
func (opt TCloudDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudUpdateOption ...
type TCloudUpdateOption struct {
	Region string `json:"region" validate:"required"`
	// 负载均衡的唯一ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`

	// 负载均衡实例名称
	LoadBalancerName *string `json:"load_balancer_name"`

	// 设置负载均衡跨地域绑定1.0的后端服务信息
	TargetRegionInfo *tclb.TargetRegionInfo `json:"target_region_info" `

	// 网络计费相关参数
	InternetChargeType      *string `json:"internet_charge_type"`
	InternetMaxBandwidthOut *int64  `json:"internet_max_bandwidth_out" `
	BandwidthpkgSubType     *string `json:"bandwidthpkg_sub_type" validate:"omitempty"`

	// Target是否放通来自CLB的流量。开启放通（true）：只验证CLB上的安全组；不开启放通（false）：需同时验证CLB和后端实例上的安全组。
	LoadBalancerPassToTarget *bool `json:"load_balancer_pass_to_target"`

	// 是否开启跨地域绑定2.0功能
	SnatPro *bool `json:"snat_pro" `

	// 是否开启删除保护
	DeleteProtect *bool `json:"delete_protect" `

	// 将负载均衡二级域名由mycloud.com改为tencentclb.com，子域名也会变换。修改后mycloud.com域名将失效。
	ModifyClassicDomain *bool `json:"modify_classic_domain" `
}

// Validate ...
func (opt TCloudUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudDescribeTaskStatusOption 查询异步任务状态
type TCloudDescribeTaskStatusOption struct {
	Region string `json:"region" validate:"required"`
	// TaskId 请求ID，即接口返回的 RequestId 参数。
	TaskId   string `json:"task_id"`
	DealName string `json:"deal_name"`
}

// Validate ...
func (opt TCloudDescribeTaskStatusOption) Validate() error {
	if len(opt.TaskId)+len(opt.DealName) == 0 {
		return errors.New("both task_id and deal_name is empty")
	}
	return validator.Validate.Struct(opt)
}

// -------------------------- Create Listener --------------------------

// TCloudCreateListenerOption defines options to create tcloud listener instances.
type TCloudCreateListenerOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例 ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ListenerName 要创建的监听器名称
	ListenerName string `json:"listener_name" validate:"required"`
	// Protocol 监听器协议： TCP | UDP | HTTP | HTTPS | TCP_SSL | QUIC
	Protocol enumor.ProtocolType `json:"protocol" validate:"required"`
	// Port 要将监听器创建到哪个端口
	Port int64 `json:"port" validate:"required"`
	// HealthCheck 健康检查相关参数，此参数仅适用于TCP/UDP/TCP_SSL/QUIC监听器
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check"`
	// Certificate 证书相关信息，此参数仅适用于TCP_SSL监听器和未开启SNI特性的HTTPS监听器。
	Certificate *corelb.TCloudCertificateInfo `json:"certificate"`
	// SessionExpireTime 会话保持时间，单位：秒。可选值：30~3600，默认 0，表示不开启。此参数仅适用于TCP/UDP监听器
	SessionExpireTime int64 `json:"session_expire_time"`
	// 监听器转发的方式。可选值：WRR、LEAST_CONN
	// Scheduler 分别表示按权重轮询、最小连接数， 默认为 WRR。此参数仅适用于TCP/UDP/TCP_SSL/QUIC监听器。
	Scheduler string `json:"scheduler"`
	// SniSwitch 是否开启SNI特性，此参数仅适用于HTTPS监听器。
	SniSwitch enumor.SniType `json:"sni_switch"`
	// TargetType 后端目标类型，NODE表示绑定普通节点，TARGETGROUP表示绑定目标组
	TargetType string `json:"target_type"`
	// 会话保持类型。不传或传NORMAL表示默认会话保持类型。
	// SessionType QUIC_CID 表示根据Quic Connection ID做会话保持。QUIC_CID只支持UDP协议
	SessionType *string `json:"session_type"`
	// KeepaliveEnable 是否开启长连接，此参数仅适用于HTTP/HTTPS监听器，0:关闭；1:开启， 默认关闭
	KeepaliveEnable int64 `json:"keepalive_enable"`
	// 创建端口段监听器时必须传入此参数，用以标识结束端口。同时，入参Ports只允许传入一个成员，用以标识开始端口。
	// EndPort 【如果您需要体验端口段功能，请通过 [工单申请](https://console.cloud.tencent.com/workorder/category)】。
	EndPort uint64 `json:"end_port"`
}

// Validate tcloud listener create option.
func (opt TCloudCreateListenerOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Listener --------------------------

// TCloudUpdateListenerOption ...
type TCloudUpdateListenerOption struct {
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ListenerId 负载均衡监听器ID
	ListenerId string `json:"listener_id" validate:"required"`
	// ListenerName 新的监听器名称
	ListenerName string `json:"listener_name"`
	// SessionExpireTime 会话保持时间，单位：秒。可选值：30~3600，默认 0，表示不开启。此参数仅适用于TCP/UDP监听器。
	SessionExpireTime int64 `json:"session_expire_time"`
	// HealthCheck 健康检查相关参数，此参数仅适用于TCP/UDP/TCP_SSL/QUIC监听器。
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check"`
	// Certificate 证书相关信息，此参数仅适用于HTTPS/TCP_SSL/QUIC监听器；此参数和MultiCertInfo不能同时传入。
	Certificate *corelb.TCloudCertificateInfo `json:"certificate"`
	// 监听器转发的方式。可选值：WRR、LEAST_CONN
	// Scheduler 分别表示按权重轮询、最小连接数， 默认为 WRR。
	Scheduler string `json:"scheduler"`
	// SniSwitch 是否开启SNI特性，此参数仅适用于HTTPS监听器。注意：未开启SNI的监听器可以开启SNI；已开启SNI的监听器不能关闭SNI。
	SniSwitch enumor.SniType `json:"sni_switch"`
	// TargetType 后端目标类型，NODE表示绑定普通节点，TARGETGROUP表示绑定目标组。
	TargetType string `json:"target_type"`
	// KeepaliveEnable 是否开启长连接，此参数仅适用于HTTP/HTTPS监听器。
	KeepaliveEnable int64 `json:"keepalive_enable"`
	// SessionType 会话保持类型。NORMAL表示默认会话保持类型。QUIC_CID表示根据Quic Connection ID做会话保持。QUIC_CID只支持UDP协议。
	SessionType string `json:"session_type"`
}

// Validate ...
func (opt TCloudUpdateListenerOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Delete Listener --------------------------

// TCloudDeleteListenerOption 批量删除监听器
type TCloudDeleteListenerOption struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerId string   `json:"load_balancer_id" validate:"required"`
	CloudIDs       []string `json:"cloud_ids" validate:"required,min=1"`
}

// Validate ...
func (opt TCloudDeleteListenerOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create Rule --------------------------

// TCloudCreateRuleOption defines options to create tcloud rule instances.
type TCloudCreateRuleOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ListenerId监听器ID
	ListenerId string `json:"listener_id" validate:"required"`
	// Rules 新建转发规则的信息
	Rules []*RuleInfo `json:"rules" validate:"required,min=1"`
}

// RuleInfo 规则基本信息
type RuleInfo struct {
	// Url 转发规则的路径。长度限制为：1~200。
	Url *string `json:"url,omitempty"  validate:"required,max=200"`
	// Domain 转发规则的域名。长度限制为：1~80。Domain和Domains只需要传一个，单域名规则传Domain，多域名规则传Domains。
	Domain *string `json:"domain,omitempty"`
	// SessionExpireTime 会话保持时间。设置为0表示关闭会话保持，开启会话保持可取值30~86400，单位：秒。
	SessionExpireTime *int64 `json:"session_expire_time,omitempty" validate:"omitempty,min=30,max=86400"`
	// HealthCheck 健康检查信息。详情请参见：[健康检查](https://cloud.tencent.com/document/product/214/6097)
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check,omitempty"`
	// Certificate 证书信息；此参数和MultiCertInfo不能同时传入。
	Certificate *corelb.TCloudCertificateInfo `json:"certificate,omitempty"`
	// 规则的请求转发方式，可选值：WRR、LEAST_CONN、IP_HASH
	// Scheduler 分别表示按权重轮询、最小连接数、按IP哈希， 默认为 WRR。
	Scheduler *string `json:"scheduler,omitempty"`
	// ForwardType 负载均衡与后端服务之间的转发协议，目前支持 HTTP/HTTPS/GRPC/TRPC，TRPC暂未对外开放，默认HTTP。
	ForwardType *string `json:"forward_type,omitempty"`
	// DefaultServer 是否将该域名设为默认域名，注意，一个监听器下只能设置一个默认域名。
	DefaultServer *bool `json:"default_server,omitempty"`
	// Http2 是否开启Http2，注意，只有HTTPS域名才能开启Http2。
	Http2 *bool `json:"http2,omitempty"`
	// TargetType 后端目标类型，NODE表示绑定普通节点，TARGETGROUP表示绑定目标组
	TargetType *string `json:"target_type,omitempty"`
	// TrpcCallee TRPC被调服务器路由，ForwardType为TRPC时必填。目前暂未对外开放。
	TrpcCallee *string `json:"trpc_callee,omitempty"`
	// TrpcFunc TRPC调用服务接口，ForwardType为TRPC时必填。目前暂未对外开放
	TrpcFunc *string `json:"trpc_func,omitempty"`
	// Quic 是否开启QUIC，注意，只有HTTPS域名才能开启QUIC
	Quic *bool `json:"quic,omitempty"`
	// Domains 转发规则的域名列表。每个域名的长度限制为：1~80。Domain和Domains只需要传一个，单域名规则传Domain，多域名规则传Domains。
	Domains []*string `json:"domains,omitempty"`
}

// Validate tcloud rule create option.
func (opt TCloudCreateRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Rule --------------------------

// TCloudUpdateRuleOption defines options to update tcloud rule instances.
type TCloudUpdateRuleOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ListenerId监听器ID
	ListenerId string `json:"listener_id" validate:"required"`
	// LocationId 规则id
	LocationId string `json:"location_id" validate:"required"`
	// 转发规则的新的转发路径，如不需修改Url，则不需提供此参数。
	Url *string `json:"url,omitempty"`

	// 规则的请求转发方式，可选值：WRR、LEAST_CONN、IP_HASH
	// 分别表示按权重轮询、最小连接数、按IP哈希， 默认为 WRR。
	Scheduler *string `json:"scheduler,omitempty"`
	// 会话保持时间。
	SessionExpireTime *int64 `json:"session_expire_time,omitempty"`
	// 负载均衡实例与后端服务之间的转发协议，默认HTTP，可取值：HTTP、HTTPS、TRPC。
	ForwardType *string `json:"forward_type,omitempty"`
	// TRPC被调服务器路由，ForwardType为TRPC时必填。目前暂未对外开放。
	TrpcCallee *string `json:"trpc_callee,omitempty"`
	// TRPC调用服务接口，ForwardType为TRPC时必填。目前暂未对外开放。
	TrpcFunc *string `json:"trpc_func,omitempty"`

	// 健康检查信息。
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check"`
}

// Validate tcloud rule update option.
func (opt TCloudUpdateRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Domain Attr --------------------------

// TCloudUpdateDomainAttrOption defines options to update tcloud domain attr instances.
type TCloudUpdateDomainAttrOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ListenerId监听器ID
	ListenerId string `json:"listener_id" validate:"required"`
	// 域名（必须是已经创建的转发规则下的域名），如果是多域名，可以指定多域名列表中的任意一个。
	Domain string `json:"domain,required"`
	// 要修改的新域名。NewDomain和NewDomains只能传一个。
	NewDomain string `json:"new_domain,omitempty"`
	// 域名相关的证书信息，注意，仅对启用SNI的监听器适用。
	Certificate *corelb.TCloudCertificateInfo `json:"certificate,omitempty"`
	// 是否开启Http2，注意，只有HTTPS域名才能开启Http2。
	Http2 bool `json:"http2"`
	// 是否设为默认域名，注意，一个监听器下只能设置一个默认域名。
	DefaultServer *bool `json:"default_server,omitempty"`
	// 是否开启Quic，注意，只有HTTPS域名才能开启Quic
	Quic bool `json:"quic,omitempty"`
	// 监听器下必须配置一个默认域名，若要关闭原默认域名，必须同时指定另一个域名作为新的默认域名，如果新的默认域名是多域名，可以指定多域名列表中的任意一个。
	NewDefaultServerDomain string `json:"new_default_server_domain,omitempty"`
	// 要修改的新域名列表。NewDomain和NewDomains只能传一个。
	NewDomains []*string `json:"new_domains,omitempty"`
}

// Validate tcloud domain attr update option.
func (opt TCloudUpdateDomainAttrOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Delete Rule --------------------------

// TCloudDeleteRuleOption 批量删除规则
type TCloudDeleteRuleOption struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerId string   `json:"load_balancer_id" validate:"required"`
	ListenerId     string   `json:"listener_id" validate:"required"`
	CloudIDs       []string `json:"cloud_ids" validate:"omitempty"`
	// Domain 要删除的转发规则的域名，如果是多域名，可以指定多域名列表中的任意一个。
	Domain *string `json:"domain,omitempty"`
	// Url 要删除的转发规则的转发路径。
	Url string `json:"url,omitempty"`
	// NewDefaultServerDomain 监听器下必须配置一个默认域名，当需要删除默认域名时，可以指定另一个域名作为新的默认域名，
	// 如果新的默认域名是多域名，可以指定多域名列表中的任意一个。
	NewDefaultServerDomain *string `json:"new_default_server_domain,omitempty"`
}

// Validate ...
func (opt TCloudDeleteRuleOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudUrlRule ...
type TCloudUrlRule struct {
	*tclb.RuleOutput
}

// GetCloudID get cloud id
func (rule TCloudUrlRule) GetCloudID() string {
	return cvt.PtrToVal(rule.LocationId)
}

// -------------------------- Register Targets --------------------------

// TCloudRegisterTargetsOption defines options to tcloud register targets instances.
type TCloudRegisterTargetsOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例 ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// 绑定目标。
	Targets []*BatchTarget `json:"targets" validate:"required"`
}

// Validate validate option.
func (opt TCloudRegisterTargetsOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// BatchTarget 批量操作Target
type BatchTarget struct {
	// 监听器 ID。
	ListenerId *string `json:"listener_id"`
	// 后端服务的监听端口。
	// 注意：绑定CVM（云服务器）或ENI（弹性网卡）时必传此参数
	Port *int64 `json:"port,omitnil" `
	// 后端服务的类型，可取：CVM（云服务器）、ENI（弹性网卡）；作为入参时，目前本参数暂不生效。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Type *string `json:"Type,omitnil"`
	// 绑定CVM时需要传入此参数，代表CVM的唯一 ID，可通过 DescribeInstances 接口返回字段中的 InstanceId 字段获取。表示绑定主网卡主IP。
	// 注意：参数 InstanceId、EniIp 有且只能传入其中一个参数。
	InstanceId *string `json:"instance_id,omitnil"`
	// 绑定 IP 时需要传入此参数，支持弹性网卡的 IP 和其他内网 IP，如果是弹性网卡则必须先绑定至CVM，然后才能绑定到负载均衡实例。
	// 注意：参数 InstanceId、EniIp 只能传入一个且必须传入一个。如果绑定双栈IPV6子机，必须传该参数。
	EniIp *string `json:"eni_ip,omitnil"`
	// 子机权重，范围[0, 100]。绑定时如果不存在，则默认为10。
	Weight *int64 `json:"weight,omitnil"`
	// 七层规则 ID。
	LocationId *string `json:"location_id,omitnil"`
}

// -------------------------- Update Target Port --------------------------

// TCloudTargetPortUpdateOption defines options to update tcloud target port instances.
type TCloudTargetPortUpdateOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例 ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ListenerId 负载均衡监听器ID。
	ListenerId string `json:"listener_id" validate:"required"`
	// Targets 要修改端口的后端服务列表。
	Targets []*BatchTarget `json:"targets" validate:"required"`
	// NewPort 后端服务绑定到监听器或转发规则的新端口。
	NewPort int64 `json:"new_port" validate:"required"`
	// LocationId 转发规则的ID，当后端服务绑定到七层转发规则时，必须提供此参数或Domain+Url两者之一。
	LocationId *string `json:"location_id,omitnil"`
	// Domain 目标规则的域名，提供LocationId参数时本参数不生效。
	Domain *string `json:"domain,omitnil"`
	// Url 目标规则的URL，提供LocationId参数时本参数不生效。
	Url *string `json:"url,omitnil"`
}

// Validate validate option.
func (opt TCloudTargetPortUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Target Weight --------------------------

// TCloudTargetWeightUpdateOption defines options to update tcloud target weight instances.
type TCloudTargetWeightUpdateOption struct {
	// Region 地域
	Region string `json:"region" validate:"required"`
	// LoadBalancerId 负载均衡实例 ID
	LoadBalancerId string `json:"load_balancer_id" validate:"required"`
	// ModifyList 要批量修改权重的列表。
	ModifyList []*TargetWeightRule `json:"modify_list" validate:"required"`
}

// Validate validate option.
func (opt TCloudTargetWeightUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TargetWeightRule Target Weight rule
type TargetWeightRule struct {
	// 负载均衡监听器 ID。
	ListenerId *string `json:"listener_id" validate:"required"`
	// 要修改权重的后端机器列表。
	Targets []*BatchTarget `json:"targets" validate:"required"`
	// 转发规则的ID，七层规则时需要此参数，4层规则不需要。
	LocationId *string `json:"locationI_id,omitnil"`
	// 目标规则的域名，提供LocationId参数时本参数不生效。
	Domain *string `json:"Domain,omitnil"`
	// 目标规则的URL，提供LocationId参数时本参数不生效。
	Url *string `json:"url,omitnil"`
	// 后端服务修改后的转发权重，取值范围：[0，100]。
	// 此参数的优先级低于前述[Target](https://cloud.tencent.com/document/api/214/30694#Target)中的Weight参数，
	// 即最终的权重值以Target中的Weight参数值为准，仅当Target中的Weight参数为空时，才以RsWeightRule中的Weight参数为准。
	Weight *int64 `json:"weight,omitnil"`
}

// Backend ...
type Backend struct {
	*tclb.Backend
}

// GetIP comma separated ip addresses
func (b Backend) GetIP() string {
	if len(b.PrivateIpAddresses) == 0 {
		return ""
	}
	if len(b.PrivateIpAddresses) == 1 {
		return cvt.PtrToVal(b.PrivateIpAddresses[0])
	}
	builder := strings.Builder{}
	builder.WriteString(cvt.PtrToVal(b.PrivateIpAddresses[0]))
	for _, ip := range b.PrivateIpAddresses[1:] {
		builder.WriteByte(',')
		builder.WriteString(cvt.PtrToVal(ip))
	}
	return builder.String()
}

// GetCloudID 内网ip+端口 作为唯一键
func (b Backend) GetCloudID() string {
	return fmt.Sprintf("%s-%d", b.GetIP(), cvt.PtrToVal(b.Port))
}

// -------------------------- List Target Health --------------------------

// TCloudListTargetHealthOption defines options to list tcloud target health instances.
type TCloudListTargetHealthOption struct {
	Region          string   `json:"region" validate:"required"`
	LoadBalancerIDs []string `json:"load_balancer_ids" validate:"required"`
}

// Validate tcloud target health list option.
func (opt TCloudListTargetHealthOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudTargetHealth for target health Instance
type TCloudTargetHealth struct {
	*tclb.LoadBalancerHealth
}

// ItemPrice Price for tcloud load balancer
type ItemPrice struct {
	// 后付费单价，单位：元。
	// 注意：此字段可能返回 null，表示取不到有效值。
	UnitPrice *float64 `json:"unit_price,omitnil"`

	// 后续计价单元，可取值范围：
	// HOUR：表示计价单元是按每小时来计算。当前涉及该计价单元的场景有：实例按小时后付费（POSTPAID_BY_HOUR）、带宽按小时后付费（BANDWIDTH_POSTPAID_BY_HOUR）；
	// GB：表示计价单元是按每GB来计算。当前涉及该计价单元的场景有：流量按小时后付费（TRAFFIC_POSTPAID_BY_HOUR）。
	// 注意：此字段可能返回 null，表示取不到有效值。
	ChargeUnit *string `json:"charge_unit,omitnil"`

	// 预支费用的原价，单位：元。
	// 注意：此字段可能返回 null，表示取不到有效值。
	OriginalPrice *float64 `json:"original_price,omitnil"`

	// 预支费用的折扣价，单位：元。
	// 注意：此字段可能返回 null，表示取不到有效值。
	DiscountPrice *float64 `json:"discount_price,omitnil"`

	// 后付费的折扣单价，单位:元
	// 注意：此字段可能返回 null，表示取不到有效值。
	UnitPriceDiscount *float64 `json:"unit_price_discount,omitnil"`

	// 折扣，如20.0代表2折。
	// 注意：此字段可能返回 null，表示取不到有效值。
	Discount *float64 `json:"discount,omitnil"`
}

// TCloudLBPrice tcloud load balancer price info
type TCloudLBPrice struct {
	// 描述了实例价格。
	// 注意：此字段可能返回 null，表示取不到有效值。
	InstancePrice *ItemPrice `json:"instance_price,omitnil"`

	// 描述了网络价格。
	// 注意：此字段可能返回 null，表示取不到有效值。
	BandwidthPrice *ItemPrice `json:"bandwidth_price,omitnil"`

	// 描述了lcu价格。
	// 注意：此字段可能返回 null，表示取不到有效值。
	LcuPrice *ItemPrice `json:"lcu_price,omitnil"`
}

// -------------------------- List Load Balancer Quota --------------------------

// ListTCloudLoadBalancerQuotaOption ...
type ListTCloudLoadBalancerQuotaOption struct {
	Region string `json:"region" validate:"required"`
}

// Validate ...
func (opt *ListTCloudLoadBalancerQuotaOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudLoadBalancerQuota ...
type TCloudLoadBalancerQuota struct {
	// 配额名称，取值范围：
	// TOTAL_OPEN_CLB_QUOTA：用户当前地域下的公网CLB配额
	// TOTAL_INTERNAL_CLB_QUOTA：用户当前地域下的内网CLB配额
	// TOTAL_LISTENER_QUOTA：一个CLB下的监听器配额
	// TOTAL_LISTENER_RULE_QUOTA：一个监听器下的转发规则配额
	// TOTAL_TARGET_BIND_QUOTA：一条转发规则下可绑定设备的配额
	// TOTAL_SNAP_IP_QUOTA： 一个CLB实例下跨地域2.0的SNAT IP配额
	// TOTAL_ISP_CLB_QUOTA：用户当前地域下的三网CLB配额
	QuotaId string `json:"quota_id,omitnil"`
	// 当前使用数量，为 null 时表示无意义。
	QuotaCurrent *int64 `json:"quota_current,omitnil"`
	// 配额数量。
	QuotaLimit int64 `json:"quota_limit,omitnil"`
}

// TCloudCreateSnatIpOpt ...
type TCloudCreateSnatIpOpt struct {
	Region         string           `json:"region" validate:"required"`
	LoadBalancerId string           `json:"load_balancer_id" validate:"required"`
	SnatIps        []*corelb.SnatIp `json:"snat_ips" validate:"required,min=1,dive,required"`
}

// Validate ...
func (opt *TCloudCreateSnatIpOpt) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudDeleteSnatIpOpt ...
type TCloudDeleteSnatIpOpt struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerId string   `json:"load_balancer_id" validate:"required"`
	Ips            []string `json:"ips" validate:"required,min=1"`
}

// Validate ...
func (opt *TCloudDeleteSnatIpOpt) Validate() error {
	return validator.Validate.Struct(opt)
}
