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
	"fmt"
	"net/http"

	logicsadmin "hcm/cmd/cloud-server/logics/admin"
	"hcm/cmd/cloud-server/service/capability"
	apisysteminit "hcm/pkg/api/cloud-server/system-init"
	"hcm/pkg/client"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/emicklei/go-restful/v3"
)

// InitAdminService initialize the system init service.
func InitAdminService(c *capability.Capability) {
	svc := &adminService{
		client:      c.ApiClient,
		adminLogics: c.Logics.Admin,
	}

	svc.registerAdminService(c.WebService)
}

func (s *adminService) registerAdminService(c *restful.WebService) {
	adminH := rest.NewHandler()
	adminH.Path("/admin/system/")
	defer adminH.Load(c)

	// 这里注册的接口都无法被webserver访问，只能被系统内部调用，无需鉴权
	adminH.Add("TenantInit", http.MethodPost, "/tenant-init", s.TenantInit)
}

type adminService struct {
	client      *client.ClientSet
	adminLogics logicsadmin.Interface
}

// TenantInit 租户初始化
func (s *adminService) TenantInit(cts *rest.Contexts) (any, error) {
	// 0. 租户表初始化
	result, err := s.adminLogics.TenantInit(cts.Kit)
	if err != nil {
		logs.Errorf("init tenant failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, fmt.Errorf("tenant data init error: %w", err)
	}

	resp := apisysteminit.SystemInitResult{
		TenantInitResult: result,
	}
	return resp, nil
}
