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

// VpcCvmRelBatchCreateReq batch create vpc cvm rel request.
type VpcCvmRelBatchCreateReq struct {
	Rels []VpcCvmRelCreate `json:"rels" validate:"required"`
}

// VpcCvmRelCreate create vpc cvm rel request.
type VpcCvmRelCreate struct {
	VpcID string `json:"vpc_id" validate:"required"`
	CvmID string `json:"cvm_id" validate:"required"`
}

// Validate VpcCvmRelBatchCreateReq.
func (req *VpcCvmRelBatchCreateReq) Validate() error {
	if len(req.Rels) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("rels count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// VpcCvmRelListResult defines list vpc cvm rel result.
type VpcCvmRelListResult struct {
	Count   uint64                `json:"count,omitempty"`
	Details []corecloud.VpcCvmRel `json:"details,omitempty"`
}

// VpcCvmRelListResp defines list vpc cvm rel resp.
type VpcCvmRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *VpcCvmRelListResult `json:"data"`
}

// VpcCvmRelWithVpcListReq defines vpc cvm rel request.
type VpcCvmRelWithVpcListReq struct {
	CvmIDs []string `json:"cvm_ids" validate:"required"`
}

// Validate VpcCvmRelWithVpcListReq.
func (req *VpcCvmRelWithVpcListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// VpcCvmRelWithVpcListResp define list resp.
type VpcCvmRelWithVpcListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []corecloud.VpcCvmRelWithBaseVpc `json:"data"`
}
