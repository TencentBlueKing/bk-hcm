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

// GetHuaWeiAccountZoneQuotaOption ...
type GetHuaWeiAccountZoneQuotaOption struct {
	Region string `json:"region" validate:"required"`
}

// Validate ...
func (opt *GetHuaWeiAccountZoneQuotaOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// HuaWeiAccountQuota define huawei account quota.
type HuaWeiAccountQuota struct {
	// 镜像元数据最大的长度。
	MaxImageMeta int32 `json:"max_image_meta"`

	// 可注入文件的最大个数。
	MaxPersonality int32 `json:"max_personality"`

	// 注入文件内容的最大长度（单位：Byte）。
	MaxPersonalitySize int32 `json:"max_personality_size"`

	// 安全组中安全组规则最大的配置个数。   > 说明：  - 具体配额限制请以VPC配额限制为准。
	MaxSecurityGroupRules int32 `json:"max_security_group_rules"`

	// 安全组最大使用个数。  > 说明：  - 具体配额限制请以VPC配额限制为准。
	MaxSecurityGroups int32 `json:"max_security_groups"`

	// 服务器组中的最大虚拟机数。
	MaxServerGroupMembers int32 `json:"max_server_group_members"`

	// 服务器组的最大个数。
	MaxServerGroups int32 `json:"max_server_groups"`

	// 可输入元数据的最大长度。
	MaxServerMeta int32 `json:"max_server_meta"`

	// CPU核数最大申请数量。
	MaxTotalCores int32 `json:"max_total_cores"`

	// 最大的浮动IP使用个数。
	MaxTotalFloatingIps int32 `json:"max_total_floating_ips"`

	// 云服务器最大申请数量。
	MaxTotalInstances int32 `json:"max_total_instances"`

	// 可以申请的SSH密钥对最大数量。
	MaxTotalKeypairs int32 `json:"max_total_keypairs"`

	// 内存最大申请容量（单位：MB）。
	MaxTotalRAMSize int32 `json:"max_total_ram_size"`

	// 竞价实例的最大申请数量。
	MaxTotalSpotInstances *int32 `json:"max_total_spot_instances"`

	// 竞价实例的CPU核数最大申请数量。
	MaxTotalSpotCores *int32 `json:"max_total_spot_cores"`

	// 竞价实例的内存最大申请容量（单位：MB）。
	MaxTotalSpotRAMSize *int32 `json:"max_total_spot_ram_size"`
}
