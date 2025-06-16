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

// TCloudRouteColumns defines all the tcloud route table's columns.
var TCloudRouteColumns = utils.MergeColumns(nil, TCloudRouteColumnDescriptor)

// TCloudRouteColumnDescriptor is tcloud route's column descriptors.
var TCloudRouteColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "route_table_id", NamedC: "route_table_id", Type: enumor.String},
	{Column: "cloud_route_table_id", NamedC: "cloud_route_table_id", Type: enumor.String},
	{Column: "destination_cidr_block", NamedC: "destination_cidr_block", Type: enumor.String},
	{Column: "destination_ipv6_cidr_block", NamedC: "destination_ipv6_cidr_block", Type: enumor.String},
	{Column: "gateway_type", NamedC: "gateway_type", Type: enumor.String},
	{Column: "cloud_gateway_id", NamedC: "cloud_gateway_id", Type: enumor.String},
	{Column: "enabled", NamedC: "enabled", Type: enumor.Boolean},
	{Column: "route_type", NamedC: "route_type", Type: enumor.String},
	{Column: "published_to_vbc", NamedC: "published_to_vbc", Type: enumor.Boolean},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// TCloudRouteTable tcloud 路由的DB表
type TCloudRouteTable struct {
	// ID 路由ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=64" json:"cloud_id"`
	// RouteTableID 路由表ID
	RouteTableID string `db:"route_table_id" validate:"max=64" json:"route_table_id"`
	// CloudRouteTableID 路由表的云上ID
	CloudRouteTableID string `db:"cloud_route_table_id" validate:"max=64" json:"cloud_route_table_id"`
	// DestinationCidrBlock 目的网段
	DestinationCidrBlock string `db:"destination_cidr_block" validate:"omitempty,cidrv4" json:"destination_cidr_block"`
	// DestinationIpv6CidrBlock 目的IPv6网段
	DestinationIpv6CidrBlock *string `db:"destination_ipv6_cidr_block" validate:"omitempty,cidrv6" json:"destination_ipv6_cidr_block"`
	// GatewayType 下一跳类型
	GatewayType string `db:"gateway_type" validate:"max=32" json:"gateway_type"`
	// CloudGatewayID 下一跳地址
	CloudGatewayID string `db:"cloud_gateway_id" validate:"max=255" json:"cloud_gateway_id"`
	// Enabled 是否启用
	Enabled *bool `db:"enabled" validate:"-" json:"enabled"`
	// RouteType 路由类型
	RouteType string `db:"route_type" validate:"max=32" json:"route_type"`
	// PublishedToVbc 路由策略是否发布到云联网
	PublishedToVbc *bool `db:"published_to_vbc" validate:"-" json:"published_to_vbc"`
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

// TableName return tcloud route's table name.
func (r TCloudRouteTable) TableName() table.Name {
	return table.TCloudRouteTable
}

// InsertValidate validate tcloud route table on insert.
func (r TCloudRouteTable) InsertValidate() error {
	if r.CloudID == "" {
		return errors.New("cloud id can not be empty")
	}

	if r.RouteTableID == "" {
		return errors.New("route table id can not be empty")
	}

	if r.CloudRouteTableID == "" {
		return errors.New("cloud route table id can not be empty")
	}

	if r.DestinationCidrBlock == "" {
		return errors.New("destination cidr block can not be empty")
	}

	if r.GatewayType == "" {
		return errors.New("gateway type can not be empty")
	}

	if r.CloudGatewayID == "" {
		return errors.New("cloud gateway id can not be empty")
	}

	if r.RouteType == "" {
		return errors.New("route type can not be empty")
	}

	if r.Creator == "" {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate tcloud route table on update.
func (r TCloudRouteTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if r.DestinationCidrBlock == "" && r.DestinationIpv6CidrBlock == nil && r.GatewayType == "" &&
		r.CloudGatewayID == "" && r.RouteType == "" && r.Enabled == nil && r.PublishedToVbc == nil && r.Memo == nil {
		return errors.New("at least one of the update fields must be set")
	}

	if r.CloudID != "" {
		return errors.New("cloud id can not update")
	}

	if r.RouteTableID != "" {
		return errors.New("route table id can not update")
	}

	if r.CloudRouteTableID != "" {
		return errors.New("cloud route table id can not update")
	}

	if r.Creator != "" {
		return errors.New("creator can not update")
	}

	if r.Reviser == "" {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}
