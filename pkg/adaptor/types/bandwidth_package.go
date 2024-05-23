/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package types

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// BwPkgChargeType 带宽包的计费类型
type BwPkgChargeType string

const (
	// TOP5_POSTPAID_BY_MONTH 按月后付费TOP5计费
	TOP5_POSTPAID_BY_MONTH BwPkgChargeType = "TOP5_POSTPAID_BY_MONTH"
	// PERCENT95_POSTPAID_BY_MONTH 按月后付费月95计费
	PERCENT95_POSTPAID_BY_MONTH BwPkgChargeType = "PERCENT95_POSTPAID_BY_MONTH"
	// ENHANCED95_POSTPAID_BY_MONTH 按月后付费增强型95计费
	ENHANCED95_POSTPAID_BY_MONTH BwPkgChargeType = "ENHANCED95_POSTPAID_BY_MONTH"
	// FIXED_PREPAID_BY_MONTH 包月预付费计费
	FIXED_PREPAID_BY_MONTH BwPkgChargeType = "FIXED_PREPAID_BY_MONTH"
	// PEAK_BANDWIDTH_POSTPAID_BY_DAY 后付费日结按带宽计费
	PEAK_BANDWIDTH_POSTPAID_BY_DAY BwPkgChargeType = "PEAK_BANDWIDTH_POSTPAID_BY_DAY"
)

// BwPkgNetworkType 带宽包网络类型
type BwPkgNetworkType string

const (
	// BGP bgp
	BGP BwPkgNetworkType = "BGP"
	// SINGLEISP single isp
	SINGLEISP BwPkgNetworkType = "SINGLEISP"
	// HIGH_QUALITY_BGP high quality bgp
	HIGH_QUALITY_BGP BwPkgNetworkType = "HIGH_QUALITY_BGP"
	// ANYCAST any cast
	ANYCAST BwPkgNetworkType = "ANYCAST"
	// SINGLEISP_CMCC cmcc
	SINGLEISP_CMCC BwPkgNetworkType = "SINGLEISP_CMCC"
	// SINGLEISP_CTCC ctcc
	SINGLEISP_CTCC BwPkgNetworkType = "SINGLEISP_CTCC"
	// SINGLEISP_CUCC cucc
	SINGLEISP_CUCC BwPkgNetworkType = "SINGLEISP_CUCC"
)

// TCloudBandwidthPackage 腾讯云共享带宽包
type TCloudBandwidthPackage struct {
	// bwp-xxxxxx
	ID   string `json:"id"`
	Name string `json:"name"`
	// 带宽包类型，包括'BGP','SINGLEISP','ANYCAST','SINGLEISP_CMCC','SINGLEISP_CTCC','SINGLEISP_CUCC'
	NetworkType BwPkgNetworkType `json:"network_type"`
	// 带宽包的计费类型
	ChargeType BwPkgChargeType `json:"charge_type"`
	// 带宽包状态，包括'CREATING','CREATED','DELETING','DELETED'
	Status string `json:"status"`
	// 带宽包限速大小。单位：Mbps，-1表示不限速。
	Bandwidth int64 `json:"bandwidth"`
	// 网络出口
	Egress string `json:"egress"`
	// 带宽包创建时间
	CreateTime string `json:"create_time"`
	// 带宽包到期时间，只有预付费会返回，按量计费返回为null
	Deadline string `json:"deadline"`
	// 带宽包资源信息
	ResourceSet []Resource `json:"resource_set"`
}

// Resource BandwidthPackage resource type
type Resource struct {
	// 带宽包资源类型，包括'Address'和'LoadBalance'
	ResourceType string `json:"resource_type"`
	// 带宽包资源Id，形如'eip-xxxx', 'lb-xxxx'
	ResourceID string `json:"resource_id"`
	// 带宽包资源Ip
	AddressIP string `json:"address_ip"`
}

// TCloudListBwPkgOption 查询带宽包
type TCloudListBwPkgOption struct {
	Region string           `json:"region" validate:"required" `
	Page   *core.TCloudPage `json:"page" validate:"required"`

	// 参数不支持同时指定PkgCloudIds和其他条件
	PkgCloudIds   []string           `json:"pkg_cloud_ids"`
	PkgNames      []string           `json:"pkg_names,omitempty"`
	NetworkTypes  []BwPkgNetworkType `json:"network_types,omitempty"`
	ChargeTypes   []BwPkgChargeType  `json:"charge_types,omitempty"`
	ResourceTypes []string           `json:"resource_types,omitempty"`
	ResourceIds   []string           `json:"resource_ids,omitempty"`
	ResAddressIps []string           `json:"res_address_ips,omitempty"`
}

// Validate ...
func (opt *TCloudListBwPkgOption) Validate() error {

	return validator.Validate.Struct(opt)
}

// TCloudListBwPkgResult 查询带宽包结果
type TCloudListBwPkgResult struct {
	TotalCount uint64                   `json:"total_count"`
	Packages   []TCloudBandwidthPackage `json:"packages"`
}
