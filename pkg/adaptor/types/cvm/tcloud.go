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
	"errors"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	tcvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// -------------------------- List --------------------------

// TCloudListOption defines options to list tcloud cvm instances.
type TCloudListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud cvm list option.
func (opt TCloudListOption) Validate() error {
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

// TCloudDeleteOption defines options to operation tcloud cvm instances.
type TCloudDeleteOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate tcloud cvm operation option.
func (opt TCloudDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Start --------------------------

// TCloudStartOption defines options to operation tcloud cvm instances.
type TCloudStartOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate tcloud cvm operation option.
func (opt TCloudStartOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Stop --------------------------

// TCloudStopOption defines options to operation tcloud cvm instances.
type TCloudStopOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	// 实例的关闭模式。取值范围：<br>
	// <li>SOFT_FIRST：表示在正常关闭失败后进行强制关闭<br>
	// <li>HARD：直接强制关闭<br>
	// <li>SOFT：仅软关机<br>
	// 默认取值：SOFT。
	StopType StopType `json:"stop_type" validate:"required"`

	// 按量计费实例关机收费模式。
	// 取值范围：<br>
	// <li>KEEP_CHARGING：关机继续收费<br>
	// <li>STOP_CHARGING：关机停止收费<br>默认取值：KEEP_CHARGING。
	// 该参数只针对部分按量计费云硬盘实例生效，详情参考[按量计费实例关机不收费说明](https://cloud.tencent.com/document/product/213/19918)
	StoppedMode StoppedMode `json:"stopped_mode" validate:"required"`
}

// Validate tcloud cvm operation option.
func (opt TCloudStopOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// StopType stop cvm type.
type StopType string

const (
	SoftFirst StopType = "SOFT_FIRST"
	Hard      StopType = "HARD"
	Soft      StopType = "SOFT"
)

// StoppedMode stop cvm type.
type StoppedMode string

const (
	KeepCharging StoppedMode = "KEEP_CHARGING"
	StopCharging StoppedMode = "STOP_CHARGING"
)

// -------------------------- Reboot --------------------------

// TCloudRebootOption defines options to operation tcloud cvm instances.
type TCloudRebootOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	StopType StopType `json:"stop_type" validate:"required"`
}

// Validate tcloud cvm operation option.
func (opt TCloudRebootOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- ResetPwd --------------------------

// TCloudResetPwdOption defines options to restart pwd tcloud cvm instances.
type TCloudResetPwdOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
	UserName string   `json:"user_name" validate:"required"`
	Password string   `json:"password" validate:"required"`
	// 是否对运行中的实例选择强制关机。建议对运行中的实例先手动关机，然后再重置用户密码。
	// 取值范围：<br>
	// <li>TRUE：表示在正常关机失败后进行强制关机<br>
	// <li>FALSE：表示在正常关机失败后不进行强制关机<br>
	// <br>默认取值：FALSE。
	// <br><br>强制关机的效果等同于关闭物理计算机的电源开关。强制关机可能会导致数据丢失或文件系统损坏，请仅在服务器不能正常关机时使用。
	ForceStop bool `json:"force_stop" validate:"required"`
}

// Validate tcloud cvm operation option.
func (opt TCloudResetPwdOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create --------------------------
//

// TCloudInternetChargeType CVM网络计费类型.
// 默认取值：非带宽包用户默认与子机付费类型保持一致，比如子机付费类型为预付费，网络计费类型默认为预付费；
// 子机付费类型为后付费，网络计费类型默认为后付费
type TCloudInternetChargeType string

const (
	// TCloudInternetBandwidthPrepaid BANDWIDTH_PREPAID 预付费按带宽结算
	TCloudInternetBandwidthPrepaid TCloudInternetChargeType = "BANDWIDTH_PREPAID"
	// TCloudInternetTrafficPostpaidByHour TRAFFIC_POSTPAID_BY_HOUR 流量按小时后付费
	TCloudInternetTrafficPostpaidByHour TCloudInternetChargeType = "TRAFFIC_POSTPAID_BY_HOUR"
	// TCloudInternetBandwidthPostpaidByHour BANDWIDTH_POSTPAID_BY_HOUR 带宽按小时后付费
	TCloudInternetBandwidthPostpaidByHour TCloudInternetChargeType = "BANDWIDTH_POSTPAID_BY_HOUR"
	// TCloudInternetBandwidthPackage BANDWIDTH_PACKAGE 带宽包用户
	TCloudInternetBandwidthPackage TCloudInternetChargeType = "BANDWIDTH_PACKAGE"
)

// TCloudCreateOption defines options to create aws cvm instances.
type TCloudCreateOption struct {
	DryRun                  bool                         `json:"dry_run" validate:"omitempty"`
	Region                  string                       `json:"region" validate:"required"`
	Name                    string                       `json:"name" validate:"required"`
	Zone                    string                       `json:"zone" validate:"required"`
	InstanceType            string                       `json:"instance_type" validate:"required"`
	CloudImageID            string                       `json:"cloud_image_id" validate:"required"`
	Password                string                       `json:"password" validate:"required"`
	RequiredCount           int64                        `json:"required_count" validate:"required"`
	CloudSecurityGroupIDs   []string                     `json:"cloud_security_group_ids" validate:"required"`
	ClientToken             *string                      `json:"client_token" validate:"omitempty"`
	CloudVpcID              string                       `json:"cloud_vpc_id" validate:"required"`
	CloudSubnetID           string                       `json:"cloud_subnet_id" validate:"required"`
	InstanceChargeType      TCloudInstanceChargeType     `json:"instance_charge_type" validate:"required"`
	InstanceChargePrepaid   *TCloudInstanceChargePrepaid `json:"instance_charge_prepaid" validate:"omitempty"`
	SystemDisk              *TCloudSystemDisk            `json:"system_disk" validate:"required"`
	DataDisk                []TCloudDataDisk             `json:"data_disk" validate:"omitempty"`
	PublicIPAssigned        bool                         `json:"public_ip_assigned" validate:"omitempty"`
	InternetMaxBandwidthOut int64                        `json:"internet_max_bandwidth_out" validate:"omitempty"`
	InternetChargeType      TCloudInternetChargeType     `json:"internet_charge_type" validate:"omitempty"`
	BandwidthPackageID      *string                      `json:"bandwidth_package_id" validate:"omitempty"`
}

// Validate aws cvm operation option.
func (opt TCloudCreateOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if opt.PublicIPAssigned && opt.InternetMaxBandwidthOut == 0 {
		return errors.New("assign public ip, internet_max_bandwidth_out is required")
	}

	return nil
}

// TCloudInstanceChargeType 实例计费类型。
type TCloudInstanceChargeType string

const (
	// Prepaid 预付费，即包年包月
	Prepaid TCloudInstanceChargeType = "PREPAID"
	// PostpaidByHour 按小时后付费
	PostpaidByHour TCloudInstanceChargeType = "POSTPAID_BY_HOUR"
	// Cdhpaid 独享子机（基于专用宿主机创建，宿主机部分的资源不收费）
	Cdhpaid TCloudInstanceChargeType = "CDHPAID"
	// Spotpaid 竞价付费
	Spotpaid TCloudInstanceChargeType = "SPOTPAID"
	// Cdcpaid 专用集群付费
	Cdcpaid TCloudInstanceChargeType = "CDCPAID"
)

// TCloudInstanceChargePrepaid define tcloud instance charge prepaid
type TCloudInstanceChargePrepaid struct {
	Period    *int64          `json:"period" validate:"omitempty"`
	RenewFlag TCloudRenewFlag `json:"renew_flag" validate:"omitempty"`
}

// TCloudRenewFlag 自动续费标识。
type TCloudRenewFlag string

const (
	// NotifyAndAutoRenew 通知过期且自动续费
	NotifyAndAutoRenew TCloudRenewFlag = "NOTIFY_AND_AUTO_RENEW"
	// NotifyAndManualRenew 通知过期不自动续费
	NotifyAndManualRenew TCloudRenewFlag = "NOTIFY_AND_MANUAL_RENEW"
	// DisableNotifyAndManualRenew 不通知过期不自动续费
	DisableNotifyAndManualRenew TCloudRenewFlag = "DISABLE_NOTIFY_AND_MANUAL_RENEW"
)

// TCloudSystemDisk tcloud system disk.
type TCloudSystemDisk struct {
	DiskType    TCloudSystemDiskType `json:"disk_type" validate:"omitempty"`
	CloudDiskID *string              `json:"cloud_disk_id" validate:"omitempty"`
	DiskSizeGB  *int64               `json:"disk_size_gb" validate:"omitempty"`
}

// TCloudSystemDiskType 硬盘类型。
type TCloudSystemDiskType string

const (
	// LocalBasic 本地硬盘
	LocalBasic TCloudSystemDiskType = "LOCAL_BASIC"
	// LocalSsd 本地SSD硬盘
	LocalSsd TCloudSystemDiskType = "LOCAL_SSD"
	// CloudBasic 普通云硬盘
	CloudBasic TCloudSystemDiskType = "CLOUD_BASIC"
	// CloudSsd SSD云硬盘
	CloudSsd TCloudSystemDiskType = "CLOUD_SSD"
	// CloudPremium 高性能云硬盘
	CloudPremium TCloudSystemDiskType = "CLOUD_PREMIUM"
	// CloudBssd 通用性SSD云硬盘
	CloudBssd TCloudSystemDiskType = "CLOUD_BSSD"
)

// TCloudDataDisk tencent cloud cvm instance data disk information
type TCloudDataDisk struct {
	DiskSizeGB  *int64             `json:"disk_size_gb" validate:"omitempty"`
	DiskType    TCloudDataDiskType `json:"disk_type" validate:"omitempty"`
	CloudDiskID *string            `json:"cloud_disk_id" validate:"omitempty"`
}

// TCloudDataDiskType 硬盘类型。
type TCloudDataDiskType string

const (
	// LocalBasicDataDiskType 本地硬盘
	LocalBasicDataDiskType TCloudDataDiskType = "LOCAL_BASIC"
	// LocalSsdDataDiskType 本地SSD硬盘
	LocalSsdDataDiskType TCloudDataDiskType = "LOCAL_SSD"
	// LocalNvmeDataDiskType 本地NVME硬盘，与InstanceType强相关，不支持指定
	LocalNvmeDataDiskType TCloudDataDiskType = "LOCAL_NVME"
	// LocalProDataDiskType 本地HDD硬盘，与InstanceType强相关，不支持指定
	LocalProDataDiskType TCloudDataDiskType = "LOCAL_PRO"
	// CloudBasicDataDiskType 普通云硬盘
	CloudBasicDataDiskType TCloudDataDiskType = "CLOUD_BASIC"
	// CloudPremiumDataDiskType 高性能云硬盘
	CloudPremiumDataDiskType TCloudDataDiskType = "CLOUD_PREMIUM"
	// CloudSsdDataDiskType SSD云硬盘
	CloudSsdDataDiskType TCloudDataDiskType = "CLOUD_SSD"
	// CloudHssdDataDiskType 增强型SSD云硬盘
	CloudHssdDataDiskType TCloudDataDiskType = "CLOUD_HSSD"
	// CloudTssdDataDiskType 极速型SSD云硬盘
	CloudTssdDataDiskType TCloudDataDiskType = "CLOUD_TSSD"
	// CloudBssdDataDiskType 通用型SSD云硬盘
	CloudBssdDataDiskType TCloudDataDiskType = "CLOUD_BSSD"
)

// TCloudCvm for cvm Instance
type TCloudCvm struct {
	*tcvm.Instance
}

// GetCloudID ...
func (cvm TCloudCvm) GetCloudID() string {
	return converter.PtrToVal(cvm.InstanceId)
}

// InquiryPriceResult define tcloud inquiry price result.
type InquiryPriceResult struct {
	DiscountPrice float64 `json:"discount_price"`
	OriginalPrice float64 `json:"original_price"`
}

// ListCvmWithCountOption returns count
type ListCvmWithCountOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
	// 基于安全组id过滤
	SGIDs []string         `json:"security_groups_ids" validate:"omitempty"`
	Page  *core.TCloudPage `json:"page" validate:"omitempty"`
}

// Validate tcloud cvm list option.
func (opt ListCvmWithCountOption) Validate() error {
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

// CvmWithCountResp ...
type CvmWithCountResp struct {
	TotalCount int64
	Cvms       []TCloudCvm
}
