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

// PermissionTemplateColumns defines all the permission_template table's columns.
var PermissionTemplateColumns = utils.MergeColumns(nil, PermissionTemplateColumnDescriptor)

// PermissionTemplateColumnDescriptor is permission_template's column descriptors.
var PermissionTemplateColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "policy_library_id", NamedC: "policy_library_id", Type: enumor.String},
	{Column: "policy_library_version", NamedC: "policy_library_version", Type: enumor.Numeric},
	{Column: "policy_library_sync_time", NamedC: "policy_library_sync_time", Type: enumor.Time},
	{Column: "policy_document", NamedC: "policy_document", Type: enumor.String},
	{Column: "policy_hash", NamedC: "policy_hash", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// PermissionTemplateTable 权限模板表
type PermissionTemplateTable struct {
	// ID 本地模板ID
	ID string `db:"id" json:"id"`
	// CloudID 云上策略ID
	CloudID string `db:"cloud_id" json:"cloud_id"`
	// Name 模板名称
	Name string `db:"name" json:"name"`
	// AccountID 所属二级账号ID
	AccountID string `db:"account_id" json:"account_id"`
	// PolicyLibraryID 来源权限策略库ID
	PolicyLibraryID *string `db:"policy_library_id" json:"policy_library_id"`
	// PolicyLibraryVersion 权限策略库版本
	PolicyLibraryVersion *int `db:"policy_library_version" json:"policy_library_version"`
	// PolicyLibrarySyncTime 权限策略库同步时间
	PolicyLibrarySyncTime *string `db:"policy_library_sync_time" json:"policy_library_sync_time"`
	// PolicyDocument 策略JSON内容
	PolicyDocument string `db:"policy_document" json:"policy_document"`
	// PolicyHash 策略内容哈希值
	PolicyHash string `db:"policy_hash" json:"policy_hash"`
	// Memo 描述
	Memo *string `db:"memo" json:"memo"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
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

// TableName return permission_template table name.
func (t PermissionTemplateTable) TableName() table.Name {
	return table.PermissionTemplateTable
}

// InsertValidate validate permission_template on insert.
func (t PermissionTemplateTable) InsertValidate() error {
	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(t.PolicyDocument) == 0 {
		return errors.New("policy_document is required")
	}

	if len(t.PolicyHash) == 0 {
		return errors.New("policy_hash is required")
	}

	if len(t.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(t.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(t.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate validate permission_template on update.
func (t PermissionTemplateTable) UpdateValidate() error {
	if len(t.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
