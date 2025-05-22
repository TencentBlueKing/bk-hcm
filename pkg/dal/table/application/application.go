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

// Package application ...
package application

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ApplicationColumns defines all the application table's columns.
var ApplicationColumns = utils.MergeColumns(nil, ApplicationColumnDescriptor)

// ApplicationColumnDescriptor is Application's column descriptors.
var ApplicationColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "source", NamedC: "source", Type: enumor.String},
	{Column: "sn", NamedC: "sn", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "bk_biz_ids", NamedC: "bk_biz_ids", Type: enumor.Json},

	{Column: "applicant", NamedC: "applicant", Type: enumor.String},
	{Column: "content", NamedC: "content", Type: enumor.Json},
	{Column: "delivery_detail", NamedC: "delivery_detail", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ApplicationTable 申请表
type ApplicationTable struct {
	// ID 申请ID
	ID string `db:"id" json:"id" validate:"max=64"`
	// Source 单据来源
	Source string `db:"source" json:"source" validate:"max=64"`
	// SN 单据号
	SN string `db:"sn" json:"sn" validate:"max=64"`
	// Type 申请类型（新增账号、新增CVM等）
	Type string `db:"type" json:"type" validate:"max=64"`
	// Status 单据状态
	Status string `db:"status" json:"status" validate:"max=32"`
	// BkBizID 业务ID
	BkBizIDs types.Int64Array `db:"bk_biz_ids" json:"bk_biz_ids"`

	// Applicant 申请人
	Applicant string `db:"applicant" json:"applicant" validate:"max=64"`
	// Content 申请的内容，不同类型的申请单，内容不一样
	Content types.JsonField `db:"content" json:"content"`
	// DeliveryDetail 交付细节，主要是包括一些交付资源ID
	DeliveryDetail types.JsonField `db:"delivery_detail" json:"delivery_detail"`
	// Memo 备注或申请理由
	Memo *string `db:"memo" json:"memo" validate:"omitempty,max=255"`

	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator" validate:"max=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at" validate:"excluded_unless"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// TableName return application table name.
func (a ApplicationTable) TableName() table.Name {
	return table.ApplicationTable
}

// InsertValidate aws security group rule table when insert
func (a ApplicationTable) InsertValidate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(a.SN) == 0 {
		return errors.New("sn is required")
	}

	if len(a.Type) == 0 {
		return errors.New("type is required")
	}

	if len(a.Status) == 0 {
		return errors.New("status is required")
	}

	if len(a.Applicant) == 0 {
		return errors.New("applicant is required")
	}

	if len(a.Content) == 0 {
		return errors.New("content is required")
	}

	if len(a.DeliveryDetail) == 0 {
		return errors.New("delivery_detail is required")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(a.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate application table when update
func (a ApplicationTable) UpdateValidate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
