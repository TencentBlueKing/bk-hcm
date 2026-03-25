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

package hssubaccount

import (
	"hcm/pkg/criteria/validator"
)

// UpdateSubAccountReq define update sub account request for hc-service.
// Pointer fields use nil to indicate "no change".
type UpdateSubAccountReq struct {
	AccountID   string  `json:"account_id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Remark      *string `json:"remark" validate:"omitempty"`
	Email       *string `json:"email" validate:"omitempty"`
	PhoneNum    *string `json:"phone_num" validate:"omitempty"`
	CountryCode *string `json:"country_code" validate:"omitempty"`
}

// Validate update sub account request.
func (req *UpdateSubAccountReq) Validate() error {
	return validator.Validate.Struct(req)
}
