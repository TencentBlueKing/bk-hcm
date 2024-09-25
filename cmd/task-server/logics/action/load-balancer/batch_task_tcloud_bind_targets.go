/*
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
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// --------------------------[批量操作-绑定RS]-----------------------------

var _ action.Action = new(BatchTaskBindTargetAction)
var _ action.ParameterAction = new(BatchTaskBindTargetAction)

// BatchTaskBindTargetAction 批量操作-绑定RS 将目标组中的RS应用到监听器或者规则
type BatchTaskBindTargetAction struct{}

// BatchTaskBindTargetOption ...
type BatchTaskBindTargetOption struct {
	LoadBalancerID string `json:"lb_id" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和批量绑定的Targets参数长度对应
	ManagementDetailIDs                []string `json:"management_detail_ids" validate:"required,max=500"`
	*hclb.BatchRegisterTCloudTargetReq `json:",inline"`
}

// Validate validate option.
func (opt BatchTaskBindTargetOption) Validate() error {
	if opt.BatchRegisterTCloudTargetReq == nil {
		return errf.New(errf.InvalidParameter, "batch_register_tcloud_target_req is required")
	}
	if len(opt.ManagementDetailIDs) != len(opt.BatchRegisterTCloudTargetReq.Targets) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and targets length not match, %d != %d",
			len(opt.ManagementDetailIDs), len(opt.BatchRegisterTCloudTargetReq.Targets))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskBindTargetAction) ParameterNew() (params any) {
	return new(BatchTaskBindTargetOption)
}

// Name return action name
func (act BatchTaskBindTargetAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskTCloudBindTarget
}

// Run 将目标组中的RS绑定到监听器/规则中
func (act BatchTaskBindTargetAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskBindTargetOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	detailList, err := listTaskDetail(kt.Kit(), opt.ManagementDetailIDs)
	if err != nil {
		return fmt.Sprintf("task detail query failed"), err
	}
	for _, detail := range detailList {
		if detail.State == enumor.TaskDetailCancel {
			// 任务被取消，跳过该批次
			return fmt.Sprintf("task detail %s canceled", detail.ID), nil
		}
		if detail.State != enumor.TaskDetailInit {
			return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
				detail.ID, detail.State)
		}
	}
	// 更新任务状态为 running
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		return fmt.Sprintf("fail to update detail to running"), err
	}

	defer func() {
		// 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if taskErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := batchUpdateTaskDetailResultState(kt.Kit(), opt.ManagementDetailIDs, targetState, nil, taskErr)
		if err != nil {
			logs.Errorf("fail to set detail to %s after cloud operation finished, err: %v, rid: %s",
				targetState, err, kt.Kit().Rid)
		}
	}()

	err = actcli.GetHCService().TCloud.Clb.BatchRegisterTargetToListenerRule(kt.Kit(), opt.LoadBalancerID,
		opt.BatchRegisterTCloudTargetReq)
	if err != nil {
		logs.Errorf("fail to register target to listener rule, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, err
	}

	return nil, nil
}

// Rollback 添加rs支持重入，无需回滚
func (act BatchTaskBindTargetAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskBindTargetAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
