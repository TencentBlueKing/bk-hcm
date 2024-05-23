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

// Package asynctask ...
package asynctask

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initialize the async task service.
func InitService(c *capability.Capability) {
	svc := &asyncTaskSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// async task apis in resource
	h.Add("GetFlow", http.MethodGet, "/async_task/flows/{id}", svc.GetFlow)
	h.Add("ListTask", http.MethodGet, "/async_task/flows/{id}/tasks/list", svc.ListTask)

	h.Load(c.WebService)
}

type asyncTaskSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
