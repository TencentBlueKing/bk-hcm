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

package cloud

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// SubnetColumns defines all the subnet table's columns.
var SubnetColumns = utils.MergeColumns(nil, SubnetColumnDescriptor)

// SubnetColumnDescriptor is subnet table column descriptors.
var SubnetColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "cloud_route_table_id", NamedC: "cloud_route_table_id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "ipv4_cidr", NamedC: "ipv4_cidr", Type: enumor.Json},
	{Column: "ipv6_cidr", NamedC: "ipv6_cidr", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "route_table_id", NamedC: "route_table_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// SubnetTable subnet表
type SubnetTable struct {
	// ID subnet ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-" json:"vendor"`
	// AccountID 账号ID
	AccountID string `db:"account_id" validate:"max=64" json:"account_id"`
	// CloudVpcID 云上vpc的ID
	CloudVpcID string `db:"cloud_vpc_id" validate:"max=255" json:"cloud_vpc_id"`
	// CloudRouteTableID 云上的路由表ID
	CloudRouteTableID *string `db:"cloud_route_table_id" validate:"omitempty,max=255" json:"cloud_route_table_id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=255" json:"cloud_id"`
	// Name subnet名称
	Name *string `db:"name" validate:"omitempty,max=128" json:"name"`
	// Region 地域
	Region string `db:"region" validate:"max=255" json:"region"`
	// Zone 可用区
	Zone string `db:"zone" validate:"max=255" json:"zone"`
	// Ipv4Cidr ipv4 cidr
	Ipv4Cidr types.StringArray `db:"ipv4_cidr" validate:"-" json:"ipv4_cidr"`
	// Ipv6Cidr ipv6 cidr
	Ipv6Cidr types.StringArray `db:"ipv6_cidr" validate:"-" json:"ipv6_cidr"`
	// Memo 备注
	Memo *string `db:"memo" validate:"omitempty,max=255" json:"memo"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" validate:"-" json:"extension"`
	// VpcID vpc的ID
	VpcID string `db:"vpc_id" validate:"max=64" json:"vpc_id"`
	// RouteTableID 路由表ID
	RouteTableID *string `db:"route_table_id" validate:"omitempty,max=64" json:"route_table_id"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" validate:"min=-1" json:"bk_biz_id"`
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

// TableName return subnet table name.
func (s SubnetTable) TableName() table.Name {
	return table.SubnetTable
}

// InsertValidate validate subnet table on insert.
func (s SubnetTable) InsertValidate() error {
	if err := s.Vendor.Validate(); err != nil {
		return err
	}

	if len(s.AccountID) == 0 {
		return errors.New("account id can not be empty")
	}

	if len(s.CloudVpcID) == 0 {
		return errors.New("cloud vpc id can not be empty")
	}

	if len(s.CloudID) == 0 {
		return errors.New("cloud id can not be empty")
	}

	if s.Name == nil {
		return errors.New("name can not be nil")
	}

	return validator.Validate.Struct(s)
}

// UpdateValidate validate subnet table on update.
func (s SubnetTable) UpdateValidate() error {
	if err := validator.Validate.Struct(s); err != nil {
		return err
	}

	if s.Name == nil && len(s.Ipv4Cidr) == 0 && s.Ipv6Cidr == nil && len(s.Extension) == 0 && s.RouteTableID == nil &&
		s.CloudRouteTableID == nil && s.BkBizID == 0 && s.Memo == nil {
		return errors.New("at least one of the update fields must be set")
	}

	if len(s.AccountID) != 0 {
		return errors.New("account id can not update")
	}

	if len(s.CloudID) != 0 {
		return errors.New("cloud id can not update")
	}

	if len(s.Zone) != 0 {
		return errors.New("zone can not update")
	}

	if len(s.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(s.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(s)
}
