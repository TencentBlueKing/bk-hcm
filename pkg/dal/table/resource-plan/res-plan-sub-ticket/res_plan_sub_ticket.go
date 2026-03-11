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

// Package resplansubticket ...
package resplansubticket

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResPlanSubTicketColumns defines all the res_plan_sub_ticket table's columns.
var ResPlanSubTicketColumns = utils.MergeColumns(nil, ResPlanSubTicketColumnDescriptor)

// ResPlanSubTicketColumnDescriptor is ResPlanSubTicketTable's column descriptors.
var ResPlanSubTicketColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "ticket_id", NamedC: "ticket_id", Type: enumor.String},
	{Column: "sub_type", NamedC: "sub_type", Type: enumor.String},
	{Column: "sub_demands", NamedC: "sub_demands", Type: enumor.Json},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_biz_name", NamedC: "bk_biz_name", Type: enumor.String},
	{Column: "op_product_id", NamedC: "op_product_id", Type: enumor.Numeric},
	{Column: "op_product_name", NamedC: "op_product_name", Type: enumor.String},
	{Column: "plan_product_id", NamedC: "plan_product_id", Type: enumor.Numeric},
	{Column: "plan_product_name", NamedC: "plan_product_name", Type: enumor.String},
	{Column: "virtual_dept_id", NamedC: "virtual_dept_id", Type: enumor.Numeric},
	{Column: "virtual_dept_name", NamedC: "virtual_dept_name", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "message", NamedC: "message", Type: enumor.String},
	{Column: "stage", NamedC: "stage", Type: enumor.String},
	{Column: "admin_audit_status", NamedC: "admin_audit_status", Type: enumor.String},
	{Column: "admin_audit_operator", NamedC: "admin_audit_operator", Type: enumor.String},
	{Column: "admin_audit_at", NamedC: "admin_audit_at", Type: enumor.String},
	{Column: "operate_info", NamedC: "operate_info", Type: enumor.String},
	{Column: "crp_sn", NamedC: "crp_sn", Type: enumor.String},
	{Column: "crp_url", NamedC: "crp_url", Type: enumor.String},
	{Column: "sub_original_os", NamedC: "sub_original_os", Type: enumor.Numeric},
	{Column: "sub_original_cpu_core", NamedC: "sub_original_cpu_core", Type: enumor.Numeric},
	{Column: "sub_original_memory", NamedC: "sub_original_memory", Type: enumor.Numeric},
	{Column: "sub_original_disk_size", NamedC: "sub_original_disk_size", Type: enumor.Numeric},
	{Column: "sub_updated_os", NamedC: "sub_updated_os", Type: enumor.Numeric},
	{Column: "sub_updated_cpu_core", NamedC: "sub_updated_cpu_core", Type: enumor.Numeric},
	{Column: "sub_updated_memory", NamedC: "sub_updated_memory", Type: enumor.Numeric},
	{Column: "sub_updated_disk_size", NamedC: "sub_updated_disk_size", Type: enumor.Numeric},
	{Column: "submitted_at", NamedC: "submitted_at", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// ResPlanSubTicketTable is used to save resource plan sub_ticket information.
type ResPlanSubTicketTable struct {
	// ID 主键
	ID string `db:"id" json:"id" validate:"lte=64"`
	// TicketID 父单据ID
	TicketID string `db:"ticket_id" json:"ticket_id" validate:"lte=64"`
	// SubType 子单据类型
	SubType enumor.RPTicketType `db:"sub_type" json:"sub_type" validate:"lte=64"`
	// SubDemands 子单据需求列表
	SubDemands types.JsonField `db:"sub_demands" json:"sub_demands"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// BkBizName 业务名称
	BkBizName string `db:"bk_biz_name" json:"bk_biz_name" validate:"lte=64"`
	// OpProductID 运营产品ID
	OpProductID int64 `db:"op_product_id" json:"op_product_id"`
	// OpProductName 运营产品名称
	OpProductName string `db:"op_product_name" json:"op_product_name" validate:"lte=64"`
	// PlanProductID 规划产品ID
	PlanProductID int64 `db:"plan_product_id" json:"plan_product_id"`
	// PlanProductName 规划产品名称
	PlanProductName string `db:"plan_product_name" json:"plan_product_name" validate:"lte=64"`
	// VirtualDeptID 虚拟部门ID
	VirtualDeptID int64 `db:"virtual_dept_id" json:"virtual_dept_id"`
	// VirtualDeptName 虚拟部门名称
	VirtualDeptName string `db:"virtual_dept_name" json:"virtual_dept_name" validate:"lte=64"`
	// Status 子单据状态
	Status enumor.RPSubTicketStatus `db:"status" json:"status" validate:"lte=64"`
	// Message 子单据失败信息
	Message *string `db:"message" json:"message" validate:"omitempty"`
	// Stage 子单据审批阶段
	Stage enumor.RPSubTicketStage `db:"stage" json:"stage" validate:"lte=64"`
	// AdminAuditStatus 管理员审批结果
	AdminAuditStatus enumor.RPAdminAuditStatus `db:"admin_audit_status" json:"admin_audit_status" validate:"lte=64"`
	// AdminAuditOperator 管理员审批过单人
	AdminAuditOperator string `db:"admin_audit_operator" json:"admin_audit_operator"`
	// AdminAuditAt 管理员审批时间
	AdminAuditAt string `db:"admin_audit_at" json:"admin_audit_at"`
	// OperateInfo 管理员审批意见
	OperateInfo *string `db:"operate_info" json:"operate_info" validate:"omitempty"`
	// CrpSN CRP单据ID
	CrpSN string `db:"crp_sn" json:"crp_sn" validate:"lte=64"`
	// CrpURL CRP单据审批链接
	CrpURL string `db:"crp_url" json:"crp_url" validate:"lte=64"`
	// SubOriginalOS 原始OS数
	SubOriginalOS *float64 `db:"sub_original_os" json:"sub_original_os"`
	// SubOriginalCPUCore 原始总CPU核心数
	SubOriginalCPUCore *int64 `db:"sub_original_cpu_core" json:"sub_original_cpu_core"`
	// SubOriginalMemory 原始总内存大小
	SubOriginalMemory *int64 `db:"sub_original_memory" json:"sub_original_memory"`
	// SubOriginalDiskSize 原始总云盘大小
	SubOriginalDiskSize *int64 `db:"sub_original_disk_size" json:"sub_original_disk_size"`
	// SubUpdatedOS 更新后OS数
	SubUpdatedOS *float64 `db:"sub_updated_os" json:"sub_updated_os"`
	// SubUpdatedCPUCore 更新后总CPU核心数
	SubUpdatedCPUCore *int64 `db:"sub_updated_cpu_core" json:"sub_updated_cpu_core"`
	// SubUpdatedMemory 更新后总内存大小
	SubUpdatedMemory *int64 `db:"sub_updated_memory" json:"sub_updated_memory"`
	// SubUpdatedDiskSize 更新总云盘大小
	SubUpdatedDiskSize *int64 `db:"sub_updated_disk_size" json:"sub_updated_disk_size"`
	// SubmittedAt 提单或改单的时间
	SubmittedAt string `db:"submitted_at" json:"submitted_at"`
	// Creator 创建人
	Creator string `db:"creator" json:"creator" validate:"lte=64"`
	// Reviser 更新人
	Reviser string `db:"reviser" json:"reviser" validate:"lte=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// TableName is the ResPlanSubTicketTable's database table name.
func (r ResPlanSubTicketTable) TableName() table.Name {
	return table.ResPlanSubTicketTable
}

// InsertValidate validate resource plan sub_ticket on insertion.
func (r ResPlanSubTicketTable) InsertValidate() error {
	if len(r.ID) == 0 || len(r.TicketID) == 0 {
		return errors.New("id and ticket_id can not be empty")
	}

	if err := r.SubType.Validate(); err != nil {
		return err
	}

	if err := r.Status.Validate(); err != nil {
		return err
	}

	if err := r.Stage.Validate(); err != nil {
		return err
	}

	if len(r.SubDemands) == 0 {
		return errors.New("sub_demands can not be empty")
	}

	if r.BkBizID <= 0 {
		return errors.New("bk biz id should be > 0")
	}

	if len(r.BkBizName) == 0 {
		return errors.New("bk biz name can not be empty")
	}

	if r.OpProductID <= 0 {
		return errors.New("op product id should be > 0")
	}

	if len(r.OpProductName) == 0 {
		return errors.New("op product name can not be empty")
	}

	if r.PlanProductID <= 0 {
		return errors.New("plan product id should be > 0")
	}

	if len(r.PlanProductName) == 0 {
		return errors.New("plan product name can not be empty")
	}

	if r.VirtualDeptID <= 0 {
		return errors.New("virtual dept id should be > 0")
	}

	if len(r.VirtualDeptName) == 0 {
		return errors.New("virtual dept name can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	if len(r.SubmittedAt) == 0 {
		return errors.New("submitted can not be empty")
	}

	return validator.Validate.Struct(r)
}

// UpdateValidate validate resource plan sub_ticket on update.
func (r ResPlanSubTicketTable) UpdateValidate() error {
	// 父单据不可变更
	if r.TicketID != "" {
		return errors.New("ticket_id can not update")
	}

	if r.BkBizID < 0 {
		return errors.New("bk biz id should be >= 0")
	}

	if r.OpProductID < 0 {
		return errors.New("op product id should be >= 0")
	}

	if r.PlanProductID < 0 {
		return errors.New("plan product id should be >= 0")
	}

	if r.VirtualDeptID < 0 {
		return errors.New("virtual dept id should be >= 0")
	}

	if len(r.SubType) > 0 {
		if err := r.SubType.Validate(); err != nil {
			return err
		}
	}

	if len(r.Status) > 0 {
		if err := r.Status.Validate(); err != nil {
			return err
		}
	}

	if len(r.Stage) > 0 {
		if err := r.Stage.Validate(); err != nil {
			return err
		}
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}

// IsInAdminAuditing return true if the sub ticket is in admin auditing
func (r ResPlanSubTicketTable) IsInAdminAuditing() bool {
	if r.Status != enumor.RPSubTicketStatusAuditing {
		return false
	}
	if r.Stage != enumor.RPSubTicketStageAdminAudit {
		return false
	}
	if r.AdminAuditStatus != enumor.RPAdminAuditStatusAuditing {
		return false
	}
	return true
}
