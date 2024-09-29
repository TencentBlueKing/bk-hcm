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

package core

import (
	"hcm/pkg/rest"
)

// BatchCreateResp ...
type BatchCreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *BatchCreateResult `json:"data"`
}

// BatchCreateResult ...
type BatchCreateResult struct {
	IDs []string `json:"ids"`
}

// ListResult define list result.
type ListResult struct {
	Count   uint64        `json:"count"`
	Details []interface{} `json:"details"`
}

// ListResultT generic list result
type ListResultT[T any] struct {
	Count   uint64 `json:"count"`
	Details []T    `json:"details"`
}

// BaseResp define base resp.
type BaseResp[T any] struct {
	rest.BaseResp `json:",inline"`
	Data          T `json:"data"`
}

// CloudCreateResult 调用云上接口创建结果
type CloudCreateResult struct {
	// 本地ID，可能为空
	ID string `json:"id"`
	// 云上ID
	CloudID string `json:"cloud_id"`
}
