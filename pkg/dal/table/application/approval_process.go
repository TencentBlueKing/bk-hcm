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

package application

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ApprovalProcessColumns defines all the approval process table's columns.
var ApprovalProcessColumns = utils.MergeColumns(nil, ApprovalProcessColumnDescriptor)

// ApprovalProcessColumnDescriptor is Approval Process's column descriptors.
var ApprovalProcessColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "application_type", NamedC: "application_type", Type: enumor.String},
	{Column: "service_id", NamedC: "service_id", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
	{Column: "managers", NamedC: "managers", Type: enumor.String},
}

// ApprovalProcessTable 审批流程表
type ApprovalProcessTable struct {
	// ID 申请ID
	ID string `db:"id" json:"id" validate:"max=64"`
	// ApplicationType 申请类型（新增账号、新增CVM等）
	ApplicationType string `db:"application_type" json:"application_type" validate:"max=64"`
	// ServiceID ITSM流程的服务ID
	ServiceID int64 `db:"service_id" json:"service_id" validate:"min=1"`
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
	// Managers 审批人
	Managers string `db:"managers" json:"managers" validate:"max=255"`
}

// TableName return approval process table name.
func (a ApprovalProcessTable) TableName() table.Name {
	return table.ApprovalProcessTable
}

// InsertValidate aws security group rule table when insert
func (a ApprovalProcessTable) InsertValidate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(a.ApplicationType) == 0 {
		return errors.New("application type is required")
	}

	if a.ServiceID <= 0 {
		return errors.New("service id should be gt 0")
	}

	if len(a.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(a.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if len(a.Managers) == 0 {
		return errors.New("managers is required")
	}

	return nil
}

// UpdateValidate approval process table when update
func (a ApprovalProcessTable) UpdateValidate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
