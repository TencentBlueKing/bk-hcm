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

// Package accountsecret defines account secret core structures.
package accountsecret

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// BaseAccountSecret 账号密钥基础信息
type BaseAccountSecret struct {
	ID             string                     `json:"id"`
	AccountID      string                     `json:"account_id"`
	Vendor         enumor.Vendor              `json:"vendor"`
	Type           enumor.AccountSecretType   `json:"type"`
	Status         enumor.AccountSecretStatus `json:"status"`
	*core.Revision `json:",inline"`
}

// AccountSecret 账号密钥（带 Extension）
type AccountSecret[Ext Extension] struct {
	BaseAccountSecret `json:",inline"`
	Extension         *Ext `json:"extension"`
}

// GetID 获取 ID
func (a AccountSecret[T]) GetID() string {
	return a.BaseAccountSecret.ID
}

// Extension 账号密钥扩展字段接口
type Extension interface {
	TCloudAccountSecretExtension
}
