/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package actionlb

import (
	actcli "hcm/cmd/task-server/logics/action/cli"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// --------------------------[将目标组中的RS应用到监听器或者规则]-----------------------------

var _ action.Action = new(ListenerRuleAddTargetAction)
var _ action.ParameterAction = new(ListenerRuleAddTargetAction)

// ListenerRuleAddTargetAction 将目标组中的RS应用到监听器或者规则
type ListenerRuleAddTargetAction struct{}

// ListenerRuleAddTargetOption ...
type ListenerRuleAddTargetOption struct {
	LoadBalancerID                     string `json:"lb_id" validate:"required"`
	*hclb.BatchRegisterTCloudTargetReq `json:",inline"`
}

// Validate validate option.
func (opt ListenerRuleAddTargetOption) Validate() error {

	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act ListenerRuleAddTargetAction) ParameterNew() (params any) {
	return new(ListenerRuleAddTargetOption)
}

// Name return action name
func (act ListenerRuleAddTargetAction) Name() enumor.ActionName {
	return enumor.ActionListenerRuleAddTarget
}

// Run 将目标组中的RS绑定到监听器/规则中
func (act ListenerRuleAddTargetAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*ListenerRuleAddTargetOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	err := actcli.GetHCService().TCloud.Clb.BatchRegisterTargetToListenerRule(
		kt.Kit(), opt.LoadBalancerID, opt.BatchRegisterTCloudTargetReq)
	if err != nil {
		logs.Errorf("fail to create register target to listener rule, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, err
	}

	return nil, err
}

// Rollback 添加rs支持重入，无需回滚
func (act ListenerRuleAddTargetAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- ListenerRuleAddTargetAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
