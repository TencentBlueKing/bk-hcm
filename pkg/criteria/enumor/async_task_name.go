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
	case ActionStartCvm, ActionStopCvm, ActionRebootCvm, ActionDeleteCvm, ActionCreateTCloudCvm,
		ActionCreateAwsCvm, ActionCreateHuaWeiCvm, ActionCreateGcpCvm, ActionCreateAzureCvm:

	case ActionDeleteFirewallRule:

	case ActionDeleteSubnet:
	case ActionDeleteSecurityGroup, ActionCreateHuaweiSGRule:

	case VirRoot:
	case ActionCreateFactoryTest, ActionProduceTest, ActionAssembleTest, ActionSleep:
	default:
		return fmt.Errorf("unsupported action name type: %s", v)
	}

	return nil
}

// 主机相关Action
const (
	ActionStartCvm        ActionName = "start_cvm"
	ActionStopCvm         ActionName = "stop_cvm"
	ActionRebootCvm       ActionName = "reboot_cvm"
	ActionDeleteCvm       ActionName = "delete_cvm"
	ActionCreateTCloudCvm ActionName = "create_tcloud_cvm"
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
