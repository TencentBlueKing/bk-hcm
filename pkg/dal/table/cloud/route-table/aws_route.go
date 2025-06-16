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

// AwsRouteColumns defines all the aws route table's columns.
var AwsRouteColumns = utils.MergeColumns(nil, AwsRouteColumnDescriptor)

// AwsRouteColumnDescriptor is AwsRoute's column descriptors.
var AwsRouteColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "route_table_id", NamedC: "route_table_id", Type: enumor.String},
	{Column: "cloud_route_table_id", NamedC: "cloud_route_table_id", Type: enumor.String},
	{Column: "destination_cidr_block", NamedC: "destination_cidr_block", Type: enumor.String},
	{Column: "destination_ipv6_cidr_block", NamedC: "destination_ipv6_cidr_block", Type: enumor.String},
	{Column: "cloud_carrier_gateway_id", NamedC: "cloud_carrier_gateway_id", Type: enumor.String},
	{Column: "core_network_arn", NamedC: "core_network_arn", Type: enumor.String},
	{Column: "cloud_destination_prefix_list_id", NamedC: "cloud_destination_prefix_list_id", Type: enumor.String},
	{Column: "cloud_egress_only_internet_gateway_id", NamedC: "cloud_egress_only_internet_gateway_id",
		Type: enumor.String},
	{Column: "cloud_gateway_id", NamedC: "cloud_gateway_id", Type: enumor.String},
	{Column: "cloud_instance_id", NamedC: "cloud_instance_id", Type: enumor.String},
	{Column: "cloud_instance_owner_id", NamedC: "cloud_instance_owner_id", Type: enumor.String},
	{Column: "cloud_local_gateway_id", NamedC: "cloud_local_gateway_id", Type: enumor.String},
	{Column: "cloud_nat_gateway_id", NamedC: "cloud_nat_gateway_id", Type: enumor.String},
	{Column: "cloud_network_interface_id", NamedC: "cloud_network_interface_id", Type: enumor.String},
	{Column: "cloud_transit_gateway_id", NamedC: "cloud_transit_gateway_id", Type: enumor.String},
	{Column: "cloud_vpc_peering_connection_id", NamedC: "cloud_vpc_peering_connection_id", Type: enumor.String},
	{Column: "state", NamedC: "state", Type: enumor.String},
	{Column: "propagated", NamedC: "propagated", Type: enumor.Boolean},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AwsRouteTable aws 路由的DB表
type AwsRouteTable struct {
	// ID 路由ID
	ID string `db:"id" validate:"len=0" json:"id"`
	// RouteTableID 路由表ID
	RouteTableID string `db:"route_table_id" validate:"max=64" json:"route_table_id"`
	// CloudRouteTableID 路由表的云上ID
	CloudRouteTableID string `db:"cloud_route_table_id" validate:"max=64" json:"cloud_route_table_id"`
	// DestinationCidrBlock 目的网段
	DestinationCidrBlock *string `db:"destination_cidr_block" validate:"omitempty,cidrv4" json:"destination_cidr_block,omitempty"`
	// DestinationIpv6CidrBlock 目的IPv6网段
	DestinationIpv6CidrBlock *string `db:"destination_ipv6_cidr_block" validate:"omitempty,cidrv6" json:"destination_ipv6_cidr_block,omitempty"`
	// CloudDestinationPrefixListID 目的AWS前缀的云上ID
	CloudDestinationPrefixListID *string `db:"cloud_destination_prefix_list_id" validate:"omitempty,max=255" json:"cloud_destination_prefix_list_id,omitempty"`
	// CloudCarrierGatewayID 运营商网关的云上ID
	CloudCarrierGatewayID *string `db:"cloud_carrier_gateway_id" validate:"omitempty,max=255" json:"cloud_carrier_gateway_id,omitempty"`
	// CoreNetworkArn 核心网络的Amazon资源名称(ARN)
	CoreNetworkArn *string `db:"core_network_arn" validate:"omitempty,max=255" json:"core_network_arn,omitempty"`
	// CloudEgressOnlyInternetGatewayID 仅用于出站的Internet网关的云上ID
	CloudEgressOnlyInternetGatewayID *string `db:"cloud_egress_only_internet_gateway_id" validate:"omitempty,max=255" json:"cloud_egress_only_internet_gateway_id,omitempty"`
	// CloudGatewayID 网关的云上ID
	CloudGatewayID *string `db:"cloud_gateway_id" validate:"omitempty,max=255" json:"cloud_gateway_id,omitempty"`
	// CloudInstanceID NAT实例的云上ID
	CloudInstanceID *string `db:"cloud_instance_id" validate:"omitempty,max=255" json:"cloud_instance_id,omitempty"`
	// CloudInstanceOwnerID 实例所属账户的云上ID
	CloudInstanceOwnerID *string `db:"cloud_instance_owner_id" validate:"omitempty,max=255" json:"cloud_instance_owner_id,omitempty"`
	// CloudLocalGatewayID 本地网关的云上ID
	CloudLocalGatewayID *string `db:"cloud_local_gateway_id" validate:"omitempty,max=255" json:"cloud_local_gateway_id,omitempty"`
	// CloudNatGatewayID NAT网关的云上ID
	CloudNatGatewayID *string `db:"cloud_nat_gateway_id" validate:"omitempty,max=255" json:"cloud_nat_gateway_id,omitempty"`
	// CloudNetworkInterfaceID 网络接口的云上ID
	CloudNetworkInterfaceID *string `db:"cloud_network_interface_id" validate:"omitempty,max=255" json:"cloud_network_interface_id,omitempty"`
	// CloudTransitGatewayID 中转网关的云上ID
	CloudTransitGatewayID *string `db:"cloud_transit_gateway_id" validate:"omitempty,max=255" json:"cloud_transit_gateway_id,omitempty"`
	// CloudVpcPeeringConnectionID VPC对等连接的云上ID
	CloudVpcPeeringConnectionID *string `db:"cloud_vpc_peering_connection_id" validate:"omitempty,max=255" json:"cloud_vpc_peering_connection_id,omitempty"` // State 状态
	// State 状态
	State string `db:"state" validate:"max=32" json:"state"`
	// Propagated 是否已传播
	Propagated *bool `db:"propagated" validate:"-" json:"propagated"`
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

// TableName return aws route's table name.
func (r AwsRouteTable) TableName() table.Name {
	return table.AwsRouteTable
}

// InsertValidate validate aws route table on insert.
func (r AwsRouteTable) InsertValidate() error {
	if r.RouteTableID == "" {
		return errors.New("route table id can not be empty")
	}

	if r.CloudRouteTableID == "" {
		return errors.New("cloud route table id can not be empty")
	}

	if r.DestinationCidrBlock == nil && r.DestinationIpv6CidrBlock == nil && r.CloudDestinationPrefixListID == nil {
		return errors.New("one of the destinations must be set")
	}

	if r.DestinationCidrBlock != nil && r.DestinationIpv6CidrBlock != nil && r.CloudDestinationPrefixListID != nil {
		return errors.New("only one of the destinations can be set")
	}

	if !isAwsRouteTableTargetSet(&r) {
		return errors.New("one of the targets must be set")
	}

	if r.Creator == "" {
		return errors.New("creator can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate aws route table on update.
func (r AwsRouteTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if !isAwsRouteTableTargetSet(&r) && r.State == "" && r.Propagated == nil {
		return errors.New("one of the targets must be set")
	}

	if r.RouteTableID != "" {
		return errors.New("route table id can not update")
	}

	if r.CloudRouteTableID != "" {
		return errors.New("cloud route table id can not update")
	}

	if r.DestinationCidrBlock != nil {
		return errors.New("destination cidr can not update")
	}

	if r.DestinationIpv6CidrBlock != nil {
		return errors.New("destination ipv6 cidr can not update")
	}

	if r.CloudDestinationPrefixListID != nil {
		return errors.New("cloud destination prefix list id can not update")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}

func isAwsRouteTableTargetSet(r *AwsRouteTable) bool {
	return r.CloudCarrierGatewayID != nil || r.CoreNetworkArn != nil || r.CloudEgressOnlyInternetGatewayID != nil ||
		r.CloudGatewayID != nil || r.CloudInstanceID != nil || r.CloudInstanceOwnerID != nil ||
		r.CloudLocalGatewayID != nil || r.CloudNatGatewayID != nil || r.CloudNetworkInterfaceID != nil ||
		r.CloudTransitGatewayID != nil || r.CloudVpcPeeringConnectionID != nil
}
