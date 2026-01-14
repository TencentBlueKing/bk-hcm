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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/device-capacity"
	"hcm/pkg/thirdparty/cvmapi"
)

// ListDeviceCapacities list device capacities.
type ListDeviceCapacities struct {
	Count            uint64                               `json:"count,omitempty"`
	DeviceCapacities []devicecapacity.DeviceCapacityTable `json:"device_capacities,omitempty"`
}

// CapacityWithDeviceInfo device capacity with device type details.
type CapacityWithDeviceInfo struct {
	RequireType     int64                    `db:"require_type" json:"require_type"`
	Region          string                   `db:"region" json:"region"`
	Zone            string                   `db:"zone" json:"zone"`
	DeviceFamily    string                   `db:"device_family" json:"device_family"`
	DeviceType      string                   `db:"device_type" json:"device_type"`
	CPUCore         int64                    `db:"cpu_core" json:"cpu_core"`
	Memory          int64                    `db:"memory" json:"memory"`
	Capacity        int64                    `db:"capacity" json:"capacity"`
	CoreType        enumor.CoreType          `db:"core_type" json:"core_type"`
	DeviceTypeClass cvmapi.InstanceTypeClass `db:"device_type_class" json:"device_type_class"`
}

// ListCapacitiesWithDeviceInfo list device capacities with device info.
type ListCapacitiesWithDeviceInfo struct {
	Count   uint64                   `json:"count,omitempty"`
	Details []CapacityWithDeviceInfo `json:"details,omitempty"`
}
