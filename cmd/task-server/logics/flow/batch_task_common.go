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

// Package actionflow ...
package actionflow

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	coretask "hcm/pkg/api/core/task"
	datatask "hcm/pkg/api/data-service/task"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
	"hcm/pkg/tools/slice"
)

const (
	// BatchTaskDefaultRetryTimes 批量任务默认重试次数
	BatchTaskDefaultRetryTimes = 3
	// BatchTaskDefaultRetryDelayMinMS 批量任务默认重试最小延迟时间
	BatchTaskDefaultRetryDelayMinMS = 600
	// BatchTaskDefaultRetryDelayMaxMS 批量任务默认重试最大延迟时间
	BatchTaskDefaultRetryDelayMaxMS = 1000
)

// ListTaskDetail list task detail
func ListTaskDetail(kt *kit.Kit, ids []string) ([]coretask.Detail, error) {
	result := make([]coretask.Detail, 0, len(ids))
	for _, idBatch := range slice.Split(ids, int(core.DefaultMaxPageLimit)) {
		// 查询任务状态
		detailListReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", idBatch),
			Page:   core.NewDefaultBasePage(),
		}
		detailResp, err := actcli.GetDataService().Global.TaskDetail.List(kt, detailListReq)
		if err != nil {
			logs.Errorf("fail to query task detail, err: %v, ids: %s, rid: %s", err, ids, kt.Rid)
			return nil, err
		}
		if len(detailResp.Details) != len(idBatch) {
			return nil, fmt.Errorf("some of task management detail ids not found, want: %d, got: %d",
				len(ids), len(detailResp.Details))
		}
		result = append(result, detailResp.Details...)
	}

	return result, nil
}

// BatchUpdateTaskDetailState batch update task detail state
func BatchUpdateTaskDetailState(kt *kit.Kit, ids []string, state enumor.TaskDetailState) error {
	detailUpdates := make([]datatask.UpdateTaskDetailField, min(len(ids), constant.BatchOperationMaxLimit))
	for _, idBatch := range slice.Split(ids, constant.BatchOperationMaxLimit) {
		for i := range idBatch {
			detailUpdates[i] = datatask.UpdateTaskDetailField{ID: ids[i], State: state}
		}
		updateTaskReq := &datatask.UpdateDetailReq{Items: detailUpdates[:len(idBatch)]}
		rangeMS := [2]uint{BatchTaskDefaultRetryDelayMinMS, BatchTaskDefaultRetryDelayMaxMS}
		policy := retry.NewRetryPolicy(0, rangeMS)
		err := policy.BaseExec(kt, func() error {
			err := actcli.GetDataService().Global.TaskDetail.Update(kt, updateTaskReq)
			if err != nil {
				logs.Errorf("fail to update task detail state to %s, err: %v, ids: %s, rid: %s",
					state, err, idBatch, kt.Rid)
				return err
			}
			return nil
		})
		if err != nil {
			logs.Errorf("fail to update task detail state to %s after retry, err: %v, ids: %s, rid: %s",
				state, err, idBatch, kt.Rid)
			return err
		}
	}

	return nil
}

// BatchUpdateTaskDetailResultState batch update task detail result state
func BatchUpdateTaskDetailResultState(kt *kit.Kit, ids []string, state enumor.TaskDetailState,
	result any, reason error) error {

	detailUpdates := make([]datatask.UpdateTaskDetailField, min(len(ids), constant.BatchOperationMaxLimit))
	for _, idBatch := range slice.Split(ids, constant.BatchOperationMaxLimit) {
		for i := range idBatch {
			field := datatask.UpdateTaskDetailField{ID: ids[i], State: state, Result: result}
			if reason != nil {
				// 需要截取否则超出DB字段长度限制，会更新状态失败
				runesReason := []rune(reason.Error())
				field.Reason = string(runesReason[:min(1000, len(runesReason))])
			}
			detailUpdates[i] = field
		}
		updateTaskReq := &datatask.UpdateDetailReq{Items: detailUpdates[:len(idBatch)]}
		rangeMS := [2]uint{BatchTaskDefaultRetryDelayMinMS, BatchTaskDefaultRetryDelayMaxMS}
		policy := retry.NewRetryPolicy(0, rangeMS)
		err := policy.BaseExec(kt, func() error {
			err := actcli.GetDataService().Global.TaskDetail.Update(kt, updateTaskReq)
			if err != nil {
				logs.Errorf("fail to update task detail result state to %s, err: %v, ids: %s, rid: %s",
					state, err, idBatch, kt.Rid)
				return err
			}
			return nil
		})
		if err != nil {
			logs.Errorf("fail to update task detail result state to %s after retry, err: %v, ids: %s, rid: %s",
				state, err, idBatch, kt.Rid)
			return err
		}
	}
	return nil
}
