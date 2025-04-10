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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// DiskColumns ...
var DiskColumns = utils.MergeColumns(nil, DiskColumnDescriptor)

// DiskColumnDescriptor ...
var DiskColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "disk_size", NamedC: "disk_size", Type: enumor.Numeric},
	{Column: "disk_type", NamedC: "disk_type", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "recycle_status", NamedC: "recycle_status", Type: enumor.String},
	{Column: "is_system_disk", NamedC: "is_system_disk", Type: enumor.Boolean},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// DiskModel ...
type DiskModel struct {
	// Disk ID
	ID string `db:"id" json:"id"`
	// Vendor 云厂商
	Vendor string `db:"vendor" json:"vendor"`
	// AccountID 账号ID
	AccountID string `db:"account_id" json:"account_id"`
	// 云上对应的资源 ID
	CloudID string `db:"cloud_id" json:"cloud_id"`
	// 分配到的业务 ID. 如果是 UnassignedBiz, 表示未分配
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// 云盘名
	Name string `db:"name" json:"name"`
	// Region 地域
	Region string `db:"region" json:"region"`
	// 可用区
	Zone string `db:"zone" json:"zone"`
	// 云盘大小
	DiskSize uint64 `db:"disk_size" json:"disk_size"`
	// 云盘类型
	DiskType string `db:"disk_type" json:"disk_type"`
	// 云盘状态
	Status string `db:"status" json:"status"`
	// 云盘回收状态
	RecycleStatus string `db:"recycle_status" json:"recycle_status"`
	// 是否是系统盘
	IsSystemDisk *bool `db:"is_system_disk" json:"is_system_disk"`
	// Memo 备注
	Memo *string `db:"memo" json:"memo"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}
