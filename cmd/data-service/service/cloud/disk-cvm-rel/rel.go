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

package diskcvmrel

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService ...
func InitService(cap *capability.Capability) {
	svc := &relSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()
	h.Add("BatchCreate", http.MethodPost, "/disk_cvm_rels/batch/create", svc.BatchCreate)
	h.Add("ListDiskCvmRel", http.MethodPost, "/disk_cvm_rels/list", svc.ListDiskCvmRel)
	h.Add("ListDiskWithoutCvm", http.MethodPost, "/disk_cvm_rels/with/disks/without/cvm/list", svc.ListDiskWithoutCvm)
	h.Add("ListWithDisk", http.MethodPost, "/disk_cvm_rels/with/disks/list", svc.ListWithDisk)
	h.Add("ListWithCvm", http.MethodPost, "/disk_cvm_rels/with/cvms/list", svc.ListWithCvm)
	h.Add("ListWithDiskExt", http.MethodPost, "/vendors/{vendor}/disk_cvm_rels/with/disks/list", svc.ListWithDiskExt)
	h.Add("BatchDelete", http.MethodDelete, "/disk_cvm_rels/batch", svc.BatchDelete)

	h.Load(cap.WebService)
}

type relSvc struct {
	dao dao.Set
}
