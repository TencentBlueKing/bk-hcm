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

package disk

import (
	"time"

	"hcm/pkg/rest"
)

// DiskListResp ...
type DiskListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *DiskListResult `json:"data"`
}

// DiskListResult ...
type DiskListResult struct {
	Count   *uint64       `json:"count,omitempty"`
	Details []*DiskResult `json:"details"`
}

// DiskResult 查询云盘列表时的单条云盘数据
type DiskResult struct {
	ID           string     `json:"id,omitempty"`
	Vendor       string     `json:"vendor,omitempty"`
	AccountID    string     `json:"account_id,omitempty"`
	Name         string     `json:"name,omitempty"`
	BkBizID      int64      `json:"bk_biz_id,omitempty'"`
	CloudID      string     `json:"cloud_id,omitempty"`
	Region       string     `json:"region,omitempty"`
	Zone         string     `json:"zone,omitempty"`
	DiskSize     uint64     `json:"disk_size,omitempty"`
	DiskType     string     `json:"disk_type,omitempty"`
	Status       string     `json:"status,omitempty"`
	IsSystemDisk bool       `json:"is_system_disk"`
	Memo         *string    `json:"memo,omitempty"`
	Creator      string     `json:"creator,omitempty"`
	Reviser      string     `json:"reviser,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

// DiskExtListResp ...
type DiskExtListResp[T DiskExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          *DiskExtListResult[T] `json:"data"`
}

// DiskExtListResult ...
type DiskExtListResult[T DiskExtensionResult] struct {
	Count   *uint64             `json:"count,omitempty"`
	Details []*DiskExtResult[T] `json:"details"`
}

// DiskExtRetrieveResp 返回单个云盘详情
type DiskExtRetrieveResp[T DiskExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          *DiskExtResult[T] `json:"data"`
}

// DiskExtResult 单个云盘时的详情数据
// TODO move to core
type DiskExtResult[T DiskExtensionResult] struct {
	ID           string     `json:"id,omitempty"`
	Vendor       string     `json:"vendor,omitempty"`
	AccountID    string     `json:"account_id,omitempty"`
	Name         string     `json:"name,omitempty"`
	BkBizID      int64      `json:"bk_biz_id,omitempty'"`
	CloudID      string     `json:"cloud_id,omitempty"`
	Region       string     `json:"region,omitempty"`
	Zone         string     `json:"zone,omitempty"`
	DiskSize     uint64     `json:"disk_size,omitempty"`
	DiskType     string     `json:"disk_type,omitempty"`
	Status       string     `json:"status,omitempty"`
	IsSystemDisk bool       `json:"is_system_disk"`
	Memo         *string    `json:"memo,omitempty"`
	Creator      string     `json:"creator,omitempty"`
	Reviser      string     `json:"reviser,omitempty"`
	CreatedAt    *time.Time `json:"created_at,omitempty"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
	Extension    *T         `json:"extension,omitempty"`
}

// DiskCountResp ...
type DiskCountResp struct {
	rest.BaseResp `json:",inline"`
	Data          *DiskCountResult `json:"data"`
}

// DiskCountResult ...
type DiskCountResult struct {
	Count uint64 `json:"count"`
}

// DiskExtensionResult ...
type DiskExtensionResult interface {
	TCloudDiskExtensionResult | AwsDiskExtensionResult | AzureDiskExtensionResult | GcpDiskExtensionResult | HuaWeiDiskExtensionResult
}
