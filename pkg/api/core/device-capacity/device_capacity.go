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

// Package devicecapacity ...
package devicecapacity

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/thirdparty/cvmapi"
)

// DeviceCapacity device capacity information.
type DeviceCapacity struct {
	ID            string             `json:"id"`
	RequireType   enumor.RequireType `json:"require_type"`
	Region        string             `json:"region"`
	Zone          string             `json:"zone"`
	DeviceType    string             `json:"device_type"`
	Capacity      *int64             `json:"capacity"`
	Extension     types.JsonField    `json:"extension"`
	core.Revision `json:",inline"`
}

// CapacityWithDeviceInfo device capacity with device type details information.
type CapacityWithDeviceInfo struct {
	RequireType     enumor.RequireType       `json:"require_type"`
	Region          string                   `json:"region"`
	Zone            string                   `json:"zone"`
	DeviceFamily    string                   `json:"device_family"`
	DeviceType      string                   `json:"device_type"`
	CPUCore         int64                    `json:"cpu_core"`
	Memory          int64                    `json:"memory"`
	Capacity        int64                    `json:"capacity"`
	CoreType        enumor.CoreType          `json:"core_type"`
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class"`
}
