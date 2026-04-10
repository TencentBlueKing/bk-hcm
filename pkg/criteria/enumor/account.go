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

// AccountType is account type.
type AccountType string

// Validate the AccountType is valid or not
func (a AccountType) Validate() error {
	switch a {
	case ResourceAccount:
	case RegistrationAccount:
	case SecurityAuditAccount:
	default:
		return fmt.Errorf("unsupported account type: %s", a)

	}

	return nil
}

const (
	// ResourceAccount 资源账号是可以用于管理该账号资源的账号。
	ResourceAccount AccountType = "resource"
	// RegistrationAccount 登记账号仅用于账号管理，不用与管理该账号下的资源。
	RegistrationAccount AccountType = "registration"
	// SecurityAuditAccount 安全审计账号，仅用于安全所需审计，不管理账号下的资源
	SecurityAuditAccount AccountType = "security_audit"
)

// AccountSiteType is site type.
type AccountSiteType string

// Validate the AccountSiteType is valid or not
func (a AccountSiteType) Validate() error {
	switch a {
	case ChinaSite:
	case InternationalSite:
	default:
		return fmt.Errorf("unsupported account site type: %s", a)

	}

	return nil
}

const (
	// ChinaSite is china site.
	ChinaSite AccountSiteType = "china"
	// InternationalSite is international site.
	InternationalSite AccountSiteType = "international"
)

// AccountSyncStatus is account sync status.
type AccountSyncStatus string

const (
	// NotStart is account not start sync status.
	// TODO: 同步时候考虑未同步时使用什么名称
	NotStart = "not_start"
)

var (

	// AccountTypeNameMap 账号类型和对应中文名
	AccountTypeNameMap = map[AccountType]string{
		RegistrationAccount:  "登记账号",
		ResourceAccount:      "资源账号",
		SecurityAuditAccount: "安全审计账号",
	}

	// AccountSiteTypeNameMap 站点类型中文名
	AccountSiteTypeNameMap = map[AccountSiteType]string{
		InternationalSite: "国际站",
		ChinaSite:         "中国站",
	}
)

// AccountSecretType is account secret type.
type AccountSecretType string

// Validate the AccountSecretType is valid or not
func (a AccountSecretType) Validate() error {
	switch a {
	case ResourceSecretType:
	case SecuritySecretType:
	default:
		return fmt.Errorf("unsupported account secret type: %s", a)
	}

	return nil
}

const (
	// ResourceSecretType 资源管理密钥
	ResourceSecretType AccountSecretType = "resource"
	// SecuritySecretType 安全管理密钥
	SecuritySecretType AccountSecretType = "security"
)

// AccountSecretStatus is account secret status.
type AccountSecretStatus string

// Validate the AccountSecretStatus is valid or not
func (a AccountSecretStatus) Validate() error {
	switch a {
	case NormalSecretStatus:
	case InvalidSecretStatus:
	default:
		return fmt.Errorf("unsupported account secret status: %s", a)
	}

	return nil
}

const (
	// NormalSecretStatus 正常状态
	NormalSecretStatus AccountSecretStatus = "normal"
	// InvalidSecretStatus 失效状态
	InvalidSecretStatus AccountSecretStatus = "invalid"
)

// AccountProtectionFlag is account protection flag.
type AccountProtectionFlag string

// Validate the AccountProtectionFlag is valid or not
func (a AccountProtectionFlag) Validate() error {
	switch a {
	case PhoneProtection:
	case TokenProtection:
	case StokenProtection:
	case WechatProtection:
	case CustomProtection:
	case MailProtection:
	case U2FTokenProtection:
	default:
		return fmt.Errorf("unsupported account protection flag: %s", a)
	}

	return nil
}

const (
	// PhoneProtection 安全手机
	PhoneProtection AccountProtectionFlag = "phone"
	// TokenProtection 硬token
	TokenProtection AccountProtectionFlag = "token"
	// StokenProtection MFA字段
	StokenProtection AccountProtectionFlag = "stoken"
	// WechatProtection 微信
	WechatProtection AccountProtectionFlag = "wechat"
	// CustomProtection 自定义
	CustomProtection AccountProtectionFlag = "custom"
	// MailProtection 邮箱
	MailProtection AccountProtectionFlag = "mail"
	// U2FTokenProtection u2f硬件token
	U2FTokenProtection AccountProtectionFlag = "u2FToken"
)
