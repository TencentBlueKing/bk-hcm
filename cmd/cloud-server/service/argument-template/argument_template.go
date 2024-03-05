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

// Package argstpl ...
package argstpl

import (
	"net/http"

	"hcm/cmd/cloud-server/logics/audit"
	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitArgsTplService initialize the argument template service.
func InitArgsTplService(c *capability.Capability) {
	svc := &argsTplSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
	}
	h := rest.NewHandler()

	// apis in biz
	h.Add("ListBizArgsTpl", http.MethodPost, "/bizs/{bk_biz_id}/argument_templates/list", svc.ListBizArgsTpl)
	h.Add("ListBizArgsTplBindInstanceRule", http.MethodPost, "/bizs/{bk_biz_id}/argument_templates/instance/rule/list",
		svc.ListBizArgsTplBindInstanceRule)
	h.Add("CreateBizArgsTpl", http.MethodPost, "/bizs/{bk_biz_id}/argument_templates/create", svc.CreateBizArgsTpl)
	h.Add("UpdateBizArgsTpl", http.MethodPut, "/bizs/{bk_biz_id}/argument_templates/{id}", svc.UpdateBizArgsTpl)
	h.Add("DeleteBizArgsTpl", http.MethodDelete, "/bizs/{bk_biz_id}/argument_templates/batch", svc.DeleteBizArgsTpl)

	// apis in resource
	h.Add("ListArgsTpl", http.MethodPost, "/argument_templates/list", svc.ListArgsTpl)
	h.Add("ListArgsTplBindInstanceRule", http.MethodPost, "/argument_templates/instance/rule/list",
		svc.ListArgsTplBindInstanceRule)
	h.Add("AssignArgsTplToBiz", http.MethodPost, "/argument_templates/assign/bizs", svc.AssignArgsTplToBiz)
	h.Add("CreateArgsTpl", http.MethodPost, "/argument_templates/create", svc.CreateArgsTpl)

	h.Load(c.WebService)
}

type argsTplSvc struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
	audit      audit.Interface
}
