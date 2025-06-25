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
	coretask "hcm/pkg/api/core/task"
	"hcm/pkg/api/data-service/task"
	"hcm/pkg/client/common"
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
func (t *TaskDetailClient) List(kt *kit.Kit, req *core.ListReq) (*core.ListResultT[coretask.Detail], error) {
	return common.Request[core.ListReq, core.ListResultT[coretask.Detail]](
		t.client, rest.POST, kt, req, "/task_details/list")
}

// Create task detail.
func (t *TaskDetailClient) Create(kt *kit.Kit, req *task.CreateDetailReq) (*core.BatchCreateResult, error) {
	return common.Request[task.CreateDetailReq, core.BatchCreateResult](
		t.client, rest.POST, kt, req, "/task_details/create")
}

// Update update task detail.
func (t *TaskDetailClient) Update(kt *kit.Kit, req *task.UpdateDetailReq) error {
	return common.RequestNoResp[task.UpdateDetailReq](t.client, rest.PATCH, kt, req, "/task_details/update")
}

// Delete task detail.
func (t *TaskDetailClient) Delete(kt *kit.Kit, req *task.DeleteDetailReq) error {
	return common.RequestNoResp[task.DeleteDetailReq](t.client, rest.DELETE, kt, req, "/task_details/delete")
}

// BatchUpdate 批量更新指定字段为相同的值
func (t *TaskDetailClient) BatchUpdate(kt *kit.Kit,
	req *task.BatchUpdateTaskDetailReq) error {

	return common.RequestNoResp[task.BatchUpdateTaskDetailReq](t.client, rest.PATCH, kt, req,
		"/task_details/update/state_reason")
}
