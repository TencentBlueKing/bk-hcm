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

package task

import (
	cloudtask "hcm/pkg/api/cloud-server/task"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListBizTaskDetail list biz task detail.
func (svc *service) ListBizTaskDetail(cts *rest.Contexts) (interface{}, error) {
	req := new(core.ListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("req is invalid, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	return svc.listTaskDetail(cts, handler.ListBizAuthRes, req)
}

func (svc *service) listTaskDetail(cts *rest.Contexts, authHandler handler.ListAuthResHandler, req *core.ListReq) (
	interface{}, error) {

	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TaskManagement,
		Action:     meta.Find,
		Filter:     req.Filter,
	})
	if err != nil {
		return nil, err
	}
	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Page:   req.Page,
		Fields: req.Fields,
	}
	return svc.client.DataService().Global.TaskDetail.List(cts.Kit, listReq)
}

// CountBizTaskDetailState count biz task detail state.
func (svc *service) CountBizTaskDetailState(cts *rest.Contexts) (interface{}, error) {
	return svc.countTaskDetailState(cts, handler.ListBizAuthRes)
}

func (svc *service) countTaskDetailState(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{},
	error) {

	req := new(cloudtask.DetailStateCountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("req is invalid, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TaskManagement,
		Action:     meta.Find,
		Filter:     tools.ContainersExpression("task_management_id", req.IDs),
	})
	if err != nil {
		return nil, err
	}
	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	countMap := make(map[string]cloudtask.DetailStateSummary, len(req.IDs))
	for _, id := range req.IDs {
		countMap[id] = cloudtask.DetailStateSummary{ID: id}
	}

	listReq := &core.ListReq{
		Filter: expr,
		Fields: []string{"task_management_id", "state"},
		Page:   core.NewDefaultBasePage(),
	}
	for {
		list, err := svc.client.DataService().Global.TaskDetail.List(cts.Kit, listReq)
		if err != nil {
			logs.Errorf("list task detail failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
			return "", err
		}
		for _, detail := range list.Details {
			detailStateCount := countMap[detail.TaskManagementID]

			detailStateCount.Total++
			switch detail.State {
			case enumor.TaskDetailSuccess:
				detailStateCount.Success++
			case enumor.TaskDetailFailed:
				detailStateCount.Failed++
			case enumor.TaskDetailInit:
				detailStateCount.Init++
			case enumor.TaskDetailRunning:
				detailStateCount.Running++
			case enumor.TaskDetailCancel:
				detailStateCount.Cancel++
			}

			countMap[detail.TaskManagementID] = detailStateCount
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}

		listReq.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	details := make([]cloudtask.DetailStateSummary, 0)
	for _, detail := range countMap {
		details = append(details, detail)
	}

	return cloudtask.DetailStateCountResult{Details: details}, nil
}

// ListBizTaskDetailByCond list biz task detail by cond.
func (svc *service) ListBizTaskDetailByCond(cts *rest.Contexts) (interface{}, error) {
	req := new(cloudtask.DetailListByCondReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		logs.Errorf("list biz task management cond req is invalid, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	listReq := &core.ListReq{
		Filter: tools.ContainersExpression("task_management_id", req.TaskManagementIDs),
		Page:   core.NewDefaultBasePage(),
	}
	return svc.listTaskDetail(cts, handler.ListBizAuthRes, listReq)
}
