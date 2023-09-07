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

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

// TCloudDiskChargeType 腾讯云盘计费模式
var TCloudDiskChargeTypeEnum = struct {
	PREPAID          string
	POSTPAID_BY_HOUR string
}{PREPAID: "PREPAID", POSTPAID_BY_HOUR: "POSTPAID_BY_HOUR"}

// TCloudDiskCreateOption 腾讯云创建云盘参数
// reference: https://cloud.tencent.com/document/api/362/16312
type TCloudDiskCreateOption struct {
	DiskName          *string `json:"disk_name"`
	Region            string  `json:"region" validate:"required"`
	Zone              string  `json:"zone" validate:"required"`
	DiskType          string  `json:"disk_type" validate:"required"`
	DiskSize          *uint64 `json:"disk_size"`
	DiskCount         *uint64 `json:"disk_count"`
	DiskChargeType    string  `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid *TCloudDiskChargePrepaid
}

// Validate ...
func (opt *TCloudDiskCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ToCreateDisksRequest 转换成接口需要的 CreateDisksRequest 结构
func (opt *TCloudDiskCreateOption) ToCreateDisksRequest() (*cbs.CreateDisksRequest, error) {
	if err := opt.Validate(); err != nil {
		return nil, err
	}

	req := cbs.NewCreateDisksRequest()
	req.DiskName = opt.DiskName
	req.Placement = &cbs.Placement{Zone: common.StringPtr(opt.Zone)}
	req.DiskType = common.StringPtr(opt.DiskType)
	req.DiskCount = opt.DiskCount
	req.DiskSize = opt.DiskSize
	req.DiskChargeType = common.StringPtr(opt.DiskChargeType)
	// 预付费模式需要设定 DiskChargePrepaid
	if *req.DiskChargeType == TCloudDiskChargeTypeEnum.PREPAID {
		req.DiskChargePrepaid = &cbs.DiskChargePrepaid{
			Period:              opt.DiskChargePrepaid.Period,
			RenewFlag:           opt.DiskChargePrepaid.RenewFlag,
			CurInstanceDeadline: opt.DiskChargePrepaid.CurInstanceDeadline,
		}
	}

	return req, nil
}

// TCloudDiskChargePrepaid 云盘预付费参数
type TCloudDiskChargePrepaid struct {
	// Period 购买云盘的时长，默认单位为月. 取值范围：1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 24, 36
	Period *uint64
	// RenewFlag 和 CurInstanceDeadline 取值参考 https://cloud.tencent.com/document/api/362/15669#DiskChargePrepaid
	RenewFlag           *string
	CurInstanceDeadline *string
}

// TCloudDiskListOption define tcloud disk list option.
type TCloudDiskListOption struct {
	Region   string           `json:"region" validate:"required"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
}

// Validate tcloud disk option.
func (opt TCloudDiskListOption) Validate() error {
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

// TCloudDiskDeleteOption ...
type TCloudDiskDeleteOption struct {
	Region   string   `json:"region"  validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"required"`
}

// Validate ...
func (o *TCloudDiskDeleteOption) Validate() error {
	return validator.Validate.Struct(o)
}

// ToTerminateDisksRequest ...
func (o *TCloudDiskDeleteOption) ToTerminateDisksRequest() (*cbs.TerminateDisksRequest, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	req := cbs.NewTerminateDisksRequest()
	req.DiskIds = common.StringPtrs(o.CloudIDs)

	return req, nil
}

// TCloudDiskAttachOption ...
type TCloudDiskAttachOption struct {
	Region       string   `json:"region"  validate:"required"`
	CloudCvmID   string   `json:"cloud_cvm_id" validate:"required"`
	CloudDiskIDs []string `json:"cloud_disk_ids" validate:"required"`
	// 可选参数，不传该参数则仅执行挂载操作。传入`True`时，会在挂载成功后将云硬盘设置为随云主机销毁模式，仅对按量计费云硬盘有效。
	DeleteWithInstance *bool `json:"delete_with_instance"`
	// 可选参数，用于控制云盘挂载时使用的挂载模式，目前仅对黑石裸金属机型有效。取值范围：<br><li>PF<br><li>VF
	AttachMode *string `json:"attach_mode"`
}

// Validate ...
func (o *TCloudDiskAttachOption) Validate() error {
	return validator.Validate.Struct(o)
}

// ToAttachDisksRequest ...
func (o *TCloudDiskAttachOption) ToAttachDisksRequest() (*cbs.AttachDisksRequest, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	req := cbs.NewAttachDisksRequest()
	req.InstanceId = &o.CloudCvmID
	req.DiskIds = common.StringPtrs(o.CloudDiskIDs)
	req.AttachMode = o.AttachMode
	req.DeleteWithInstance = o.DeleteWithInstance

	return req, nil
}

// TCloudDiskDetachOption ...
type TCloudDiskDetachOption struct {
	Region       string   `json:"region"  validate:"required"`
	CloudCvmID   string   `json:"cloud_cvm_id" validate:"required"`
	CloudDiskIDs []string `json:"cloud_disk_ids" validate:"required"`
}

// Validate ...
func (o *TCloudDiskDetachOption) Validate() error {
	return validator.Validate.Struct(o)
}

// ToDetachDisksRequest ...
func (o *TCloudDiskDetachOption) ToDetachDisksRequest() (*cbs.DetachDisksRequest, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}

	req := cbs.NewDetachDisksRequest()
	req.InstanceId = common.StringPtr(o.CloudCvmID)
	req.DiskIds = common.StringPtrs(o.CloudDiskIDs)
	return req, nil
}

// TCloudDisk for cbs Disk
type TCloudDisk struct {
	*cbs.Disk
}

// GetCloudID ...
func (disk TCloudDisk) GetCloudID() string {
	return converter.PtrToVal(disk.DiskId)
}

// InquiryPriceResult define tcloud inquiry price result.
type InquiryPriceResult struct {
	DiscountPrice float64 `json:"discount_price"`
	OriginalPrice float64 `json:"original_price"`
}
