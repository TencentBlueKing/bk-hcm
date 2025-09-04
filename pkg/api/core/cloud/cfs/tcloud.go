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
// api core
package cfs

import (
	cfs "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cfs/v20190719"

	"hcm/pkg/api/core"
)

// TCloudCfsExtension cfs extension.
type TCloudCfsExtension struct {
	// Tags  文件系统标签列表, 计费使用
	Tags core.TagMap `json:"tags"`
	// FSID 挂载根目录
	FSID string `json:"fsid,omitempty"`
	// IpAddress 挂载点 IP
	IpAddress string `json:"ip_address,omitempty"`
	// PGroupId 权限组 ID. pgroupbasic 是默认权限组.
	// 通过控制查询权限组列表接口获取[DescribeCfsPGroups](https://cloud.tencent.com/document/product/582/38157)
	PGroupId string `json:"p_group_id"`
	// MountInfo 挂载点信息
	MountInfo *cfs.MountInfo `json:"mount_info"`
	// AutoScaleUpRule 文件系统自动扩容策略
	AutoScaleUpRule *cfs.AutoScaleUpRule `json:"auto_scale_up_rule,omitempty"`

	// StorageResourcePkg 文件系统绑定的预付费存储包
	StorageResourcePkg *string `json:"storage_resource_pkg,omitempty"`
	// BandwidthResourcePkg 文件系统绑定的预付费带宽包（暂未支持）
	BandwidthResourcePkg *string `json:"bandwidth_resource_pkg,omitempty"`
	// SnapStatus 文件系统处理快照状态,snapping: 快照中,normal: 正常状态
	SnapStatus *string `json:"snap_status,omitempty"`
	// AutoSnapshotPolicyId 文件系统关联的快照策略
	AutoSnapshotPolicyId *string `json:"auto_snapshot_policy_id,omitempty"`
	// TieringState 文件系统生命周期管理状态; 如: NotAvailable:不可用; Available:可用.
	TieringState *string `json:"tiering_state,omitempty"`
	// TieringDetail 分层存储详情
	TieringDetail *cfs.TieringDetailInfo `json:"tiering_detail,omitempty"`
	// Version 文件系统版本
	Version *string `json:"version,omitempty"`
	//// MetaType 类型; basic: 标准版元数据类型; enhanced: 增项版元数据类型.
	//MetaType *string `json:"meta_type,omitempty"`
	//// ExstraPerformanceInfo 额外性能信息; 注意: 此字段可能返回 null,表示取不到有效值.
	//ExstraPerformanceInfo []*ExstraPerformanceInfo `json:"exstra_performance_info,omitempty"`
}
