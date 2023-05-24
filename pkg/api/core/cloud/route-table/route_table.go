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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// RouteTable defines route table info.
type RouteTable[T RouteTableExtension] struct {
	BaseRouteTable `json:",inline"`
	Extension      *T `json:"extension"`
}

// BaseRouteTable defines base route table info.
type BaseRouteTable struct {
	ID             string        `json:"id"`
	Vendor         enumor.Vendor `json:"vendor"`
	AccountID      string        `json:"account_id"`
	CloudID        string        `json:"cloud_id"`
	CloudVpcID     string        `json:"cloud_vpc_id"`
	Name           string        `json:"name"`
	Region         string        `json:"region"`
	Memo           *string       `json:"memo,omitempty"`
	VpcID          string        `json:"vpc_id"`
	BkBizID        int64         `json:"bk_biz_id"`
	*core.Revision `json:",inline"`
}

// RouteTableExtension defines route table extensional info.
type RouteTableExtension interface {
	TCloudRouteTableExtension | AwsRouteTableExtension | AzureRouteTableExtension | HuaWeiRouteTableExtension
}

// TCloudRouteTableExtension defines tencent cloud route table extensional info.
type TCloudRouteTableExtension struct {
	Main bool `json:"main"`
}

// TCloudRouteTableAsst defines tencent cloud route table association info.
type TCloudRouteTableAsst struct {
	CloudSubnetID string `json:"cloud_subnet_id"`
}

// AwsRouteTableExtension defines aws route table extensional info.
type AwsRouteTableExtension struct {
	Main bool `json:"main"`
}

// AwsRouteTableAsst defines aws route table association info.
type AwsRouteTableAsst struct {
	AssociationState string  `json:"association_state,omitempty"`
	CloudGatewayID   *string `json:"cloud_gateway_id,omitempty"`
	CloudSubnetID    *string `json:"cloud_subnet_id,omitempty"`
}

// AzureRouteTableExtension defines azure route table extensional info.
type AzureRouteTableExtension struct {
	ResourceGroupName   string `json:"resource_group_name"`
	CloudSubscriptionID string `json:"cloud_subscription_id"`
}

// HuaWeiRouteTableExtension defines huawei route table extensional info.
type HuaWeiRouteTableExtension struct {
	Default  bool   `json:"default"`
	TenantID string `json:"tenant_id"`
}

// TCloudRouteTable defines tencent cloud route table.
type TCloudRouteTable RouteTable[TCloudRouteTableExtension]

// GetID ...
func (routeTable TCloudRouteTable) GetID() string {
	return routeTable.BaseRouteTable.ID
}

// GetCloudID ...
func (routeTable TCloudRouteTable) GetCloudID() string {
	return routeTable.BaseRouteTable.CloudID
}

// AwsRouteTable defines aws route table.
type AwsRouteTable RouteTable[AwsRouteTableExtension]

// GetID ...
func (routeTable AwsRouteTable) GetID() string {
	return routeTable.BaseRouteTable.ID
}

// GetCloudID ...
func (routeTable AwsRouteTable) GetCloudID() string {
	return routeTable.BaseRouteTable.CloudID
}

// GcpRouteTable defines gcp route table.
type GcpRouteTable BaseRouteTable

// AzureRouteTable defines azure route table.
type AzureRouteTable RouteTable[AzureRouteTableExtension]

// GetID ...
func (routeTable AzureRouteTable) GetID() string {
	return routeTable.BaseRouteTable.ID
}

// GetCloudID ...
func (routeTable AzureRouteTable) GetCloudID() string {
	return routeTable.BaseRouteTable.CloudID
}

// HuaWeiRouteTable defines huawei route table.
type HuaWeiRouteTable RouteTable[HuaWeiRouteTableExtension]

// GetID ...
func (routeTable HuaWeiRouteTable) GetID() string {
	return routeTable.BaseRouteTable.ID
}

// GetCloudID ...
func (routeTable HuaWeiRouteTable) GetCloudID() string {
	return routeTable.BaseRouteTable.CloudID
}
