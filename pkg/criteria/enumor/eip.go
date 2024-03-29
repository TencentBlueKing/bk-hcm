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

// EipBindStatus is eipBindStatus.
type EipBindStatus string

// Validate EipBindStatus.
func (v EipBindStatus) Validate() error {
	switch v {
	case EipBind:
	case EipUnBind:
	default:
		return fmt.Errorf("unsupported eip bind status: %s", v)
	}

	return nil
}

const (
	// EipBind status
	EipBind EipBindStatus = "BIND"
	// EipUnBind status
	EipUnBind EipBindStatus = "UNBIND"
)

// EipBindType is eipBindType.
type EipBindType string

// Validate EipBindType.
func (v EipBindType) Validate() error {
	switch v {
	case EipBindCvm:
	default:
		return fmt.Errorf("unsupported eip bind type: %s", v)
	}

	return nil
}

const (
	// EipBindCvm eip bind cvm
	EipBindCvm EipBindType = "CVM"
)
