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

// Package eip ...
package eip

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitEipService ...
func InitEipService(cap *capability.Capability) {
	svc := &eipSvc{dao: cap.Dao}

	h := rest.NewHandler()

	h.Add("BatchCreateEipExt", http.MethodPost, "/vendors/{vendor}/eips/batch/create", svc.BatchCreateEipExt)
	h.Add("RetrieveEipExt", http.MethodGet, "/vendors/{vendor}/eips/{id}", svc.RetrieveEipExt)
	h.Add("ListEip", http.MethodPost, "/eips/list", svc.ListEip)
	h.Add("ListEipExt", http.MethodPost, "/vendors/{vendor}/eips/list", svc.ListEipExt)
	h.Add("BatchUpdateEipExt", http.MethodPatch, "/vendors/{vendor}/eips", svc.BatchUpdateEipExt)
	h.Add("BatchUpdateEip", http.MethodPatch, "/eips", svc.BatchUpdateEip)
	h.Add("BatchDeleteEip", http.MethodDelete, "/eips/batch", svc.BatchDeleteEip)

	h.Load(cap.WebService)
}

type eipSvc struct {
	dao dao.Set
}
