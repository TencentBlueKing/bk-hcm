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
	"fmt"

	corecloud "hcm/pkg/api/core/cloud"
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// SGCommonRelBatchCreateReq ...
type SGCommonRelBatchCreateReq struct {
	Rels []SGCommonRelCreate `json:"rels" validate:"required"`
}

// SGCommonRelCreate ...
type SGCommonRelCreate struct {
	SecurityGroupID string                   `json:"security_group_id" validate:"required"`
	Vendor          enumor.Vendor            `json:"vendor" validate:"required"`
	ResID           string                   `json:"res_id" validate:"required"`
	ResType         enumor.CloudResourceType `json:"res_type" validate:"omitempty"`
	Priority        int64                    `json:"priority" validate:"omitempty"`
}

// Validate security group common rel create request.
func (req *SGCommonRelBatchCreateReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Upsert --------------------------

// SGCommonRelBatchUpsertReq ...
type SGCommonRelBatchUpsertReq struct {
	Rels      []SGCommonRelCreate       `json:"rels" validate:"required,min=1,dive"`
	DeleteReq *dataproto.BatchDeleteReq `json:"delete_req"`
}

// Validate security group common rel upsert request.
func (req *SGCommonRelBatchUpsertReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// SGCommonRelListResult define sg common rels list result.
type SGCommonRelListResult struct {
	Count   uint64                             `json:"count,omitempty"`
	Details []corecloud.SecurityGroupCommonRel `json:"details,omitempty"`
}

// SGCommonRelWithSecurityGroupListReq ...
type SGCommonRelWithSecurityGroupListReq struct {
	ResIDs  []string                 `json:"res_ids" validate:"required,min=1"`
	ResType enumor.CloudResourceType `json:"res_type" validate:"required"`
}

// Validate SGCommonRelWithSecurityGroupListReq.
func (req *SGCommonRelWithSecurityGroupListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.ResIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("res_ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	if len(req.ResType) == 0 {
		return fmt.Errorf("res_type is required")
	}

	return nil
}
