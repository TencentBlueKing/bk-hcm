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

// Package base defines basic api call protocols.
package base

import "hcm/pkg/rest"

// CreateResp is a standard create operation http response.
type CreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CreateResult `json:"data"`
}

// CreateResult is a standard create operation result.
type CreateResult struct {
	ID uint64 `json:"id"`
}

// BatchDeleteReq is a standard batch delete operation http request.
type BatchDeleteReq struct {
	IDs []uint64 `json:"ids"`
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
