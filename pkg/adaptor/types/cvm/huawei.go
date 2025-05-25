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
	"hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/validator"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
)

// -------------------------- List --------------------------

// HuaWeiListOption defines options to list huawei cvm instances.
type HuaWeiListOption struct {
	Region   string                    `json:"region" validate:"required"`
	CloudIDs []string                  `json:"cloud_ids" validate:"omitempty"`
	Page     *core.HuaWeiCvmOffsetPage `json:"page" validate:"omitempty"`
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
	DryRun                bool                  `json:"dry_run" validate:"omitempty"`
	Region                string                `json:"region" validate:"required"`
	Name                  string                `json:"name" validate:"required"`
	Zone                  string                `json:"zone" validate:"required"`
	InstanceType          string                `json:"instance_type" validate:"required"`
	CloudImageID          string                `json:"cloud_image_id" validate:"required"`
	Password              string                `json:"password" validate:"required"`
	RequiredCount         int32                 `json:"required_count" validate:"required"`
	CloudSecurityGroupIDs []string              `json:"cloud_security_group_ids" validate:"required"`
	ClientToken           *string               `json:"client_token" validate:"omitempty"`
	CloudVpcID            string                `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID         string                `json:"cloud_subnet_id" validate:"required"`
	Description           *string               `json:"description" validate:"omitempty"`
	RootVolume            *HuaWeiVolume         `json:"root_volume" validate:"required"`
	DataVolume            []HuaWeiVolume        `json:"data_volume" validate:"omitempty"`
	InstanceCharge        *HuaWeiInstanceCharge `json:"instance_charge" validate:"required"`
	PublicIPAssigned      bool                  `json:"public_ip_assigned" validate:"omitempty"`
	Eip                   *HuaWeiEip            `json:"eip" validate:"omitempty"`
}

// Validate aws cvm operation option.
func (opt HuaWeiCreateOption) Validate() error {
	if opt.PublicIPAssigned {
		if err := opt.Eip.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(opt)
}

// HuaWeiEip define huawei eip option.
type HuaWeiEip struct {
	Type         EipType            `json:"type" validate:"required"`
	ChargingMode HuaWeiChargingMode `json:"charging_mode" validate:"required"`
	Size         int32              `json:"size" validate:"required"`
}

// Validate HuaWeiEip.
func (opt HuaWeiEip) Validate() error {
	return validator.Validate.Struct(opt)
}

// EipType eip type.
type EipType string

const (
	// BGP 全动态BGP
	BGP EipType = "5_bgp"
	// SBGP 静态BGP
	SBGP EipType = "5_sbgp"
	// YouXuanNBGP 优选BGP
	YouXuanNBGP EipType = "5_youxuanbgp"
)

// HuaWeiInstanceCharge 计费相关参数
type HuaWeiInstanceCharge struct {
	ChargingMode HuaWeiChargingMode `json:"charging_mode" validate:"required"`
	PeriodType   *PeriodType        `json:"period_type" validate:"omitempty"`
	// PeriodNum 订购周期数。
	// periodType=month（周期类型为月）时，取值为[1，9]；
	// periodType=year（周期类型为年）时，取值为[1，3]；
	PeriodNum   *int32 `json:"period_num" validate:"omitempty"`
	IsAutoRenew *bool  `json:"is_auto_renew" validate:"omitempty"`
}

// HuaWeiChargingMode 计费模式
type HuaWeiChargingMode string

// EipChargingMode eip charging mode.
func (mod *HuaWeiChargingMode) EipChargingMode() (model.PrePaidServerEipExtendParamChargingMode, error) {
	switch *mod {
	case PrePaid:
		return model.GetPrePaidServerEipExtendParamChargingModeEnum().PRE_PAID, nil
	case PostPaid:
		return model.GetPrePaidServerEipExtendParamChargingModeEnum().POST_PAID, nil
	default:
		return model.PrePaidServerEipExtendParamChargingMode{}, fmt.Errorf("unknown %s charging model", *mod)
	}
}

// ChargingMode charging mode.
func (mod *HuaWeiChargingMode) ChargingMode() (model.PrePaidServerExtendParamChargingMode, error) {
	switch *mod {
	case PrePaid:
		return model.GetPrePaidServerExtendParamChargingModeEnum().PRE_PAID, nil
	case PostPaid:
		return model.GetPrePaidServerExtendParamChargingModeEnum().POST_PAID, nil
	default:
		return model.PrePaidServerExtendParamChargingMode{}, fmt.Errorf("unknown %s charging model", *mod)
	}
}

const (
	// PrePaid 预付费，即包年包月；
	PrePaid HuaWeiChargingMode = "prePaid"
	// PostPaid 后付费，即按需付费；
	PostPaid HuaWeiChargingMode = "postPaid"
)

// PeriodType 订购周期
type PeriodType string

// PeriodType period type.
func (typ *PeriodType) PeriodType() (model.PrePaidServerExtendParamPeriodType, error) {
	switch *typ {
	case Month:
		return model.GetPrePaidServerExtendParamPeriodTypeEnum().MONTH, nil
	case Year:
		return model.GetPrePaidServerExtendParamPeriodTypeEnum().YEAR, nil
	default:
		return model.PrePaidServerExtendParamPeriodType{}, fmt.Errorf("unknown %s period type", *typ)
	}
}

const (
	// Month ...
	Month PeriodType = "month"
	// Year ...
	Year PeriodType = "year"
)

// HuaWeiVolume ...
type HuaWeiVolume struct {
	VolumeType HuaWeiVolumeType `json:"volume_type" validate:"required"`
	SizeGB     int32            `json:"size_gb" validate:"required"`
}

// HuaWeiVolumeType 系统盘对应的磁盘类型，需要与系统所提供的磁盘类型相匹配。
type HuaWeiVolumeType string

// RootVolumeType return huawei root volume type.
func (vol *HuaWeiVolumeType) RootVolumeType() (model.PrePaidServerRootVolumeVolumetype, error) {
	switch *vol {
	case Sata:
		return model.GetPrePaidServerRootVolumeVolumetypeEnum().SATA, nil
	case Sas:
		return model.GetPrePaidServerRootVolumeVolumetypeEnum().SAS, nil
	case Gpssd:
		return model.GetPrePaidServerRootVolumeVolumetypeEnum().GPSSD, nil
	case Ssd:
		return model.GetPrePaidServerRootVolumeVolumetypeEnum().SSD, nil
	case Essd:
		return model.GetPrePaidServerRootVolumeVolumetypeEnum().ESSD, nil
	default:
		return model.PrePaidServerRootVolumeVolumetype{}, fmt.Errorf("unknown %s volume type", *vol)
	}
}

// DataVolumeType return huawei data volume type.
func (vol *HuaWeiVolumeType) DataVolumeType() (model.PrePaidServerDataVolumeVolumetype, error) {
	switch *vol {
	case Sata:
		return model.GetPrePaidServerDataVolumeVolumetypeEnum().SATA, nil
	case Sas:
		return model.GetPrePaidServerDataVolumeVolumetypeEnum().SAS, nil
	case Gpssd:
		return model.GetPrePaidServerDataVolumeVolumetypeEnum().GPSSD, nil
	case Ssd:
		return model.GetPrePaidServerDataVolumeVolumetypeEnum().SSD, nil
	case Essd:
		return model.GetPrePaidServerDataVolumeVolumetypeEnum().ESSD, nil
	default:
		return model.PrePaidServerDataVolumeVolumetype{}, fmt.Errorf("unknown %s volume type", *vol)
	}
}

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
	IPv6Enable *bool   `json:"ipv6_enable" validate:"omitempty"`
}

// HuaWeiCvm for model ServerDetail
type HuaWeiCvm struct {
	model.ServerDetail
	CloudOSDiskID        string
	CLoudDataDiskIDs     []string
	PrivateIPv4Addresses []string
	PublicIPv4Addresses  []string
	PrivateIPv6Addresses []string
	PublicIPv6Addresses  []string
	CloudLaunchedTime    string
	Flavor               *cvm.HuaWeiFlavor
}

// HuaWeiCvmWrapper for model ServerDetail, with extra info for sync
type HuaWeiCvmWrapper struct {
	HuaWeiCvm
	CloudSubnetIDs []string
	SubnetIDs      []string
}

// GetCloudID ...
func (cvm HuaWeiCvm) GetCloudID() string {
	return cvm.Id
}

// GetCloudVpcID ...
func (cvm HuaWeiCvm) GetCloudVpcID() string {
	return cvm.Metadata["vpc_id"]
}
