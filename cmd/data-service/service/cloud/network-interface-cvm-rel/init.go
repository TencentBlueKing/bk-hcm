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

package networkcvmrel

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/rest"
)

// InitService ...
func InitService(cap *capability.Capability) {
	svc := &relSvc{
		Set: cap.Dao,
	}
	svc.Init()

	h := rest.NewHandler()
	h.Add("BatchCreate", http.MethodPost, "/network_cvm_rels/batch/create", svc.BatchCreate)
	h.Add("List", http.MethodPost, "/network_cvm_rels/list", svc.List)
	h.Add("ListWithExtension", http.MethodPost, "/vendors/{vendor}/network_cvm_rels/with/interfaces/list",
		svc.ListWithExtension)
	h.Add("BatchDelete", http.MethodDelete, "/network_cvm_rels/batch", svc.BatchDelete)

	h.Load(cap.WebService)
}
