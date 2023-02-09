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
	Region            string
	Zone              string
	DiskType          string
	DiskSize          int32
	DiskCount         *int32
	DiskChargeType    *string
	DiskChargePrepaid *HuaWeiDiskChargePrepaid
}

// ToCreateVolumeRequest 转换成接口需要的 CreateVolumeRequest 结构
func (opt *HuaWeiDiskCreateOption) ToCreateVolumeRequest() (*model.CreateVolumeRequest, error) {
	req := &model.CreateVolumeRequest{}
	req.Body = &model.CreateVolumeRequestBody{}

	volumeType, err := getCreateVolumeOptionVolumeType(opt.DiskType)
	if err != nil {
		return nil, err
	}

	req.Body.Volume = &model.CreateVolumeOption{
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
