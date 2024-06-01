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

package sgcomrel

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initial the security group common rel service
func InitService(cap *capability.Capability) {
	svc := &sgComRelSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreate", http.MethodPost, "/security_group_common_rels/batch/create", svc.BatchCreate)
	h.Add("BatchUpsert", http.MethodPost, "/security_group_common_rels/batch/upsert", svc.BatchUpsert)
	h.Add("BatchDelete", http.MethodDelete, "/security_group_common_rels/batch", svc.BatchDelete)
	h.Add("List", http.MethodPost, "/security_group_common_rels/list", svc.List)
	h.Add("ListWithSecurityGroup", http.MethodPost, "/security_group_common_rels/with/security_group/list",
		svc.ListWithSecurityGroup)

	h.Load(cap.WebService)
}

type sgComRelSvc struct {
	dao dao.Set
}
