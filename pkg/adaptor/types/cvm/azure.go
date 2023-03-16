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

package cvm

import (
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// AzureListOption defines options to list azure cvm instances.
type AzureListOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
}

// Validate azure cvm list option.
func (opt AzureListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	return nil
}

// -------------------------- Delete --------------------------

// AzureDeleteOption defines options to operation huawei cvm instances.
type AzureDeleteOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Name              string `json:"name" validate:"required"`
	Force             bool   `json:"force" validate:"required"`
}

// Validate cvm operation option.
func (opt AzureDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Start --------------------------

// AzureStartOption defines options to operation huawei cvm instances.
type AzureStartOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Name              string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt AzureStartOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reboot --------------------------

// AzureRebootOption defines options to operation huawei cvm instances.
type AzureRebootOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Name              string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt AzureRebootOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reboot --------------------------

// AzureStopOption defines options to operation huawei cvm instances.
type AzureStopOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Name              string `json:"name" validate:"required"`
	// SkipShutdown The parameter to request non-graceful VM shutdown. True value for this flag
	// indicates non-graceful shutdown whereas false indicates otherwise.
	// Default value for this flag is false if not specified
	SkipShutdown bool `json:"skip_shutdown" validate:"required"`
}

// Validate cvm operation option.
func (opt AzureStopOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create --------------------------

// AzureCreateOption defines options to create azure cvm instances.
type AzureCreateOption struct {
	ResourceGroupName    string          `json:"resource_group_name" validate:"required"`
	Region               string          `json:"region" validate:"required"`
	Name                 string          `json:"name" validate:"required"`
	Zones                []string        `json:"zones" validate:"required"`
	InstanceType         string          `json:"instance_type" validate:"required"`
	Image                *AzureImage     `json:"image" validate:"required"`
	Username             string          `json:"username" validate:"required"`
	Password             string          `json:"password" validate:"required"`
	CloudSubnetID        string          `json:"cloud_subnet_id" validate:"required"`
	CloudSecurityGroupID string          `json:"cloud_security_group_id" validate:"required"`
	OSDisk               *AzureOSDisk    `json:"os_disk" validate:"required"`
	DataDisk             []AzureDataDisk `json:"data_disk" validate:"omitempty"`
}

// AzureImage ...
type AzureImage struct {
	Offer     string `json:"offer" validate:"required"`
	Publisher string `json:"publisher" validate:"required"`
	Sku       string `json:"skus" validate:"required"`
	Version   string `json:"version" validate:"required"`
}

// Validate azure cvm operation option.
func (opt AzureCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureOSDisk azure os disk.
type AzureOSDisk struct {
	Name   string `json:"name" validate:"required"`
	SizeGB int32  `json:"size_gb" validate:"required"`
}

// AzureDataDisk azure data disk.
type AzureDataDisk struct {
	Name   string `json:"name" validate:"required"`
	SizeGB int32  `json:"size_gb" validate:"required"`
}

// AzureGetOption ...
type AzureGetOption struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	Name              string `json:"name" validate:"required"`
}

// Validate ...
func (opt *AzureGetOption) Validate() error {
	return validator.Validate.Struct(opt)
}
