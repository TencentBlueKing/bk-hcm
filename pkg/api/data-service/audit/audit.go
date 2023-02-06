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

package audit

import (
	"fmt"

	"hcm/pkg/api/core/audit"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Update Audit --------------------------

// CloudResourceUpdateAuditReq define cloud create audit request when cloud resource update.
type CloudResourceUpdateAuditReq struct {
	Updates []CloudResourceUpdateInfo `json:"updates" validate:"required"`
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
	Deletes []CloudResourceDeleteInfo `json:"deletes" validate:"required"`
}

// Validate cloud create audit request when cloud resource update.
func (req *CloudResourceDeleteAuditReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.Deletes) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("deletes shuold <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// CloudResourceDeleteInfo defines cloud resource deletes info for audit.
type CloudResourceDeleteInfo struct {
	ResType enumor.AuditResourceType `json:"res_type" validate:"required"`
	ResID   string                   `json:"res_id" validate:"required"`
}

// -------------------------- Assign --------------------------

// CloudResourceAssignInfo defines cloud resource updates info for audit.
type CloudResourceAssignInfo struct {
	ResType enumor.AuditResourceType `json:"res_type" validate:"required"`
	ResID   string                   `json:"res_id" validate:"required"`
	BkBizID int64                    `json:"bk_biz_id" validate:"required"`
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
