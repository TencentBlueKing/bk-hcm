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

// Package eip ...
package eip

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// EipColumns ...
var EipColumns = utils.MergeColumns(nil, EipColumnDescriptor)

// EipColumnDescriptor ...
var EipColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "recycle_status", NamedC: "recycle_status", Type: enumor.String},
	{Column: "public_ip", NamedC: "public_ip", Type: enumor.String},
	{Column: "private_ip", NamedC: "private_ip", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// EipModel ...
type EipModel struct {
	ID            string          `db:"id" json:"id"`
	Vendor        string          `db:"vendor" json:"vendor"`
	AccountID     string          `db:"account_id" json:"account_id"`
	CloudID       string          `db:"cloud_id" json:"cloud_id"`
	BkBizID       int64           `db:"bk_biz_id" json:"bk_biz_id"`
	Name          *string         `db:"name" json:"name"`
	Region        string          `db:"region" json:"region"`
	Status        string          `db:"status" json:"status"`
	RecycleStatus string          `db:"recycle_status" json:"recycle_status,omitempty"`
	PublicIp      string          `db:"public_ip" json:"public_ip"`
	PrivateIp     string          `db:"private_ip" json:"private_ip"`
	Extension     types.JsonField `db:"extension" json:"extension" validate:"-"`
	Creator       string          `db:"creator" json:"creator"`
	Reviser       string          `db:"reviser" json:"reviser"`
	CreatedAt     types.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     types.Time      `db:"updated_at" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// InsertValidate ...
func (m *EipModel) InsertValidate() error {

	if len(m.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(m.Status) == 0 {
		return errors.New("status is required")
	}

	if len(m.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(m.Extension) == 0 {
		return errors.New("extension is required")
	}

	if len(m.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(m.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return validator.Validate.Struct(m)
}

// UpdateValidate ...
func (m *EipModel) UpdateValidate() error {
	return validator.Validate.Struct(m)
}
