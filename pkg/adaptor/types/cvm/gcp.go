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
	"fmt"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// GcpListOption defines options to list gcp cvm instances.
type GcpListOption struct {
	Region   string        `json:"region" validate:"required"`
	Zone     string        `json:"zone" validate:"required"`
	CloudIDs []string      `json:"cloud_ids" validate:"omitempty"`
	Names    []string      `json:"names" validate:"omitempty"`
	Page     *core.GcpPage `json:"page" validate:"omitempty"`
}

// Validate gcp cvm list option.
func (opt GcpListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if len(opt.CloudIDs) != 0 && opt.Page != nil {
		return fmt.Errorf("list by cloud_ids not support page")
	}

	if len(opt.CloudIDs) > core.GcpQueryLimit {
		return fmt.Errorf("cloud_ids should <= %d", core.GcpQueryLimit)
	}

	if len(opt.Names) != 0 && opt.Page != nil {
		return fmt.Errorf("list by names not support page")
	}

	if len(opt.Names) > core.GcpQueryLimit {
		return fmt.Errorf("nnames should <= %d", core.GcpQueryLimit)
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// -------------------------- Delete --------------------------

// GcpDeleteOption defines options to operation cvm instances.
type GcpDeleteOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Stop --------------------------

// GcpStopOption defines options to operation cvm instances.
type GcpStopOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpStopOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Start --------------------------

// GcpStartOption defines options to operation cvm instances.
type GcpStartOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpStartOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reset --------------------------

// GcpResetOption defines options to operation cvm instances.
type GcpResetOption struct {
	Zone string `json:"zone" validate:"required"`
	Name string `json:"name" validate:"required"`
}

// Validate cvm operation option.
func (opt GcpResetOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create --------------------------

// GcpCreateOption defines options to create gcp cvm instances.
type GcpCreateOption struct {
	NamePrefix         string `json:"name_prefix" validate:"required"`
	Zone               string `json:"zone" validate:"required"`
	InstanceType       string `json:"instance_type" validate:"required"`
	CloudImageSelfLink string `json:"cloud_image_self_link" validate:"required"`
	Password           string `json:"password" validate:"required"`
	RequiredCount      int64  `json:"required_count" validate:"required"`
	// RequestID 唯一标识支持生产请求
	RequestID           string `json:"request_id" validate:"omitempty"`
	CloudVpcSelfLink    string `json:"cloud_vpc_self_link" validate:"required"`
	CloudSubnetSelfLink string `json:"cloud_subnet_self_link" validate:"required"`
	Description         string `json:"description" validate:"omitempty"`
	// ImageProjectType 用于判断是 linux 还是 windows 机器。
	ImageProjectType GcpImageProjectType `json:"image_project_type" validate:"required"`
	SystemDisk       *GcpOsDisk          `json:"system_disk" validate:"required"`
	DataDisk         []GcpDataDisk       `json:"data_disk" validate:"omitempty"`
}

// Validate gcp cvm operation option.
func (opt GcpCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// GcpOsDisk gcp os disk.
type GcpOsDisk struct {
	DiskType GcpDiskType `json:"disk_type" validate:"required"`
	SizeGb   int64       `json:"size_gb" validate:"required"`
}

// GcpDataDisk gcp disk.
type GcpDataDisk struct {
	DiskType   GcpDiskType `json:"disk_type" validate:"required"`
	SizeGb     int64       `json:"size_gb" validate:"required"`
	Mode       GcpDiskMode `json:"mode" validate:"required"`
	AutoDelete bool        `json:"auto_delete" validate:"omitempty"`
}

// GcpDiskType ...
type GcpDiskType string

const (
	// PdStandard 标准永久性磁盘
	PdStandard GcpDiskType = "pd-standard"
	// PdBalanced 均衡永久性磁盘
	PdBalanced GcpDiskType = "pd-balanced"
	// PdSsd 性能(SSD)永久性磁盘
	PdSsd GcpDiskType = "pd-ssd"
	// PdExtreme 极端永久性磁盘
	PdExtreme GcpDiskType = "pd-extreme"
)

// GcpDiskMode gcp disk model.
type GcpDiskMode string

const (
	ReadOnly  GcpDiskMode = "READ_ONLY"
	ReadWrite GcpDiskMode = "READ_WRITE"
)

// GcpImageProjectType gcp image project type.
type GcpImageProjectType string

// StartupScript return image project type's start up script.
func (typ *GcpImageProjectType) StartupScript(passwd string) (string, error) {
	switch *typ {
	case Windows:
		return fmt.Sprintf(`<script>
net user administrator %s
</script>`, passwd), nil
	case Linux:
		return fmt.Sprintf(`#!/bin/bash
echo key:%s|chpasswd
sed -i 's/PasswordAuthentication/\# PasswordAuthentication/g' /etc/ssh/sshd_config
sudo sed -i 's/PermitRootLogin no/PermitRootLogin prohibit-password/g' /etc/ssh/sshd_config
sed -i '20 a PasswordAuthentication yes' /etc/ssh/sshd_config
systemctl restart sshd`, passwd), nil
	default:
		return "", fmt.Errorf("unknown %s image project type", &typ)
	}
}

const (
	Windows GcpImageProjectType = "windows"
	Linux   GcpImageProjectType = "linux"
)
