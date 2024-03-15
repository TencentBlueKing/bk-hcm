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

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
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
	return converter.PtrToVal(clb.LoadBalancerId)
}

// -------------------------- List Listeners--------------------------

// TCloudListListenersOption defines options to list tcloud listeners instances.
type TCloudListListenersOption struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerId string   `json:"load_balancer_id" validate:"required"`
	CloudIDs       []string `json:"cloud_ids" validate:"omitempty"`
	Protocol       string   `json:"protocol" validate:"omitempty"`
	Port           int64    `json:"port" validate:"omitempty"`
}

// Validate tcloud listeners list option.
func (opt TCloudListListenersOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}

// TCloudListeners for listeners Instance
type TCloudListeners struct {
	*tclb.Listener
}

// GetCloudID get cloud id
func (clb TCloudListeners) GetCloudID() string {
	return converter.PtrToVal(clb.ListenerId)
}

// -------------------------- List Targets--------------------------

// TCloudListTargetsOption defines options to list tcloud targets instances.
type TCloudListTargetsOption struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerId string   `json:"load_balancer_id" validate:"required"`
	CloudIDs       []string `json:"cloud_ids" validate:"omitempty"`
	Protocol       string   `json:"protocol" validate:"omitempty"`
	Port           int64    `json:"port" validate:"omitempty"`
}

// Validate tcloud targets list option.
func (opt TCloudListTargetsOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}

// TCloudListenerTargets for listener targets Instance
type TCloudListenerTargets struct {
	*tclb.ListenerBackend
}

// -------------------------- Create Clb--------------------------

// TCloudCreateClbOption defines options to create tloud clb instances.
type TCloudCreateClbOption struct {
	Region                   string                     `json:"region" validate:"required"`
	LoadBalancerType         TCloudLoadBalancerType     `json:"load_balancer_type" validate:"required"`
	Forward                  TCloudLoadBalancerInstType `json:"forward" validate:"omitempty"`
	LoadBalancerName         *string                    `json:"load_balancer_name" validate:"omitempty,max=60"`
	VpcID                    *string                    `json:"vpc_id" validate:"omitempty"`
	SubnetID                 *string                    `json:"subnet_id" validate:"omitempty"`
	ProjectID                *int64                     `json:"project_id" validate:"omitempty"`
	AddressIPVersion         *TCloudAddressIPVersion    `json:"address_ip_version" validate:"omitempty"`
	Number                   *uint64                    `json:"number" validate:"omitempty"`
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

	InternetChargeType      *string `json:"internet_charge_type"`
	InternetMaxBandwidthOut *int64  `json:"internet_max_bandwidth_out" `
	BandwidthpkgSubType     *string `json:"bandwidthpkg_sub_type" validate:"omitempty"`
}

// Validate tcloud clb create option.
func (opt TCloudCreateClbOption) Validate() error {
	return validator.Validate.Struct(opt)
}

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

// TCloudAddressIPVersion 仅适用于公网负载均衡。IP版本，可取值：IPV4、IPV6、IPv6FullChain，不区分大小写，默认值 IPV4。
// 说明：取值为IPV6表示为IPV6 NAT64版本；取值为IPv6FullChain，表示为IPv6版本。
type TCloudAddressIPVersion string

const (
	// IPV4AddressVersion IPV4版本
	IPV4AddressVersion TCloudAddressIPVersion = "IPV4"
	// IPV6AddressVersion IPV6版本
	IPV6AddressVersion TCloudAddressIPVersion = "IPV6"
	// IPV6FullChainAddressVersion IPv6FullChain版本
	IPV6FullChainAddressVersion TCloudAddressIPVersion = "IPv6FullChain"
)

// TCloudLoadBalancerStatus 负载均衡实例的状态
type TCloudLoadBalancerStatus uint64

const (
	// IngStatus 负载均衡实例的状态-创建中
	IngStatus TCloudLoadBalancerStatus = 0
	// SuccessStatus 负载均衡实例的状态-正常运行
	SuccessStatus TCloudLoadBalancerStatus = 1
)

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
