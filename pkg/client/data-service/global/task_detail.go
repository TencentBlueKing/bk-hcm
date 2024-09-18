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

// TaskDetailClient is data service task detail api client.
type TaskDetailClient struct {
	client rest.ClientInterface
}

// NewTaskDetailClient create a new task detail api client.
func NewTaskDetailClient(client rest.ClientInterface) *TaskDetailClient {
	return &TaskDetailClient{
		client: client,
	}
}

// List task detail.
func (t *TaskDetailClient) List(kt *kit.Kit, req *core.ListReq) (*task.ListDetailResult, error) {
	resp := new(task.DetailListResp)

	err := t.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_details/list").
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

// Create task detail.
func (t *TaskDetailClient) Create(kt *kit.Kit, req *task.CreateDetailReq) (*core.BatchCreateResult, error) {
	resp := new(core.BatchCreateResp)

	err := t.client.Post().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_details/create").
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

// Update update task detail.
func (t *TaskDetailClient) Update(kt *kit.Kit, req *task.UpdateDetailReq) error {
	resp := new(rest.BaseResp)

	err := t.client.Patch().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_details/update").
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

// Delete task detail.
func (t *TaskDetailClient) Delete(kt *kit.Kit, req *task.DeleteDetailReq) error {
	resp := new(rest.BaseResp)

	err := t.client.Delete().
		WithContext(kt.Ctx).
		Body(req).
		SubResourcef("/task_details/delete").
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
