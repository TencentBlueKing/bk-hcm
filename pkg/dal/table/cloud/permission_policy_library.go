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

package cloud

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// PermissionPolicyLibraryColumns defines all the permission_policy_library table's columns.
var PermissionPolicyLibraryColumns = utils.MergeColumns(nil, PermissionPolicyLibraryColumnDescriptor)

// PermissionPolicyLibraryColumnDescriptor is permission_policy_library's column descriptors.
var PermissionPolicyLibraryColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "policy_document", NamedC: "policy_document", Type: enumor.String},
	{Column: "policy_hash", NamedC: "policy_hash", Type: enumor.String},
	{Column: "version", NamedC: "version", Type: enumor.Numeric},
	{Column: "bk_biz_ids", NamedC: "bk_biz_ids", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// PermissionPolicyLibraryTable 权限策略库表
type PermissionPolicyLibraryTable struct {
	// ID 策略库ID
	ID string `db:"id" json:"id"`
	// Name 策略库名称
	Name string `db:"name" json:"name"`
	// PolicyDocument 当前版本的权限策略JSON内容
	PolicyDocument string `db:"policy_document" json:"policy_document"`
	// PolicyHash 策略内容SHA256哈希值
	PolicyHash string `db:"policy_hash" json:"policy_hash"`
	// Version 当前版本号，从1开始递增
	Version int `db:"version" json:"version"`
	// BkBizIDs 允许使用的业务ID列表
	BkBizIDs types.JsonField `db:"bk_biz_ids" json:"bk_biz_ids"`
	// Memo 策略库描述
	Memo *string `db:"memo" json:"memo"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName return permission_policy_library table name.
func (p PermissionPolicyLibraryTable) TableName() table.Name {
	return table.PermissionPolicyLibraryTable
}

// InsertValidate validate permission_policy_library on insert.
func (p PermissionPolicyLibraryTable) InsertValidate() error {
	if len(p.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(p.Name) == 0 {
		return errors.New("name is required")
	}

	if len(p.PolicyDocument) == 0 {
		return errors.New("policy_document is required")
	}

	if len(p.PolicyHash) == 0 {
		return errors.New("policy_hash is required")
	}

	if p.Version == 0 {
		return errors.New("version is required")
	}

	if len(p.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(p.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(p.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(p.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(p.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate validate permission_policy_library on update.
func (p PermissionPolicyLibraryTable) UpdateValidate() error {
	if len(p.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(p.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
