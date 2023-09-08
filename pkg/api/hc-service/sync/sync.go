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

// Package sync ...
package sync

import "hcm/pkg/criteria/validator"

// TCloudRegionSyncReq tcloud sync request
type TCloudRegionSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate tcloud sync request.
func (req *TCloudRegionSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudSyncReq tcloud sync request
type TCloudSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate tcloud sync request.
func (req *TCloudSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsRegionSyncReq aws sync request
type AwsRegionSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate aws sync request.
func (req *AwsRegionSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsSyncReq aws sync request
type AwsSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate aws sync request.
func (req *AwsSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiRegionSyncReq huawei sync request
type HuaWeiRegionSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate huawei sync request.
func (req *HuaWeiRegionSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiSyncReq huawei sync request
type HuaWeiSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate huawei sync request.
func (req *HuaWeiSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiSubnetSyncReq huawei subnet sync request
type HuaWeiSubnetSyncReq struct {
	AccountID  string `json:"account_id" validate:"required"`
	CloudVpcID string `json:"cloud_vpc_id" validate:"required"`
	Region     string `json:"region" validate:"required"`
}

// Validate huawei sync request.
func (req *HuaWeiSubnetSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpCvmSyncReq gcp sync request
type GcpCvmSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
}

// Validate gcp sync request.
func (req *GcpCvmSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpSyncReq gcp sync request
type GcpSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate gcp sync request.
func (req *GcpSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpGlobalRegionResSyncReq gcp sync request
type GcpGlobalRegionResSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate gcp sync request.
func (req *GcpGlobalRegionResSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpDiskSyncReq gcp disk sync request
type GcpDiskSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
}

// Validate gcp disk sync request.
func (req *GcpDiskSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpRouteSyncReq gcp route sync request
type GcpRouteSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Zone      string `json:"zone" validate:"required"`
}

// Validate gcp route sync request.
func (req *GcpRouteSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// GcpFireWallSyncReq gcp firewall sync request
type GcpFireWallSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate gcp firewall sync request.
func (req *GcpFireWallSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureResourceGroupSyncReq azure sync request
type AzureResourceGroupSyncReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate azure sync request.
func (req *AzureResourceGroupSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureSyncReq azure sync request
type AzureSyncReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate azure sync request.
func (req *AzureSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureSubnetSyncReq azure sync request
type AzureSubnetSyncReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	CloudVpcID        string `json:"cloud_vpc_id" validate:"required"`
}

// Validate azure sync request.
func (req *AzureSubnetSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureImageReq azure image sync request
type AzureImageReq struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
}

// Validate azure sync request.
func (req *AzureImageReq) Validate() error {
	return validator.Validate.Struct(req)
}
