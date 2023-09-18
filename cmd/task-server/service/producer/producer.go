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
	tsproducer "hcm/cmd/task-server/logics/async/producer"
	"hcm/cmd/task-server/service/capability"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// InitAsyncService initial the async service
func InitAsyncService(cap *capability.Capability) {
	svc := &producer{
		cs:  cap.ApiClient,
		pro: cap.Producer,
	}

	h := rest.NewHandler()

	h.Add("AddAsyncTplFlow", "POST", "/async/flows/tpls/add", svc.AddAsyncTplFlow)
	h.Add("ListAsyncFlow", "POST", "/async/flows/list", svc.ListAsyncFlow)
	h.Add("GetAsyncFlow", "GET", "/async/flows/{flow_id}", svc.GetAsyncFlow)

	h.Load(cap.WebService)
}

type producer struct {
	cs  *client.ClientSet
	pro *tsproducer.Producer
}

// AddAsyncTplFlow add async flow
func (p producer) AddAsyncTplFlow(cts *rest.Contexts) (interface{}, error) {
	p.pro.GetBackend().SetBackendKit(cts.Kit)

	// 1. 解析请求体
	req := new(taskserver.AddFlowReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 2. 按照模板添加异步任务流
	return p.pro.AddAsyncTplFlow(req)
}

// ListAsyncFlow list async flow
func (p producer) ListAsyncFlow(cts *rest.Contexts) (interface{}, error) {
	p.pro.GetBackend().SetBackendKit(cts.Kit)

	// 1. 解析请求体
	req := new(taskserver.FlowListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 2. 列出异步任务流
	return p.pro.ListAsyncFlow(req)
}

// GetAsyncFlow get async flow
func (p producer) GetAsyncFlow(cts *rest.Contexts) (interface{}, error) {
	p.pro.GetBackend().SetBackendKit(cts.Kit)

	// 1. 解析url获取flow_id
	flowID := cts.PathParameter("flow_id").String()

	// 2. 根据flowid获取异步任务流
	return p.pro.GetAsyncFlow(flowID)
}
