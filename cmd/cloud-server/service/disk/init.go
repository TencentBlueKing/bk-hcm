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
 */

package disk

import (
	"net/http"

	"hcm/cmd/cloud-server/service/capability"
	"hcm/pkg/rest"
)

// InitDiskService initialize the disk service.
func InitDiskService(c *capability.Capability) {
	svc := &diskSvc{
		client:     c.ApiClient,
		authorizer: c.Authorizer,
		audit:      c.Audit,
		diskLgc:    c.Logics.Disk,
	}

	h := rest.NewHandler()

	h.Add("ListDisk", http.MethodPost, "/disks/list", svc.ListDisk)

	h.Add("AttachDisk", http.MethodPost, "/disks/attach", svc.AttachDisk)
	h.Add("DetachDisk", http.MethodPost, "/disks/detach", svc.DetachDisk)
	h.Add("AssignDisk", http.MethodPost, "/disks/assign/bizs", svc.AssignDisk)

	h.Add("RetrieveDisk", http.MethodGet, "/disks/{id}", svc.RetrieveDisk)
	h.Add("DeleteDisk", http.MethodDelete, "/disks/{id}", svc.DeleteDisk)
	h.Add("CreateDisk", http.MethodPost, "/disks/create", svc.CreateDisk)

	h.Add("ListDiskExtByCvmID", http.MethodGet, "/vendors/{vendor}/disks/cvms/{cvm_id}", svc.ListDiskExtByCvmID)
	h.Add("ListRelWithCvm", http.MethodPost, "/disk_cvm_rels/with/cvms/list", svc.ListRelWithCvm)
	h.Add("ListDiskCvmRel", http.MethodPost, "/disk_cvm_rels/list", svc.ListDiskCvmRel)
	h.Add("ListRelDiskWithoutCvm", http.MethodPost, "/disk_cvm_rels/with/disks/without/cvm/list",
		svc.ListRelDiskWithoutCvm)

	// disk apis in biz
	h.Add("ListBizDisk", http.MethodPost, "/bizs/{bk_biz_id}/disks/list", svc.ListBizDisk)
	h.Add("ListBizRelWithCvm", http.MethodPost, "/bizs/{bk_biz_id}/disk_cvm_rels/with/cvms/list", svc.ListBizRelWithCvm)
	h.Add("ListBizDiskExtByCvmID", http.MethodGet, "/bizs/{bk_biz_id}/vendors/{vendor}/disks/cvms/{cvm_id}",
		svc.ListBizDiskExtByCvmID)
	h.Add("ListBizRelDiskWithoutCvm", http.MethodPost, "/bizs/{bk_biz_id}/disk_cvm_rels/with/disks/without/cvm/list",
		svc.ListBizRelDiskWithoutCvm)
	h.Add("RetrieveBizDisk", http.MethodGet, "/bizs/{bk_biz_id}/disks/{id}", svc.RetrieveBizDisk)
	h.Add("DeleteBizDisk", http.MethodDelete, "/bizs/{bk_biz_id}/disks/{id}", svc.DeleteBizDisk)
	h.Add("AttachBizDisk", http.MethodPost, "/bizs/{bk_biz_id}/disks/attach", svc.AttachBizDisk)
	h.Add("DetachBizDisk", http.MethodPost, "/bizs/{bk_biz_id}/disks/detach", svc.DetachBizDisk)

	// recycle operation related apis
	h.Add("RecycleDisk", http.MethodPost, "/disks/recycle", svc.RecycleDisk)
	h.Add("RecycleBizDisk", http.MethodPost, "/bizs/{bk_biz_id}/disks/recycle", svc.RecycleBizDisk)
	h.Add("RecoverDisk", http.MethodPost, "/disks/recover", svc.RecoverDisk)
	h.Add("RetrieveRecycledDisk", http.MethodGet, "/recycled/disks/{id}", svc.RetrieveRecycledDisk)
	h.Add("BatchDeleteRecycledDisk", http.MethodDelete, "/recycled/disks/batch", svc.BatchDeleteRecycledDisk)

	h.Load(c.WebService)
}
