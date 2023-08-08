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
	ID            string                 `json:"id"`
	CloudID       string                 `json:"cloud_id"`
	Name          string                 `json:"name"`
	Vendor        enumor.Vendor          `json:"vendor"`
	Site          enumor.AccountSiteType `json:"site"`
	AccountID     string                 `json:"account_id"`
	Managers      []string               `json:"managers"`
	BkBizIDs      []int64                `json:"bk_biz_ids"`
	Memo          *string                `json:"memo"`
	core.Revision `json:",inline"`
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
