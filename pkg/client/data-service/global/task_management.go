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

package global

import (
	"hcm/pkg/api/core"
	"hcm/pkg/api/data-service/task"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/rest"
)

// TaskManagementClient is data service task management api client.
type TaskManagementClient struct {
	client rest.ClientInterface
}

// NewTaskManagementClient create a new task management api client.
func NewTaskManagementClient(client rest.ClientInterface) *TaskManagementClient {
	return &TaskManagementClient{
		client: client,
	}
}

// List task management.
func (t *TaskManagementClient) List(kt *kit.Kit, req *core.ListReq) (*task.ListManagementResult, error) {
	resp := new(task.ManagementListResp)

	err := t.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_managements/list").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// Create task management.
func (t *TaskManagementClient) Create(kt *kit.Kit, req *task.CreateManagementReq) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := t.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_managements/create").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return nil, err
	}

	if resp.Code != errf.OK {
		return nil, errf.New(resp.Code, resp.Message)
	}

	return resp.Data, nil
}

// Update update task management.
func (t *TaskManagementClient) Update(kt *kit.Kit, req *task.UpdateManagementReq) error {
	resp := new(rest.BaseResp)

	err := t.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_managements/update").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// Delete task management.
func (t *TaskManagementClient) Delete(kt *kit.Kit, req *task.DeleteManagementReq) error {
	resp := new(rest.BaseResp)

	err := t.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_managements/delete").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}

// Cancel task management.
func (t *TaskManagementClient) Cancel(kt *kit.Kit, req *task.CancelReq) error {
	resp := new(rest.BaseResp)

	err := t.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_managements/cancel").
		WithHeaders(kt.Header()).
		Do().
		Into(resp)
	if err != nil {
		return err
	}

	if resp.Code != errf.OK {
		return errf.New(resp.Code, resp.Message)
	}

	return nil
}
