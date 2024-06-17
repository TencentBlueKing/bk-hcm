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

// ActionName is action name.
type ActionName string

// Validate ActionName.
func (v ActionName) Validate() error {
	switch v {
	case ActionAssignCvm, ActionStartCvm, ActionStopCvm, ActionRebootCvm, ActionDeleteCvm, ActionCreateCvm,
		ActionCreateAwsCvm, ActionCreateHuaWeiCvm, ActionCreateGcpCvm, ActionCreateAzureCvm:

	case ActionDeleteFirewallRule:

	case ActionDeleteSubnet:
	case ActionDeleteSecurityGroup, ActionCreateHuaweiSGRule:
	case ActionDeleteEIP:

	case VirRoot:
	case ActionCreateFactoryTest, ActionProduceTest, ActionAssembleTest, ActionSleep:
	case ActionTargetGroupAddRS, ActionTargetGroupRemoveRS, ActionTargetGroupModifyPort, ActionTargetGroupModifyWeight:
	case ActionLoadBalancerOperateWatch:
	case ActionListenerRuleAddTarget:
	case ActionDeleteLoadBalancer:
	case ActionPullDailyRawBill, ActionDailyAccountSplit, ActionDailyAccountSummary:
	default:
		return fmt.Errorf("unsupported action name type: %s", v)
	}

	return nil
}

// 主机相关Action
const (
	ActionAssignCvm       ActionName = "assign_cvm"
	ActionStartCvm        ActionName = "start_cvm"
	ActionStopCvm         ActionName = "stop_cvm"
	ActionRebootCvm       ActionName = "reboot_cvm"
	ActionDeleteCvm       ActionName = "delete_cvm"
	ActionCreateCvm       ActionName = "create_cvm"
	ActionCreateAwsCvm    ActionName = "create_aws_cvm"
	ActionCreateHuaWeiCvm ActionName = "create_huawei_cvm"
	ActionCreateGcpCvm    ActionName = "create_gcp_cvm"
	ActionCreateAzureCvm  ActionName = "create_azure_cvm"
)

// 防火墙相关Action
const (
	ActionDeleteFirewallRule ActionName = "delete_firewall_rule"
)

// 子网相关Action
const (
	ActionDeleteSubnet ActionName = "delete_subnet"
)

// 框架测试和框架中使用到的Action
const (
	// VirRoot vir root
	VirRoot ActionName = "root"

	// ActionCreateFactoryTest 测试相关Action
	ActionCreateFactoryTest ActionName = "create_factory"
	ActionProduceTest       ActionName = "produce"
	ActionAssembleTest      ActionName = "assemble"
	ActionSleep             ActionName = "sleep"
)

// Security Group
const (
	ActionDeleteSecurityGroup ActionName = "delete_security_group"
	ActionCreateHuaweiSGRule  ActionName = "create_huawei_sg_rule"
)

// EIP related action
const (
	// ActionDeleteEIP ...
	ActionDeleteEIP ActionName = "delete_eip"
)

// Flow相关Action
const (
	ActionLoadBalancerOperateWatch ActionName = "load_balancer_operate_watch"
)

// 负载均衡相关Action
const (
	ActionTargetGroupAddRS        ActionName = "tg_add_rs"
	ActionTargetGroupRemoveRS     ActionName = "tg_remove_rs"
	ActionTargetGroupModifyPort   ActionName = "tg_modify_port"
	ActionTargetGroupModifyWeight ActionName = "tg_modify_weight"

	// ActionListenerRuleAddTarget 直接将RS绑定到 监听器/规则 上
	ActionListenerRuleAddTarget ActionName = "listener_rule_add_target"

	ActionDeleteLoadBalancer = "delete_load_balancer"
)

// 账单相关Action
const (
	ActionPullDailyRawBill    = "bill_pull_daily_raw"
	ActionMainAccountSummary  = "bill_main_account_summary"
	ActionDailyAccountSplit   = "bill_daily_account_split"
	ActionDailyAccountSummary = "bill_daily_account_summary"
)
