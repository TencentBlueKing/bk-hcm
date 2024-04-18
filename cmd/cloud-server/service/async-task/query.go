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

package asynctask

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/hooks/handler"
)

// GetFlow 根据异步任务FlowID，获取异步任务Flow详情.
func (svc *asyncTaskSvc) GetFlow(cts *rest.Contexts) (any, error) {
	return svc.getFlow(cts, handler.ListResourceAuthRes)
}

func (svc *asyncTaskSvc) getFlow(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	flowInfo, err := svc.client.TaskServer().GetFlow(cts.Kit, id)
	if err != nil {
		logs.Errorf("fail to call task-server get flow info, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	// 仅支持查询负载均衡的Flow信息
	if err = flowInfo.Name.ValidateLoadBalancer(); err != nil {
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get load balancer")
	}

	flowInfo.Worker = nil
	flowInfo.Memo = ""
	flowInfo.ShareData = nil

	return flowInfo, nil
}

// ListTask 根据异步任务FlowID，获取异步任务Task列表.
func (svc *asyncTaskSvc) ListTask(cts *rest.Contexts) (any, error) {
	return svc.listTask(cts, handler.ListResourceAuthRes)
}

func (svc *asyncTaskSvc) listTask(cts *rest.Contexts, validHandler handler.ListAuthResHandler) (any, error) {
	id := cts.PathParameter("id").String()
	if len(id) == 0 {
		return nil, errf.New(errf.InvalidParameter, "id is required")
	}

	flowInfo, err := svc.client.TaskServer().GetFlow(cts.Kit, id)
	if err != nil {
		logs.Errorf("fail to call task-server get flow info, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	// 仅支持查询负载均衡的Flow信息
	if err = flowInfo.Name.ValidateLoadBalancer(); err != nil {
		return nil, err
	}

	// validate biz and authorize
	_, noPerm, err := validHandler(cts,
		&handler.ListAuthResOption{Authorizer: svc.authorizer, ResType: meta.LoadBalancer, Action: meta.Find})
	if err != nil {
		return nil, err
	}
	if noPerm {
		return nil, errf.New(errf.PermissionDenied, "permission denied for get load balancer")
	}

	taskReq := &core.ListReq{
		Filter: tools.EqualExpression("flow_id", id),
		Page:   core.NewDefaultBasePage(),
	}
	taskList, err := svc.client.TaskServer().ListTask(cts.Kit, taskReq)
	if err != nil {
		logs.Errorf("fail to call task-server get task list, err: %v, id: %s, rid: %s", err, id, cts.Kit.Rid)
		return nil, err
	}

	return taskList, nil
}
