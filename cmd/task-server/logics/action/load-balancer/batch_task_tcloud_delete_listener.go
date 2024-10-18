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
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// --------------------------[批量操作-删除监听器(需要所有RS都已解绑)]-----------------------------

var _ action.Action = new(BatchTaskDeleteListenerAction)
var _ action.ParameterAction = new(BatchTaskDeleteListenerAction)

// BatchTaskDeleteListenerAction 批量操作-删除监听器
type BatchTaskDeleteListenerAction struct{}

// BatchTaskDeleteListenerOption ...
type BatchTaskDeleteListenerOption struct {
	Vendor         enumor.Vendor `json:"vendor" validate:"required"`
	LoadBalancerID string        `json:"lb_id" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和批量绑定的Targets参数长度对应
	ManagementDetailIDs  []string `json:"management_detail_ids" validate:"required,max=500"`
	*core.BatchDeleteReq `json:",inline"`
}

// Validate validate option.
func (bdl BatchTaskDeleteListenerOption) Validate() error {

	switch bdl.Vendor {
	case enumor.TCloud:
	default:
		return fmt.Errorf("unsupport vendor for batch delete listener: %s", bdl.Vendor)
	}

	if bdl.BatchDeleteReq == nil {
		return errf.New(errf.InvalidParameter, "batch_tcloud_delete_listener_req is required")
	}
	if len(bdl.ManagementDetailIDs) != len(bdl.BatchDeleteReq.IDs) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and deleteListenerIDs length "+
			"is not match, %d != %d", len(bdl.ManagementDetailIDs), len(bdl.BatchDeleteReq.IDs))
	}
	return validator.Validate.Struct(bdl)
}

// ParameterNew return request params.
func (act BatchTaskDeleteListenerAction) ParameterNew() (params any) {
	return new(BatchTaskDeleteListenerOption)
}

// Name return action name
func (act BatchTaskDeleteListenerAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskDeleteListener
}

// Run 批量删除监听器-支持幂等
func (act BatchTaskDeleteListenerAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskDeleteListenerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type is not BatchTaskDeleteListenerOption")
	}

	// 批量查询并检查任务详情状态
	detailList, err := listTaskDetail(kt.Kit(), opt.ManagementDetailIDs)
	if err != nil {
		return fmt.Sprintf("task detail query failed, mdIDs: %v", opt.ManagementDetailIDs), err
	}

	for _, detail := range detailList {
		if detail.State == enumor.TaskDetailCancel {
			// 任务被取消，跳过该批次
			return fmt.Sprintf("task detail task: %s is canceled", detail.ID), nil
		}
		if detail.State != enumor.TaskDetailInit {
			return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
				detail.ID, detail.State)
		}
	}

	// 查询监听器列表
	lblList, err := batchListListenerByIDs(kt.Kit(), opt.BatchDeleteReq.IDs)
	if err != nil {
		logs.Errorf("failed to batch list listener by ids, err: %v, lblIDs: %v, rid: %s",
			err, opt.BatchDeleteReq.IDs, kt.Kit().Rid)
		return nil, err
	}
	// 监听器不存在，直接返回
	if len(lblList) == 0 {
		logs.Infof("delete listener list query empty, lbID: %s, lblIDs: %v, detailIDs: %v, rid: %s",
			opt.LoadBalancerID, opt.BatchDeleteReq.IDs, opt.ManagementDetailIDs, kt.Kit().Rid)
		return nil, nil
	}

	// 汇总需要删除的监听器ID
	delIDs := slice.Map(lblList, func(lbl corelb.BaseListener) string {
		return lbl.ID
	})

	// 更新任务状态为 running
	if err = batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		return fmt.Sprintf("failed to update detail task to running, mdIDs: %v", opt.ManagementDetailIDs), err
	}

	defer func() {
		// 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if taskErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err = batchUpdateTaskDetailResultState(kt.Kit(), opt.ManagementDetailIDs, targetState, nil, taskErr)
		if err != nil {
			logs.Errorf("failed to set detail to %s after cloud operation finished, err: %v, rid: %s",
				targetState, err, kt.Kit().Rid)
		}
	}()

	// 分批删除监听器
	parts := slice.Split(delIDs, constant.BatchDeleteListenerCloudMaxLimit)
	for _, partIDs := range parts {
		delIDsReq := &core.BatchDeleteReq{IDs: partIDs}
		err = actcli.GetHCService().TCloud.Clb.DeleteListener(kt.Kit(), delIDsReq)
		if err != nil {
			logs.Errorf("failed to batch delete listener, err: %v, lblPartIDs: %v, rid: %s", err, partIDs, kt.Kit().Rid)
			return nil, err
		}
	}

	return nil, nil
}

// Rollback 批量删除监听器支持重入，无需回滚
func (act BatchTaskDeleteListenerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskDeleteListenerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
