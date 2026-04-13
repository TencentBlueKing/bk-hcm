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

// Package audit ...
package audit

import (
	"fmt"

	coreasync "hcm/pkg/api/core/async"
	"hcm/pkg/api/core/audit"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Update Audit --------------------------

// CloudResourceUpdateAuditReq define cloud create audit request when cloud resource update.
type CloudResourceUpdateAuditReq struct {
	ParentID string                    `json:"parent_id" validate:"omitempty"`
	Updates  []CloudResourceUpdateInfo `json:"updates" validate:"required"`
}

// Validate cloud create audit request when cloud resource update.
func (req *CloudResourceUpdateAuditReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Updates) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("updates shuold <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// CloudResourceUpdateInfo defines cloud resource updates info for audit.
type CloudResourceUpdateInfo struct {
	ResType      enumor.AuditResourceType `json:"res_type" validate:"required"`
	ResID        string                   `json:"res_id" validate:"required"`
	UpdateFields map[string]interface{}   `json:"update_fields" validate:"required"`
}

// -------------------------- Delete Audit --------------------------

// CloudResourceDeleteAuditReq define cloud create audit request when cloud resource delete.
type CloudResourceDeleteAuditReq struct {
	ParentID string                    `json:"parent_id" validate:"omitempty"`
	Deletes  []CloudResourceDeleteInfo `json:"deletes" validate:"required"`
}

// Validate cloud create audit request when cloud resource update.
func (req *CloudResourceDeleteAuditReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CloudResourceDeleteInfo defines cloud resource deletes info for audit.
type CloudResourceDeleteInfo struct {
	ResType enumor.AuditResourceType `json:"res_type" validate:"required"`
	ResID   string                   `json:"res_id" validate:"required"`
}

// -------------------------- Assign --------------------------

// CloudResourceAssignInfo defines cloud resource updates info for audit.
type CloudResourceAssignInfo struct {
	ResType         enumor.AuditResourceType    `json:"res_type" validate:"required"`
	ResID           string                      `json:"res_id" validate:"required"`
	AssignedResType enumor.AuditAssignedResType `json:"assigned_res_type" validate:"required"`
	AssignedResID   int64                       `json:"assigned_res_id" validate:"required"`
}

// CloudResourceAssignAuditReq cloud resource assign audit request.
type CloudResourceAssignAuditReq struct {
	Assigns []CloudResourceAssignInfo `json:"assigns" validate:"required"`
}

// Validate cloud create audit request when cloud resource assign.
func (req *CloudResourceAssignAuditReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Assigns) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("assign shuold <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// CloudResourceOperationInfo define cloud resource operation info.
type CloudResourceOperationInfo struct {
	ResType           enumor.AuditResourceType `json:"res_type" validate:"required"`
	ResID             string                   `json:"res_id" validate:"required"`
	Action            OperationAction          `json:"action" validate:"required"`
	AssociatedResType enumor.AuditResourceType `json:"associated_res_type" validate:"omitempty"`
	AssociatedResID   string                   `json:"associated_res_id" validate:"omitempty"`
}

// OperationAction define operation action.
type OperationAction string

// ConvAuditAction conv audit action from operation action.
func (o *OperationAction) ConvAuditAction() (enumor.AuditAction, error) {
	switch *o {
	case Start:
		return enumor.Start, nil
	case Stop:
		return enumor.Stop, nil
	case Reboot:
		return enumor.Reboot, nil
	case ResetPwd:
		return enumor.ResetPwd, nil
	case Associate:
		return enumor.Associate, nil
	case Disassociate:
		return enumor.Disassociate, nil
	case ApplyOp:
		return enumor.Apply, nil

	default:
		return "", fmt.Errorf("action is not corresponding audit action")
	}
}

// OperationAction 操作动作
const (
	Start    OperationAction = "start"
	Stop     OperationAction = "stop"
	Reboot   OperationAction = "reboot"
	ResetPwd OperationAction = "reset_pwd"
	// Associate 绑定、挂载等操作
	Associate OperationAction = "associate"
	// Disassociate 解绑、解挂载等操作
	Disassociate OperationAction = "disassociate"
	// ApplyOp 应用操作
	ApplyOp OperationAction = "apply"
)

// CloudResourceOperationAuditReq define cloud resource operation audit req.
type CloudResourceOperationAuditReq struct {
	Operations []CloudResourceOperationInfo `json:"operations" validate:"required"`
}

// Validate cloud create audit request when cloud resource operate.
func (req *CloudResourceOperationAuditReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Operations) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("assign shuold <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// -------------------------- Recycle Audit --------------------------

// CloudResourceRecycleAuditReq defines create cloud resource recycle audit request.
type CloudResourceRecycleAuditReq struct {
	ResType enumor.AuditResourceType   `json:"res_type" validate:"required"`
	Action  RecycleAction              `json:"action" validate:"required"`
	Infos   []CloudResRecycleAuditInfo `json:"infos" validate:"min=1,max=100"`
}

// CloudResRecycleAuditInfo defines create cloud resource recycle audit info.
type CloudResRecycleAuditInfo struct {
	ResID string      `json:"res_id" validate:"required"`
	Data  interface{} `json:"data" validate:"required"`
}

// Validate CloudResourceRecycleAuditReq.
func (r *CloudResourceRecycleAuditReq) Validate() error {
	return validator.Validate.Struct(r)
}

// RecycleAction define recycle action.
type RecycleAction string

// ConvAuditAction conv audit action from recycle action.
func (r *RecycleAction) ConvAuditAction() (enumor.AuditAction, error) {
	switch *r {
	case Recycle:
		return enumor.Recycle, nil
	case Recover:
		return enumor.Recover, nil
	default:
		return "", fmt.Errorf("action has no corresponding audit action")
	}
}

// RecycleAction 回收动作
const (
	Recycle RecycleAction = "recycle"
	Recover RecycleAction = "recover"
)

// -------------------------- List --------------------------

// ListResp defines list audit response.
type ListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListResult `json:"data"`
}

// ListResult defines list audit result.
type ListResult struct {
	Count   uint64        `json:"count"`
	Details []audit.Audit `json:"details"`
}

// -------------------------- Get --------------------------

// GetResp defines get audit response.
type GetResp struct {
	rest.BaseResp `json:",inline"`
	Data          *audit.Audit `json:"data"`
}

// -------------------------- Get Audit Async Task --------------------------

// GetAsyncTaskResp defines get audit async task response.
type GetAsyncTaskResp struct {
	Flow  *coreasync.AsyncFlow      `json:"flow"`
	Tasks []coreasync.AsyncFlowTask `json:"tasks"`
}

// -------------------------- Batch Create Audit --------------------------

// BatchCreateAuditInfo defines a single audit record for batch creation.
// This allows cloud-server to construct complete audit records with all fields (including bk_biz_id)
// and submit them directly, bypassing the limitation of existing audit interfaces that cannot carry bk_biz_id.
type BatchCreateAuditInfo struct {
	ResID      string                   `json:"res_id" validate:"required,lte=64"`
	CloudResID string                   `json:"cloud_res_id" validate:"omitempty,lte=255"`
	ResName    string                   `json:"res_name" validate:"omitempty,lte=255"`
	ResType    enumor.AuditResourceType `json:"res_type" validate:"required"`
	Action     enumor.AuditAction       `json:"action" validate:"required"`
	BkBizID    int64                    `json:"bk_biz_id"`
	Vendor     enumor.Vendor            `json:"vendor" validate:"omitempty,lte=16"`
	AccountID  string                   `json:"account_id" validate:"omitempty,lte=64"`
	Detail     interface{}              `json:"detail,omitempty"`
}

// BatchCreateAuditReq defines the request for batch creating audit records directly.
type BatchCreateAuditReq struct {
	Audits []BatchCreateAuditInfo `json:"audits" validate:"required,min=1,max=100"`
}

// Validate BatchCreateAuditReq.
func (r *BatchCreateAuditReq) Validate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	for _, a := range r.Audits {
		if !a.ResType.Exist() {
			return fmt.Errorf("res_type %s not supported", a.ResType)
		}
		if !a.Action.Exist() {
			return fmt.Errorf("action %s not supported", a.Action)
		}
	}

	return nil
}
