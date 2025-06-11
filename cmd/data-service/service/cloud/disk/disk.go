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

// Package disk ...
package disk

import (
	"net/http"

	"hcm/cmd/data-service/service/capability"
	"hcm/pkg/dal/dao"
	"hcm/pkg/rest"
)

// InitService ...
func InitService(cap *capability.Capability) {
	svc := &diskSvc{
		dao: cap.Dao,
	}

	h := rest.NewHandler()

	// 批量创建云盘(支持 extension 字段)
	h.Add("BatchCreateDiskExt", http.MethodPost, "/vendors/{vendor}/disks/batch/create", svc.BatchCreateDiskExt)
	// 获取单个云盘
	h.Add("RetrieveDiskExt", http.MethodGet, "/vendors/{vendor}/disks/{id}", svc.RetrieveDiskExt)
	// 查询云盘列表 (不带 extension 字段)
	h.Add("ListDisk", http.MethodPost, "/disks/list", svc.ListDisk)
	// 查询云盘列表 (带 extension 字段)
	h.Add("ListDiskExt", http.MethodPost, "/vendors/{vendor}/disks/list", svc.ListDiskExt)
	// 批量更新云盘数据(支持 extension 字段)
	h.Add("BatchUpdateDiskExt", http.MethodPatch, "/vendors/{vendor}/disks", svc.BatchUpdateDiskExt)
	// 批量更新云盘基础数据
	h.Add("BatchUpdateDisk", http.MethodPatch, "/disks", svc.BatchUpdateDisk)
	h.Add("BatchDeleteDisk", http.MethodDelete, "/disks/batch", svc.BatchDeleteDisk)
	h.Add("CountDisk", http.MethodPost, "/disks/count", svc.CountDisk)

	h.Load(cap.WebService)
}

type diskSvc struct {
	dao dao.Set
}
