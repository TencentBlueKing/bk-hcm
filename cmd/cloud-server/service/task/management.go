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
	"fmt"

	cloudtask "hcm/pkg/api/cloud-server/task"
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// ListBizTaskManagement list biz task management.
func (svc *service) ListBizTaskManagement(cts *rest.Contexts) (interface{}, error) {
	return svc.listTaskManagement(cts, handler.ListBizAuthRes)
}

func (svc *service) listTaskManagement(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{},
	error) {

	req := new(core.ListReq)
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
	return svc.client.DataService().Global.TaskManagement.List(cts.Kit, listReq)
}

// CancelBizTaskManagement cancel biz task management.
func (svc *service) CancelBizTaskManagement(cts *rest.Contexts) (interface{}, error) {
	return svc.cancelTaskManagement(cts, handler.BizOperateAuth)
}

func (svc *service) cancelTaskManagement(cts *rest.Contexts, validHandler handler.ValidWithAuthHandler) (interface{},
	error) {

	req := new(task.CancelReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	filter, err := tools.And(tools.ContainersExpression("id", req.IDs),
		tools.EqualExpression("state", enumor.TaskManagementRunning))
	if err != nil {
		logs.Errorf("get cancel filter failed, err: %v, req: %v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}
	listReq := &core.ListReq{
		Filter: filter,
		Fields: []string{"id", "bk_biz_id"},
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.client.DataService().Global.TaskManagement.List(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list task management failed, err: %v, req: %+v, rid: %s", err, listReq, cts.Kit.Rid)
		return nil, err
	}
	if len(list.Details) != len(req.IDs) {
		logs.Errorf("task management ids are invalid, req: %+v, count: %d", listReq, len(list.Details))
		return nil, fmt.Errorf("ids(%v) are invalid", req.IDs)
	}

	// validate biz and authorize
	basicInfos := make(map[string]types.CloudResourceBasicInfo, len(list.Details))
	for _, management := range list.Details {
		basicInfos[management.ID] = types.CloudResourceBasicInfo{ID: management.ID, BkBizID: management.BkBizID}
	}
	err = validHandler(cts, &handler.ValidWithAuthOption{Authorizer: svc.authorizer, ResType: meta.TaskManagement,
		Action: meta.Update, BasicInfos: basicInfos})
	if err != nil {
		return nil, err
	}

	if err = svc.client.DataService().Global.TaskManagement.Cancel(cts.Kit, req); err != nil {
		logs.Errorf("cancel task management failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// ListBizTaskManagementState list biz task management state.
func (svc *service) ListBizTaskManagementState(cts *rest.Contexts) (interface{}, error) {
	return svc.listTaskManagementState(cts, handler.ListBizAuthRes)
}

func (svc *service) listTaskManagementState(cts *rest.Contexts, authHandler handler.ListAuthResHandler) (interface{},
	error) {

	req := new(cloudtask.ManagementListStateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	expr, noPermFlag, err := authHandler(cts, &handler.ListAuthResOption{
		Authorizer: svc.authorizer,
		ResType:    meta.TaskManagement,
		Action:     meta.Find,
		Filter:     tools.ContainersExpression("id", req.IDs),
	})
	if err != nil {
		return nil, err
	}
	if noPermFlag {
		return &core.ListResult{Count: 0, Details: make([]interface{}, 0)}, nil
	}

	listReq := &core.ListReq{
		Filter: expr,
		Fields: []string{"id", "state", "flow_ids"},
		Page:   core.NewDefaultBasePage(),
	}
	list, err := svc.client.DataService().Global.TaskManagement.List(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("list task management failed, err: %v, req: %+v, rid: %s", err, listReq, cts.Kit.Rid)
		return nil, err
	}

	details := make([]cloudtask.ManagementState, 0)
	for _, management := range list.Details {
		// 由于任务管理的状态是根据后台协程周期同步更新的，所以可能会存在任务执行完，但是还没有更新最终状态的情况，因此需要刷新下获取最新状态
		state, err := refreshTaskMgmtState(cts.Kit, svc.client, management)
		if err != nil {
			logs.Errorf("refresh task management state failed, err: %v, data: %+v, rid: %s", err, management,
				cts.Kit.Rid)
			return nil, err
		}
		details = append(details, cloudtask.ManagementState{ID: management.ID, State: state})
	}

	return &cloudtask.ManagementListStateResult{Details: details}, nil
}
