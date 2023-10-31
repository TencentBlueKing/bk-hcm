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

import "hcm/pkg/criteria/validator"

// RouteTableUpdateReq defines update route table request.
type RouteTableUpdateReq struct {
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate RouteTableUpdateReq.
func (u *RouteTableUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Sync --------------------------

// TCloudRouteTableSyncReq defines sync route table request.
type TCloudRouteTableSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids,omitempty"`
}

// Validate validate sync route table request.
func (r *TCloudRouteTableSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HuaWeiRouteTableSyncReq defines sync route table request.
type HuaWeiRouteTableSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids,omitempty"`
}

// Validate validate sync route table request.
func (r *HuaWeiRouteTableSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureRouteTableSyncReq defines sync route table request.
type AzureRouteTableSyncReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids,omitempty"`
}

// Validate validate sync route table request.
func (r *AzureRouteTableSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AwsRouteTableSyncReq defines sync route table request.
type AwsRouteTableSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids,omitempty"`
}

// Validate validate sync route table request.
func (r *AwsRouteTableSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}
