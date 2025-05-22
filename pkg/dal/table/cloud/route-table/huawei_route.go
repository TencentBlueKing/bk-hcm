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

// HuaWeiRouteColumns defines all the huawei route table's columns.
var HuaWeiRouteColumns = utils.MergeColumns(nil, HuaWeiRouteColumnDescriptor)

// HuaWeiRouteColumnDescriptor is HuaWeiRoute's column descriptors.
var HuaWeiRouteColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "route_table_id", NamedC: "route_table_id", Type: enumor.String},
	{Column: "cloud_route_table_id", NamedC: "cloud_route_table_id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "destination", NamedC: "destination", Type: enumor.String},
	{Column: "nexthop", NamedC: "nexthop", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// HuaWeiRouteTable huawei 路由的DB表
type HuaWeiRouteTable struct {
	// ID 路由ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// RouteTableID 路由表ID
	RouteTableID string `db:"route_table_id" validate:"max=64" json:"route_table_id"`
	// CloudRouteTableID 路由表的云上ID
	CloudRouteTableID string `db:"cloud_route_table_id" validate:"max=64" json:"cloud_route_table_id"`
	// Type 路由类型
	Type string `db:"type" validate:"max=32" json:"type"`
	// Destination 目的网段
	Destination string `db:"destination" validate:"omitempty,cidr" json:"destination"`
	// NextHop 下一跳对象的ID
	NextHop string `db:"nexthop" validate:"max=255" json:"nexthop"`
	// Memo 备注
	Memo *string `db:"memo" validate:"omitempty,max=255" json:"memo"`
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

// TableName return huawei route's table name.
func (r HuaWeiRouteTable) TableName() table.Name {
	return table.HuaWeiRouteTable
}

// InsertValidate validate huawei route table on insert.
func (r HuaWeiRouteTable) InsertValidate() error {
	if len(r.RouteTableID) == 0 {
		return errors.New("route table id can not be empty")
	}

	if len(r.CloudRouteTableID) == 0 {
		return errors.New("cloud route table id can not be empty")
	}

	if len(r.Type) == 0 {
		return errors.New("type can not be empty")
	}

	if len(r.Destination) == 0 {
		return errors.New("destination can not be empty")
	}

	if len(r.NextHop) == 0 {
		return errors.New("next hop can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate huawei route table on update.
func (r HuaWeiRouteTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Type) == 0 && len(r.Destination) == 0 && len(r.NextHop) == 0 && r.Memo == nil {
		return errors.New("at least one of the update fields must be set")
	}

	if len(r.RouteTableID) != 0 {
		return errors.New("route table id can not update")
	}

	if len(r.CloudRouteTableID) != 0 {
		return errors.New("cloud route table id can not update")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}
