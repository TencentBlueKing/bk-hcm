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

// HuaWeiListOption defines options to list huawei cvm instances.
type HuaWeiListOption struct {
	Region   string                 `json:"region" validate:"required"`
	CloudIDs []string               `json:"cloud_ids" validate:"omitempty"`
	Page     *core.HuaWeiOffsetPage `json:"page" validate:"omitempty"`
}

// Validate huawei cvm list option.
func (opt HuaWeiListOption) Validate() error {
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

// HuaWeiDeleteOption defines options to operation huawei cvm instances.
type HuaWeiDeleteOption struct {
	Region         string   `json:"region" validate:"required"`
	CloudIDs       []string `json:"cloud_ids" validate:"required"`
	DeletePublicIP bool     `json:"delete_public_ip" validate:"required"`
	DeleteVolume   bool     `json:"delete_volume" validate:"required"`
}

// Validate huawei cvm operation option.
func (opt HuaWeiDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Start --------------------------

// HuaWeiStartOption defines options to operation cvm instances.
type HuaWeiStartOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate cvm operation option.
func (opt HuaWeiStartOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Stop --------------------------

// HuaWeiStopOption defines options to operation cvm instances.
type HuaWeiStopOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	Force    bool     `json:"force" validate:"required"`
}

// Validate cvm operation option.
func (opt HuaWeiStopOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reboot --------------------------

// HuaWeiRebootOption defines options to operation cvm instances.
type HuaWeiRebootOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	Force    bool     `json:"force" validate:"required"`
}

// Validate cvm operation option.
func (opt HuaWeiRebootOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Reset PWD --------------------------

// HuaWeiResetPwdOption defines options to operation cvm instances.
type HuaWeiResetPwdOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	Password string   `json:"password" validate:"required"`
}

// Validate cvm operation option.
func (opt HuaWeiResetPwdOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create --------------------------

// HuaWeiCreateOption defines options to create aws cvm instances.
type HuaWeiCreateOption struct {
	Region           string                   `json:"region" validate:"required"`
	Name             string                   `json:"name" validate:"required"`
	Zone             string                   `json:"zone" validate:"required"`
	InstanceType     string                   `json:"instance_type" validate:"required"`
	ImageID          string                   `json:"image_id" validate:"required"`
	Password         string                   `json:"password" validate:"required"`
	RequiredCount    int32                    `json:"required_count" validate:"required"`
	SecurityGroupIDs []string                 `json:"security_group_i_ds" validate:"omitempty"`
	ClientToken      *string                  `json:"client_token" validate:"omitempty"`
	VpcID            string                   `json:"vpc_id" validate:"required"`
	Nics             []HuaWeiNetworkInterface `json:"nics" validate:"required"`
	Description      *string                  `json:"description" validate:"omitempty"`
	RootVolume       *HuaWeiVolume            `json:"root_volume" validate:"required"`
	DataVolume       []HuaWeiVolume           `json:"data_volume" validate:"omitempty"`
}

// Validate aws cvm operation option.
func (opt HuaWeiCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// HuaWeiVolume ...
type HuaWeiVolume struct {
	VolumeType HuaWeiVolumeType `json:"volume_type" validate:"required"`
	SizeGB     int64            `json:"size_gb" validate:"omitempty"`
}

// HuaWeiVolumeType 系统盘对应的磁盘类型，需要与系统所提供的磁盘类型相匹配。
type HuaWeiVolumeType string

const (
	// Sata 普通IO云硬盘
	Sata HuaWeiVolumeType = "SATA"
	// Sas 高IO云硬盘
	Sas HuaWeiVolumeType = "SAS"
	// Gpssd 通用型SSD云硬盘
	Gpssd HuaWeiVolumeType = "GPSSD"
	// Ssd 超高IO云硬盘
	Ssd HuaWeiVolumeType = "SSD"
	// Essd 极速IO云硬盘
	Essd HuaWeiVolumeType = "ESSD"
)

// HuaWeiNetworkInterface ...
type HuaWeiNetworkInterface struct {
	SubnetID   string  `json:"subnet_id" validate:"required"`
	IPAddress  *string `json:"ip_address" validate:"omitempty"`
	IPv6Enable *bool   `json:"i_pv_6_enable" validate:"omitempty"`
}
