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

// Package recycle ...
package recycle

import (
	recyclerecord "hcm/pkg/api/core/recycle-record"
	"hcm/pkg/criteria/enumor"
)

// ------------------------ Recycle ------------------------

// RecycleResult defines recycle resource result.
type RecycleResult struct {
	TaskID string `json:"task_id"`
}

// -------------------------- List --------------------------

// RecycleRecordListResult defines list recycle record result.
type RecycleRecordListResult struct {
	Count   uint64                        `json:"count"`
	Details []recyclerecord.RecycleRecord `json:"details"`
}

// DiskAttachInfo 磁盘挂载信息
type DiskAttachInfo struct {
	Err         error
	DiskID      string `json:"disk_id,omitempty"`
	CachingType string `json:"caching_type,omitempty"`
	DeviceName  string `json:"device_name,omitempty"`
}

// EipBindInfo eip 绑定信息
type EipBindInfo struct {
	EipID string `json:"eip_id"`
	NicID string `json:"nic_id"`
	Err   error
}

// CvmRecycleDetail Cvm回收详情，会在数据库中保存回收时的快照
type CvmRecycleDetail struct {
	WithDisk  bool              `json:"with_disk"`
	WithEip   bool              `json:"with_eip"`
	DiskList  []*DiskAttachInfo `json:"disk_list"`
	EipList   []*EipBindInfo    `json:"eip_list"`
	FailedAt  enumor.CloudResourceType
	Vendor    enumor.Vendor
	CvmID     string
	AccountID string
}
