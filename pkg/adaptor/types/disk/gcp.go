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
	"fmt"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"

	"google.golang.org/api/compute/v1"
)

// GcpDiskCreateOption ...
type GcpDiskCreateOption struct {
	DiskName  string  `json:"disk_name" validate:"required"`
	Region    string  `json:"region" validate:"required"`
	Zone      string  `json:"zone" validate:"required"`
	DiskType  string  `json:"disk_type" validate:"required"`
	DiskSize  int64   `json:"disk_size" validate:"required"`
	DiskCount *uint64 `json:"disk_count" validate:"required"`
}

// Validate ...
func (opt *GcpDiskCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToCreateDiskRequest 转换成接口需要的 *compute.Disk 结构
func (opt *GcpDiskCreateOption) ToCreateDiskRequest(cloudProjectID string) (*compute.Disk, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	return &compute.Disk{
		Region: opt.Region,
		Name:   opt.DiskName,
		Type: fmt.Sprintf("projects/%s/zones/%s/diskTypes/%s", cloudProjectID, opt.Zone,
			opt.DiskType),
		SizeGb: opt.DiskSize,
	}, nil
}

// GcpDiskListOption define gcp disk list option.
type GcpDiskListOption struct {
	Zone      string        `json:"zone" validate:"required"`
	CloudIDs  []string      `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string      `json:"self_links" validate:"omitempty"`
	Names     []string      `json:"names" validate:"omitempty"`
	Page      *core.GcpPage `json:"page" validate:"omitempty"`
}

// Validate gcp disk list option.
func (opt GcpDiskListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}
	if len(opt.CloudIDs) != 0 && opt.Page != nil {
		return fmt.Errorf("list by cloud_ids not support page")
	}

	if len(opt.CloudIDs) > core.GcpQueryLimit {
		return fmt.Errorf("cloud_ids should <= %d", core.GcpQueryLimit)
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// GcpDiskDeleteOption ...
type GcpDiskDeleteOption struct {
	Zone     string `json:"zone" validate:"required"`
	DiskName string `json:"disk_name" validate:"required"`
}

// Validate ...
func (opt *GcpDiskDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpDiskAttachOption ...
type GcpDiskAttachOption struct {
	Zone       string `json:"zone" validate:"required"`
	CvmName    string `json:"cvm_name" validate:"required"`
	DiskName   string `json:"disk_name" validate:"required"`
	DeviceName string `json:"device_name"`
}

// Validate ...
func (opt *GcpDiskAttachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAttachDiskRequest ...
func (opt *GcpDiskAttachOption) ToAttachDiskRequest() (*compute.AttachedDisk, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &compute.AttachedDisk{}
	if opt.DeviceName != "" {
		req.DeviceName = opt.DeviceName
	}

	return req, nil
}

// GcpDiskDetachOption ...
type GcpDiskDetachOption struct {
	Zone       string `json:"zone" validate:"required"`
	CvmName    string `json:"cvm_name" validate:"required"`
	DiskName   string `json:"disk_name" validate:"required"`
	DeviceName string `json:"device_name" validate:"required"`
}

// Validate ...
func (opt *GcpDiskDetachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpDisk for compute Disk
type GcpDisk struct {
	*compute.Disk
}

// GetCloudID ...
func (disk GcpDisk) GetCloudID() string {
	return fmt.Sprint(disk.Id)
}
