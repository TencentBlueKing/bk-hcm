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

// TCloudPolicyType 腾讯云策略类型
type TCloudPolicyType int

const (
	// TCloudCustomPolicy 自定义策略
	TCloudCustomPolicy TCloudPolicyType = 1
	// TCloudPresetPolicy 预设策略
	TCloudPresetPolicy TCloudPolicyType = 2
)

// Validate whether the TCloudPolicyType is valid.
func (t TCloudPolicyType) Validate() error {
	switch t {
	case TCloudCustomPolicy, TCloudPresetPolicy:
		return nil
	default:
		return fmt.Errorf("unsupported tcloud policy type: %d", t)
	}
}

// OperatePermTemplateAction is the action type for operate permission template application.
type OperatePermTemplateAction string

const (
	// PermTemplateActionCreate creates a new permission template from a policy library.
	PermTemplateActionCreate OperatePermTemplateAction = "create"
	// PermTemplateActionUpdate updates an existing custom permission template to use a new policy library.
	PermTemplateActionUpdate OperatePermTemplateAction = "update"
	// PermTemplateActionDelete deletes an existing custom permission template.
	PermTemplateActionDelete OperatePermTemplateAction = "delete"
)

// Validate checks whether the OperatePermTemplateAction is valid.
func (a OperatePermTemplateAction) Validate() error {
	switch a {
	case PermTemplateActionCreate, PermTemplateActionUpdate, PermTemplateActionDelete:
		return nil
	default:
		return fmt.Errorf("unsupported operate permission template action: %s", a)
	}
}
