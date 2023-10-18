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

package actionfirewall

import (
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
)

var _ action.Action = new(DeleteAction)
var _ action.ParameterAction = new(DeleteAction)

// DeleteAction define delete cvm action.
type DeleteAction struct{}

// ParameterNew return delete params.
func (act DeleteAction) ParameterNew() (params interface{}) {
	return new(string)
}

// Name return action name.
func (act DeleteAction) Name() enumor.ActionName {
	return enumor.ActionDeleteFirewallRule
}

// Run delete firewall rule.
func (act DeleteAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	idPtr, ok := params.(*string)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type not right")
	}

	if idPtr == nil || len(*idPtr) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	id := *idPtr
	if err := actcli.GetHCService().Gcp.Firewall.DeleteFirewallRule(kt.Kit(), id); err != nil {
		logs.Errorf("delete firewall rule failed, err: %v, id: %s, rid: %s", err, id, kt.Kit().Rid)
		return nil, err
	}

	return nil, nil
}
