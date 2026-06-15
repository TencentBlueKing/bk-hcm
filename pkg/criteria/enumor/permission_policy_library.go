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

import "fmt"

// PermPolicyLibAction is the action type for permission policy library application.
type PermPolicyLibAction string

const (
	// PermPolicyLibActionApplyCreate applies a policy library by creating new CAM policies.
	PermPolicyLibActionApplyCreate PermPolicyLibAction = "apply_create"
	// PermPolicyLibActionApplyUpdate applies a policy library by updating existing CAM policies.
	PermPolicyLibActionApplyUpdate PermPolicyLibAction = "apply_update"
)

// Validate checks whether the PermPolicyLibAction is valid.
func (a PermPolicyLibAction) Validate() error {
	switch a {
	case PermPolicyLibActionApplyCreate, PermPolicyLibActionApplyUpdate:
		return nil
	default:
		return fmt.Errorf("unsupported permission policy library action: %s", a)
	}
}
