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

// Package networkinterface 包提供各类云资源的请求与返回序列化器
package networkinterface

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	coreni "hcm/pkg/api/core/cloud/network-interface"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// NetworkInterfaceBatchCreateReq defines batch create network interface request.
type NetworkInterfaceBatchCreateReq[T NetworkInterfaceCreateExtension] struct {
	NetworkInterfaces []NetworkInterfaceReq[T] `json:"network_interfaces" validate:"required,max=100"`
}

// NetworkInterfaceReq defines create network interface request.
type NetworkInterfaceReq[T NetworkInterfaceCreateExtension] struct {
	Vendor        string   `json:"vendor" validate:"required"`
	Name          string   `json:"name" validate:"required"`
	AccountID     string   `json:"account_id" validate:"required"`
	Region        string   `json:"region" validate:"omitempty"`
	Zone          string   `json:"zone" validate:"omitempty"`
	CloudID       string   `json:"cloud_id" validate:"omitempty"`
	VpcID         string   `json:"vpc_id" validate:"omitempty"`
	CloudVpcID    string   `json:"cloud_vpc_id" validate:"omitempty"`
	SubnetID      string   `json:"subnet_id" validate:"omitempty"`
	CloudSubnetID string   `json:"cloud_subnet_id" validate:"omitempty"`
	PrivateIPv4   []string `json:"private_ipv4,omitempty" validate:"omitempty"`
	PrivateIPv6   []string `json:"private_ipv6,omitempty" validate:"omitempty"`
	PublicIPv4    []string `json:"public_ipv4,omitempty" validate:"omitempty"`
	PublicIPv6    []string `json:"public_ipv6,omitempty" validate:"omitempty"`
	BkBizID       int64    `json:"bk_biz_id" validate:"omitempty"`
	InstanceID    string   `json:"instance_id,omitempty" validate:"omitempty"`
	Extension     *T       `json:"extension" validate:"required"`
}

// NetworkInterfaceCreateExtension defines create network interface extensional info.
type NetworkInterfaceCreateExtension interface {
	AzureNICreateExt | GcpNICreateExt | HuaWeiNICreateExt
}

// AzureNICreateExt defines azure network interface extensional info.
type AzureNICreateExt struct {
	Type              string `json:"type" validate:"omitempty"`
	ResourceGroupName string `json:"resource_group_name" validate:"omitempty"`
	MacAddress        string `json:"mac_address"`
	// EnableAcceleratedNetworking 是否加速网络
	EnableAcceleratedNetworking *bool `json:"enable_accelerated_networking"`
	// EnableIPForwarding 是否允许IP转发
	EnableIPForwarding *bool `json:"enable_ip_forwarding"`
	// DNSSettings DNS设置
	DNSSettings *coreni.InterfaceDNSSettings `json:"dns_settings"`
	// CloudGatewayLoadBalancerID 网关负载均衡器ID
	CloudGatewayLoadBalancerID *string `json:"cloud_gateway_load_balancer_id"`
	// CloudSecurityGroupID 网络安全组ID
	CloudSecurityGroupID *string `json:"cloud_security_group_id"`
	SecurityGroupID      *string `json:"security_group_id,omitempty"`
	// IPConfigurations IP配置列表
	IPConfigurations []*coreni.InterfaceIPConfiguration `json:"ip_configurations"`
}

// GcpNICreateExt defines gcp network interface extensional info.
type GcpNICreateExt struct {
	VpcSelfLink    string          `json:"vpc_self_link,omitempty"`
	SubnetSelfLink string          `json:"subnet_self_link,omitempty"`
	CanIpForward   bool            `json:"can_ip_forward,omitempty"`
	Status         string          `json:"status,omitempty"`
	StackType      string          `json:"stack_type,omitempty"`
	AccessConfigs  []*AccessConfig `json:"access_configs,omitempty"`
}

// AccessConfig An access configuration attached to an instance's
// network interface. Only one access config per instance is supported.
type AccessConfig struct {
	// Name: The name of this access configuration. The default and
	// recommended name is External NAT, but you can use any arbitrary
	// string, such as My external IP or Network Access.
	Name string `json:"name,omitempty"`

	// NatIP: An external IP address associated with this instance. Specify
	// an unused static external IP address available to the project or
	// leave this field undefined to use an IP from a shared ephemeral IP
	// address pool. If you specify a static external IP address, it must
	// live in the same region as the zone of the instance.
	NatIP string `json:"nat_ip,omitempty"`

	// NetworkTier: This signifies the networking tier used for configuring
	// this access configuration and can only take the following values:
	// PREMIUM, STANDARD. If an AccessConfig is specified without a valid
	// external IP address, an ephemeral IP will be created with this
	// networkTier. If an AccessConfig with a valid external IP address is
	// specified, it must match that of the networkTier associated with the
	// Address resource owning that IP.
	//
	// Possible values:
	//   "FIXED_STANDARD" - Public internet quality with fixed bandwidth.
	//   "PREMIUM" - High quality, Google-grade network tier, support for
	// all networking products.
	//   "STANDARD" - Public internet quality, only limited support for
	// other networking products.
	//   "STANDARD_OVERRIDES_FIXED_STANDARD" - (Output only) Temporary tier
	// for FIXED_STANDARD when fixed standard tier is expired or not
	// configured.
	NetworkTier string `json:"network_tier,omitempty"`

	// Type: The type of configuration. The default and only option is
	// ONE_TO_ONE_NAT.
	//
	// Possible values:
	//   "DIRECT_IPV6"
	//   "ONE_TO_ONE_NAT" (default)
	Type string `json:"type,omitempty"`
}

// HuaWeiNICreateExt defines huawei network interface extensional info.
type HuaWeiNICreateExt struct {
	// FixedIps 网卡私网IP信息列表。
	FixedIps []ServerInterfaceFixedIp `json:"fixed_ips,omitempty"`
	// MacAddr 网卡Mac地址信息。
	MacAddr *string `json:"mac_addr,omitempty"`
	// NetId 网卡端口所属网络ID。
	NetId *string `json:"net_id,omitempty"`
	// PortState 网卡端口状态。
	PortState *string `json:"port_state,omitempty"`
	// DeleteOnTermination 卸载网卡时，是否删除网卡。
	DeleteOnTermination *bool `json:"delete_on_termination,omitempty"`
	// DriverMode 从guest os中，网卡的驱动类型。可选值为virtio和hinic，默认为virtio
	DriverMode *string `json:"driver_mode,omitempty"`
	// MinRate 网卡带宽下限。
	MinRate *int32 `json:"min_rate,omitempty"`
	// MultiqueueNum 网卡多队列个数。
	MultiqueueNum *int32 `json:"multiqueue_num,omitempty"`
	// PciAddress 弹性网卡在Linux GuestOS里的BDF号
	PciAddress *string `json:"pci_address,omitempty"`
	// IpV6 IpV6地址
	IpV6 *string `json:"ipv6,omitempty"`
	// VirtualIPList 虚拟IP地址数组
	VirtualIPList []NetVirtualIP `json:"virtual_ip_list,omitempty"`
	// Addresses 云服务器对应的网络地址信息
	Addresses *EipNetwork `json:"addresses,omitempty"`
	// CloudSecurityGroupIDs 云服务器所属安全组ID
	CloudSecurityGroupIDs []string `json:"cloud_security_group_ids"`
}

// EipNetwork 华为云主机绑定的弹性IP
type EipNetwork struct {
	// IPVersion IP地址类型，值为4或6(4：IP地址类型是IPv4 6：IP地址类型是IPv6)
	IPVersion int32 `json:"ip_version"`
	// PublicIPAddress IP地址
	PublicIPAddress string `json:"public_ip_address"`
	// PublicIPV6Address IPV6地址
	PublicIPV6Address string `json:"public_ipv6_address"`
	// BandwidthID 带宽ID
	BandwidthID string `json:"bandwidth_id"`
	// BandwidthSize 带宽大小
	BandwidthSize int32 `json:"bandwidth_size"`
	// BandwidthType 带宽类型，示例:5_bgp(全动态BGP)
	BandwidthType string `json:"bandwidth_type"`
}

// NetVirtualIP 网络接口的虚拟IP
type NetVirtualIP struct {
	// IP 虚拟IP
	IP string `json:"ip,omitempty"`
	// ElasticityIP 弹性公网IP
	ElasticityIP string `json:"elasticity_ip,omitempty"`
}

// ServerInterfaceFixedIp ...
type ServerInterfaceFixedIp struct {
	// IpAddress 网卡私网IP信息。
	IpAddress *string `json:"ip_address,omitempty"`
	// SubnetId 网卡私网IP对应子网信息。
	SubnetId *string `json:"subnet_id,omitempty"`
}

// Validate NetworkInterfaceBatchCreateReq.
func (c *NetworkInterfaceBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureNIBatchCreate define azure network interface when create.
type AzureNIBatchCreate struct {
	Vendor        string   `json:"vendor"`
	Name          string   `json:"name"`
	AccountID     string   `json:"account_id"`
	Region        string   `json:"region"`
	Zone          string   `json:"zone"`
	CloudID       string   `json:"cloud_id"`
	VpcID         string   `json:"vpc_id"`
	CloudVpcID    string   `json:"cloud_vpc_id"`
	SubnetID      string   `json:"subnet_id"`
	CloudSubnetID string   `json:"cloud_subnet_id"`
	PrivateIPv4   []string `json:"private_ipv4"`
	PrivateIPv6   []string `json:"private_ipv6"`
	PublicIPv4    []string `json:"public_ipv4"`
	PublicIPv6    []string `json:"public_ipv6"`
	BkBizID       int64    `json:"bk_biz_id"`
	InstanceID    string   `json:"instance_id"`
}

// -------------------------- Update --------------------------

// NetworkInterfaceBatchUpdateReq define batch update network interface request.
type NetworkInterfaceBatchUpdateReq[T NetworkInterfaceCreateExtension] struct {
	NetworkInterfaces []NetworkInterfaceUpdateReq[T] `json:"network_interfaces" validate:"required,max=100"`
}

// NetworkInterfaceUpdateReq resource group batch update option.
type NetworkInterfaceUpdateReq[T NetworkInterfaceCreateExtension] struct {
	ID string `json:"id" validate:"required"`

	Vendor        string   `json:"vendor" validate:"required"`
	Name          string   `json:"name" validate:"required"`
	AccountID     string   `json:"account_id" validate:"required"`
	Region        string   `json:"region" validate:"omitempty"`
	Zone          string   `json:"zone" validate:"omitempty"`
	CloudID       string   `json:"cloud_id" validate:"omitempty"`
	VpcID         string   `json:"vpc_id" validate:"omitempty"`
	CloudVpcID    string   `json:"cloud_vpc_id" validate:"omitempty"`
	SubnetID      string   `json:"subnet_id" validate:"omitempty"`
	CloudSubnetID string   `json:"cloud_subnet_id" validate:"omitempty"`
	PrivateIPv4   []string `json:"private_ipv4,omitempty" validate:"omitempty"`
	PrivateIPv6   []string `json:"private_ipv6,omitempty" validate:"omitempty"`
	PublicIPv4    []string `json:"public_ipv4,omitempty" validate:"omitempty"`
	PublicIPv6    []string `json:"public_ipv6,omitempty" validate:"omitempty"`
	BkBizID       int64    `json:"bk_biz_id" validate:"omitempty"`
	InstanceID    string   `json:"instance_id" validate:"omitempty"`
	Extension     *T       `json:"extension" validate:"required"`
}

// Validate network interface batch update request.
func (req *NetworkInterfaceBatchUpdateReq[T]) Validate() error {
	if len(req.NetworkInterfaces) == 0 {
		return errors.New("network interface is required")
	}

	if len(req.NetworkInterfaces) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("network interface count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Update Common --------------------------

// NetworkInterfaceCommonInfoBatchUpdateReq define network interface common info batch update req.
type NetworkInterfaceCommonInfoBatchUpdateReq struct {
	IDs     []string `json:"ids" validate:"required"`
	BkBizID int64    `json:"bk_biz_id" validate:"required"`
}

// Validate network interface common info batch update req.
func (req *NetworkInterfaceCommonInfoBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.IDs) == 0 {
		return errors.New("ids required")
	}

	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Delete --------------------------

// AzureNIBatchDeleteReq azure network interface delete request.
type AzureNIBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate azure network interface delete request.
func (req *AzureNIBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Get --------------------------

// NetworkInterfaceGetResp defines get network interface response.
type NetworkInterfaceGetResp[T coreni.NetworkInterfaceExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *coreni.NetworkInterface[T] `json:"data"`
}

// -------------------------- List --------------------------

// NetworkInterfaceListResp defines list network interface response.
type NetworkInterfaceListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *NetworkInterfaceListResult `json:"data"`
}

// NetworkInterfaceListResult defines list network interface result.
type NetworkInterfaceListResult struct {
	Count   uint64                        `json:"count"`
	Details []coreni.BaseNetworkInterface `json:"details"`
}

// NetworkInterfaceAssociateListResp defines list network interface associate response.
type NetworkInterfaceAssociateListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *NetworkInterfaceAssociateListResult `json:"data"`
}

// NetworkInterfaceAssociateListResult defines list network interface associate result.
type NetworkInterfaceAssociateListResult struct {
	Count   uint64                             `json:"count"`
	Details []coreni.NetworkInterfaceAssociate `json:"details"`
}

// NetworkInterfaceExtListResult define network interface with extension list result.
type NetworkInterfaceExtListResult[T coreni.NetworkInterfaceExtension] struct {
	Count   uint64                       `json:"count,omitempty"`
	Details []coreni.NetworkInterface[T] `json:"details,omitempty"`
}

// NetworkInterfaceExtListResp define network interface with extension list response.
type NetworkInterfaceExtListResp[T coreni.NetworkInterfaceExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *NetworkInterfaceExtListResult[T] `json:"data"`
}

// NetworkInterfaceListReq is network interface list operation http request.
type NetworkInterfaceListReq struct {
	core.ListReq `json:",inline"`
	IsAssociate  bool `json:"is_associate"` // true:获取已关联的列表 false:获取未关联的列表，默认不传
}

// Validate ListReq.
func (l *NetworkInterfaceListReq) Validate() error {
	if l.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}

	if l.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	return nil
}
