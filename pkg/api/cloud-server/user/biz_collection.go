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

package csuser

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// BizCollectionReq define biz collection request.
type BizCollectionReq struct {
	BkBizID int64 `json:"bk_biz_id" validate:"required"`
}

// Validate BizCollectionReq.
func (req BizCollectionReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CreateCollectionReq define create collection request.
type CreateCollectionReq struct {
	ResType enumor.UserCollectionResType `json:"res_type" validate:"required"`
	ResID   string                       `json:"res_id" validate:"required"`
}

// Validate CreateCollectionReq.
func (req CreateCollectionReq) Validate() error {
	return validator.Validate.Struct(req)
}
