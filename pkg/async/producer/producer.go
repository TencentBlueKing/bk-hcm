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

// Package producer 异步任务生产者
package producer

import (
	"errors"
	"fmt"

	"hcm/pkg/async/action"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/backend/model"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
)

// Producer 异步任务生产者，提供异步任务下发相关功能。。
type Producer interface {
	AddFlow(kt *kit.Kit, opt *AddFlowOption) (id string, err error)
}

var _ Producer = new(producer)

// NewProducer new producer.
func NewProducer(bd backend.Backend, register prometheus.Registerer) (Producer, error) {
	if bd == nil {
		return nil, errors.New("backend is required")
	}

	if register == nil {
		return nil, errors.New("metrics register is required")
	}

	return &producer{
		backend: bd,
		mc:      initMetric(register),
	}, nil
}

// producer ...
type producer struct {
	backend backend.Backend
	mc      *metric
}

// AddFlow add async flow
func (p *producer) AddFlow(kt *kit.Kit, opt *AddFlowOption) (id string, err error) {
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

func buildFlow(tpl action.FlowTemplate, opt *AddFlowOption) *model.Flow {
	flow := &model.Flow{
		Name:      tpl.Name,
		ShareData: tpl.ShareData,
		Memo:      opt.Memo,
		Tasks:     make([]model.Task, 0, len(tpl.Tasks)),
	}

	m := make(map[string]types.JsonField, len(opt.Tasks))
	for _, one := range opt.Tasks {
		m[one.ActionID] = one.Params
	}

	for _, one := range tpl.Tasks {
		flow.Tasks = append(flow.Tasks, model.Task{
			FlowName:   tpl.Name,
			ActionID:   one.ActionID,
			ActionName: one.ActionName,
			Params:     m[one.ActionID],
			CanRetry:   one.CanRetry,
			DependOn:   one.DependOn,
		})
	}

	return flow
}

// validateTplUseParam 校验任务流执行动作所需参数满足要求
func validateTplUseParam(kt *kit.Kit, template action.FlowTemplate, opt *AddFlowOption) error {

	// 校验Action请求参数都已经提供
	m := make(map[string]types.JsonField, len(opt.Tasks))
	for _, one := range opt.Tasks {
		m[one.ActionID] = one.Params
	}

	for _, task := range template.Tasks {
		if !task.NeedParam {
			continue
		}

		fields, exist := m[task.ActionID]
		if !exist {
			return fmt.Errorf("action: %s need params", task.ActionName)
		}

		act, exist := action.GetAction(task.ActionName)
		if !exist {
			return fmt.Errorf("action: %s not exist", task.ActionName)
		}

		paramAct, ok := act.(action.ParameterAction)
		if !ok {
			return fmt.Errorf("action: %s need params, but not impl ParameterAction", task.ActionName)
		}

		params := paramAct.ParameterNew()
		if err := action.Unmarshal(string(fields), params); err != nil {
			logs.Errorf("action: %s can not unmarshal params, err: %v, field: %s, type: %T, rid: %s", task.ActionName,
				err, fields, params, kt.Rid)
			return fmt.Errorf("action: %s can not unmarshal param, err: %v", task.ActionName, err)
		}
	}

	return nil
}
