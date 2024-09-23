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
	return common.Request[core.ListReq, core.ListResultT[coretask.Management]](
		t.client, rest.POST, kt, req, "/task_managements/list")
}

// Create task management.
func (t *TaskManagementClient) Create(kt *kit.Kit, req *task.CreateManagementReq) (*core.BatchCreateResult, error) {
	return common.Request[task.CreateManagementReq, core.BatchCreateResult](
		t.client, rest.POST, kt, req, "/task_managements/create")
}

// Update update task management.
func (t *TaskManagementClient) Update(kt *kit.Kit, req *task.UpdateManagementReq) error {
	return common.RequestNoResp[task.UpdateManagementReq](t.client, rest.PATCH, kt, req, "/task_managements/update")
}

// Delete task management.
func (t *TaskManagementClient) Delete(kt *kit.Kit, req *task.DeleteManagementReq) error {
	return common.RequestNoResp[task.DeleteManagementReq](t.client, rest.DELETE, kt, req, "/task_managements/delete")
}

// Cancel task management.
func (t *TaskManagementClient) Cancel(kt *kit.Kit, req *task.CancelReq) error {
	return common.RequestNoResp[task.CancelReq](t.client, rest.PATCH, kt, req, "/task_managements/cancel")
}
