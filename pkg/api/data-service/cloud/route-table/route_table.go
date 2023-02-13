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
	routetable "hcm/pkg/api/core/cloud/route-table"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------

// RouteTableBatchCreateReq defines batch create route table request.
type RouteTableBatchCreateReq[T RouteTableCreateExtension] struct {
	RouteTables []RouteTableCreateReq[T] `json:"route_tables" validate:"min=1,max=100"`
}

// RouteTableCreateReq defines create route table request.
type RouteTableCreateReq[T RouteTableCreateExtension] struct {
	AccountID  string  `json:"account_id" validate:"required"`
	CloudID    string  `json:"cloud_id" validate:"required"`
	Name       *string `json:"name,omitempty" validate:"omitempty"`
	Region     string  `json:"region" validate:"required"`
	CloudVpcID string  `json:"cloud_vpc_id" validate:"omitempty"`
	Memo       *string `json:"memo,omitempty" validate:"omitempty"`
	Extension  *T      `json:"extension" validate:"required"`
}

// RouteTableCreateExtension defines create route table extensional info.
type RouteTableCreateExtension interface {
	TCloudRouteTableCreateExt | AwsRouteTableCreateExt | AzureRouteTableCreateExt | HuaWeiRouteTableCreateExt
}

// TCloudRouteTableCreateExt defines create tencent cloud route table extensional info.
type TCloudRouteTableCreateExt struct {
	Main bool `json:"main"`
}

// AwsRouteTableCreateExt defines create aws route table extensional info.
type AwsRouteTableCreateExt struct {
	Main bool `json:"main"`
}

// AzureRouteTableCreateExt defines azure route table extensional info.
type AzureRouteTableCreateExt struct {
	CloudSubscriptionID string `json:"cloud_subscription_id"`
	ResourceGroup       string `json:"resource_group"`
}

// HuaWeiRouteTableCreateExt defines huawei route table extensional info.
type HuaWeiRouteTableCreateExt struct {
	Default  bool   `json:"default"`
	TenantID string `json:"tenant_id"`
}

// Validate RouteTableBatchCreateReq.
func (c *RouteTableBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(c)
}

// -------------------------- Update --------------------------

// RouteTableUpdateBaseInfo defines update route table request base info.
type RouteTableUpdateBaseInfo struct {
	Name    *string `json:"name,omitempty" validate:"omitempty"`
	Memo    *string `json:"memo,omitempty" validate:"omitempty"`
	BkBizID int64   `json:"bk_biz_id,omitempty" validate:"omitempty"`
}

// RouteTableBaseInfoBatchUpdateReq defines batch update route table base info request.
type RouteTableBaseInfoBatchUpdateReq struct {
	RouteTables []RouteTableBaseInfoUpdateReq `json:"route_tables" validate:"required"`
}

// Validate RouteTableBaseInfoBatchUpdateReq.
func (u *RouteTableBaseInfoBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// RouteTableBaseInfoUpdateReq defines update route table base info request.
type RouteTableBaseInfoUpdateReq struct {
	IDs  []string                  `json:"id" validate:"required"`
	Data *RouteTableUpdateBaseInfo `json:"data" validate:"required"`
}

// -------------------------- Get --------------------------

// RouteTableGetResp defines get route table response.
type RouteTableGetResp[T routetable.RouteTableExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *routetable.RouteTable[T] `json:"data"`
}

// -------------------------- List --------------------------

// RouteTableListResp defines list route table response.
type RouteTableListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *RouteTableListResult `json:"data"`
}

// RouteTableListResult defines list route table result.
type RouteTableListResult struct {
	Count   uint64                      `json:"count"`
	Details []routetable.BaseRouteTable `json:"details"`
}

// RouteTableSubnetsCountResp defines count route tables' subnets response.
type RouteTableSubnetsCountResp struct {
	rest.BaseResp `json:",inline"`
	Data          []RouteTableSubnetsCountResult `json:"data"`
}

// RouteTableSubnetsCountResult defines count route tables' subnets result.
type RouteTableSubnetsCountResult struct {
	Count uint64 `json:"count"`
	ID    string `json:"id"`
}
