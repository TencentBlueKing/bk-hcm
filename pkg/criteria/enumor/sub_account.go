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

// SubAccountType is sub account type.
type SubAccountType string

// Validate the AccountType is valid or not
func (a SubAccountType) Validate() error {
	switch a {
	case MainAccount:
	case CurrentAccount:
	default:
		return fmt.Errorf("unsupported sub account type: %s", a)
	}

	return nil
}

const (
	// MainAccount 主账号
	MainAccount SubAccountType = "main_account"
	// CurrentAccount 当前账号
	CurrentAccount SubAccountType = "current_account"
)

// SubAccountSecretStatus is sub account secret status.
type SubAccountSecretStatus string

// Validate the SubAccountSecretStatus is valid or not
func (s SubAccountSecretStatus) Validate() error {
	switch s {
	case EnabledSecretStatus:
	case DisabledSecretStatus:
	default:
		return fmt.Errorf("unsupported sub account secret status: %s", s)
	}

	return nil
}

const (
	// EnabledSecretStatus 启用状态
	EnabledSecretStatus SubAccountSecretStatus = "enabled"
	// DisabledSecretStatus 禁用状态
	DisabledSecretStatus SubAccountSecretStatus = "disabled"
)

// SubAccountConsoleLogin is sub account console login type.
type SubAccountConsoleLogin int64

// Validate the SubAccountConsoleLogin is valid or not
func (s SubAccountConsoleLogin) Validate() error {
	switch s {
	case ProgramAccount:
	case ConsoleAccount:
	default:
		return fmt.Errorf("unsupported sub account console login type: %d", s)
	}

	return nil
}

// GetNameZh get the chinese name of the sub account console login type.
func (s SubAccountConsoleLogin) GetNameZh() string {
	switch s {
	case ProgramAccount:
		return "编程账号"
	case ConsoleAccount:
		return "控制台账号"
	}

	return ""
}

const (
	// ProgramAccount 编程账号，无法登录控制台
	ProgramAccount SubAccountConsoleLogin = 0
	// ConsoleAccount 控制台账号，可登录控制台
	ConsoleAccount SubAccountConsoleLogin = 1
)
