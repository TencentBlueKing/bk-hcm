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

// Package viewer 查看异步任务
package viewer

import (
	"hcm/cmd/task-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// Init initial the async service
func Init(cap *capability.Capability) {
	svc := &service{
		cs:  cap.ApiClient,
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("ListFlow", "POST", "/flows/list", svc.ListFlow)
	h.Add("GetFlow", "GET", "/flows/{id}", svc.GetFlow)
	h.Add("ListTask", "POST", "/tasks/list", svc.ListTask)
	h.Add("GetTask", "GET", "/tasks/{id}", svc.GetTask)

	h.Load(cap.WebService)
}

type service struct {
	cs  *client.ClientSet
	dao dao.Set
}
