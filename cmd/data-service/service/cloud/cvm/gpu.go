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

package cvm

import (
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

// buildGPUMachineTypes loads the GPU machine type list of the vendor from global_config.
func (svc *cvmSvc) buildGPUMachineTypes(kt *kit.Kit, vendor enumor.Vendor) (map[string]struct{}, error) {
	key, err := getGPUMachineKey(vendor)
	if err != nil {
		logs.Errorf("get gpu machine key failed, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
		return nil, err
	}

	opt := &types.ListOption{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("config_type", string(enumor.GlobalConfigTypeGPUMachineType)),
			tools.RuleEqual("config_key", string(key)),
		),
		Page: &core.BasePage{Limit: 1},
	}

	result, err := svc.dao.GlobalConfig().List(kt, opt)
	if err != nil {
		logs.Errorf("list gpu machine type config failed, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
		return nil, err
	}

	machineTypes := make(map[string]struct{})
	if len(result.Details) == 0 {
		logs.Warnf("gpu machine type config not found, vendor: %s, rid: %s", vendor, kt.Rid)
		return machineTypes, nil
	}

	rawList := make([]string, 0)
	if err = json.Unmarshal([]byte(result.Details[0].ConfigValue), &rawList); err != nil {
		logs.Errorf("unmarshal gpu machine type config failed, err: %v, vendor: %s, rid: %s", err, vendor, kt.Rid)
		return nil, err
	}

	for _, one := range rawList {
		machineTypes[one] = struct{}{}
	}
	logs.Infof("loaded gpu machine type config, vendor: %s, count: %d, rid: %s", vendor, len(machineTypes), kt.Rid)
	return machineTypes, nil
}

// isGPUMachine reports whether the given machine type belongs to the GPU machine type
// list configured in the global_config table for the vendor. extension is the cvm extension
// json, which is only consumed by GCP (to read the attached GPU card count) and ignored otherwise.
func isGPUMachine(vendor enumor.Vendor, machineType string, machineTypes map[string]struct{},
	extension string) bool {

	switch vendor {
	case enumor.Gcp:
		return isGcpGPUMachine(machineType, machineTypes, extension)
	case enumor.HuaWei, enumor.TCloud, enumor.Azure:
		// 这些厂商的 GPU 机型均按机型前缀识别
		return matchGPUMachineTypeByPrefix(machineType, machineTypes)
	case enumor.Aws:
		// AWS 按机型全名精确匹配
		_, hit := machineTypes[machineType]
		return hit
	default:
		return false
	}
}

// isGcpGPUMachine reports whether a GCP instance is a GPU machine.
// GCP 需双重判断：A2/A3/A4/G2/G4 等专用 GPU 机型的 GPU 与机型绑定（实例不返回 guestAccelerators），
// 按机型前缀识别；N1 等通用机型可额外挂载 GPU，其机型不含 GPU 前缀，只能依据挂载的 GPU 卡数识别。
func isGcpGPUMachine(machineType string, machineTypes map[string]struct{}, extension string) bool {
	if matchGPUMachineTypeByPrefix(machineType, machineTypes) {
		return true
	}
	return gcpGPUCardCount(extension) > 0
}

// gcpGPUCardCount sums the attached GPU card count from the GCP cvm extension json. It returns 0
// for an empty extension or when the instance has no attached guest accelerators.
func gcpGPUCardCount(extension string) int {
	if len(extension) == 0 {
		return 0
	}

	ext := new(corecvm.GcpCvmExtension)
	if err := json.Unmarshal([]byte(extension), ext); err != nil {
		logs.Errorf("unmarshal gcp cvm extension for gpu card count failed, err: %v", err)
		return 0
	}

	// 累加实例上挂载的所有 guest accelerator（GPU）卡数
	count := 0
	for _, acc := range ext.GuestAccelerators {
		count += int(acc.AcceleratorCount)
	}
	return count
}

func getGPUMachineKey(vendor enumor.Vendor) (enumor.GlobalConfigKeyGPUMachineType, error) {
	switch vendor {
	case enumor.HuaWei:
		return enumor.GlobalConfigKeyHuaweiGPUPrefix, nil
	case enumor.TCloud:
		return enumor.GlobalConfigKeyTcloudGPUPrefix, nil
	case enumor.Gcp:
		return enumor.GlobalConfigKeyGcpGPUPrefix, nil
	case enumor.Aws:
		return enumor.GlobalConfigKeyAws, nil
	case enumor.Azure:
		return enumor.GlobalConfigKeyAzureGPUPrefix, nil
	default:
		return "", errf.New(errf.InvalidParameter, fmt.Sprintf("unsupported vendor: %s", vendor))
	}
}

// matchGPUMachineTypeByPrefix reports whether machineType matches any configured GPU machine type prefix.
func matchGPUMachineTypeByPrefix(machineType string, prefixes map[string]struct{}) bool {
	// 忽略大小写
	machineType = strings.ToLower(machineType)
	for prefix := range prefixes {
		if strings.HasPrefix(machineType, strings.ToLower(prefix)) {
			return true
		}
	}
	return false
}
