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

// Package permissionpolicylibrary defines permission policy library service for cloud server.
package permissionpolicylibrary

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initialize the permission policy library service.
func InitService(c *capability.Capability) {
	svc := &svc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("CreatePermissionPolicyLibrary", http.MethodPost,
		"/vendors/{vendor}/permission_policy_libraries/create", svc.CreatePermissionPolicyLibrary)
	h.Add("ListPermissionPolicyLibrary", http.MethodPost,
		"/vendors/{vendor}/permission_policy_libraries/list", svc.ListPermissionPolicyLibrary)
	h.Add("UpdatePermissionPolicyLibrary", http.MethodPatch,
		"/vendors/{vendor}/permission_policy_libraries/{id}", svc.UpdatePermissionPolicyLibrary)
	h.Add("DeletePermissionPolicyLibrary", http.MethodDelete,
		"/vendors/{vendor}/permission_policy_libraries/{id}", svc.DeletePermissionPolicyLibrary)

	h.Load(c.WebService)
}

type svc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
