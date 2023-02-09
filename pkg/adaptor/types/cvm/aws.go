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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// AwsListOption defines options to list aws cvm instances.
type AwsListOption struct {
	Region   string        `json:"region" validate:"required"`
	CloudIDs []string      `json:"cloud_ids" validate:"omitempty"`
	Page     *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate aws cvm list option.
func (opt AwsListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Delete --------------------------

// AwsDeleteOption defines options to operation aws cvm instances.
type AwsDeleteOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate aws cvm operation option.
func (opt AwsDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Start --------------------------

// AwsStartOption defines options to operation aws cvm instances.
type AwsStartOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate aws cvm operation option.
func (opt AwsStartOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Stop --------------------------

// AwsStopOption defines options to operation aws cvm instances.
type AwsStopOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	// Forces the instances to stop. The instances do not have an opportunity to
	// flush file system caches or file system metadata. If you use this option,
	// you must perform file system check and repair procedures. This option is
	// not recommended for Windows instances.
	//
	// Default: false
	Force bool `json:"force" validate:"required"`
	// Hibernates the instance if the instance was enabled for hibernation at launch.
	// If the instance cannot hibernate successfully, a normal shutdown occurs.
	// For more information, see Hibernate your instance
	// (https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/Hibernate.html)
	// in the Amazon EC2 User Guide.
	//
	// Default: false
	Hibernate bool `json:"hibernate" validate:"required"`
}

// Validate aws cvm operation option.
func (opt AwsStopOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reboot --------------------------

// AwsRebootOption defines options to operation aws cvm instances.
type AwsRebootOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate aws cvm operation option.
func (opt AwsRebootOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create --------------------------

// AwsCreateOption defines options to create aws cvm instances.
type AwsCreateOption struct {
	Region             string                  `json:"region" validate:"required"`
	Name               *string                 `json:"name" validate:"omitempty"`
	Zone               *string                 `json:"zone" validate:"omitempty"`
	InstanceType       string                  `json:"instance_type" validate:"required"`
	ImageID            string                  `json:"image_id" validate:"required"`
	Password           string                  `json:"password" validate:"required"`
	RequiredCount      int64                   `json:"required_count" validate:"required"`
	SecurityGroupIDs   []string                `json:"security_group_i_ds" validate:"omitempty"`
	ClientToken        *string                 `json:"client_token" validate:"omitempty"`
	SubnetID           string                  `json:"subnet_id" validate:"required"`
	BlockDeviceMapping []AwsBlockDeviceMapping `json:"block_device_mapping" validate:"required"`
}

// AwsBlockDeviceMapping ...
type AwsBlockDeviceMapping struct {
	// DeviceName 设备名称，如 /dev/sdh 或 xvdh。
	DeviceName *string `json:"device_name" validate:"omitempty"`
	Ebs        *AwsEbs `json:"ebs" validate:"required"`
}

// AwsEbs ...
type AwsEbs struct {
	VolumeSizeGB int64  `json:"volume_size_gb" validate:"required"`
	VolumeType   string `json:"volume_type" validate:"required"`
	// Iops volumeType为 io1 和 io2 时必填
	Iops *int64 `json:"iops" validate:"omitempty"`
}

// Validate aws cvm operation option.
func (opt AwsCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}
