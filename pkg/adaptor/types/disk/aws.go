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

	"github.com/aws/aws-sdk-go/service/ec2"
)

// AwsDiskCreateOption AWS 创建云盘参数
// reference: https://docs.amazonaws.cn/AWSEC2/latest/APIReference/API_CreateVolume.html
type AwsDiskCreateOption struct {
	Region   string
	Zone     *string
	DiskType *string
	DiskSize *int64
}

// ToCreateVolumeInput 转换成接口需要的 CreateVolumeInput 结构
func (opt *AwsDiskCreateOption) ToCreateVolumeInput() (*ec2.CreateVolumeInput, error) {
	return &ec2.CreateVolumeInput{
		AvailabilityZone: opt.Zone,
		Size:             opt.DiskSize,
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
