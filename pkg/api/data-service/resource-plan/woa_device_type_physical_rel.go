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

package resourceplan

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	wd "hcm/pkg/dal/table/resource-plan/woa-device-type"
)

// WoaDeviceTypeSyncReq sync device type request
type WoaDeviceTypeSyncReq struct {
	DeviceTypes []string `json:"device_types" validate:"required,min=1"`
}

// Validate validates the WoaDeviceTypeSyncReq
func (req *WoaDeviceTypeSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

// WoaDeviceTypePhysicalRelListResult woa device type physical rel list result.
type WoaDeviceTypePhysicalRelListResult types.ListResult[wd.WoaDeviceTypePhysicalRelTable]

// WoaDeviceTypePhysicalRel woa device type physical rel.
type WoaDeviceTypePhysicalRel struct {
	ID                   string `json:"id"`
	DeviceType           string `json:"device_type"`
	PhysicalDeviceFamily string `json:"physical_device_family"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
}

// WoaDeviceTypePhysicalRelCreateReq defines the request for creating a woa device type physical rel record.
type WoaDeviceTypePhysicalRelCreateReq struct {
	DeviceType           string `json:"device_type" validate:"required"`
	PhysicalDeviceFamily string `json:"physical_device_family" validate:"required"`
}

// Validate validates the WoaDeviceTypePhysicalRelCreateReq.
func (req *WoaDeviceTypePhysicalRelCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// WoaDeviceTypePhysicalRelUpdateReq defines the request for updating a woa device type physical rel record.
type WoaDeviceTypePhysicalRelUpdateReq struct {
	ID                   string `json:"id" validate:"required"`
	DeviceType           string `json:"device_type,omitempty"`
	PhysicalDeviceFamily string `json:"physical_device_family,omitempty"`
}

// Validate validates the WoaDeviceTypePhysicalRelUpdateReq.
func (req *WoaDeviceTypePhysicalRelUpdateReq) Validate() error {
	if req.DeviceType == "" && req.PhysicalDeviceFamily == "" {
		return errf.New(errf.InvalidParameter, "at least one field to update is required")
	}
	return validator.Validate.Struct(req)
}

// WoaDeviceTypePhysicalRelBatchCreateReq defines the request for batch creating woa device type physical rel records.
type WoaDeviceTypePhysicalRelBatchCreateReq struct {
	Records []WoaDeviceTypePhysicalRelCreateReq `json:"records" validate:"required,min=1,max=100"`
}

// Validate validates the WoaDeviceTypePhysicalRelBatchCreateReq.
func (req *WoaDeviceTypePhysicalRelBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}
	for _, record := range req.Records {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// WoaDeviceTypePhysicalRelBatchUpdateReq defines the request for batch updating woa device type physical rel records.
type WoaDeviceTypePhysicalRelBatchUpdateReq struct {
	Records []WoaDeviceTypePhysicalRelUpdateReq `json:"records"`
}

// Validate validates the WoaDeviceTypePhysicalRelBatchUpdateReq.
func (req *WoaDeviceTypePhysicalRelBatchUpdateReq) Validate() error {
	if len(req.Records) == 0 {
		return errf.New(errf.InvalidParameter, "records is required")
	}
	for _, record := range req.Records {
		if err := record.Validate(); err != nil {
			return err
		}
	}
	return nil
}
