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

	ecsmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ecs/v2/model"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/evs/v2/model"
)

// HuaWeiDiskTypeEnum 华为云盘类型
var HuaWeiDiskTypeEnum = struct {
	SSD   string
	GPSSD string
	SAS   string
	SATA  string
	ESSD  string
}{SSD: "SSD", GPSSD: "GPSSD", SAS: "SAS", SATA: "SATA", ESSD: "ESSD"}

// HuaWeiDiskChargeTypeEnum 华为云盘计费模式
var HuaWeiDiskChargeTypeEnum = struct {
	PRE_PAID  string
	POST_PAID string
}{PRE_PAID: "prePaid", POST_PAID: "postPaid"}

// HuaWeiDiskPeriodTypeEnum 华为云盘计费周期
var HuaWeiDiskPeriodTypeEnum = struct {
	MONTH string
	YEAR  string
}{MONTH: "month", YEAR: "year"}

// HuaWeiDiskIsAutoRenewEnum 华为云盘是否自动续费标记
var HuaWeiDiskIsAutoRenewEnum = struct {
	FALSE string
	TRUE  string
}{FALSE: "false", TRUE: "true"}

// HuaWeiDiskCreateOption 华为云创建云盘参数
// reference: https://support.huaweicloud.com/api-evs/evs_04_2003.html
type HuaWeiDiskCreateOption struct {
	DiskName          *string                  `json:"disk_name"`
	Region            string                   `json:"region" validate:"required"`
	Zone              string                   `json:"zone" validate:"required"`
	DiskType          string                   `json:"disk_type" validate:"required"`
	DiskSize          int32                    `json:"disk_size" validate:"required"`
	DiskCount         *int32                   `json:"disk_count"`
	DiskChargeType    *string                  `json:"disk_charge_type"`
	DiskChargePrepaid *HuaWeiDiskChargePrepaid `json:"disk_charge_prepaid"`
}

// Validate ...
func (opt *HuaWeiDiskCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToCreateVolumeRequest 转换成接口需要的 CreateVolumeRequest 结构
func (opt *HuaWeiDiskCreateOption) ToCreateVolumeRequest() (*model.CreateVolumeRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := &model.CreateVolumeRequest{}
	req.Body = &model.CreateVolumeRequestBody{}

	volumeType, err := getCreateVolumeOptionVolumeType(opt.DiskType)
	if err != nil {
		return nil, err
	}

	req.Body.Volume = &model.CreateVolumeOption{
		Name:             opt.DiskName,
		AvailabilityZone: opt.Zone,
		VolumeType:       *volumeType,
		Size:             opt.DiskSize,
		Count:            opt.DiskCount,
	}

	chargingMode, err := getCreateVolumeChargingMode(*opt.DiskChargeType)
	req.Body.BssParam = &model.BssParamForCreateVolume{ChargingMode: chargingMode}
	// 预付费时, 需要设置订购周期等参数
	if chargingMode.Value() == HuaWeiDiskChargeTypeEnum.PRE_PAID {
		req.Body.BssParam.PeriodNum = opt.DiskChargePrepaid.PeriodNum
		req.Body.BssParam.PeriodType, _ = getCreateVolumePeriodType(*opt.DiskChargePrepaid.PeriodType)
		req.Body.BssParam.IsAutoRenew, _ = getCreateVolumeIsAutoRenew(*opt.DiskChargePrepaid.IsAutoRenew)
	}

	return req, nil
}

// HuaWeiDiskChargePrepaid 云盘预付费参数
type HuaWeiDiskChargePrepaid struct {
	PeriodNum   *int32
	PeriodType  *string
	IsAutoRenew *string
}

func getCreateVolumeOptionVolumeType(diskType string) (*model.CreateVolumeOptionVolumeType, error) {
	enum := model.GetCreateVolumeOptionVolumeTypeEnum()
	switch diskType {
	case HuaWeiDiskTypeEnum.SATA:
		return &enum.SATA, nil
	case HuaWeiDiskTypeEnum.ESSD:
		return &enum.ESSD, nil
	case HuaWeiDiskTypeEnum.SAS:
		return &enum.SAS, nil
	case HuaWeiDiskTypeEnum.SSD:
		return &enum.SSD, nil
	case HuaWeiDiskTypeEnum.GPSSD:
		return &enum.GPSSD, nil
	default:
		return nil, fmt.Errorf("invalid disk type %s", diskType)
	}
}

func getCreateVolumeChargingMode(chargeType string) (*model.BssParamForCreateVolumeChargingMode, error) {
	enum := model.GetBssParamForCreateVolumeChargingModeEnum()
	switch chargeType {
	case HuaWeiDiskChargeTypeEnum.PRE_PAID:
		return &enum.PRE_PAID, nil
	case HuaWeiDiskChargeTypeEnum.POST_PAID:
		return &enum.POST_PAID, nil
	default:
		return nil, fmt.Errorf("invalid charge type %s", chargeType)
	}
}

func getCreateVolumePeriodType(periodType string) (*model.BssParamForCreateVolumePeriodType, error) {
	enum := model.GetBssParamForCreateVolumePeriodTypeEnum()
	switch periodType {
	case HuaWeiDiskPeriodTypeEnum.YEAR:
		return &enum.YEAR, nil
	case HuaWeiDiskPeriodTypeEnum.MONTH:
		return &enum.MONTH, nil
	default:
		return nil, fmt.Errorf("invalid period type %s", periodType)
	}
}

func getCreateVolumeIsAutoRenew(isAutoRenew string) (*model.BssParamForCreateVolumeIsAutoRenew, error) {
	enum := model.GetBssParamForCreateVolumeIsAutoRenewEnum()
	switch isAutoRenew {
	case HuaWeiDiskIsAutoRenewEnum.TRUE:
		return &enum.TRUE, nil
	case HuaWeiDiskIsAutoRenewEnum.FALSE:
		return &enum.FALSE, nil
	default:
		return nil, fmt.Errorf("invalid autorenew flag %s", isAutoRenew)
	}
}

// HuaWeiDiskListOption define huawei disk list option.
type HuaWeiDiskListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.HuaWeiPage `json:"page" validate:"omitempty"`
}

// Validate huawei disk list option.
func (opt HuaWeiDiskListOption) Validate() error {
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

// HuaWeiDiskDeleteOption ...
type HuaWeiDiskDeleteOption struct {
	Region  string `json:"region" validate:"required"`
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate ...
func (opt *HuaWeiDiskDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDeleteVolumeRequest ...
func (opt *HuaWeiDiskDeleteOption) ToDeleteVolumeRequest() (*model.DeleteVolumeRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &model.DeleteVolumeRequest{VolumeId: opt.CloudID}, nil
}

// HuaWeiDiskAttachOption ...
type HuaWeiDiskAttachOption struct {
	Region     string `json:"region" validate:"required"`
	CloudCvmID string `json:"cloud_cvm_id" validate:"required"`
	// 待挂载磁盘的磁盘ID，UUID格式。
	CloudDiskID string `json:"cloud_disk_id" validate:"required"`
	// 磁盘挂载点。  > 说明：  - 新增加的磁盘挂载点不能和已有的磁盘挂载点相同。
	// - 对于采用XEN虚拟化类型的弹性云服务器，device为必选参数；系统盘挂载点请指定/dev/sda；
	//数据盘挂载点请按英文字母顺序依次指定，如/dev/sdb，/dev/sdc，如果指定了以“/dev/vd”开头的挂载点，系统默认改为“/dev/sd”。
	//- 对于采用KVM虚拟化类型的弹性云服务器，系统盘挂载点请指定/dev/vda；数据盘挂载点可不用指定，也可按英文字母顺序依次指定，
	//如/dev/vdb，/dev/vdc，如果指定了以“/dev/sd”开头的挂载点，系统默认改为“/dev/vd”。
	DeviceName *string `json:"device_name"`
}

// Validate ...
func (opt *HuaWeiDiskAttachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToAttachServerVolumeRequest ...
func (opt *HuaWeiDiskAttachOption) ToAttachServerVolumeRequest() (*ecsmodel.AttachServerVolumeRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	req := &ecsmodel.AttachServerVolumeRequest{}
	req.ServerId = opt.CloudCvmID
	req.Body = &ecsmodel.AttachServerVolumeRequestBody{
		VolumeAttachment: &ecsmodel.AttachServerVolumeOption{Device: opt.DeviceName, VolumeId: opt.CloudDiskID},
	}

	return req, nil
}

// HuaWeiDiskDetachOption ...
type HuaWeiDiskDetachOption struct {
	Region      string `json:"region" validate:"required"`
	CloudCvmID  string `json:"cloud_cvm_id" validate:"required"`
	CloudDiskID string `json:"cloud_disk_id" validate:"required"`
}

// Validate ...
func (opt *HuaWeiDiskDetachOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToDetachServerVolumeRequest ...
func (opt *HuaWeiDiskDetachOption) ToDetachServerVolumeRequest() (*ecsmodel.DetachServerVolumeRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}
	return &ecsmodel.DetachServerVolumeRequest{ServerId: opt.CloudCvmID, VolumeId: opt.CloudDiskID}, nil
}
