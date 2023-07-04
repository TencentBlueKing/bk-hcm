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

package csdisk

import (
	"errors"

	hcproto "hcm/pkg/api/hc-service/disk"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// TCloudDiskAttachReq ...
type TCloudDiskAttachReq struct {
	DiskID string `json:"disk_id" validate:"required"`
	CvmID  string `json:"cvm_id" validate:"required"`
}

// Validate ...
func (req *TCloudDiskAttachReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudDiskCreateReq ...
type TCloudDiskCreateReq struct {
	AccountID         string                           `json:"account_id" validate:"required"`
	BkBizID           int64                            `json:"bk_biz_id" validate:"omitempty"`
	DiskName          string                           `json:"disk_name" validate:"required"`
	Region            string                           `json:"region" validate:"required"`
	Zone              string                           `json:"zone" validate:"required"`
	DiskSize          uint64                           `json:"disk_size" validate:"required"`
	DiskType          string                           `json:"disk_type" validate:"required"`
	DiskCount         uint32                           `json:"disk_count" validate:"required"`
	DiskChargeType    string                           `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid *hcproto.TCloudDiskChargePrepaid `json:"disk_charge_prepaid" validate:"required"`
	Memo              *string                          `json:"memo" validate:"omitempty"`
}

// Validate ...
func (req *TCloudDiskCreateReq) Validate(bizRequired bool) error {
	if req.DiskCount > constant.BatchOperationMaxLimit {
		return errors.New("disk count should <= 100")
	}

	if bizRequired && req.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	return validator.Validate.Struct(req)
}
