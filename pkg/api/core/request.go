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

// Package core defines basic api call protocols.
package core

import (
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// CreateResp is a standard create operation http response.
type CreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CreateResult `json:"data"`
}

// CreateResult is a standard create operation result.
type CreateResult struct {
	ID string `json:"id"`
}

// BatchDeleteReq is a standard batch delete operation http request.
type BatchDeleteReq struct {
	IDs []string `json:"ids"`
}

// Validate BatchDeleteReq.
func (d *BatchDeleteReq) Validate() error {
	if len(d.IDs) == 0 {
		return errf.New(errf.InvalidParameter, "ids are required")
	}

	return nil
}

// BatchDeleteResp is a standard batch delete operation http response.
type BatchDeleteResp struct {
	Succeeded []string    `json:"succeeded,omitempty"`
	Failed    *FailedInfo `json:"failed,omitempty"`
}

// FailedInfo is a standard operation failed info.
type FailedInfo struct {
	ID    string `json:"id"`
	Error string `json:"error"`
}

// UpdateResp ...
type UpdateResp struct {
	rest.BaseResp `json:",inline"`
	Data          interface{} `json:"data"`
}

// DeleteResp ...
type DeleteResp struct {
	rest.BaseResp `json:",inline"`
	Data          interface{} `json:"data"`
}

// ListReq is a standard list operation http request.
type ListReq struct {
	Filter *filter.Expression `json:"filter"`
	Page   *BasePage          `json:"page"`
	Fields []string           `json:"fields"`
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
