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

package hcservice

import "hcm/pkg/criteria/validator"

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq struct {
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate VpcUpdateReq.
func (u *VpcUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- Sync --------------------------

// TCloudResourceSyncReq defines sync resource request.
type TCloudResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *TCloudResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HuaWeiResourceSyncReq defines sync resource request.
type HuaWeiResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	VpcID     string   `json:"vpc_id" validate:"omitempty"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *HuaWeiResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// GcpResourceSyncReq defines sync resource request.
type GcpResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *GcpResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AzureResourceSyncReq defines sync resource request.
type AzureResourceSyncReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudVpcID        string   `json:"cloud_vpc_id" validate:"omitempty"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *AzureResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// AwsResourceSyncReq defines sync resource request.
type AwsResourceSyncReq struct {
	AccountID string   `json:"account_id" validate:"required"`
	Region    string   `json:"region" validate:"required"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate validate sync vpc request.
func (r *AwsResourceSyncReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ResourceSyncResult defines sync vpc result.
type ResourceSyncResult struct {
	TaskID string `json:"task_id"`
}
