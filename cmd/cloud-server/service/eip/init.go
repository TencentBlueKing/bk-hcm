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

package eip

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/rest"
)

// InitEipService initialize the eip service.
func InitEipService(c *capability.Capability) {
	svc := &eipSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	h.Add("ListEip", http.MethodPost, "/eips/list", svc.ListEip)
	h.Add("RetrieveEip", http.MethodGet, "/eips/{id}", svc.RetrieveEip)
	h.Add("AssignEip", http.MethodPost, "/eips/assign/bizs", svc.AssignEip)
	h.Add("DeleteEip", http.MethodDelete, "/eips/{id}", svc.DeleteEip)

	h.Add("ListEipExtByCvmID", http.MethodGet, "/vendors/{vendor}/eips/cvms/{cvm_id}", svc.ListEipExtByCvmID)

	h.Add("AssociateEip", http.MethodPost, "/vendors/{vendor}/eips/associate", svc.AssociateEip)
	h.Add("DisassociateEip", http.MethodPost, "/vendors/{vendor}/eips/disassociate", svc.DisassociateEip)

	h.Load(c.WebService)
}
