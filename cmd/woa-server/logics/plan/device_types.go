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

package plan

import (
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// IsDeviceMatched return whether each device type in deviceTypeSlice can use deviceType's resource plan.
func (c *Controller) IsDeviceMatched(kt *kit.Kit, deviceTypeSlice []string, deviceType string) ([]bool, error) {
	// get device type map.
	deviceTypeMap, err := c.deviceTypesMap.GetDeviceTypes(kt)
	if err != nil {
		logs.Errorf("failed to get device type map, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	result := make([]bool, len(deviceTypeSlice))
	for idx, ele := range deviceTypeSlice {
		// if ele and device type are equal, then they are matched.
		if ele == deviceType {
			result[idx] = true
		}

		if _, ok := deviceTypeMap[ele]; !ok {
			continue
		}

		if _, ok := deviceTypeMap[deviceType]; !ok {
			continue
		}

		// if technical_class of ele and core type are equal, then they are matched.
		if deviceTypeMap[ele].TechnicalClass == deviceTypeMap[deviceType].TechnicalClass &&
			deviceTypeMap[ele].CoreType == deviceTypeMap[deviceType].CoreType {
			result[idx] = true
		}
	}

	return result, nil
}
