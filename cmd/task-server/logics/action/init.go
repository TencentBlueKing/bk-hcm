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
	actcli "hcm/cmd/task-server/logics/action/cli"
	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	actionfirewall "hcm/cmd/task-server/logics/action/firewall"
	actionsg "hcm/cmd/task-server/logics/action/security-group"
	actionsubnet "hcm/cmd/task-server/logics/action/subnet"
	"hcm/pkg/async/action"
	"hcm/pkg/client"
)

// Init init action.
func Init(cli *client.ClientSet) {
	actcli.SetClientSet(cli)

	register()
}

func register() {
	action.RegisterAction(actioncvm.NewStartAction())
	action.RegisterAction(actioncvm.NewStopAction())
	action.RegisterAction(actioncvm.NewRebootAction())
	action.RegisterAction(actioncvm.NewDeleteAction())
	action.RegisterAction(actioncvm.CreateTCloudCvmAction{})

	action.RegisterAction(actionfirewall.DeleteAction{})

	action.RegisterAction(actionsubnet.DeleteAction{})
	action.RegisterAction(actionsg.DeleteSgAction{})
}
