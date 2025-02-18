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

package usermgr

import (
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// ListDepartmentParams is list department parameter.
type ListDepartmentParams struct {
	ID           int64  `json:"id"`
	Page         int64  `json:"page"`
	PageSize     int64  `json:"page_size"`
	Fields       string `json:"fields"`
	LookupField  string `json:"lookup_field"`
	ExactLookups string `json:"exact_lookups"`
}

// ListDepartmentResp is list department response.
type ListDepartmentResp struct {
	apigateway.BaseResponse
	*ListDeptResult `json:"data"`
}

// ListDeptResult is usermgr list department response.
type ListDeptResult struct {
	Count   int64       `json:"count"`
	Results []*DeptInfo `json:"results"`
}

// DeptInfo is usermgr department info.
type DeptInfo struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	HasChildren bool       `json:"has_children"`
	Level       int64      `json:"level"`
	FullName    string     `json:"full_name"`
	Parent      int64      `json:"parent"`
	Enabled     bool       `json:"enabled"`
	CategoryID  int64      `json:"category_id"`
	Order       int64      `json:"order"`
	Left        int64      `json:"lft"`
	Right       int64      `json:"rght"`
	Extras      *DeptExtra `json:"extras"`
	Children    []DeptInfo `json:"children"`
}

// DeptExtra is usermgr department extra info.
type DeptExtra struct {
	Code string `json:"code"`
}
