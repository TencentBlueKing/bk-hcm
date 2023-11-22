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

package csselection

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/client"
	"hcm/pkg/iam/auth"
	"hcm/pkg/rest"
)

// InitService initialize the vpc service.
func InitService(c *capability.Capability) {
	svc := &service{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
	}

	h := rest.NewHandler()

	// 方案相关接口
	h.Add("BatchDeleteScheme", http.MethodDelete, "/selections/schemes/batch", svc.BatchDeleteScheme)
	h.Add("CreateScheme", http.MethodPost, "/selections/schemes/create", svc.CreateScheme)
	h.Add("GetScheme", http.MethodGet, "/selections/schemes/{id}", svc.GetScheme)
	h.Add("ListScheme", http.MethodPost, "/selections/schemes/list", svc.ListScheme)
	h.Add("UpdateScheme", http.MethodPatch, "/selections/schemes/{id}", svc.UpdateScheme)

	// 业务类型接口
	h.Add("ListBizType", http.MethodPost, "/selections/biz_types/list", svc.ListBizType)

	// IDC接口
	h.Add("ListIdc", http.MethodPost, "/selections/idcs/list", svc.ListIdc)

	h.Load(c.WebService)
}

type service struct {
	client     *client.ClientSet
	authorizer auth.Authorizer
}
