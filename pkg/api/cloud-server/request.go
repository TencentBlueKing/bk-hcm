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
	"encoding/json"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"

	"github.com/tidwall/gjson"
)

// AccountReq ...
type AccountReq struct {
	AccountID string `json:"account_id" validate:"required"`
}

// Validate ...
func (req *AccountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListReq is a standard list operation http request.
type ListReq struct {
	Filter *filter.Expression `json:"filter"`
	Page   *core.BasePage     `json:"page"`
}

// Validate ListReq.
func (l *ListReq) Validate() error {
	if l.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter is required")
	}

	if l.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	return nil
}

// ResourceListReq raw list request, only account_id is decoded, others are raw json.
type ResourceListReq struct {
	AccountID string
	Data      json.RawMessage
}

// UnmarshalJSON ...
func (r *ResourceListReq) UnmarshalJSON(raw []byte) error {
	r.AccountID = gjson.GetBytes(raw, "account_id").String()
	r.Data = raw
	return nil
}

// -------------------------- Delete --------------------------

// BatchDeleteReq security group update request.
type BatchDeleteReq struct {
	IDs []string `json:"ids" validate:"required"`
}

// Validate security group delete request.
func (req *BatchDeleteReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// ResourceDeleteReq raw delete request, only account_id is decoded, others are raw json.
type ResourceDeleteReq struct {
	AccountID string
	Data      json.RawMessage
}

// UnmarshalJSON unmarshal raw json to RawCreateReq
func (r *ResourceDeleteReq) UnmarshalJSON(raw []byte) error {
	r.AccountID = gjson.GetBytes(raw, "account_id").String()
	r.Data = raw
	return nil
}

// -------------------------- Create --------------------------

// RawCreateReq raw create request, only vendor is decoded, others are raw json.
type RawCreateReq struct {
	Vendor enumor.Vendor
	Data   json.RawMessage
}

// UnmarshalJSON unmarshal raw json to RawCreateReq
func (r *RawCreateReq) UnmarshalJSON(raw []byte) error {
	r.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())
	r.Data = raw
	return nil
}

// ResourceCreateReq raw create request, only account_id is decoded, others are raw json.
type ResourceCreateReq struct {
	AccountID string
	Data      json.RawMessage
}

// UnmarshalJSON unmarshal raw json to RawCreateReq
func (r *ResourceCreateReq) UnmarshalJSON(raw []byte) error {
	r.AccountID = gjson.GetBytes(raw, "account_id").String()
	r.Data = raw
	return nil
}
