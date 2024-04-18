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

// Package producer 生产者相关接口
package producer

import (
	"hcm/cmd/task-server/service/capability"
	"hcm/pkg/api/core"
	"hcm/pkg/async/producer"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// Init initial the async service
func Init(cap *capability.Capability) {
	svc := &service{
		cs:  cap.ApiClient,
		pro: cap.Async.GetProducer(),
	}

	h := rest.NewHandler()

	h.Add("CreateTemplateFlow", "POST", "/template_flows/create", svc.CreateTemplateFlow)
	h.Add("CreateCustomFlow", "POST", "/custom_flows/create", svc.CreateCustomFlow)
	h.Add("CloneFlow", "POST", "/flows/{flow_id}/clone", svc.CloneFlow)

	h.Load(cap.WebService)
}

type service struct {
	cs  *client.ClientSet
	pro producer.Producer
}

// CreateTemplateFlow add async flow
func (p service) CreateTemplateFlow(cts *rest.Contexts) (interface{}, error) {

	// 1. 解析请求体。
	// 请求体使用的是 taskserver.AddTemplateFlowReq，但解析使用的是 producer.AddTemplateFlowOption，是想通过http请求去自动序列化
	// task.Params，而不需要手动 Marshal 请求参数。
	opt := new(producer.AddTemplateFlowOption)
	if err := cts.DecodeInto(opt); err != nil {
		return nil, err
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 2. 按照模板添加异步任务流
	id, err := p.pro.AddTemplateFlow(cts.Kit, opt)
	if err != nil {
		logs.Errorf("add template flow failed, err: %v, opt: %+v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	return &core.CreateResult{ID: id}, nil
}

// CreateCustomFlow add custom flow
func (p service) CreateCustomFlow(cts *rest.Contexts) (interface{}, error) {

	// 1. 解析请求体
	// 请求体使用的是 taskserver.AddCustomFlowReq，但解析使用的是 producer.AddCustomFlowOption，是想通过http请求去自动序列化
	// task.Params，而不需要手动 Marshal 请求参数。
	opt := new(producer.AddCustomFlowOption)
	if err := cts.DecodeInto(opt); err != nil {
		return nil, err
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 2. 添加异步任务流
	id, err := p.pro.AddCustomFlow(cts.Kit, opt)
	if err != nil {
		logs.Errorf("add custom flow failed, err: %v, opt: %+v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	return &core.CreateResult{ID: id}, nil
}

// CloneFlow 按原参数重新发起一次任务
func (p service) CloneFlow(cts *rest.Contexts) (any, error) {
	flowId := cts.PathParameter("flow_id").String()
	if len(flowId) == 0 {
		return nil, errf.New(errf.InvalidParameter, "flow_id is required")
	}

	opt := new(producer.CloneFlowOption)
	if err := cts.DecodeInto(opt); err != nil {
		return nil, err
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 复制一份flow 和task
	id, err := p.pro.CloneFlow(cts.Kit, flowId, opt)
	if err != nil {
		logs.Errorf("fail to clone flow(%s), err: %v, rid: %s", flowId, err, cts.Kit.Rid)
		return nil, err
	}

	return &core.CreateResult{ID: id}, nil
}
