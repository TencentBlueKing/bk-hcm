/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package cos cos table.
package cos

import (
	"errors"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// CosColumns cos table columns.
var CosColumns = utils.MergeColumns(nil, CosColumnsDescriptor)

// CosColumnsDescriptor cos table columns.
var CosColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id"},
	{Column: "cloud_id", NamedC: "cloud_id"},
	{Column: "name", NamedC: "name"},
	{Column: "vendor", NamedC: "vendor"},
	{Column: "account_id", NamedC: "account_id"},
	{Column: "bk_biz_id", NamedC: "bk_biz_id"},
	{Column: "region", NamedC: "region"},
	{Column: "acl", NamedC: "acl"},
	{Column: "grant_full_control", NamedC: "grant_full_control"},
	{Column: "grant_read", NamedC: "grant_read"},
	{Column: "grant_write", NamedC: "grant_write"},
	{Column: "grant_read_acp", NamedC: "grant_read_acp"},
	{Column: "grant_write_acp", NamedC: "grant_write_acp"},
	{Column: "create_bucket_configuration", NamedC: "create_bucket_configuration"},
	{Column: "domain", NamedC: "domain"},
	{Column: "status", NamedC: "status"},
	{Column: "cloud_created_time", NamedC: "cloud_created_time"},
	{Column: "cloud_status_time", NamedC: "cloud_status_time"},
	{Column: "cloud_expired_time", NamedC: "cloud_expired_time"},
	{Column: "sync_time", NamedC: "sync_time"},
	{Column: "tags", NamedC: "tags"},
	{Column: "extension", NamedC: "extension"},
	{Column: "creator", NamedC: "creator"},
	{Column: "reviser", NamedC: "reviser"},
	{Column: "created_at", NamedC: "created_at"},
	{Column: "updated_at", NamedC: "updated_at"},
}

// CosTable cos table.
type CosTable struct {
	ID        string        `db:"id" validate:"lte=64" json:"id"`
	CloudID   string        `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name      string        `db:"name" validate:"lte=255" json:"name"`
	Vendor    enumor.Vendor `db:"vendor" validate:"lte=16"  json:"vendor"`
	AccountID string        `db:"account_id" validate:"lte=64" json:"account_id"`
	BkBizID   int64         `db:"bk_biz_id" json:"bk_biz_id"`
	Region    string        `db:"region" validate:"lte=20" json:"region"`

	ACL                       string          `db:"acl" json:"acl"`
	GrantFullControl          string          `db:"grant_full_control" json:"grant_full_control"`
	GrantRead                 string          `db:"grant_read" json:"grant_read"`
	GrantWrite                string          `db:"grant_write" json:"grant_write"`
	GrantReadACP              string          `db:"grant_read_acp" json:"grant_read_acp"`
	GrantWriteACP             string          `db:"grant_write_acp" json:"grant_write_acp"`
	CreateBucketConfiguration types.JsonField `db:"create_bucket_configuration" json:"create_bucket_configuration"`

	Domain           string          `db:"domain" json:"domain"`
	Status           string          `db:"status" json:"status"`
	CloudCreatedTime string          `db:"cloud_created_time" json:"cloud_created_time"`
	CloudStatusTime  string          `db:"cloud_status_time" json:"cloud_status_time"`
	CloudExpiredTime string          `db:"cloud_expired_time" json:"cloud_expired_time"`
	SyncTime         string          `db:"sync_time" json:"sync_time"`
	Tags             types.StringMap `db:"tags" json:"tags"`
	Extension        types.JsonField `db:"extension" json:"extension"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return load_balancer table name.
func (cos *CosTable) TableName() table.Name {
	return table.LoadBalancerTable
}

// InsertValidate load_balancer table when insert.
func (cos *CosTable) InsertValidate() error {
	if err := validator.Validate.Struct(cos); err != nil {
		return err
	}

	if len(cos.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(cos.Name) == 0 {
		return errors.New("name is required")
	}

	if len(cos.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(cos.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(cos.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate load_balancer table when update.
func (cos *CosTable) UpdateValidate() error {
	if err := validator.Validate.Struct(cos); err != nil {
		return err
	}

	if len(cos.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(cos.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
