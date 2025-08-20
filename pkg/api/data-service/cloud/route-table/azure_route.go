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
	routetable "hcm/pkg/api/core/cloud/route-table"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------

// AzureRouteBatchCreateReq defines batch create azure route request.
type AzureRouteBatchCreateReq struct {
	AzureRoutes []AzureRouteCreateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate AzureRouteBatchCreateReq.
func (c *AzureRouteBatchCreateReq) Validate() error {
	return validator.Validate.Struct(c)
}

// AzureRouteCreateReq defines create azure route request.
type AzureRouteCreateReq struct {
	CloudID           string  `json:"cloud_id" validate:"required"`
	CloudRouteTableID string  `json:"cloud_route_table_id" validate:"required"`
	Name              string  `json:"name" validate:"required"`
	AddressPrefix     string  `json:"address_prefix" validate:"required"`
	NextHopType       string  `json:"next_hop_type" validate:"required"`
	NextHopIPAddress  *string `json:"next_hop_ip_address,omitempty" validate:"omitempty"`
	ProvisioningState string  `json:"provisioning_state" validate:"required"`
}

// -------------------------- Update --------------------------

// AzureRouteBatchUpdateReq defines batch update azure route request.
type AzureRouteBatchUpdateReq struct {
	AzureRoutes []AzureRouteUpdateReq `json:"routes" validate:"min=1,max=100"`
}

// Validate AzureRouteBatchUpdateReq.
func (u *AzureRouteBatchUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// AzureRouteUpdateReq defines update azure route request.
type AzureRouteUpdateReq struct {
	ID                   string `json:"id" validate:"required"`
	AzureRouteUpdateInfo `json:",inline" validate:"omitempty"`
}

// AzureRouteUpdateInfo defines update azure route request base info.
type AzureRouteUpdateInfo struct {
	AddressPrefix     string  `json:"address_prefix" validate:"omitempty"`
	NextHopType       string  `json:"next_hop_type" validate:"omitempty"`
	NextHopIPAddress  *string `json:"next_hop_ip_address,omitempty" validate:"omitempty"`
	ProvisioningState string  `json:"provisioning_state" validate:"omitempty"`
}

// -------------------------- List --------------------------

// AzureRouteListReq defines list azure route request.
type AzureRouteListReq struct {
	*core.ListReq `json:",inline"`
	RouteTableID  string `json:"route_table_id"`
}

// Validate ...
func (r AzureRouteListReq) Validate() error {
	if r.ListReq == nil {
		return errf.New(errf.InvalidParameter, "list request is required")
	}

	if err := r.ListReq.Validate(); err != nil {
		return err
	}

	return nil
}

// AzureRouteListResp defines list azure route response.
type AzureRouteListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AzureRouteListResult `json:"data"`
}

// AzureRouteListResult defines list azure route result.
type AzureRouteListResult struct {
	Count   uint64                  `json:"count"`
	Details []routetable.AzureRoute `json:"details"`
}
