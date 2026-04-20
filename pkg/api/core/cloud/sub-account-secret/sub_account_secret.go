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

// Package subaccountsecret defines sub account secret core structures.
package subaccountsecret

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// BaseSubAccountSecret 子账号密钥基础信息
type BaseSubAccountSecret struct {
	ID             string                        `json:"id"`
	Vendor         enumor.Vendor                 `json:"vendor"`
	Status         enumor.SubAccountSecretStatus `json:"status"`
	AccountID      string                        `json:"account_id"`
	SubAccountID   string                        `json:"sub_account_id"`
	CloudCreatedAt *string                       `json:"cloud_created_at"`
	DisabledTime   *string                       `json:"disabled_time"`
	LastUsedTime   *string                       `json:"last_used_time"`
	*core.Revision `json:",inline"`
}

// SubAccountSecret 子账号密钥（带 Extension）
type SubAccountSecret[Ext Extension] struct {
	BaseSubAccountSecret `json:",inline"`
	Extension            *Ext `json:"extension"`
}

// GetID 获取 ID
func (a SubAccountSecret[T]) GetID() string {
	return a.BaseSubAccountSecret.ID
}

// GetCloudID 获取云侧唯一标识
func (a SubAccountSecret[T]) GetCloudID() string {
	if a.Extension != nil {
		return (*a.Extension).GetCloudSecretID()
	}
	return ""
}

// Extension 子账号密钥扩展字段接口
type Extension interface {
	GetCloudSecretID() string
}

// TCloudSubAccountSecretListExt defines Tencent Cloud-only filter fields for biz-scoped sub account secret
// join list. Used in data-service request extension JSON and in DAO join filter options (same shape).
type TCloudSubAccountSecretListExt struct {
	CloudSecretIDs      []string                       `json:"cloud_secret_ids" validate:"omitempty,max=500,dive,lte=255"`
	CloudMainAccountIDs []string                       `json:"cloud_main_account_ids" validate:"omitempty,max=500,dive,lte=255"`
	CloudSubAccountIDs  []string                       `json:"cloud_sub_account_ids" validate:"omitempty,max=500,dive,lte=255"`
	ConsoleLogin        *enumor.SubAccountConsoleLogin `json:"console_login,omitempty" validate:"omitempty,min=0,max=1"`
}
