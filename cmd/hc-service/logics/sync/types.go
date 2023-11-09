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

package sync

// HandlerName define handler name.
type HandlerName string

// SourceData define source data.
type SourceData interface {
	Data

	any
}

// TargetData define target data.
type TargetData interface {
	Data
	// GetID 获取目标数据的ID，因为 GetUUID() 获取的唯一标识不是目标数据的ID，更新数据时还需要转换一次。
	GetID() string

	any
}

// Data 定义数据拥有的接口
type Data interface {
	// GetUUID 数据唯一标识。
	GetUUID() string
}

// Result define sync result.
type Result struct {
	DeleteIDs []string `json:"delete_ids"`
	CreateIDs []string `json:"create_ids"`
	UpdateIDs []string `json:"update_ids"`
}
