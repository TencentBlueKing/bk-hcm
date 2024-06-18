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

package logicsaction

import (
	actionbilldailypull "hcm/cmd/task-server/logics/action/bill/dailypull"
	actionbillsplit "hcm/cmd/task-server/logics/action/bill/dailysplit"
	actiondailysummary "hcm/cmd/task-server/logics/action/bill/dailysummary"
	actcli "hcm/cmd/task-server/logics/action/cli"
	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	actioneip "hcm/cmd/task-server/logics/action/eip"
	actionfirewall "hcm/cmd/task-server/logics/action/firewall"
	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionsg "hcm/cmd/task-server/logics/action/security-group"
	actionsubnet "hcm/cmd/task-server/logics/action/subnet"
	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/pkg/async/action"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao"
)

// Init init action.
func Init(cli *client.ClientSet, dao dao.Set) {
	actcli.SetClientSet(cli)
	actcli.SetDaoSet(dao)

	register()
}

func register() {
	action.RegisterAction(actioncvm.NewStartAction())
	action.RegisterAction(actioncvm.NewStopAction())
	action.RegisterAction(actioncvm.NewRebootAction())
	action.RegisterAction(actioncvm.NewDeleteAction())
	action.RegisterAction(actioncvm.CreateCvmAction{})
	action.RegisterAction(actioncvm.AssignCvmAction{})

	action.RegisterAction(actionfirewall.DeleteAction{})

	action.RegisterAction(actionsubnet.DeleteAction{})
	action.RegisterAction(actionsg.DeleteSgAction{})
	action.RegisterAction(actionsg.CreateHuaweiSGRuleAction{})
	action.RegisterAction(actioneip.DeleteEIPAction{})

	action.RegisterAction(actionlb.AddTargetToGroupAction{})
	action.RegisterAction(actionflow.LoadBalancerOperateWatchAction{})
	action.RegisterTpl(actionflow.FlowLoadBalancerOperateWatchTpl)
	action.RegisterAction(actionlb.RemoveTargetAction{})
	action.RegisterAction(actionlb.ModifyTargetPortAction{})
	action.RegisterAction(actionlb.ModifyTargetWeightAction{})

	action.RegisterAction(actionlb.ListenerRuleAddTargetAction{})
	action.RegisterAction(actionlb.DeleteLoadBalancerAction{})

	action.RegisterAction(actionbilldailypull.PullDailyBillAction{})
	action.RegisterAction(actionbillsplit.DailyAccountSplitAction{})
	action.RegisterAction(actiondailysummary.DailySummaryAction{})
}
