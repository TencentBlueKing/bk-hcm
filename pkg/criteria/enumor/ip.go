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

// IPAddressType is ip address type.
type IPAddressType string

// Validate IPAddressType.
func (i IPAddressType) Validate() error {
	switch i {
	case Ipv4:
	case Ipv6:
	case Ipv6Nat64:
	case Ipv6DualStack:
	default:
		return fmt.Errorf("unsupported ip address type: %s", i)
	}

	return nil
}

const (
	// Ipv4 is ipv4 address type.
	Ipv4 IPAddressType = "ipv4"
	// Ipv6 is ipv6 address type.
	Ipv6 IPAddressType = "ipv6"
	// Ipv6DualStack  双栈，同时支持v4和v6
	Ipv6DualStack IPAddressType = "ipv6_dual_stack"
	// Ipv6Nat64  6转4模式 腾讯云CLB支持该模式
	Ipv6Nat64 IPAddressType = "ipv6_nat64"
)
