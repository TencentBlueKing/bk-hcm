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
	"hcm/cmd/task-server/logics/async/backends/iface"
	"hcm/cmd/task-server/logics/async/flow/tpl"
	taskserver "hcm/pkg/api/task-server"
)

type Producer struct {
	backend iface.Backend
}

func (p *Producer) SetBackend(backend iface.Backend) {
	p.backend = backend
}

// AddAsyncTplFlow add async flow
func (p *Producer) AddAsyncTplFlow(req *taskserver.AddFlowReq) (interface{}, error) {
	// 1. 模版是否存在
	if err := req.FlowName.Validate(); err != nil {
		return nil, err
	}

	// 2. 添加任务流到DB
	flowID, err := p.backend.AddFlow(req)
	if err != nil {
		return nil, err
	}

	// 3. 按照模版添加任务集合到DB
	operator := new(tpl.TemplateFlowOperator)
	if err := operator.SetTemplateFlow(req.FlowName); err != nil {
		return nil, err
	}

	if _, err := operator.MakeTemplateFlowTasks(flowID, req, p.backend); err != nil {
		return nil, err
	}

	return flowID, nil
}

// ListAsyncFlow list async flow
func (p *Producer) ListAsyncFlow(req *taskserver.FlowListReq) (interface{}, error) {
	// 1. 按照过滤条件从DB中获取所有任务流
	flows, err := p.backend.GetFlows(req)
	if err != nil {
		return nil, err
	}

	// 2. 根据任务流ID从DB中获取任务信息
	for _, flow := range flows {
		taskResult, err := p.backend.GetTasksByFlowID(flow.ID)
		if err != nil {
			return nil, err
		}

		flow.Tasks = taskResult
	}

	return flows, nil
}

// GetAsyncFlow get async flow
func (p *Producer) GetAsyncFlow(flowID string) (interface{}, error) {
	// 1. 根据任务流ID从DB中获取任务流信息
	flow, err := p.backend.GetFlowByID(flowID)
	if err != nil {
		return nil, err
	}

	// 2. 根据任务流ID从DB中获取任务信息
	taskResult, err := p.backend.GetTasksByFlowID(flowID)
	if err != nil {
		return nil, err
	}

	flow.Tasks = taskResult

	return flow, nil
}
