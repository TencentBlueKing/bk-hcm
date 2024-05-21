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
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// AddTemplateFlow add template flow
func (p *producer) AddTemplateFlow(kt *kit.Kit, opt *AddTemplateFlowOption) (id string, err error) {
	if err = opt.Validate(); err != nil {
		return "", err
	}

	tpl, exist := action.GetTpl(opt.Name)
	if !exist {
		return "", fmt.Errorf("flow tempalte: %s not found", opt.Name)
	}

	if err = validateTplUseParam(kt, tpl, opt); err != nil {
		logs.Errorf("validate flow template use param failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flow := buildFlow(tpl, opt)

	id, err = p.backend.CreateFlow(kt, flow)
	if err != nil {
		logs.Errorf("create flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return id, nil
}

func buildFlow(tpl action.FlowTemplate, opt *AddTemplateFlowOption) *model.Flow {
	flow := &model.Flow{
		Name:      tpl.Name,
		ShareData: tpl.ShareData,
		Memo:      opt.Memo,
		Tasks:     make([]model.Task, 0, len(tpl.Tasks)),
	}
	if opt.IsInitState {
		flow.State = enumor.FlowInit
	}

	m := make(map[action.ActIDType]types.JsonField, len(opt.Tasks))
	for _, one := range opt.Tasks {
		m[one.ActionID] = one.Params
	}

	for _, one := range tpl.Tasks {
		if one.Retry == nil {
			one.Retry = new(tableasync.Retry)
		}

		task := model.Task{
			FlowName:   tpl.Name,
			ActionID:   one.ActionID,
			ActionName: one.ActionName,
			Params:     m[one.ActionID],
			Retry:      one.Retry,
			DependOn:   one.DependOn,
		}
		if opt.IsInitState {
			task.State = enumor.TaskInit
		}

		flow.Tasks = append(flow.Tasks, task)
	}

	return flow
}

// validateTplUseParam 校验任务流执行动作所需参数满足要求
// 1. Task参数校验
// 2. 回滚参数校验
func validateTplUseParam(kt *kit.Kit, template action.FlowTemplate, opt *AddTemplateFlowOption) error {

	// 校验Action请求参数都已经提供
	m := make(map[action.ActIDType]types.JsonField, len(opt.Tasks))
	for _, one := range opt.Tasks {
		m[one.ActionID] = one.Params
	}

	for _, task := range template.Tasks {
		act, exist := action.GetAction(task.ActionName)
		if !exist {
			return fmt.Errorf("action: %s not exist", task.ActionName)
		}

		// Task 参数校验
		if task.Params != nil && task.Params.Type != nil {
			fields, exist := m[task.ActionID]
			if !exist {
				return fmt.Errorf("action: %s need params", task.ActionName)
			}

			paramAct, ok := act.(action.ParameterAction)
			if !ok {
				return fmt.Errorf("action: %s need params, but not impl ParameterAction", task.ActionName)
			}

			params := paramAct.ParameterNew()
			if err := action.Decode(fields, params); err != nil {
				logs.Errorf("action: %s can not decode params, err: %v, field: %s, type: %T, rid: %s", task.ActionName,
					err, fields, params, kt.Rid)
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
	}

	return nil
}
