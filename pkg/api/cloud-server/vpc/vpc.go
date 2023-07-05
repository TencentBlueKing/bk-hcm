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

package csvpc

import (
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Update --------------------------

// VpcUpdateReq defines update vpc request.
type VpcUpdateReq struct {
	Memo *string `json:"memo" validate:"required"`
}

// Validate VpcUpdateReq.
func (u VpcUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- List --------------------------

// VpcListResult defines list vpc result.
type VpcListResult struct {
	Count   uint64          `json:"count"`
	Details []cloud.BaseVpc `json:"details"`
}

// VpcListResp defines list vpc response.
type VpcListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *VpcListResult `json:"data"`
}

// -------------------------- Relation ------------------------

// AssignVpcToBizReq assign vpcs to biz request.
type AssignVpcToBizReq struct {
	VpcIDs  []string `json:"vpc_ids"`
	BkBizID int64    `json:"bk_biz_id"`
}

// Validate AssignVpcToBizReq.
func (a AssignVpcToBizReq) Validate() error {
	if len(a.VpcIDs) == 0 {
		return errf.New(errf.InvalidParameter, "vpc ids are required")
	}

	if a.BkBizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is required")
	}

	return nil
}

// BindVpcWithCloudAreaReq bind vpcs with bizs request.
type BindVpcWithCloudAreaReq []VpcCloudAreaRelation

// VpcCloudAreaRelation vpc and cloud area relation.
type VpcCloudAreaRelation struct {
	VpcID     string `json:"vpc_id"`
	BkCloudID int64  `json:"bk_cloud_id"`
}

// Validate BindVpcWithCloudAreaReq.
func (b BindVpcWithCloudAreaReq) Validate() error {
	if len(b) == 0 {
		return errf.New(errf.InvalidParameter, "bind vpc with cloud area request can not be empty")
	}

	for _, relation := range b {
		if len(relation.VpcID) == 0 {
			return errf.New(errf.InvalidParameter, "vpc id is required")
		}

		if relation.BkCloudID == 0 {
			return errf.New(errf.InvalidParameter, "cloud id is required")
		}
	}

	return nil
}
