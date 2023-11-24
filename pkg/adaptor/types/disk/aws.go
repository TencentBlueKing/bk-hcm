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

package disk

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// AwsDiskCreateOption AWS 创建云盘参数
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateVolume.html
type AwsDiskCreateOption struct {
	Region    string  `json:"region" validate:"required"`
	Zone      string  `json:"zone" validate:"required"`
	DiskType  *string `json:"disk_type"`
	DiskSize  int64   `json:"disk_size" validate:"required"`
	DiskCount *uint64 `json:"disk_count" validate:"required"`
}

// Validate ...
func (opt *AwsDiskCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToCreateVolumeInput 转换成接口需要的 CreateVolumeInput 结构
func (opt *AwsDiskCreateOption) ToCreateVolumeInput() (*ec2.CreateVolumeInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &ec2.CreateVolumeInput{
		AvailabilityZone: aws.String(opt.Zone),
		Size:             aws.Int64(opt.DiskSize),
		VolumeType:       opt.DiskType,
	}, nil
}

// AwsDiskListOption define aws disk list option.
type AwsDiskListOption struct {
	Region   string        `json:"region" validate:"required"`
	CloudIDs []string      `json:"cloud_ids" validate:"omitempty"`
	Page     *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate disk list option.
func (opt AwsDiskListOption) Validate() error {
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

// AwsDiskDeleteOption ...
type AwsDiskDeleteOption struct {
	Region  string `json:"region" validate:"required"`
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate ...
func (opt *AwsDiskDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDeleteVolumeInput ...
func (opt *AwsDiskDeleteOption) ToDeleteVolumeInput() (*ec2.DeleteVolumeInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &ec2.DeleteVolumeInput{VolumeId: aws.String(opt.CloudID)}, nil
}

// AwsDiskAttachOption ...
type AwsDiskAttachOption struct {
	Region string `json:"region" validate:"required"`
	// DeviceName Device 参数，/dev/sdx, xvdh
	DeviceName  string `json:"device_name" validate:"required"`
	CloudCvmID  string `json:"cloud_cvm_id" validate:"required"`
	CloudDiskID string `json:"cloud_disk_id" validate:"required"`
}

// Validate ...
func (opt *AwsDiskAttachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAttachVolumeInput ...
func (opt *AwsDiskAttachOption) ToAttachVolumeInput() (*ec2.AttachVolumeInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &ec2.AttachVolumeInput{
		Device:     aws.String(opt.DeviceName),
		InstanceId: aws.String(opt.CloudCvmID),
		VolumeId:   aws.String(opt.CloudDiskID),
	}, nil
}

// AwsDiskDetachOption ...
type AwsDiskDetachOption struct {
	Region      string `json:"region" validate:"required"`
	CloudCvmID  string `json:"cloud_cvm_id" validate:"required"`
	CloudDiskID string `json:"cloud_disk_id" validate:"required"`
}

// Validate ...
func (opt *AwsDiskDetachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDetachVolumeInput ...
func (opt *AwsDiskDetachOption) ToDetachVolumeInput() (*ec2.DetachVolumeInput, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &ec2.DetachVolumeInput{InstanceId: aws.String(opt.CloudCvmID), VolumeId: aws.String(opt.CloudDiskID)}, nil
}

// AwsDisk for ec2 Volume
type AwsDisk struct {
	*ec2.Volume
	Boot *bool
}

// GetCloudID ...
func (disk AwsDisk) GetCloudID() string {
	return converter.PtrToVal(disk.VolumeId)
}
