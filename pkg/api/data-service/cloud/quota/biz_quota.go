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

package dsquota

import (
	corequota "hcm/pkg/api/core/cloud/quota"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	tablequota "hcm/pkg/dal/table/cloud/quota"
)

// CreateBizQuotaReq define create biz quota req.
type CreateBizQuotaReq struct {
	CloudQuotaID string                 `json:"cloud_quota_id" validate:"required"`
	Vendor       enumor.Vendor          `json:"vendor" validate:"required"`
	ResType      enumor.BizQuotaResType `json:"res_type" validate:"required"`
	AccountID    string                 `json:"account_id" validate:"required"`
	BkBizID      int64                  `json:"bk_biz_id" validate:"required"`
	Region       string                 `json:"region" validate:"required"`
	Zone         string                 `json:"zone" validate:"required"`
	Levels       tablequota.Levels      `json:"levels" validate:"required"`
	Dimensions   tablequota.Dimensions  `json:"dimensions" validate:"required"`
	Memo         *string                `json:"memo"`
}

// Validate CreateBizQuotaReq.
func (req *CreateBizQuotaReq) Validate() error {
	return validator.Validate.Struct(req)
}

// UpdateBizQuotaReq define update biz quota req.
type UpdateBizQuotaReq struct {
	ID string `json:"id" validate:"required"`

	CloudQuotaID string                `json:"cloud_quota_id"`
	Vendor       enumor.Vendor         `json:"vendor"`
	AccountID    string                `json:"account_id"`
	Dimensions   tablequota.Dimensions `json:"dimensions"`
	Memo         *string               `json:"memo"`
}

// Validate UpdateBizQuotaReq.
func (req *UpdateBizQuotaReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListBizQuotaResult define list biz quota result.
type ListBizQuotaResult struct {
	Count   uint64               `json:"count"`
	Details []corequota.BizQuota `json:"details"`
}
