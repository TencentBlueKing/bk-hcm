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

package clb

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
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty,max=200"`
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
	LoadBalancerName         string                     `json:"load_balancer_name" validate:"omitempty,max=60"`
	VpcID                    string                     `json:"vpc_id" validate:"omitempty"`
	SubnetID                 string                     `json:"subnet_id" validate:"omitempty"`
	ProjectID                int64                      `json:"project_id" validate:"omitempty"`
	AddressIPVersion         TCloudAddressIPVersion     `json:"address_ip_version" validate:"omitempty"`
	Number                   uint64                     `json:"number" validate:"omitempty"`
	MasterZoneID             string                     `json:"master_zone_id" validate:"omitempty"`
	ZoneID                   string                     `json:"zone_id" validate:"omitempty"`
	InternetAccessible       *tclb.InternetAccessible   `json:"internet_accessible" validate:"omitempty"`
	VipIsp                   string                     `json:"vip_isp" validate:"omitempty"`
	Tags                     []*tclb.TagInfo            `json:"tags" validate:"omitempty"`
	Vip                      string                     `json:"vip" validate:"omitempty"`
	BandwidthPackageID       string                     `json:"bandwidth_package_id" validate:"omitempty"`
	ExclusiveCluster         *tclb.ExclusiveCluster     `json:"exclusive_cluster" validate:"omitempty"`
	SlaType                  string                     `json:"sla_type" validate:"omitempty"`
	ClusterIds               []*string                  `json:"cluster_ids" validate:"omitempty"`
	ClientToken              string                     `json:"client_token" validate:"omitempty"`
	SnatPro                  bool                       `json:"snat_pro" validate:"omitempty"`
	SnatIps                  []*tclb.SnatIp             `json:"snat_ips" validate:"omitempty"`
	ClusterTag               string                     `json:"cluster_tag" validate:"omitempty"`
	SlaveZoneID              string                     `json:"slave_zone_id" validate:"omitempty"`
	EipAddressID             string                     `json:"eip_address_id" validate:"omitempty"`
	LoadBalancerPassToTarget bool                       `json:"load_balancer_pass_to_target" validate:"omitempty"`
	DynamicVip               bool                       `json:"dynamic_vip" validate:"omitempty"`
	Egress                   string                     `json:"egress" validate:"omitempty"`
}

// Validate tcloud clb create option.
func (opt TCloudCreateClbOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.LoadBalancerType != OpenLoadBalancerType && opt.LoadBalancerType != InternalLoadBalancerType {
		return errors.New("load_balancer_type is illegal")
	}

	if opt.Forward != 0 && opt.Forward != DefaultLoadBalancerInstType {
		return errors.New("forward is illegal")
	}

	if len(opt.AddressIPVersion) != 0 && opt.AddressIPVersion != IPV4AddressVersion &&
		opt.AddressIPVersion != IPV6AddressVersion && opt.AddressIPVersion != IPV6FullChainAddressVersion {
		return errors.New("address_ip_version is illegal")
	}

	return nil
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

// --------------------------[设置负载均衡实例的安全组]--------------------------

// TCloudSetClbSecurityGroupOption defines options to set tcloud clb security-group option.
type TCloudSetClbSecurityGroupOption struct {
	Region         string   `json:"region" validate:"required"`
	LoadBalancerID string   `json:"load_balancer_id" validate:"required"`
	SecurityGroups []string `json:"security_groups" validate:"omitempty,max=50"`
}

// Validate tcloud clb security-group option.
func (opt TCloudSetClbSecurityGroupOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.LoadBalancerID) == 0 {
		return errors.New("load_balancer_id is required")
	}

	if len(opt.SecurityGroups) > constant.LoadBalancerBindSecurityGroupMaxLimit {
		return fmt.Errorf("invalid page.limit max value: %d", constant.LoadBalancerBindSecurityGroupMaxLimit)
	}

	return nil
}
