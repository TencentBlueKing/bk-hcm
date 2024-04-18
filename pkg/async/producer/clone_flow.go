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

package producer

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// CloneFlow 按给定flow id 复制一份flow,task
func (p *producer) CloneFlow(kt *kit.Kit, flowId string, opt *CloneFlowOption) (newID string, err error) {
	listInput := &backend.ListInput{
		Filter: tools.EqualExpression("id", flowId),
		Page:   core.NewDefaultBasePage(),
	}
	flowList, err := p.backend.ListFlow(kt, listInput)
	if err != nil {
		logs.Errorf("fail to list flow for clone, err: %s, flow id: %s, rid: %s", err, flowId, kt.Rid)
		return "", err
	}
	if len(flowList) == 0 {
		return "", fmt.Errorf("flow(%s) not found", flowId)
	}

	oldFlow := flowList[0]
	listInput.Filter = tools.EqualExpression("flow_id", flowId)
	oldTaskList, err := p.backend.ListTask(kt, listInput)
	if err != nil {
		logs.Errorf("fail to list task for clone, err: %s, flow id: %s, rid: %s", err, flowId, kt.Rid)
		return "", err
	}
	newFlow := clone(kt, oldFlow, oldTaskList, opt)
	newID, err = p.backend.CreateFlow(kt, newFlow)
	if err != nil {
		logs.Errorf("create flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return newID, nil
}

func clone(kt *kit.Kit, oldFlow model.Flow, oldTaskList []model.Task, opt *CloneFlowOption) (newFlow *model.Flow) {
	newFlow = &model.Flow{
		Name:      oldFlow.Name,
		ShareData: tableasync.NewShareData(oldFlow.ShareData.GetInitData()),
		Memo:      oldFlow.Memo,
		State:     enumor.FlowPending,
		Reason:    nil,
		Worker:    nil,
		Tasks:     make([]model.Task, len(oldTaskList)),
		Creator:   kt.User,
		Reviser:   kt.User,
	}

	if opt.IsInitState {
		newFlow.State = enumor.FlowInit
	}
	if len(opt.Memo) > 0 {
		newFlow.Memo = opt.Memo
	}
	for i, old := range oldTaskList {
		newFlow.Tasks[i] = model.Task{
			FlowName:   oldFlow.Name,
			ActionID:   old.ActionID,
			ActionName: old.ActionName,
			Params:     old.Params,
			Retry:      old.Retry,
			DependOn:   old.DependOn,
			State:      mapCloneTaskState(old.State),
			Reason:     nil,
			Result:     "",
			Creator:    kt.User,
			Reviser:    kt.User,
		}
	}
	return newFlow
}

func mapCloneTaskState(state enumor.TaskState) enumor.TaskState {
	if state == enumor.TaskSuccess {
		// 成功的不必再执行
		return enumor.TaskSuccess
	}
	return enumor.TaskPending

}
