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
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initial the task service
func InitService(cap *capability.Capability) {
	svc := &service{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("CreateTaskManagement", http.MethodPost, "/task_managements/create", svc.CreateTaskManagement)
	h.Add("DeleteTaskManagement", http.MethodDelete, "/task_managements/delete", svc.DeleteTaskManagement)
	h.Add("UpdateTaskManagement", http.MethodPatch, "/task_managements/update", svc.UpdateTaskManagement)
	h.Add("ListTaskManagement", http.MethodPost, "/task_managements/list", svc.ListTaskManagement)
	h.Add("CancelTaskManagement", http.MethodPatch, "/task_managements/cancel", svc.CancelTaskManagement)

	h.Add("CreateTaskDetail", http.MethodPost, "/task_details/create", svc.CreateTaskDetail)
	h.Add("UpdateTaskDetail", http.MethodPatch, "/task_details/update", svc.UpdateTaskDetail)
	h.Add("DeleteTaskDetail", http.MethodDelete, "/task_details/delete", svc.DeleteTaskDetail)
	h.Add("ListTaskDetail", http.MethodPost, "/task_details/list", svc.ListTaskDetail)

	h.Load(cap.WebService)
}

type service struct {
	dao dao.Set
}
