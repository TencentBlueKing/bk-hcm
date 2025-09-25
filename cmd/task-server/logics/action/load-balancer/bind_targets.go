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

// ListenerRuleAddTargetAction 将RS应用到监听器或者规则
type ListenerRuleAddTargetAction struct{}

// ListenerRuleAddTargetOption ...
type ListenerRuleAddTargetOption struct {
	LoadBalancerID                     string `json:"lb_id" validate:"required"`
	*hclb.BatchRegisterTCloudTargetReq `json:",inline"`
	ManagementDetailIDs                []string `json:"management_detail_ids" validate:"required,min=1"`
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

	// detail 状态检查
	reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs)
	if err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return nil, err
	}
	if len(reason) > 0 {
		return reason, nil
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	info, err := actcli.GetDataService().Global.Cloud.GetResBasicInfo(kt.Kit(), enumor.LoadBalancerCloudResType,
		opt.LoadBalancerID, "vendor")
	if err != nil {
		logs.Errorf("fail to get load balancer info, err: %v, res id: %s, rid: %s", err, opt.LoadBalancerID,
			kt.Kit().Rid)
		return nil, err
	}

	taskDetailState := enumor.TaskDetailSuccess
	defer func() {
		// 更新任务状态
		if err := batchUpdateTaskDetailResultState(kt.Kit(), opt.ManagementDetailIDs, taskDetailState,
			nil, err); err != nil {
			logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		}
	}()
	switch info.Vendor {
	case enumor.TCloud:
		if err = actcli.GetHCService().TCloud.Clb.BatchRegisterTargetToListenerRule(kt.Kit(), opt.LoadBalancerID,
			opt.BatchRegisterTCloudTargetReq); err != nil {
			taskDetailState = enumor.TaskDetailFailed
			logs.Errorf("fail to register target to listener rule, err: %v, rid: %s", err, kt.Kit().Rid)
			return nil, err
		}
	default:
		taskDetailState = enumor.TaskDetailFailed
		err = errf.Newf(errf.InvalidParameter, "vendor(%s) not supported for listener rule add target", info.Vendor)
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
