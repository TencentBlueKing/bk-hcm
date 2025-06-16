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

// GcpRouteColumns defines all the gcp route table's columns.
var GcpRouteColumns = utils.MergeColumns(nil, GcpRouteColumnDescriptor)

// GcpRouteColumnDescriptor is GcpRoute's column descriptors.
var GcpRouteColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "route_table_id", NamedC: "route_table_id", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "self_link", NamedC: "self_link", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "dest_range", NamedC: "dest_range", Type: enumor.String},
	{Column: "next_hop_gateway", NamedC: "next_hop_gateway", Type: enumor.String},
	{Column: "next_hop_ilb", NamedC: "next_hop_ilb", Type: enumor.String},
	{Column: "next_hop_instance", NamedC: "next_hop_instance", Type: enumor.String},
	{Column: "next_hop_ip", NamedC: "next_hop_ip", Type: enumor.String},
	{Column: "next_hop_network", NamedC: "next_hop_network", Type: enumor.String},
	{Column: "next_hop_peering", NamedC: "next_hop_peering", Type: enumor.String},
	{Column: "next_hop_vpn_tunnel", NamedC: "next_hop_vpn_tunnel", Type: enumor.String},
	{Column: "priority", NamedC: "priority", Type: enumor.Numeric},
	{Column: "route_status", NamedC: "route_status", Type: enumor.String},
	{Column: "route_type", NamedC: "route_type", Type: enumor.String},
	{Column: "tags", NamedC: "tags", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// GcpRouteTable gcp 路由的DB表
type GcpRouteTable struct {
	// ID 路由ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// CloudID 云上ID
	CloudID string `db:"cloud_id" validate:"max=64" json:"cloud_id"`
	// RouteTableID 路由表ID
	RouteTableID string `db:"route_table_id" validate:"max=64" json:"route_table_id"`
	// VpcID VPC的ID
	VpcID string `db:"vpc_id" validate:"max=64" json:"vpc_id"`
	// CloudVpcID VPC的云上ID
	CloudVpcID string `db:"cloud_vpc_id" validate:"max=255" json:"cloud_vpc_id"`
	// SelfLink 路由在GCP里的URL
	SelfLink string `db:"self_link" validate:"max=255" json:"self_link"`
	// Name 路由名称
	Name string `db:"name" validate:"max=128" json:"name"`
	// DestRange 目标 IP 地址范围, 可以是cidr或者具体ip地址
	DestRange string `db:"dest_range" validate:"cidr|ip" json:"dest_range"`
	// NextHopGateway 下一跳网关的URL
	NextHopGateway *string `db:"next_hop_gateway" validate:"omitempty,max=255" json:"next_hop_gateway,omitempty"`
	// NextHopIlb 下一跳内部负载均衡转发规则的URL或IP地址
	NextHopIlb *string `db:"next_hop_ilb" validate:"omitempty,max=255" json:"next_hop_ilb,omitempty"`
	// NextHopInstance 下一跳实例的URL
	NextHopInstance *string `db:"next_hop_instance" validate:"omitempty,max=255" json:"next_hop_instance,omitempty"`
	// NextHopIp 下一跳IP地址
	NextHopIp *string `db:"next_hop_ip" validate:"omitempty,max=255" json:"next_hop_ip,omitempty"`
	// NextHopNetwork 下一跳网络的URL
	NextHopNetwork *string `db:"next_hop_network" validate:"omitempty,max=255" json:"next_hop_network,omitempty"`
	// NextHopPeering 下一跳对等连接名称
	NextHopPeering *string `db:"next_hop_peering" validate:"omitempty,max=255" json:"next_hop_peering,omitempty"`
	// NextHopVpnTunnel 下一跳VPN通道的URL
	NextHopVpnTunnel *string `db:"next_hop_vpn_tunnel" validate:"omitempty,max=255" json:"next_hop_vpn_tunnel,omitempty"`
	// Priority 优先级
	Priority int64 `db:"priority" validate:"min=0,max=65535" json:"priority"`
	// RouteStatus 状态
	RouteStatus string `db:"route_status" validate:"max=32" json:"route_status"`
	// RouteType 路由类型
	RouteType string `db:"route_type" validate:"max=32" json:"route_type"`
	// Tags 实例标签列表
	Tags types.StringArray `db:"tags" validate:"omitempty" json:"tags,omitempty"`
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

// TableName return gcp route's table name.
func (r GcpRouteTable) TableName() table.Name {
	return table.GcpRouteTable
}

// InsertValidate validate gcp route table on insert.
func (r GcpRouteTable) InsertValidate() error {
	if len(r.CloudID) == 0 {
		return errors.New("cloud id can not be empty")
	}

	if len(r.RouteTableID) == 0 {
		return errors.New("route table id can not be empty")
	}

	if len(r.VpcID) == 0 {
		return errors.New("cloud route table id can not be empty")
	}

	if len(r.CloudVpcID) == 0 {
		return errors.New("cloud vpc id can not be empty")
	}

	if len(r.SelfLink) == 0 {
		return errors.New("self link can not be empty")
	}

	if len(r.Name) == 0 {
		return errors.New("name can not be empty")
	}

	if len(r.DestRange) == 0 {
		return errors.New("dest range can not be empty")
	}

	if r.NextHopGateway == nil && r.NextHopIlb == nil && r.NextHopInstance == nil && r.NextHopIp == nil &&
		r.NextHopNetwork == nil && r.NextHopPeering == nil && r.NextHopVpnTunnel == nil {
		return errors.New("one of the next hop must be set")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}
