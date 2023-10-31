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

package disk

import (
	coredisk "hcm/pkg/api/core/cloud/disk"
	"hcm/pkg/rest"
)

// ListResp ...
type ListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListResult `json:"data"`
}

// ListResult ...
type ListResult struct {
	Count   uint64               `json:"count,omitempty"`
	Details []*coredisk.BaseDisk `json:"details"`
}

// ListExtResp ...
type ListExtResp[T coredisk.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *ListExtResult[T] `json:"data"`
}

// ListExtResult ...
type ListExtResult[T coredisk.Extension] struct {
	Count   uint64              `json:"count,omitempty"`
	Details []*coredisk.Disk[T] `json:"details"`
}

// GetResp 返回单个云盘详情
type GetResp[T coredisk.Extension] struct {
	rest.BaseResp `json:",inline"`
	Data          *coredisk.Disk[T] `json:"data"`
}
