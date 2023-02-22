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

package vpccvmrel

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService initialize the vpc and cvm relation service
func InitService(cap *capability.Capability) {
	svc := &vpcCvmRelSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	h.Add("BatchCreate", http.MethodPost, "/vpc_cvm_rels/batch/create", svc.BatchCreate)
	h.Add("BatchDelete", http.MethodDelete, "/vpc_cvm_rels/batch", svc.BatchDelete)
	h.Add("List", http.MethodPost, "/vpc_cvm_rels/list", svc.List)
	h.Add("ListWithVpc", http.MethodPost, "/vpc_cvm_rels/with/vpc/list", svc.ListWithVpc)

	h.Load(cap.WebService)
}

type vpcCvmRelSvc struct {
	dao dao.Set
}
