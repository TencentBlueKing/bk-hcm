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

package producer

import (
	"fmt"

	"hcm/pkg/async/action"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AddCustomFlow add custom flow
func (p *producer) AddCustomFlow(kt *kit.Kit, opt *AddCustomFlowOption) (id string, err error) {
	if err = opt.Validate(); err != nil {
		return "", err
	}

	if err = validateCustomFlowParam(kt, opt); err != nil {
		logs.Errorf("validate flow template use param failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flow := buildCustomFlow(opt)

	id, err = p.backend.CreateFlow(kt, flow)
	if err != nil {
		logs.Errorf("create flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return id, nil
}

func validateCustomFlowParam(kt *kit.Kit, opt *AddCustomFlowOption) error {

	taskMap := make(map[action.ActIDType]bool)
	// actionID 唯一性校验
	for _, one := range opt.Tasks {
		if taskMap[one.ActionID] {
			return fmt.Errorf("actionID: %s repeat", one.ActionID)
		}
		taskMap[one.ActionID] = true
	}

	for _, task := range opt.Tasks {
		act, exist := action.GetAction(task.ActionName)
		if !exist {
			return fmt.Errorf("action: %s not exist", task.ActionName)
		}

		// 参数校验
		if !task.Params.IsEmpty() {
			paramAct, ok := act.(action.ParameterAction)
			if !ok {
				return fmt.Errorf("action: %s need params, but not impl ParameterAction", task.ActionName)
			}

			params := paramAct.ParameterNew()
			if err := action.Decode(task.Params, params); err != nil {
				logs.Errorf("action: %s can not decode params, err: %v, field: %s, type: %T, rid: %s", task.ActionName,
					err, task.Params, params, kt.Rid)
				return fmt.Errorf("action: %s can not decode param, err: %v", task.ActionName, err)
			}
		}

		// Task 是否可重试校验
		if task.Retry != nil && task.Retry.IsEnable() {
			_, ok := act.(action.RollbackAction)
			if !ok {
				return fmt.Errorf("action: %s can retry, but not impl RollbackAction", task.ActionName)
			}
		}

		// 依赖节点是否存在校验
		for _, one := range task.DependOn {
			if !taskMap[one] {
				return fmt.Errorf("dependOn's actionID: %s not exist", one)
			}
		}
	}

	return nil
}

func buildCustomFlow(opt *AddCustomFlowOption) *model.Flow {
	if opt.ShareData == nil {
		opt.ShareData = new(tableasync.ShareData)
	}

	flow := &model.Flow{
		Name:      opt.Name,
		ShareData: opt.ShareData,
		Memo:      opt.Memo,
		Tasks:     make([]model.Task, 0, len(opt.Tasks)),
	}
	if opt.IsInitState {
		flow.State = enumor.FlowInit
	}

	for _, one := range opt.Tasks {
		if one.Retry == nil {
			one.Retry = new(tableasync.Retry)
		}

		task := model.Task{
			FlowName:   opt.Name,
			ActionID:   one.ActionID,
			ActionName: one.ActionName,
			Params:     one.Params,
			Retry:      one.Retry,
			DependOn:   one.DependOn,
		}

		flow.Tasks = append(flow.Tasks, task)
	}

	return flow
}

// BatchUpdateCustomFlowState batch update custom flow state
func (p *producer) BatchUpdateCustomFlowState(kt *kit.Kit, opt *UpdateCustomFlowStateOption) error {
	if err := opt.Validate(); err != nil {
		return err
	}

	err := p.backend.BatchUpdateFlowStateByCAS(kt, opt.FlowInfos)
	if err != nil {
		logs.Errorf("batch update custom flow state failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return err
	}

	return nil
}

// RetryFlowTask retry task of flow
func (p *producer) RetryFlowTask(kt *kit.Kit, flowID, taskID string) error {

	err := p.backend.RetryTask(kt, flowID, taskID)
	if err != nil {
		logs.Errorf("retry task(%s) of flow(%s) failed, err: %v, rid: %s", taskID, flowID, err, kt.Rid)
		return err
	}
	return nil
}
