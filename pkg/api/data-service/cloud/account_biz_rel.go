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
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// -------------------------- Update --------------------------

// AccountBizRelUpdateReq ...
type AccountBizRelUpdateReq struct {
	UsageBizIDs []int64 `json:"usage_biz_ids" validate:"required"`
}

// Validate ...
func (req *AccountBizRelUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- List --------------------------

// AccountBizRelWithAccountListReq ...
type AccountBizRelWithAccountListReq struct {
	UsageBizIDs []int64 `json:"usage_biz_ids" validate:"required"`
	AccountType string  `json:"account_type" validate:"omitempty"`
}

// Validate ...
func (req *AccountBizRelWithAccountListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.AccountType != "" {
		if err := enumor.AccountType(req.AccountType).Validate(); err != nil {
			return err
		}
	}

	return nil
}

// AccountBizRelWithAccountListResp list account biz relation with account response
type AccountBizRelWithAccountListResp struct {
	rest.BaseResp `json:",inline"`
	Data          []*AccountBizRelWithAccount `json:"data"`
}

// AccountBizRelWithAccount account biz relation with account
type AccountBizRelWithAccount struct {
	corecloud.BaseAccount `json:",inline"`
	RelUsageBizID         int64  `json:"rel_usage_biz_id"`
	RelCreator            string `db:"rel_creator" json:"rel_creator"`
	RelCreatedAt          string `db:"rel_created_at" json:"rel_created_at"`
}

// AccountBizRelListResp list account biz relation response
type AccountBizRelListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AccountBizRelListResult `json:"data"`
}

// AccountBizRelListResult list account biz relation result
type AccountBizRelListResult struct {
	Count   uint64                    `json:"count,omitempty"`
	Details []corecloud.AccountBizRel `json:"details,omitempty"`
}
