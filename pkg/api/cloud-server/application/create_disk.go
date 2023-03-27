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

package application

import (
	"errors"

	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/assert"
)

// TCloudDiskCreateReq ...
type TCloudDiskCreateReq struct {
	AccountID         string                           `json:"account_id" validate:"required"`
	BkBizID           int64                            `json:"bk_biz_id" validate:"required,min=1"`
	DiskName          string                           `json:"disk_name" validate:"required"`
	Region            string                           `json:"region" validate:"required"`
	Zone              string                           `json:"zone" validate:"required"`
	DiskSize          uint64                           `json:"disk_size" validate:"required"`
	DiskType          string                           `json:"disk_type" validate:"required"`
	DiskCount         uint32                           `json:"disk_count" validate:"required"`
	DiskChargeType    string                           `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid *hcproto.TCloudDiskChargePrepaid `json:"disk_charge_prepaid"`
	Memo              *string                          `json:"memo"`
}

// Validate ...
func (req *TCloudDiskCreateReq) Validate() error {
	if req.DiskCount > requiredCountMaxLimit {
		return errors.New("disk count should <= 100")
	}

	return validator.Validate.Struct(req)
}

// HuaWeiDiskCreateReq ...
type HuaWeiDiskCreateReq struct {
	AccountID         string                           `json:"account_id" validate:"required"`
	BkBizID           int64                            `json:"bk_biz_id" validate:"required,min=1"`
	DiskName          *string                          `json:"disk_name"`
	Region            string                           `json:"region" validate:"required"`
	Zone              string                           `json:"zone" validate:"required"`
	DiskType          string                           `json:"disk_type" validate:"required"`
	DiskSize          int32                            `json:"disk_size" validate:"required"`
	DiskCount         int32                            `json:"disk_count" validate:"required"`
	DiskChargeType    *string                          `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid *hcproto.HuaWeiDiskChargePrepaid `json:"disk_charge_prepaid"`
	Memo              *string                          `json:"memo"`
}

// Validate ...
func (req *HuaWeiDiskCreateReq) Validate() error {
	if req.DiskCount > requiredCountMaxLimit {
		return errors.New("disk count should <= 100")
	}

	return validator.Validate.Struct(req)
}

// GcpDiskCreateReq ...
type GcpDiskCreateReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id" validate:"required,min=1"`
	DiskName  string  `json:"disk_name" validate:"required"`
	Region    string  `json:"region" validate:"required"`
	Zone      string  `json:"zone" validate:"required"`
	DiskType  string  `json:"disk_type" validate:"required"`
	DiskSize  int32   `json:"disk_size" validate:"required"`
	DiskCount int32   `json:"disk_count" validate:"required"`
	Memo      *string `json:"memo"`
}

// Validate ...
func (req *GcpDiskCreateReq) Validate() error {
	if req.DiskCount > requiredCountMaxLimit {
		return errors.New("disk count should <= 100")
	}

	return validator.Validate.Struct(req)
}

// AzureDiskCreateReq ...
type AzureDiskCreateReq struct {
	AccountID         string  `json:"account_id" validate:"required"`
	BkBizID           int64   `json:"bk_biz_id" validate:"required,min=1"`
	DiskName          string  `json:"disk_name" validate:"required,lowercase"`
	ResourceGroupName string  `json:"resource_group_name" validate:"required,lowercase"`
	Region            string  `json:"region" validate:"required,lowercase"`
	Zone              string  `json:"zone" validate:"required,lowercase"`
	DiskType          string  `json:"disk_type" validate:"required"`
	DiskSize          int32   `json:"disk_size" validate:"required"`
	DiskCount         int32   `json:"disk_count" validate:"required"`
	Memo              *string `json:"memo"`
}

// Validate ...
func (req *AzureDiskCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.DiskCount > requiredCountMaxLimit {
		return errors.New("disk count should <= 100")
	}

	// region can be no space lowercase
	if !assert.IsSameCaseNoSpaceString(req.Region) {
		return errf.New(errf.InvalidParameter, "region can only be lowercase")
	}

	// zone can be no space lowercase
	if !assert.IsSameCaseNoSpaceString(req.Zone) {
		return errf.New(errf.InvalidParameter, "zone can only be lowercase")
	}

	return nil
}

// AwsDiskCreateReq ...
type AwsDiskCreateReq struct {
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id" validate:"required,min=1"`
	DiskName  string  `json:"disk_name" validate:"required"`
	Region    string  `json:"region" validate:"required"`
	Zone      string  `json:"zone" validate:"required"`
	DiskType  string  `json:"disk_type" validate:"required"`
	DiskSize  int32   `json:"disk_size" validate:"required"`
	DiskCount int32   `json:"disk_count" validate:"required"`
	Memo      *string `json:"memo"`
}

// Validate ...
func (req *AwsDiskCreateReq) Validate() error {
	if req.DiskCount > requiredCountMaxLimit {
		return errors.New("disk count should <= 100")
	}

	return validator.Validate.Struct(req)
}
