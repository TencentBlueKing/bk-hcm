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

package enumor

import "fmt"

// LoadUsageTimeGranularity 使用率查询接口的时间粒度取值
type LoadUsageTimeGranularity string

const (
	// TimeGranularityDay 每天
	TimeGranularityDay LoadUsageTimeGranularity = "day"
	// TimeGranularityWeek 每周日
	TimeGranularityWeek LoadUsageTimeGranularity = "week"
	// TimeGranularityMonth 每月末
	TimeGranularityMonth LoadUsageTimeGranularity = "month"
)

// Validate LoadUsageTimeGranularity.
func (t LoadUsageTimeGranularity) Validate() error {
	switch t {
	case TimeGranularityDay:
	case TimeGranularityWeek:
	case TimeGranularityMonth:
	default:
		return fmt.Errorf("unsupported time granularity: %s, must be one of: day, week, month", t)
	}

	return nil
}

// LoadUsageDeviceENV 设备环境
type LoadUsageDeviceENV string

const (
	// DeviceENVIDC 国内环境
	DeviceENVIDC LoadUsageDeviceENV = "idc"
	// DeviceENVSG SG环境
	DeviceENVSG LoadUsageDeviceENV = "sg"
)

// Validate LoadUsageDeviceENV.
func (e LoadUsageDeviceENV) Validate() error {
	switch e {
	case DeviceENVIDC:
	case DeviceENVSG:
	default:
		return fmt.Errorf("unsupported load usage device env: %s, must be one of: idc, sg", e)
	}

	return nil
}

// FinOpsDeviceType 设备类型
type FinOpsDeviceType string

const (
	// FinOpsDeviceTypeCVM 虚拟机
	FinOpsDeviceTypeCVM FinOpsDeviceType = "cvm"
	// FinOpsDeviceTypeBareMetal 物理机
	FinOpsDeviceTypeBareMetal FinOpsDeviceType = "bareMetal"
)

// Validate FinOpsDeviceType.
func (d FinOpsDeviceType) Validate() error {
	switch d {
	case FinOpsDeviceTypeCVM:
	case FinOpsDeviceTypeBareMetal:
	default:
		return fmt.Errorf("unsupported finops device type: %s, must be one of: cvm, bareMetal", d)
	}

	return nil
}
