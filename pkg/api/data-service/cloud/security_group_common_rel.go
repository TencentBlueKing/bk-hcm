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

// SGCommonRelBatchCreateReq ...
type SGCommonRelBatchCreateReq struct {
	Rels []SGCommonRelCreate `json:"rels" validate:"required"`
}

// SGCommonRelCreate ...
type SGCommonRelCreate struct {
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	ResID           string `json:"res_id" validate:"required"`
	ResType         string `json:"res_type" validate:"omitempty"`
	Priority        int64  `json:"priority" validate:"omitempty"`
}

// Validate security group common rel create request.
func (req *SGCommonRelBatchCreateReq) Validate() error {
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

// SGCommonRelListResp define list resp.
type SGCommonRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SGCommonRelListResult `json:"data"`
}

// SGCommonRelWithSecurityGroupListReq ...
type SGCommonRelWithSecurityGroupListReq struct {
	ResIDs []string `json:"res_ids" validate:"required,min=1"`
}

// Validate SGCommonRelWithSecurityGroupListReq.
func (req *SGCommonRelWithSecurityGroupListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SGCommonRelWithSGListResp define list resp.
type SGCommonRelWithSGListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []corecloud.SGCommonRelWithBaseSecurityGroup `json:"data"`
}
