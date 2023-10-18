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

// FlowName is tpl name.
type FlowName string

// Validate FlowName.
func (v FlowName) Validate() error {
	switch v {
	case FlowStartCvm, FlowStopCvm, FlowRebootCvm, FlowDeleteCvm, FlowCreateCvm:
	case FlowDeleteFirewallRule:
	case FlowDeleteSubnet:
	case FlowNormalTest, FlowSleepTest:
	default:
		return fmt.Errorf("unsupported tpl: %s", v)
	}

	return nil
}

// 主机相关Flow
const (
	FlowStartCvm  FlowName = "start_cvm"
	FlowStopCvm   FlowName = "stop_cvm"
	FlowRebootCvm FlowName = "reboot_cvm"
	FlowDeleteCvm FlowName = "delete_cvm"
	FlowCreateCvm FlowName = "create_cvm"
)

// 防火墙相关Flow
const (
	FlowDeleteFirewallRule FlowName = "delete_firewall_rule"
)

// 子网相关Flow
const (
	FlowDeleteSubnet FlowName = "delete_subnet"
)

// 测试相关Flow
const (
	// FlowNormalTest normal flow template test.
	FlowNormalTest FlowName = "normal_test"
	// FlowSleepTest sleep flow template test.
	FlowSleepTest FlowName = "sleep_test"
)
