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

package enumor

const (
	// RecycleStatus is a special status indicating that resource is recycling
	RecycleStatus = "recycling"
	// RecoverStatus is a special status indicating that resource is recovered
	RecoverStatus = "recovered"
)

// RecycleRecordStatus is recycle record status.
type RecycleRecordStatus string

const (
	// WaitingRecycleRecordStatus is a status indicating that resource is waiting to be recycled.
	WaitingRecycleRecordStatus = "wait_recycle"
	// RecycledRecycleRecordStatus is a status indicating that resource is recycled.
	RecycledRecycleRecordStatus = "recycled"
	// RecoverRecycleRecordStatus is a status indicating that resource is recovered.
	RecoverRecycleRecordStatus = "recovered"
	// FailedRecycleRecordStatus is a status indicating that resource recycle failed.
	FailedRecycleRecordStatus = "failed"
)

// RecycleAuditResTypeMap recycle resource audit type to cloud resource type map.
var RecycleAuditResTypeMap = map[AuditResourceType]CloudResourceType{
	CvmAuditResType:  CvmCloudResType,
	DiskAuditResType: DiskCloudResType,
}

// RecycleType 回收类型
type RecycleType string

const (
	// RecycleTypeNormal 没有设置类型则默认为正常类型
	RecycleTypeNormal RecycleType = ""

	// RecycleTypeRelated  关联资源类型，作为关联资源类型不能被操作，大致等价为占位符。
	// 目前主要用于标识disk作为关联资源随cvm回收的类型。
	RecycleTypeRelated RecycleType = "related"
)
