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
	"hcm/cmd/cloud-server/service/eip/aws"
	"hcm/cmd/cloud-server/service/eip/azure"
	"hcm/cmd/cloud-server/service/eip/gcp"
	"hcm/cmd/cloud-server/service/eip/huawei"
	"hcm/cmd/cloud-server/service/eip/tcloud"
	"hcm/pkg/rest"
)

// InitEipService initialize the eip service.
func InitEipService(c *capability.Capability) {
	svc := &eipSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		tcloud:     tcloud.NewTCloud(c.ApiClient, c.Authorizer, c.Audit),
		aws:        aws.NewAws(c.ApiClient, c.Authorizer, c.Audit),
		azure:      azure.NewAzure(c.ApiClient, c.Authorizer, c.Audit),
		gcp:        gcp.NewGcp(c.ApiClient, c.Authorizer, c.Audit),
		huawei:     huawei.NewHuaWei(c.ApiClient, c.Authorizer, c.Audit),
	}

	h := rest.NewHandler()

	h.Add("ListEip", http.MethodPost, "/eips/list", svc.ListEip)
	h.Add("RetrieveEip", http.MethodGet, "/eips/{id}", svc.RetrieveEip)
	h.Add("AssignEip", http.MethodPost, "/eips/assign/bizs", svc.AssignEip)
	h.Add("BatchDeleteEip", http.MethodDelete, "/eips/batch", svc.BatchDeleteEip)
	h.Add("ListEipExtByCvmID", http.MethodGet, "/vendors/{vendor}/eips/cvms/{cvm_id}", svc.ListEipExtByCvmID)
	h.Add("ListRelEipWithoutCvm", http.MethodPost, "/eip_cvm_rels/with/eips/without/cvm/list",
		svc.ListRelEipWithoutCvm)
	h.Add("AssociateEip", http.MethodPost, "/eips/associate", svc.AssociateEip)
	h.Add("DisassociateEip", http.MethodPost, "/eips/disassociate", svc.DisassociateEip)
	h.Add("CreateEip", http.MethodPost, "/eips/create", svc.CreateEip)

	// eip apis in biz
	h.Add("ListBizEip", http.MethodPost, "/bizs/{bk_biz_id}/eips/list", svc.ListBizEip)
	h.Add("ListBizEipExtByCvmID", http.MethodGet, "/bizs/{bk_biz_id}/vendors/{vendor}/eips/cvms/{cvm_id}",
		svc.ListBizEipExtByCvmID)
	h.Add("ListBizRelEipWithoutCvm", http.MethodPost, "/bizs/{bk_biz_id}/eip_cvm_rels/with/eips/without/cvm/list",
		svc.ListBizRelEipWithoutCvm)
	h.Add("RetrieveBizEip", http.MethodGet, "/bizs/{bk_biz_id}/eips/{id}", svc.RetrieveBizEip)
	h.Add("BatchDeleteBizEip", http.MethodDelete, "/bizs/{bk_biz_id}/eips/batch", svc.BatchDeleteBizEip)
	h.Add("AssociateBizEip", http.MethodPost, "/bizs/{bk_biz_id}/eips/associate", svc.AssociateBizEip)
	h.Add("DisassociateBizEip", http.MethodPost, "/bizs/{bk_biz_id}/eips/disassociate", svc.DisassociateBizEip)
	h.Add("CreateBizEip", http.MethodPost, "/bizs/{bk_biz_id}/eips/create", svc.CreateBizEip)

	h.Load(c.WebService)
}
