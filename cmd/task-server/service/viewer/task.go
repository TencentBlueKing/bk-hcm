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

package viewer

import (
	"hcm/pkg/api/core"
	coreasync "hcm/pkg/api/core/async"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListTask list task.
func (svc *service) ListTask(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	opt := &types.ListOption{
		Fields: req.Fields,
		Filter: req.Filter,
		Page:   req.Page,
	}
	result, err := svc.dao.AsyncFlowTask().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list task failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if req.Page.Count {
		return &ts.ListTaskResult{Count: result.Count}, nil
	}

	tasks := make([]coreasync.AsyncFlowTask, 0, len(result.Details))
	for _, one := range result.Details {
		tasks = append(tasks, convCoreTask(one))
	}

	return &ts.ListTaskResult{Details: tasks}, nil
}

func convCoreTask(one tableasync.AsyncFlowTaskTable) coreasync.AsyncFlowTask {
	return coreasync.AsyncFlowTask{
		ID:         one.ID,
		FlowID:     one.FlowID,
		FlowName:   one.FlowName,
		ActionID:   one.ActionID,
		ActionName: one.ActionName,
		Params:     one.Params,
		Retry:      one.Retry,
		DependOn:   one.DependOn,
		State:      one.State,
		Reason:     one.Reason,
		Revision: core.Revision{
			Creator:   one.Creator,
			Reviser:   one.Reviser,
			CreatedAt: one.CreatedAt.String(),
			UpdatedAt: one.UpdatedAt.String(),
		},
	}
}

// GetTask get task.
func (svc *service) GetTask(cts *rest.Contexts) (interface{}, error) {
	id := cts.PathParameter("id").String()
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page:   core.NewDefaultBasePage(),
	}
	result, err := svc.dao.AsyncFlowTask().List(cts.Kit, opt)
	if err != nil {
		logs.Errorf("list task failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "task: %s not found", id)
	}

	task := convCoreTask(result.Details[0])
	return &task, nil
}
