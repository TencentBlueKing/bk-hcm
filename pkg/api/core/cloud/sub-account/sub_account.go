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

package coresubaccount

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// SubAccount define sub account.
type SubAccount[Ext Extension] struct {
	BaseSubAccount `json:",inline"`
	Extension      *Ext `json:"extension"`
}

// BizSubAccountItem defines biz sub account ext response item.
// 业务下返回的子账号详情
type BizSubAccountItem[Ext Extension] struct {
	SubAccount[Ext]       `json:",inline"`
	Operable              bool   `json:"operable"`
	AccountName           string `json:"account_name"`
	SubAccountSecretCount uint64 `json:"sub_account_secret_count"`
}

// BizSubAccountExtListResult defines biz sub account ext response.
// 业务下返回的子账号列表
type BizSubAccountExtListResult[Ext Extension] struct {
	Count   uint64                   `json:"count"`
	Details []BizSubAccountItem[Ext] `json:"details"`
}

// GetID ...
func (account SubAccount[Ext]) GetID() string {
	return account.ID
}

// GetCloudID ...
func (account SubAccount[Ext]) GetCloudID() string {
	return account.CloudID
}

// Extension account extension.
type Extension interface {
	TCloudExtension | AwsExtension | HuaWeiExtension | AzureExtension | GcpExtension
}

// BaseSubAccount 云账号
type BaseSubAccount struct {
	ID                    string                 `json:"id"`
	CloudID               string                 `json:"cloud_id"`
	Name                  string                 `json:"name"`
	Vendor                enumor.Vendor          `json:"vendor"`
	Site                  enumor.AccountSiteType `json:"site"`
	AccountID             string                 `json:"account_id"`
	AccountType           string                 `json:"account_type"`
	Managers              []string               `json:"managers"`
	PermissionTemplateIDs []string               `json:"permission_template_ids"`
	BkBizIDs              []int64                `json:"bk_biz_ids"`
	Email                 *string                `json:"email,omitempty"`
	PhoneNum              *string                `json:"phone_num,omitempty"`
	CountryCode           *string                `json:"country_code,omitempty"`
	CloudCreatedAt        *string                `json:"cloud_created_at,omitempty"`
	Memo                  *string                `json:"memo"`
	core.Revision         `json:",inline"`
}

// TCloudExtension define tcloud extension.
type TCloudExtension struct {
	// CloudMainAccountID 主账号ID
	CloudMainAccountID string `json:"cloud_main_account_id"`
	// 子用户用户 ID
	Uin *uint64 `json:"uin"`
	// 昵称
	// 注意：此字段可能返回 null，表示取不到有效值。
	NickName *string `json:"nick_name"`
	// 创建时间
	// 注意：此字段可能返回 null，表示取不到有效值。
	CreateTime *string `json:"create_time"`
	// LoginFlag 登录保护设置
	// 注意：此字段可能返回 null，表示取不到有效值。
	LoginFlag *enumor.AccountProtectionFlag `json:"login_flag,omitempty"`
	// ActionFlag 敏感操作保护设置
	// 注意：此字段可能返回 null，表示取不到有效值。
	ActionFlag *enumor.AccountProtectionFlag `json:"action_flag,omitempty"`
	// ConsoleLogin 控制台登录权限
	// 注意：此字段可能返回 null，表示取不到有效值。
	ConsoleLogin *enumor.SubAccountConsoleLogin `json:"console_login,omitempty"`
}

// AwsExtension define aws extension.
type AwsExtension struct {
	CloudAccountID string  `json:"cloud_account_id"`
	Arn            *string `json:"arn"`
	JoinedMethod   *string `json:"joined_method"`
	Status         *string `json:"status"`
}

// AzureExtension define azure extension.
type AzureExtension struct {
	DisplayNameName       *string `json:"display_name_name"`
	GivenName             *string `json:"given_name"`
	SurName               *string `json:"sur_name"`
	CloudTenantID         string  `json:"cloud_tenant_id"`
	CloudSubscriptionID   string  `json:"cloud_subscription_id"`
	CloudSubscriptionName string  `json:"cloud_subscription_name"`
}

// HuaWeiExtension define huawei extension.
type HuaWeiExtension struct {
	CloudAccountID string  `json:"cloud_account_id"`
	LastProjectID  *string `json:"last_project_id"`
	Enabled        bool    `json:"enabled"`
}

// GcpExtension define gcp extension.
type GcpExtension struct {
	CloudProjectID   string `json:"cloud_project_id"`
	CloudProjectName string `json:"cloud_project_name"`
}
