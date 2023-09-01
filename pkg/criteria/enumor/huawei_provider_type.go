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

package enumor

import (
	"fmt"
)

// HuaWeiProviderType is huawei provider type.
type HuaWeiProviderType string

// Validate the HuaWeiProviderType is valid or not
func (h HuaWeiProviderType) Validate() error {
	switch h {
	case HuaWeiCvmProviderType:
	case HuaWeiDiskProviderType:
	case HuaWeiSGProviderType:
	case HuaWeiVpcProviderType:
	case HuaWeiEipProviderType:
	default:
		return fmt.Errorf("unsupported huawei provider type: %s", h)

	}

	return nil
}

const (
	// HuaWeiCvmProviderType cvm
	HuaWeiCvmProviderType HuaWeiProviderType = "ecs.cloudservers"
	// HuaWeiDiskProviderType disk
	HuaWeiDiskProviderType HuaWeiProviderType = "evs.volumes"
	// HuaWeiSGProviderType sg
	HuaWeiSGProviderType HuaWeiProviderType = "vpc.securityGroups"
	// HuaWeiVpcProviderType vpc
	HuaWeiVpcProviderType HuaWeiProviderType = "vpc.vpcs"
	// HuaWeiEipProviderType eip
	HuaWeiEipProviderType HuaWeiProviderType = "vpc.publicips"
)
