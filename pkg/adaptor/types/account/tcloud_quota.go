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

package account

import "hcm/pkg/criteria/validator"

// GetTCloudAccountZoneQuotaOption ...
type GetTCloudAccountZoneQuotaOption struct {
	Region string `json:"region" validate:"required"`
	Zone   string `json:"zone" validate:"required"`
}

// Validate ...
func (opt *GetTCloudAccountZoneQuotaOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudAccountQuota ...
type TCloudAccountQuota struct {
	PostPaidQuotaSet          *TCloudPostPaidQuota             `json:"post_paid_quota_set,omitempty"`
	PrePaidQuota              *TCloudPrePaidQuota              `json:"pre_paid_quota,omitempty"`
	SpotPaidQuota             *TCloudSpotPaidQuota             `json:"spot_paid_quota,omitempty"`
	ImageQuota                *TCloudImageQuota                `json:"image_quota,omitempty"`
	DisasterRecoverGroupQuota *TCloudDisasterRecoverGroupQuota `json:"disaster_recover_group_quota,omitempty"`
}

// TCloudPostPaidQuota 后付费配额列表
type TCloudPostPaidQuota struct {
	// UsedQuota 累计已使用配额
	UsedQuota *uint64 `json:"used_quota"`
	// RemainingQuota 剩余配额
	RemainingQuota *uint64 `json:"remaining_quota"`
	// TotalQuota 总配额
	TotalQuota *uint64 `json:"total_quota"`
}

// TCloudPrePaidQuota 预付费配额列表
type TCloudPrePaidQuota struct {
	// UsedQuota 累计已使用配额
	UsedQuota *uint64 `json:"used_quota"`
	// OnceQuota 单次购买最大数量
	OnceQuota *uint64 `json:"once_quota"`
	// RemainingQuota 剩余配额
	RemainingQuota *uint64 `json:"remaining_quota"`
	// TotalQuota 总配额
	TotalQuota *uint64 `json:"total_quota"`
}

// TCloudSpotPaidQuota spot配额列表
type TCloudSpotPaidQuota struct {
	// UsedQuota 累计已使用配额
	UsedQuota *uint64 `json:"used_quota"`
	// RemainingQuota 剩余配额
	RemainingQuota *uint64 `json:"remaining_quota"`
	// TotalQuota 总配额
	TotalQuota *uint64 `json:"total_quota"`
}

// TCloudImageQuota 镜像配额列表
type TCloudImageQuota struct {
	// UsedQuota 累计已使用配额
	UsedQuota *uint64 `json:"used_quota"`
	// TotalQuota 总配额
	TotalQuota *uint64 `json:"total_quota"`
}

// TCloudDisasterRecoverGroupQuota 置放群组配额列表
type TCloudDisasterRecoverGroupQuota struct {
	// GroupQuota 可创建置放群组数量的上限。
	GroupQuota *int64 `json:"group_quota"`
	// CurrentNum 当前用户已经创建的置放群组数量。
	CurrentNum *int64 `json:"current_num"`
	// CvmInHostGroupQuota 物理机类型容灾组内实例的配额数。
	CvmInHostGroupQuota *int64 `json:"cvm_in_host_group_quota"`
	// CvmInSwitchGroupQuota 交换机类型容灾组内实例的配额数。
	CvmInSwitchGroupQuota *int64 `json:"cvm_in_switch_group_quota"`
	// CvmInRackGroupQuota 机架类型容灾组内实例的配额数。
	CvmInRackGroupQuota *int64 `json:"cvm_in_rack_group_quota"`
}
