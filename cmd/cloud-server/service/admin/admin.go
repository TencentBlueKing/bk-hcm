/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2025 THL A29 Limited,
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

// Package admin ...
package admin

import (
	"net/http"

	logicsadmin "hcm/cmd/cloud-server/logics/admin"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/rest"
)

// InitAdminService initialize the system init service.
func InitAdminService(c *capability.Capability) {
	svc := &adminService{
		client:      c.ApiClient,
		adminLogics: c.Logics.Admin,
	}

	h := rest.NewHandler()
	h.Add("OtherAccountInit", http.MethodPost, "/admin/system/other_account_init", svc.OtherAccountInit)

	h.Load(c.WebService)
}

type adminService struct {
	client      *client.ClientSet
	adminLogics logicsadmin.Interface
}
