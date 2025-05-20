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

// Package cvm ...
package cvm

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/cmd/data-service/service/cloud/logics/cmdb"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

var svc *cvmSvc

// InitService initial the security group service
func InitService(cap *capability.Capability) {
	svc = &cvmSvc{
		dao: cap.Dao,
	}

	svc.cmdbLogics = cmdb.NewCmdbLogics(cap.CmdbClient)

	h := rest.NewHandler()

	h.Add("CreateCvm", http.MethodPost, "/vendors/{vendor}/cvms/batch/create", svc.BatchCreateCvm)
	h.Add("BatchUpdateCvm", http.MethodPatch, "/vendors/{vendor}/cvms/batch/update", svc.BatchUpdateCvm)
	h.Add("GetCvm", http.MethodGet, "/vendors/{vendor}/cvms/{id}", svc.GetCvm)
	h.Add("ListCvm", http.MethodPost, "/cvms/list", svc.ListCvm)
	h.Add("ListCvmExt", http.MethodPost, "/vendors/{vendor}/cvms/list", svc.ListCvmExt)
	h.Add("BatchDeleteCvm", http.MethodDelete, "/cvms/batch", svc.BatchDeleteCvm)
	h.Add("BatchUpdateCvmCommonInfo", http.MethodPatch, "/cvms/common/info/batch/update", svc.BatchUpdateCvmCommonInfo)

	h.Load(cap.WebService)
}

type cvmSvc struct {
	dao        dao.Set
	cmdbLogics *cmdb.CmdbLogics
}
