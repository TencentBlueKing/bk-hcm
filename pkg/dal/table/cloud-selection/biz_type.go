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

package tableselection

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// BizTypeTableColumns defines all the biz_type table's columns.
var BizTypeTableColumns = utils.MergeColumns(nil, BizTypeTableColumnDescriptor)

// BizTypeTableColumnDescriptor is biz_type table column descriptors.
var BizTypeTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "biz_type", NamedC: "biz_type", Type: enumor.String},
	{Column: "cover_ping", NamedC: "cover_ping", Type: enumor.Numeric},
	{Column: "deployment_architecture", NamedC: "deployment_architecture", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.String},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.String},
}

// BizTypeTable biz_type表
type BizTypeTable struct {
	// ID biz_type ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// BizType 业务类型
	BizType string `db:"biz_type" json:"biz_type" validate:"max=64"`
	// CoverPing 网络延迟
	CoverPing float64 `db:"cover_ping" json:"cover_ping"`
	// DeploymentArchitecture 部署方式
	DeploymentArchitecture types.StringArray `db:"deployment_architecture" json:"deployment_architecture"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
}

// TableName return biz_type table name.
func (v BizTypeTable) TableName() table.Name {
	return table.CloudSelectionBizTypeTable
}

// InsertValidate validate biz_type table on insert.
func (v BizTypeTable) InsertValidate() error {
	if len(v.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(v)
}

// UpdateValidate validate biz_type table on update.
func (v BizTypeTable) UpdateValidate() error {
	if err := validator.Validate.Struct(v); err != nil {
		return err
	}

	if len(v.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(v.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
