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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	clb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// TCloudLBType 负载均衡类型
type TCloudLBType string

const (
	// Internal 内网型
	Internal TCloudLBType = "INTERNAL"
	// Open 公网型
	Open TCloudLBType = "OPEN"
)

// TCloudIPVersion 负载均衡ip版本
type TCloudIPVersion string

const (
	IPv4      TCloudIPVersion = "IPV4"
	IPv6      TCloudIPVersion = "IPv6FullChain"
	IPv6NAT64 TCloudIPVersion = "IPv6"
)

type TCloudInternetAccessible struct {
}

// TCloudCLBCreateOpt ...
type TCloudCLBCreateOpt struct {
	Region string `json:"region" validate:"required"`
	// 公网/内网
	LBType     TCloudLBType `json:"lb_type" validate:"required"`
	Name       string       `json:"name" validate:"max=60"`
	CloudVpcID string       `json:"vpc_id"`
	SubnetID   string       `json:"subnet_id"`

	// vip
	Vip *string
	// 默认BGP，指定其他运营商需要配置为带宽包计费
	VipIsp *string
	// 公网负载均衡版本
	AddressIPVersion *TCloudIPVersion `json:"address_ip_version"`
	// 网络计费模式
	InternetChargeType *string `json:"internet_charge_type"`
	// 最大出带宽
	InternetMaxBandwidthOut *int64 `json:"internet_max_bandwidth_out"`
	// 带宽包ID，计费模式为带宽包计费时必填
	BandwidthPackageId *string

	// 性能规格，性能型填写，共享型留空
	SlaType *string
	// true 只验证CLB上的安全组；false 需同时验证CLB和后端实例上的安全组。
	LoadBalancerPassToTarget *bool
}

// Validate ...
func (o TCloudCLBCreateOpt) Validate() error {
	return validator.Validate.Struct(o)
}

// TCloudCLBListOpt ...
type TCloudCLBListOpt struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate ...
func (o TCloudCLBListOpt) Validate() error {
	return validator.Validate.Struct(o)
}

// TCloudCLBDeleteOpt ...
type TCloudCLBDeleteOpt struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate ...
func (o TCloudCLBDeleteOpt) Validate() error {
	return validator.Validate.Struct(o)
}

// TCloudCLB ...
type TCloudCLB struct {
	*clb.LoadBalancer
}

// TCloudListenerListOpt ...
type TCloudListenerListOpt struct {
	Region      string   `json:"region" validate:"required"`
	ClbID       string   `json:"clb_id" validate:"required"`
	ListenerIds []string `json:"listener_ids" validate:"max=100"`
}

// Validate ...
func (o TCloudListenerListOpt) Validate() error {
	return validator.Validate.Struct(o)
}

// TCloudListener ...
type TCloudListener struct {
	*clb.Listener
}

// TCloudTargetListOpt ...
type TCloudTargetListOpt struct {
	Region      string   `json:"region" validate:"required"`
	ClbID       string   `json:"clb_id" validate:"required"`
	ListenerIds []string `json:"listener_ids" validate:"max=100"`
}

// Validate ...
func (o TCloudTargetListOpt) Validate() error {
	return validator.Validate.Struct(o)
}

// TCloudListenerBackend ...
type TCloudListenerBackend struct {
	*clb.ListenerBackend
}
