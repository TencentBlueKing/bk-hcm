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

// Package image ...
package image

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ImageColumns ...
var ImageColumns = utils.MergeColumns(nil, ImageColumnDescriptor)

// ImageColumnDescriptor ...
var ImageColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "architecture", NamedC: "architecture", Type: enumor.String},
	{Column: "platform", NamedC: "platform", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "os_type", NamedC: "os_type", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ImageModel ...
type ImageModel struct {
	ID           string          `db:"id" json:"id"`
	Vendor       string          `db:"vendor" json:"vendor"`
	CloudID      string          `db:"cloud_id" json:"cloud_id"`
	Name         string          `db:"name" json:"name"`
	Architecture string          `db:"architecture" json:"architecture"`
	Platform     string          `db:"platform" json:"platform"`
	State        string          `db:"state" json:"state"`
	Type         string          `db:"type" json:"type"`
	Extension    types.JsonField `db:"extension" json:"extension"`
	OsType       enumor.OsType   `db:"os_type" json:"os_type"`
	Creator      string          `db:"creator" json:"creator"`
	Reviser      string          `db:"reviser" json:"reviser"`
	CreatedAt    types.Time      `db:"created_at" json:"created_at"`
	UpdatedAt    types.Time      `db:"updated_at" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// InsertValidate ...
func (m *ImageModel) InsertValidate() error {
	return validator.Validate.Struct(m)
}

// UpdateValidate ...
func (m *ImageModel) UpdateValidate() error {
	return validator.Validate.Struct(m)
}
