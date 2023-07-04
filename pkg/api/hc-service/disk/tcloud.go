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
	"hcm/pkg/criteria/validator"
)

// TCloudDiskCreateReq ...
type TCloudDiskCreateReq struct {
	*DiskBaseCreateReq `json:",inline" validate:"required"`
	Extension          *TCloudDiskExtensionCreateReq `json:"extension" validate:"required"`
}

// Validate ...
func (req *TCloudDiskCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudDiskExtensionCreateReq ...
type TCloudDiskExtensionCreateReq struct {
	DiskChargeType    string                   `json:"disk_charge_type" validate:"required"`
	DiskChargePrepaid *TCloudDiskChargePrepaid `json:"disk_charge_prepaid"`
}

// TCloudDiskChargePrepaid ...
type TCloudDiskChargePrepaid struct {
	Period    *uint64 `json:"period" validate:"omitempty"`
	RenewFlag *string `json:"renew_flag" validate:"omitempty"`
}

// Validate ...
func (req *TCloudDiskChargePrepaid) Validate() error {
	return validator.Validate.Struct(req)
}

// TCloudDiskAttachReq ...
type TCloudDiskAttachReq struct {
	AccountID string `json:"account_id" validate:"required"`
	CvmID     string `json:"cvm_id" validate:"required"`
	DiskID    string `json:"disk_id" validate:"required"`
}

// Validate ...
func (req *TCloudDiskAttachReq) Validate() error {
	return validator.Validate.Struct(req)
}
