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
	// 	校验默认的FlowName
	if err := v.ValidateDefault(); err == nil {
		return nil
	}

	// 校验负载均衡的FlowName
	if err := v.ValidateLoadBalancer(); err == nil {
		return nil
	}

	return fmt.Errorf("unsupported flow name: %s", v)
}

// 默认的FlowName
var defaultFlowNameMap = map[FlowName]struct{}{
	FlowStartCvm:            {},
	FlowStopCvm:             {},
	FlowRebootCvm:           {},
	FlowDeleteCvm:           {},
	FlowCreateCvm:           {},
	FlowDeleteFirewallRule:  {},
	FlowDeleteSubnet:        {},
	FlowNormalTest:          {},
	FlowSleepTest:           {},
	FlowDeleteSecurityGroup: {},
	FlowCreateHuaweiSGRule:  {},
	FlowDeleteEIP:           {},
	FlowPullRawBill:         {},
	FlowSplitBill:           {},
	FlowBillDailySummary:    {},
}

// ValidateDefault validate default FlowName.
func (v FlowName) ValidateDefault() error {
	_, exist := defaultFlowNameMap[v]
	if !exist {
		return fmt.Errorf("unsupported flow name: %s", v)
	}
	return nil
}

// 负载均衡相关的FlowName
var loadBalancerFlowNameMap = map[FlowName]struct{}{
	FlowTargetGroupAddRS:               {},
	FlowTargetGroupRemoveRS:            {},
	FlowTargetGroupModifyPort:          {},
	FlowTargetGroupModifyWeight:        {},
	FlowLoadBalancerOperateWatch:       {},
	FlowApplyTargetGroupToListenerRule: {},
	FlowDeleteLoadBalancer:             {},
}

// ValidateLoadBalancer validate load balancer FlowName.
func (v FlowName) ValidateLoadBalancer() error {
	_, exist := loadBalancerFlowNameMap[v]
	if !exist {
		return fmt.Errorf("%s does not have a corresponding flow name", v)
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

// 安全组和安全组规则相关Flow
const (
	FlowDeleteSecurityGroup FlowName = "delete_security_group"
	FlowCreateHuaweiSGRule  FlowName = "create_huawei_sg_rule"
)

// EIP 相关Flow
const (
	// FlowDeleteEIP ...
	FlowDeleteEIP FlowName = "delete_eip"
)

// Flow 相关Flow
const (
	// FlowLoadBalancerOperateWatch 负载均衡操作查询
	FlowLoadBalancerOperateWatch FlowName = "load_balancer_operate_watch"
)

// 负载均衡相关Flow
const (
	FlowTargetGroupAddRS        FlowName = "tg_add_rs"
	FlowTargetGroupRemoveRS     FlowName = "tg_remove_rs"
	FlowTargetGroupModifyPort   FlowName = "tg_modify_port"
	FlowTargetGroupModifyWeight FlowName = "tg_modify_weight"

	FlowApplyTargetGroupToListenerRule FlowName = "apply_tg_listener_rule"

	FlowDeleteLoadBalancer FlowName = "delete_load_balancer"
)

// 账单相关Flow
const (
	FlowPullRawBill      FlowName = "bill_pull_daily_raw"
	FlowSplitBill        FlowName = "bill_split_daily"
	FlowBillDailySummary FlowName = "bill_daily_summary"
)
