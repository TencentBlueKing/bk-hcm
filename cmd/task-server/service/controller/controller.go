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

// Package controller 对异步任务的控制
package controller

import (
	"hcm/cmd/task-server/service/capability"
	"hcm/pkg/async/consumer"
	"hcm/pkg/async/producer"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/retry"
)

// Init initial the async service
func Init(cap *capability.Capability) {
	svc := &service{
		pro: cap.Async.GetProducer(),
		csm: cap.Async.GetConsumer(),
	}

	h := rest.NewHandler()

	h.Add("UpdateCustomFlowState", "PATCH", "/custom_flows/state/update", svc.UpdateCustomFlowState)
	h.Add("RetryFlowTask", "PATCH", "/flows/{flow_id}/tasks/{task_id}/retry", svc.RetryFlowTask)
	h.Add("CancelFlow", "POST", "/flows/{flow_id}/cancel", svc.CancelFlow)

	h.Load(cap.WebService)
}

type service struct {
	pro producer.Producer
	csm consumer.Consumer
}

// UpdateCustomFlowState update custom flow state
func (p service) UpdateCustomFlowState(cts *rest.Contexts) (interface{}, error) {
	opt := new(producer.UpdateCustomFlowStateOption)
	if err := cts.DecodeInto(opt); err != nil {
		return nil, err
	}

	if err := opt.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	rty := retry.NewRetryPolicy(consumer.DefRetryCount, consumer.DefRetryRangeMS)
	err := rty.BaseExec(cts.Kit, func() error {
		return p.pro.BatchUpdateCustomFlowState(cts.Kit, opt)
	})
	if err != nil {
		logs.Errorf("taskserver batch update flow state failed, err: %v, opt: %+v, rid: %s", err, opt, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// RetryFlowTask retry flow task, requirements: given flow and task must be `failed` state
func (p service) RetryFlowTask(cts *rest.Contexts) (any, error) {
	flowId := cts.PathParameter("flow_id").String()
	if len(flowId) == 0 {
		return nil, errf.New(errf.InvalidParameter, "flow_id is required")
	}
	taskId := cts.PathParameter("task_id").String()
	if len(taskId) == 0 {
		return nil, errf.New(errf.InvalidParameter, "task_id is required")
	}

	if err := p.pro.RetryFlowTask(cts.Kit, flowId, taskId); err != nil {
		logs.Errorf("task server retry task(%s) failed, err: %v, opt: %+v, rid: %s", taskId, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}

// CancelFlow 取消任务，无条件终止
func (p service) CancelFlow(cts *rest.Contexts) (any, error) {
	// 终止任务
	flowId := cts.PathParameter("flow_id").String()
	if len(flowId) == 0 {
		return nil, errf.New(errf.InvalidParameter, "flow_id is required")
	}
	if err := p.csm.CancelFlow(cts.Kit, flowId); err != nil {
		logs.Errorf("task server terminate flow(%s) failed, err: %v, opt: %+v, rid: %s", flowId, err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
