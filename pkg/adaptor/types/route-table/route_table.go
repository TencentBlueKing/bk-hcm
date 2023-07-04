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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/errf"
)

// -------------------------- Update --------------------------

// RouteTableUpdateOption defines update route table options.
type RouteTableUpdateOption struct {
	ResourceID string                    `json:"resource_id"`
	Data       *BaseRouteTableUpdateData `json:"data"`
}

// BaseRouteTableUpdateData defines the basic update route table instance data.
type BaseRouteTableUpdateData struct {
	Memo *string `json:"memo"`
}

// Validate BaseRouteTableUpdateData.
func (r BaseRouteTableUpdateData) Validate() error {
	if r.Memo == nil {
		return errf.New(errf.InvalidParameter, "memo is required")
	}
	return nil
}

// TCloudRouteTableUpdateOption defines tencent cloud update route table options.
type TCloudRouteTableUpdateOption struct{}

// Validate TCloudRouteTableUpdateOption.
func (r TCloudRouteTableUpdateOption) Validate() error {
	return nil
}

// AwsRouteTableUpdateOption defines aws update route table options.
type AwsRouteTableUpdateOption struct{}

// Validate AwsRouteTableUpdateOption.
func (r AwsRouteTableUpdateOption) Validate() error {
	return nil
}

// AzureRouteTableUpdateOption defines azure update route table options.
type AzureRouteTableUpdateOption struct{}

// Validate AzureRouteTableUpdateOption.
func (r AzureRouteTableUpdateOption) Validate() error {
	return nil
}

// HuaWeiRouteTableUpdateOption defines huawei update route table options.
type HuaWeiRouteTableUpdateOption struct {
	RouteTableUpdateOption `json:",inline"`
	Region                 string `json:"region"`
}

// Validate HuaWeiRouteTableUpdateOption.
func (r HuaWeiRouteTableUpdateOption) Validate() error {
	if len(r.Region) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if len(r.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource id is required")
	}

	if r.Data == nil {
		return errf.New(errf.InvalidParameter, "update data is required")
	}

	if err := r.Data.Validate(); err != nil {
		return err
	}
	return nil
}

// -------------------------- Find --------------------------

// TCloudRouteTableListResult defines tencent cloud list route table result.
type TCloudRouteTableListResult struct {
	Count   *uint64            `json:"count,omitempty"`
	Details []TCloudRouteTable `json:"details"`
}

// AwsRouteTableListOption defines aws list route table options.
type AwsRouteTableListOption struct {
	*core.AwsListOption `json:",inline"`
	SubnetIDs           []string `json:"subnet_id,omitempty"`
}

const AwsRouteTableListLimit = 100

// Validate huawei list option.
func (r AwsRouteTableListOption) Validate() error {
	if r.AwsListOption == nil {
		return errf.New(errf.InvalidParameter, "list option is required")
	}

	if err := r.AwsListOption.Validate(); err != nil {
		return err
	}

	if len(r.SubnetIDs) > AwsRouteTableListLimit || len(r.CloudIDs) > AwsRouteTableListLimit {
		return errf.Newf(errf.InvalidParameter, "ids can not exceeds max limit %d", AwsRouteTableListLimit)
	}

	if r.Page != nil && r.Page.MaxResults != nil && *r.Page.MaxResults > AwsRouteTableListLimit {
		return errf.Newf(errf.InvalidParameter, "max result can not exceeds max limit %d", AwsRouteTableListLimit)
	}
	return nil
}

// AwsRouteTableListResult defines aws list route table result.
type AwsRouteTableListResult struct {
	NextToken *string         `json:"next_token,omitempty"`
	Details   []AwsRouteTable `json:"details"`
}

// AzureRouteTableListResult defines azure list route table result.
type AzureRouteTableListResult struct {
	Details []AzureRouteTable `json:"details"`
}

// AzureRouteTableGetOption defines azure get route table option.
type AzureRouteTableGetOption struct {
	ResourceGroupName string `json:"resource_group_name"`
	Name              string `json:"name"`
}

// Validate azure get option.
func (r AzureRouteTableGetOption) Validate() error {
	if len(r.ResourceGroupName) == 0 {
		return errf.New(errf.InvalidParameter, "resource group is required")
	}

	if len(r.Name) == 0 {
		return errf.New(errf.InvalidParameter, "name is required")
	}

	return nil
}

// HuaWeiRouteTableListOption defines huawei list route table options.
type HuaWeiRouteTableListOption struct {
	Region   string           `json:"region"`
	Page     *core.HuaWeiPage `json:"page,omitempty"`
	ID       string           `json:"id,omitempty"`
	VpcID    string           `json:"vpc_id,omitempty"`
	SubnetID string           `json:"subnet_id,omitempty"`
}

// Validate huawei list option.
func (r HuaWeiRouteTableListOption) Validate() error {
	if len(r.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if r.Page != nil {
		if err := r.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// HuaWeiRouteTableListResult defines huawei list route table result.
type HuaWeiRouteTableListResult struct {
	Details []HuaWeiRouteTable `json:"details"`
}

// HuaWeiRouteTableGetOption defines huawei get route table options.
type HuaWeiRouteTableGetOption struct {
	Region string `json:"region"`
	ID     string `json:"id"`
}

// Validate huawei get option.
func (r HuaWeiRouteTableGetOption) Validate() error {
	if len(r.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(r.ID) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	return nil
}

// ----------------------- Definition -----------------------

// RouteTable defines route table struct.
type RouteTable[T RouteTableExtension] struct {
	CloudID    string  `json:"cloud_id"`
	Name       string  `json:"name"`
	CloudVpcID string  `json:"cloud_vpc_id"`
	Region     string  `json:"region"`
	Memo       *string `json:"memo,omitempty"`
	Extension  *T      `json:"extension"`
}

// RouteTableExtension defines route table extensional info.
type RouteTableExtension interface {
	TCloudRouteTableExtension | AwsRouteTableExtension | AzureRouteTableExtension | HuaWeiRouteTableExtension
}

// TCloudRouteTableExtension defines tencent cloud route table extensional info.
type TCloudRouteTableExtension struct {
	Associations []TCloudRouteTableAsst `json:"associations"`
	Routes       []TCloudRoute          `json:"routes"`
	Main         bool                   `json:"main"`
}

// TCloudRouteTableAsst defines tencent cloud route table association info.
type TCloudRouteTableAsst struct {
	CloudSubnetID string `json:"cloud_subnet_id"`
	// TODO confirm if all route table id is this route table's id
	// CloudRouteTableID string `json:"cloud_route_table_id"`
}

// AwsRouteTableExtension defines aws route table extensional info.
type AwsRouteTableExtension struct {
	Associations []AwsRouteTableAsst `json:"associations"`
	Main         bool                `json:"main"`
	Routes       []AwsRoute          `json:"routes"`
}

// AwsRouteTableAsst defines aws route table association info.
type AwsRouteTableAsst struct {
	AssociationState string  `json:"association_state,omitempty"`
	CloudGatewayID   *string `json:"cloud_gateway_id,omitempty"`
	CloudSubnetID    *string `json:"cloud_subnet_id,omitempty"`
}

// AzureRouteTableExtension defines azure route table extensional info.
type AzureRouteTableExtension struct {
	CloudSubscriptionID string       `json:"cloud_subscription_id"`
	ResourceGroupName   string       `json:"resource_group_name"`
	Routes              []AzureRoute `json:"routes,omitempty"`
	CloudSubnetIDs      []string     `json:"cloud_subnet_ids,omitempty"`
}

// HuaWeiRouteTableExtension defines huawei route table extensional info.
type HuaWeiRouteTableExtension struct {
	Default        bool          `json:"default"`
	CloudSubnetIDs []string      `json:"cloud_subnet_ids,omitempty"`
	TenantID       string        `json:"tenant_id"`
	Routes         []HuaWeiRoute `json:"routes,omitempty"`
}

// TCloudRouteTable defines tencent cloud route table.
type TCloudRouteTable RouteTable[TCloudRouteTableExtension]

// GetCloudID ...
func (touteTable TCloudRouteTable) GetCloudID() string {
	return touteTable.CloudID
}

// AwsRouteTable defines aws route table.
type AwsRouteTable RouteTable[AwsRouteTableExtension]

// GetCloudID ...
func (touteTable AwsRouteTable) GetCloudID() string {
	return touteTable.CloudID
}

// AzureRouteTable defines azure route table.
type AzureRouteTable RouteTable[AzureRouteTableExtension]

// GetCloudID ...
func (touteTable AzureRouteTable) GetCloudID() string {
	return touteTable.CloudID
}

// HuaWeiRouteTable defines huawei route table.
type HuaWeiRouteTable RouteTable[HuaWeiRouteTableExtension]

// GetCloudID ...
func (touteTable HuaWeiRouteTable) GetCloudID() string {
	return touteTable.CloudID
}
