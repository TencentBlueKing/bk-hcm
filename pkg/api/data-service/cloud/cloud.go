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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
)

// -------------------------- Get --------------------------

// GetResourceBasicInfoResp define get resource basic info resp.
type GetResourceBasicInfoResp struct {
	rest.BaseResp `json:",inline"`
	Data          *types.CloudResourceBasicInfo `json:"data"`
}

// -------------------------- List --------------------------

// ListResourceBasicInfoReq define list resource basic info req.
type ListResourceBasicInfoReq struct {
	ResourceType enumor.CloudResourceType `json:"resource_type" validate:"required"`
	IDs          []string                 `json:"ids" validate:"required"`
}

// Validate list resource vendor req.
func (req *ListResourceBasicInfoReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListResourceBasicInfoResp list resource basic info resp.
type ListResourceBasicInfoResp struct {
	rest.BaseResp `json:",inline"`
	Data          map[string]types.CloudResourceBasicInfo `json:"data"`
}

// ------------------------- Assign -------------------------

// AssignResourceToBizReq assign cloud resource to biz request.
type AssignResourceToBizReq struct {
	AccountID string                     `json:"account_id"  validate:"required"`
	BkBizID   int64                      `json:"bk_biz_id"  validate:"min=1"`
	ResTypes  []enumor.CloudResourceType `json:"res_types"  validate:"min=1"`
}

// Validate AssignResourceToBizReq.
func (a AssignResourceToBizReq) Validate() error {
	return validator.Validate.Struct(a)
}
