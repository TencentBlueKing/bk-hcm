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

package resourceplan

import (
	"errors"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	dtypes "hcm/pkg/dal/dao/types"
	tablers "hcm/pkg/dal/table/resource-plan/res-plan-sub-ticket"
	"hcm/pkg/dal/table/types"
)

// ResPlanSubTicketBatchCreateReq create request
type ResPlanSubTicketBatchCreateReq struct {
	SubTickets []ResPlanSubTicketCreateReq `json:"sub_tickets" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *ResPlanSubTicketBatchCreateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, t := range r.SubTickets {
		if err := t.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanSubTicketCreateReq create request
type ResPlanSubTicketCreateReq struct {
	TicketID            string                    `json:"ticket_id" validate:"required"`
	BkBizID             int64                     `json:"bk_biz_id" validate:"required"`
	BkBizName           string                    `json:"bk_biz_name" validate:"required"`
	OpProductID         int64                     `json:"op_product_id" validate:"required"`
	OpProductName       string                    `json:"op_product_name" validate:"required"`
	PlanProductID       int64                     `json:"plan_product_id" validate:"required"`
	PlanProductName     string                    `json:"plan_product_name" validate:"required"`
	VirtualDeptID       int64                     `json:"virtual_dept_id" validate:"required"`
	VirtualDeptName     string                    `json:"virtual_dept_name" validate:"required"`
	SubType             enumor.RPTicketType       `json:"sub_type" validate:"required"`
	SubDemands          types.JsonField           `json:"sub_demands" validate:"required"`
	Status              enumor.RPSubTicketStatus  `json:"status" validate:"required"`
	Stage               enumor.RPSubTicketStage   `json:"stage" validate:"required"`
	AdminAuditStatus    enumor.RPAdminAuditStatus `json:"admin_audit_status"`
	CrpSN               string                    `json:"crp_sn"`
	CrpURL              string                    `json:"crp_url"`
	SubOriginalOS       *float64                  `json:"sub_original_os"`
	SubOriginalCPUCore  *int64                    `json:"sub_original_cpu_core"`
	SubOriginalMemory   *int64                    `json:"sub_original_memory"`
	SubOriginalDiskSize *int64                    `json:"sub_original_disk_size"`
	SubUpdatedOS        *float64                  `json:"sub_updated_os"`
	SubUpdatedCPUCore   *int64                    `json:"sub_updated_cpu_core"`
	SubUpdatedMemory    *int64                    `json:"sub_updated_memory"`
	SubUpdatedDiskSize  *int64                    `json:"sub_updated_disk_size"`
	SubmittedAt         string                    `json:"submitted_at" validate:"required"`
}

// Validate validate
func (r *ResPlanSubTicketCreateReq) Validate() error {
	if err := r.SubType.Validate(); err != nil {
		return err
	}

	if err := r.Stage.Validate(); err != nil {
		return err
	}

	if err := r.Status.Validate(); err != nil {
		return err
	}

	return validator.Validate.Struct(r)
}

// ResPlanSubTicketBatchUpdateReq batch update request
type ResPlanSubTicketBatchUpdateReq struct {
	SubTickets []ResPlanSubTicketUpdateReq `json:"sub_tickets" validate:"required,min=1,max=100"`
}

// Validate validate
func (r *ResPlanSubTicketBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, c := range r.SubTickets {
		if err := c.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// ResPlanSubTicketUpdateReq batch update request
type ResPlanSubTicketUpdateReq struct {
	ID                  string                    `json:"id" validate:"required"`
	SubType             enumor.RPTicketType       `json:"sub_type"`
	SubDemands          *string                   `json:"sub_demands"`
	BkBizID             int64                     `json:"bk_biz_id"`
	BkBizName           string                    `json:"bk_biz_name"`
	OpProductID         int64                     `json:"op_product_id"`
	OpProductName       string                    `json:"op_product_name"`
	PlanProductID       int64                     `json:"plan_product_id"`
	PlanProductName     string                    `json:"plan_product_name"`
	VirtualDeptID       int64                     `json:"virtual_dept_id"`
	VirtualDeptName     string                    `json:"virtual_dept_name"`
	Status              enumor.RPSubTicketStatus  `json:"status"`
	Message             *string                   `json:"message"`
	Stage               enumor.RPSubTicketStage   `json:"stage"`
	AdminAuditStatus    enumor.RPAdminAuditStatus `json:"admin_audit_status"`
	AdminAuditOperator  string                    `json:"admin_audit_operator"`
	AdminAuditAt        string                    `json:"admin_audit_at"`
	OperateInfo         *string                   `json:"operate_info"`
	CrpSN               string                    `json:"crp_sn"`
	CrpURL              string                    `json:"crp_url"`
	SubOriginalOS       *float64                  `json:"sub_original_os"`
	SubOriginalCPUCore  *int64                    `json:"sub_original_cpu_core"`
	SubOriginalMemory   *int64                    `json:"sub_original_memory"`
	SubOriginalDiskSize *int64                    `json:"sub_original_disk_size"`
	SubUpdatedOS        *float64                  `json:"sub_updated_os"`
	SubUpdatedCPUCore   *int64                    `json:"sub_updated_cpu_core"`
	SubUpdatedMemory    *int64                    `json:"sub_updated_memory"`
	SubUpdatedDiskSize  *int64                    `json:"sub_updated_disk_size"`
}

// Validate validate
func (r *ResPlanSubTicketUpdateReq) Validate() error {
	if len(r.SubType) > 0 {
		if err := r.SubType.Validate(); err != nil {
			return err
		}
	}

	if len(r.Stage) > 0 {
		if err := r.Stage.Validate(); err != nil {
			return err
		}
	}

	if len(r.Status) > 0 {
		if err := r.Status.Validate(); err != nil {
			return err
		}
	}

	if len(r.AdminAuditStatus) > 0 {
		if err := r.AdminAuditStatus.Validate(); err != nil {
			return err
		}
	}

	return validator.Validate.Struct(r)
}

// ResPlanSubTicketStatusUpdateReq update res plan sub ticket status, if update to failed, message is required.
type ResPlanSubTicketStatusUpdateReq struct {
	IDs      []string                 `json:"ids" validate:"omitempty"`
	TicketID string                   `json:"ticket_id" validate:"required"`
	Source   enumor.RPSubTicketStatus `json:"source" validate:"required"`
	Target   enumor.RPSubTicketStatus `json:"target" validate:"required"`
	Message  *string                  `json:"message" validate:"omitempty"`
}

// Validate validate ResPlanSubTicketStatusUpdateReq
func (r ResPlanSubTicketStatusUpdateReq) Validate() error {
	if err := r.Source.Validate(); err != nil {
		return err
	}
	if err := r.Target.Validate(); err != nil {
		return err
	}

	if r.Target == enumor.RPSubTicketStatusFailed && r.Source != enumor.RPSubTicketStatusInvalid {
		if r.Message == nil {
			return errors.New("failed status, message is required")
		}
	}
	return validator.Validate.Struct(r)
}

// ResPlanSubTicketListResult list result
type ResPlanSubTicketListResult dtypes.ListResult[tablers.ResPlanSubTicketTable]

// ResPlanSubTicketListReq list request
type ResPlanSubTicketListReq struct {
	core.ListReq `json:",inline"`
}

// Validate validate
func (r *ResPlanSubTicketListReq) Validate() error {
	return r.ListReq.Validate()
}
