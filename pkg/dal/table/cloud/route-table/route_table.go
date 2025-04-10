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

package routetable

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// RouteTableColumns defines all the route table table's columns.
var RouteTableColumns = utils.MergeColumns(nil, RouteTableColumnDescriptor)

// RouteTableColumnDescriptor is RouteTable's column descriptors.
var RouteTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// RouteTableTable 路由表的DB表
type RouteTableTable struct {
	// ID 路由表ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-" json:"vendor"`
	// AccountID 账号ID
	AccountID string `db:"account_id" validate:"max=64" json:"account_id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=255" json:"cloud_id"`
	// CloudID VPC的云上ID
	CloudVpcID string `db:"cloud_vpc_id" validate:"max=255" json:"cloud_vpc_id"`
	// Name 路由表名称
	Name *string `db:"name" validate:"omitempty,max=128" json:"name"`
	// Region 地域
	Region string `db:"region" validate:"max=255" json:"region"`
	// Memo 备注
	Memo *string `db:"memo" validate:"omitempty,max=255" json:"memo"`
	// VpcID VPC ID
	VpcID string `db:"vpc_id" validate:"max=64" json:"vpc_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" validate:"min=-1" json:"bk_biz_id"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" validate:"-" json:"extension"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return route table's table name.
func (r RouteTableTable) TableName() table.Name {
	return table.RouteTableTable
}

// InsertValidate validate route table table on insert.
func (r RouteTableTable) InsertValidate() error {
	if err := r.Vendor.Validate(); err != nil {
		return err
	}

	if len(r.AccountID) == 0 {
		return errors.New("account id can not be empty")
	}

	if len(r.CloudID) == 0 {
		return errors.New("cloud id can not be empty")
	}

	if r.Name == nil {
		return errors.New("name can not be nil")
	}

	if r.BkBizID == 0 {
		return errors.New("biz id can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate route table table on update.
func (r RouteTableTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.Name == nil && len(r.VpcID) == 0 && len(r.Extension) == 0 && r.BkBizID == 0 && r.Memo == nil {
		return errors.New("at least one of the update fields must be set")
	}

	if len(r.AccountID) != 0 {
		return errors.New("account id can not update")
	}

	if len(r.CloudID) != 0 {
		return errors.New("cloud id can not update")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}
