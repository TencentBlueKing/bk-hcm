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

package cloudserver

import (
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Update --------------------------

// SubnetUpdateReq defines update subnet request.
type SubnetUpdateReq struct {
	Memo *string `json:"memo" validate:"required"`
}

// Validate SubnetUpdateReq.
func (u SubnetUpdateReq) Validate() error {
	return validator.Validate.Struct(u)
}

// -------------------------- List --------------------------

// SubnetListResult defines list subnet result.
type SubnetListResult struct {
	Count   uint64             `json:"count"`
	Details []cloud.BaseSubnet `json:"details"`
}

// -------------------------- Relation ------------------------

// AssignSubnetToBizReq assign subnets to biz request.
type AssignSubnetToBizReq struct {
	SubnetIDs []string `json:"subnet_ids"`
	BkBizID   int64    `json:"bk_biz_id"`
}

// Validate AssignSubnetToBizReq.
func (a AssignSubnetToBizReq) Validate() error {
	if len(a.SubnetIDs) == 0 {
		return errf.New(errf.InvalidParameter, "subnet ids are required")
	}

	if a.BkBizID == 0 {
		return errf.New(errf.InvalidParameter, "biz id is required")
	}

	return nil
}
