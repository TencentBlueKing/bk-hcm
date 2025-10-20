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

// Package cfs 如下:
// cfs handler
package cfs

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/cmd/data-service/service/cloud/logics/cmdb"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// svc cfs service
var svc *cfsSvc

// InitService initial the security group service
func InitService(cap *capability.Capability) {
	svc = &cfsSvc{
		dao: cap.Dao,
	}
	svc.cmdbLogics = cmdb.NewCmdbLogics(cap.CmdbClient)

	h := rest.NewHandler()

	h.Add("CreateCfs", http.MethodPost, "/vendors/{vendor}/cfs/create", svc.CreateCfs)
	h.Add("BatchDeleteCfs", http.MethodDelete, "/vendors/{vendor}/cfs/delete", svc.BatchDeleteCfs)
	h.Add("ListCfsExt", http.MethodPost, "/vendors/{vendor}/cfs/list", svc.ListCfsExt)
	h.Add("GetCfs", http.MethodGet, "/vendors/{vendor}/cfs/{id}", svc.GetCfs)

	h.Load(cap.WebService)
}

type cfsSvc struct {
	dao dao.Set

	cmdbLogics *cmdb.CmdbLogics
}
