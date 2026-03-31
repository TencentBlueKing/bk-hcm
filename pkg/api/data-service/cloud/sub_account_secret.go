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
	"encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	coresass "hcm/pkg/api/core/cloud/sub-account-secret"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ------------------------ Batch Create ------------------------

// SubAccountSecretBatchCreateReq defines batch create sub account secret request.
type SubAccountSecretBatchCreateReq[T coresass.Extension] struct {
	SubAccountSecrets []SubAccountSecretCreate[T] `json:"sub_account_secrets" validate:"required,min=1,max=100"`
}

// SubAccountSecretCreate defines create sub account secret.
type SubAccountSecretCreate[T coresass.Extension] struct {
	AccountID      string                        `json:"account_id" validate:"required"`
	SubAccountID   string                        `json:"sub_account_id" validate:"required"`
	Status         enumor.SubAccountSecretStatus `json:"status" validate:"required"`
	CloudCreatedAt string                        `json:"cloud_created_at" validate:"omitempty"`
	DisabledTime   string                        `json:"disabled_time" validate:"omitempty"`
	LastUsedTime   string                        `json:"last_used_time" validate:"omitempty"`
	Extension      *T                            `json:"extension" validate:"required"`
}

// Validate sub account secret batch create request.
func (req *SubAccountSecretBatchCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------ Batch Update ------------------------

// SubAccountSecretBatchUpdateReq defines batch update sub account secret request.
type SubAccountSecretBatchUpdateReq[T coresass.Extension] struct {
	SubAccountSecrets []SubAccountSecretUpdate[T] `json:"sub_account_secrets" validate:"required,min=1,max=100"`
}

// SubAccountSecretUpdate defines update sub account secret.
type SubAccountSecretUpdate[T coresass.Extension] struct {
	ID             string                         `json:"id" validate:"required"`
	Status         *enumor.SubAccountSecretStatus `json:"status,omitempty"`
	CloudCreatedAt *string                        `json:"cloud_created_at,omitempty"`
	DisabledTime   *string                        `json:"disabled_time,omitempty"`
	LastUsedTime   *string                        `json:"last_used_time,omitempty"`
	Extension      *T                             `json:"extension,omitempty"`
}

// Validate sub account secret batch update request.
func (req *SubAccountSecretBatchUpdateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------ Batch Delete ------------------------

// SubAccountSecretBatchDeleteReq defines batch delete sub account secret request.
type SubAccountSecretBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate sub account secret batch delete request.
func (req *SubAccountSecretBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ------------------------ List ------------------------

// SubAccountSecretListReq defines list sub account secret request.
type SubAccountSecretListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate sub account secret list request.
func (req *SubAccountSecretListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SubAccountSecretListResult defines list sub account secret result.
type SubAccountSecretListResult struct {
	Count   uint64                          `json:"count"`
	Details []coresass.BaseSubAccountSecret `json:"details"`
}

// SubAccountSecretListResp defines list sub account secret response.
type SubAccountSecretListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SubAccountSecretListResult `json:"data"`
}

// SubAccountSecretExtListReq defines list sub account secret with extension request.
type SubAccountSecretExtListReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate sub account secret list with extension request.
func (req *SubAccountSecretExtListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SubAccountSecretExtListResult defines list sub account secret with extension result.
type SubAccountSecretExtListResult[T coresass.Extension] struct {
	Count   uint64                         `json:"count"`
	Details []coresass.SubAccountSecret[T] `json:"details"`
}

// SubAccountSecretExtListResp defines list sub account secret with extension response.
type SubAccountSecretExtListResp[T coresass.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *SubAccountSecretExtListResult[T] `json:"data"`
}

// TCloudSubAccountSecretListExt is the data-service API name for coresass.TCloudSubAccountSecretListExt.
type TCloudSubAccountSecretListExt = coresass.TCloudSubAccountSecretListExt

// ParseTCloudBizListExtension decodes and validates Tencent Cloud biz list extension JSON.
func ParseTCloudBizListExtension(f tabletypes.JsonField) (*TCloudSubAccountSecretListExt, error) {
	if f.IsEmpty() {
		return nil, nil
	}
	var ext TCloudSubAccountSecretListExt
	if err := json.Unmarshal([]byte(f), &ext); err != nil {
		return nil, fmt.Errorf("invalid extension json: %w", err)
	}
	if err := validator.Validate.Struct(&ext); err != nil {
		return nil, err
	}
	return &ext, nil
}

// SubAccountSecretFilters defines biz-scoped list filters; Extension holds vendor-specific JSON (shape depends on vendor).
type SubAccountSecretFilters struct {
	Status             *enumor.SubAccountSecretStatus `json:"status,omitempty"`
	AccountIDs         []string                       `json:"account_ids" validate:"omitempty,max=500,dive,lte=64"`
	SubAccountIDs      []string                       `json:"sub_account_ids" validate:"omitempty,max=500,dive,lte=64"`
	AccountManagers    []string                       `json:"account_managers" validate:"omitempty,max=500,dive,lte=64"`
	SubAccountManagers []string                       `json:"sub_account_managers" validate:"omitempty,max=500,dive,lte=64"`
	Extension          tabletypes.JsonField           `json:"extension,omitempty"`
}

// SubAccountSecretJoinExtListReq defines sub account secret join-list request for data-service
// (secret joined with sub_account and account). Vendor must match the path; Extension JsonField is vendor-specific.
type SubAccountSecretJoinExtListReq struct {
	BkBizID                 int64 `json:"bk_biz_id" validate:"required"`
	SubAccountSecretFilters `json:",inline"`
	Page                    *core.BasePage `json:"page" validate:"required"`
}

// Validate join list request.
func (req *SubAccountSecretJoinExtListReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}
	return req.Page.Validate(core.NewDefaultPageOption())
}

// SubAccountSecretJoinExtDetail is one row in join+ext list response (tcloud detail shape).
type SubAccountSecretJoinExtDetail struct {
	coresass.BaseSubAccountSecret `json:",inline"`
	Extension                     *coresass.TCloudSubAccountSecretExtension `json:"extension"`
	ConsoleLogin                  *int64                                    `json:"console_login,omitempty"`
	AccountManagers               []string                                  `json:"account_managers"`
	SubAccountManagers            []string                                  `json:"sub_account_managers"`
}

// SubAccountSecretJoinExtListResult defines join list response.
type SubAccountSecretJoinExtListResult struct {
	Count   uint64                          `json:"count"`
	Details []SubAccountSecretJoinExtDetail `json:"details"`
}

// SubAccountSecretJoinListResp defines list join HTTP response.
type SubAccountSecretJoinListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SubAccountSecretJoinExtListResult `json:"data"`
}
