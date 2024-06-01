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

// Package cert ...
package cert

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitCertService initialize the cvm service.
func InitCertService(c *capability.Capability) {
	svc := &certSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}

	h := rest.NewHandler()

	// cert apis in biz
	h.Add("ListBizCvm", http.MethodPost, "/bizs/{bk_biz_id}/certs/list", svc.ListBizCert)
	h.Add("CreateBizCert", http.MethodPost, "/bizs/{bk_biz_id}/certs/create", svc.CreateBizCert)
	h.Add("DeleteBizCert", http.MethodDelete, "/bizs/{bk_biz_id}/certs/{id}", svc.DeleteBizCert)

	// cert apis in resource
	h.Add("ListCert", http.MethodPost, "/certs/list", svc.ListCert)
	h.Add("AssignCertToBiz", http.MethodPost, "/certs/assign/bizs", svc.AssignCertToBiz)
	h.Add("CreateCert", http.MethodPost, "/certs/create", svc.CreateCert)
	h.Add("DeleteCert", http.MethodDelete, "/certs/{id}", svc.DeleteCert)

	h.Load(c.WebService)
}

type certSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
