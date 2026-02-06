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

// Package devicetype ...
package devicetype

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/enumor"
	dt "hcm/pkg/dal/table/cloud/device-type"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
)

// DeviceType 机型实例，用于实现 CloudResType 接口
type DeviceType struct {
	ID              string                   `json:"id"`
	Vendor          enumor.Vendor            `json:"vendor"`
	Region          string                   `json:"region"`
	Zone            string                   `json:"zone"`
	DeviceType      string                   `json:"device_type"`
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class"`
	DeviceClass     string                   `json:"device_class"`
	DeviceFamily    string                   `json:"device_family"`
	CoreType        enumor.CoreType          `json:"core_type"`
	CpuCore         int64                    `json:"cpu_core"`
	Memory          int64                    `json:"memory"`
	TechnicalClass  string                   `json:"technical_class"`
	Disable         bool                     `json:"disable"`
	Source          enumor.DeviceTypeSource  `json:"source"`
	Creator         string                   `json:"creator"`
	Reviser         string                   `json:"reviser"`
}

// GetCloudID 返回机型的唯一标识，使用 region+zone+device_type 组合
func (d DeviceType) GetCloudID() string {
	return fmt.Sprintf("%s:%s:%s", d.Region, d.Zone, d.DeviceType)
}

// CloudIDInfo cloud id info
type CloudIDInfo struct {
	Region     string
	Zone       string
	DeviceType string
}

// GetInfoFromCloudID 从云ID中获取信息
func GetInfoFromCloudID(cloudID string) (*CloudIDInfo, error) {
	parts := strings.Split(cloudID, ":")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid cloudID: %s", cloudID)
	}
	return &CloudIDInfo{
		Region:     parts[0],
		Zone:       parts[1],
		DeviceType: parts[2],
	}, nil
}

// GetID 返回数据库ID，用于实现 DBResType 接口
func (d DeviceType) GetID() string {
	return d.ID
}

// ConvTableToDeviceType 将 DeviceTypeTable 转换为 DeviceType
func ConvTableToDeviceType(one dt.DeviceTypeTable) DeviceType {
	return DeviceType{
		ID:              one.ID,
		Vendor:          one.Vendor,
		Region:          one.Region,
		Zone:            one.Zone,
		DeviceType:      one.DeviceType,
		DeviceTypeClass: one.DeviceTypeClass,
		DeviceClass:     one.DeviceClass,
		DeviceFamily:    one.DeviceFamily,
		CoreType:        one.CoreType,
		CpuCore:         one.CpuCore,
		Memory:          one.Memory,
		TechnicalClass:  one.TechnicalClass,
		Disable:         cvt.PtrToVal(one.Disable),
		Source:          one.Source,
		Creator:         one.Creator,
		Reviser:         one.Reviser,
	}
}

// DistinctDeviceType distinct device type
type DistinctDeviceType struct {
	ID              string                   `json:"id"`
	Vendor          enumor.Vendor            `json:"vendor"`
	DeviceType      string                   `json:"device_type"`
	DeviceClass     string                   `json:"device_class"`
	DeviceFamily    string                   `json:"device_family"`
	CoreType        enumor.CoreType          `json:"core_type"`
	CpuCore         int64                    `json:"cpu_core"`
	Memory          int64                    `json:"memory"`
	TechnicalClass  string                   `json:"technical_class"`
	DeviceTypeClass cvmapi.InstanceTypeClass `json:"device_type_class"`
	Disable         bool                     `json:"disable"`
	Source          enumor.DeviceTypeSource  `json:"source"`
	Creator         string                   `json:"creator"`
	Reviser         string                   `json:"reviser"`
}

// ConvTableToDistinctDeviceType 将 DeviceTypeTable 转换为 DistinctDeviceType
func ConvTableToDistinctDeviceType(one dt.DeviceTypeTable) DistinctDeviceType {
	return DistinctDeviceType{
		ID:              one.ID,
		Vendor:          one.Vendor,
		DeviceType:      one.DeviceType,
		DeviceClass:     one.DeviceClass,
		DeviceFamily:    one.DeviceFamily,
		CoreType:        one.CoreType,
		CpuCore:         one.CpuCore,
		Memory:          one.Memory,
		TechnicalClass:  one.TechnicalClass,
		DeviceTypeClass: one.DeviceTypeClass,
		Disable:         cvt.PtrToVal(one.Disable),
		Source:          one.Source,
		Creator:         one.Creator,
		Reviser:         one.Reviser,
	}
}
