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
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Create --------------------------

// SGCvmRelBatchCreateReq ...
type SGCvmRelBatchCreateReq struct {
	Rels []SGCvmRelCreate `json:"rels" validate:"required"`
}

// SGCvmRelCreate ...
type SGCvmRelCreate struct {
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	CvmID           string `json:"cvm_id" validate:"required"`
}

// Validate security group create request.
func (req *SGCvmRelBatchCreateReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// SGCvmRelListResult define sg cvm rels list result.
type SGCvmRelListResult struct {
	Count   uint64                          `json:"count,omitempty"`
	Details []corecloud.SecurityGroupCvmRel `json:"details,omitempty"`
}

// SGCvmRelListResp define list resp.
type SGCvmRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SGCvmRelListResult `json:"data"`
}

// SGCvmRelWithSecurityGroupListReq ...
type SGCvmRelWithSecurityGroupListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate SGCvmRelWithSecurityGroupListReq.
func (req *SGCvmRelWithSecurityGroupListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SGCvmRelWithSGListResp define list resp.
type SGCvmRelWithSGListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []corecloud.SGCvmRelWithBaseSecurityGroup `json:"data"`
}
