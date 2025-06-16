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

// AzureRouteColumns defines all the azure route table's columns.
var AzureRouteColumns = utils.MergeColumns(nil, AzureRouteColumnDescriptor)

// AzureRouteColumnDescriptor is AzureRoute's column descriptors.
var AzureRouteColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "route_table_id", NamedC: "route_table_id", Type: enumor.String},
	{Column: "cloud_route_table_id", NamedC: "cloud_route_table_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "address_prefix", NamedC: "address_prefix", Type: enumor.String},
	{Column: "next_hop_type", NamedC: "next_hop_type", Type: enumor.String},
	{Column: "next_hop_ip_address", NamedC: "next_hop_ip_address", Type: enumor.String},
	{Column: "provisioning_state", NamedC: "provisioning_state", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AzureRouteTable azure 路由的DB表
type AzureRouteTable struct {
	// ID 路由ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=255" json:"cloud_id"`
	// RouteTableID 路由表ID
	RouteTableID string `db:"route_table_id" validate:"max=64" json:"route_table_id"`
	// CloudRouteTableID 路由表的云上ID
	CloudRouteTableID string `db:"cloud_route_table_id" validate:"max=255" json:"cloud_route_table_id"`
	// Name 路由名称
	Name string `db:"name" validate:"max=80" json:"name"`
	// AddressPrefix 目的网段
	AddressPrefix string `db:"address_prefix" validate:"max=64" json:"address_prefix"`
	// NextHopType 下一跳类型
	NextHopType string `db:"next_hop_type" validate:"max=32" json:"next_hop_type"`
	// NextHopIPAddress 下一跳地址
	NextHopIPAddress *string `db:"next_hop_ip_address" validate:"max=255" json:"next_hop_ip_address,omitempty"`
	// ProvisioningState 当前供应状态
	ProvisioningState string `db:"provisioning_state" validate:"max=32" json:"provisioning_state"`
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

// TableName return azure route's table name.
func (r AzureRouteTable) TableName() table.Name {
	return table.AzureRouteTable
}

// InsertValidate validate azure route table on insert.
func (r AzureRouteTable) InsertValidate() error {
	if len(r.CloudID) == 0 {
		return errors.New("cloud id can not be empty")
	}

	if len(r.RouteTableID) == 0 {
		return errors.New("route table id can not be empty")
	}

	if len(r.CloudRouteTableID) == 0 {
		return errors.New("cloud route table id can not be empty")
	}

	if len(r.Name) == 0 {
		return errors.New("name can not be empty")
	}

	if len(r.AddressPrefix) == 0 {
		return errors.New("address prefix can not be empty")
	}

	if len(r.NextHopType) == 0 {
		return errors.New("next hop type can not be empty")
	}

	if len(r.ProvisioningState) == 0 {
		return errors.New("provisioning state can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate azure route table on update.
func (r AzureRouteTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.AddressPrefix) == 0 && len(r.NextHopType) == 0 && r.NextHopIPAddress == nil &&
		len(r.ProvisioningState) == 0 {
		return errors.New("at least one of the update fields must be set")
	}

	if len(r.CloudID) != 0 {
		return errors.New("cloud id can not update")
	}

	if len(r.RouteTableID) != 0 {
		return errors.New("route table id can not update")
	}

	if len(r.CloudRouteTableID) != 0 {
		return errors.New("cloud route table id can update")
	}

	if len(r.Name) != 0 {
		return errors.New("name can not update")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}
