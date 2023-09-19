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

package tpl

import (
	"encoding/json"
	"errors"

	"hcm/cmd/task-server/logics/async/backends"
	"hcm/pkg/api/core/task"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
)

/*
	// 测试任务集合
	tasks := []task.Task{
		{ID: "00000001", ActionName: string(enumor.TestCreateSG), DependOn: []string{},
			State: enumor.TaskPending, TimeoutSecs: 10, FlowID: "1"},
		{ID: "00000002", ActionName: string(enumor.TestCreateSubnet), DependOn: []string{"00000001"},
			State: enumor.TaskPending, TimeoutSecs: 10, FlowID: "1"},
		{ID: "00000003", ActionName: string(enumor.TestCreateVpc), DependOn: []string{"00000001"},
			State: enumor.TaskPending, TimeoutSecs: 10, FlowID: "1"},
		{ID: "00000004", ActionName: string(enumor.TestCreateCvm), DependOn: []string{"00000002", "00000003"},
			State: enumor.TaskPending, TimeoutSecs: 10, FlowID: "1"},
	}
*/

// FirstTest first test
type FirstTest struct{}

func (ft *FirstTest) makeTemplateFlowTasks(flowID string, req *taskserver.AddFlowReq,
	backend backends.Backend) ([]string, error) {

	tasks := []task.AsyncFlowTask{
		{ActionName: string(enumor.TestCreateSG), TimeoutSecs: 10},
		{ActionName: string(enumor.TestCreateSubnet), TimeoutSecs: 10},
		{ActionName: string(enumor.TestCreateVpc), TimeoutSecs: 10},
		{ActionName: string(enumor.TestCreateCvm), TimeoutSecs: 10},
	}

	// 设置参数
	if len(req.Parameters) > 0 {
		var pars task.AddFlowParameters
		if err := json.Unmarshal(req.Parameters, &pars); err != nil {
			return nil, err
		}

		for _, par := range pars.Params {
			for index, task := range tasks {
				if par.ActionName == task.ActionName {
					tasks[index].Params = types.JsonField(par.Param)
				}
			}
		}
	}

	ids, err := backend.MakeTaskIDs(len(tasks))
	if err != nil {
		return nil, err
	}

	if len(tasks) != len(ids) {
		return nil, errors.New("tasks num not equal id num")
	}

	for index := range tasks {
		tasks[index].ID = ids[index]
		tasks[index].FlowID = flowID
		tasks[index].FlowName = string(req.FlowName)
		tasks[index].State = enumor.TaskPending
	}

	// 任务编排
	tasks[0].DependOn = []string{}
	tasks[1].DependOn = []string{tasks[0].ID}
	tasks[2].DependOn = []string{tasks[0].ID}
	tasks[3].DependOn = []string{tasks[1].ID, tasks[2].ID}

	if err := backend.AddTasks(tasks); err != nil {
		return nil, err
	}

	return ids, nil
}
