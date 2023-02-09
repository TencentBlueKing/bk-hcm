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

	cbs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cbs/v20170312"
)

// TCloudDiskChargeType 腾讯云盘计费模式
var TCloudDiskChargeTypeEnum = struct {
	PREPAID          string
	POSTPAID_BY_HOUR string
}{PREPAID: "PREPAID", POSTPAID_BY_HOUR: "POSTPAID_BY_HOUR"}

// TCloudDiskCreateOption 腾讯云创建云盘参数
// reference: https://cloud.tencent.com/document/api/362/16312
type TCloudDiskCreateOption struct {
	Name              *string
	Region            string
	Zone              *string
	DiskType          *string
	DiskSize          *uint64
	DiskCount         *uint64
	DiskChargeType    *string
	DiskChargePrepaid *TCloudDiskChargePrepaid
}

// ToCreateDisksRequest 转换成接口需要的 CreateDisksRequest 结构
// TODO 增加参数校验
func (opt *TCloudDiskCreateOption) ToCreateDisksRequest() (*cbs.CreateDisksRequest, error) {
	req := cbs.NewCreateDisksRequest()

	req.DiskName = opt.Name
	req.Placement = &cbs.Placement{Zone: opt.Zone}
	req.DiskType = opt.DiskType
	req.DiskCount = opt.DiskCount
	req.DiskSize = opt.DiskSize
	req.DiskChargeType = opt.DiskChargeType

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
	Region string           `json:"region" validate:"required"`
	Page   *core.TCloudPage `json:"page" validate:"omitempty"`
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
