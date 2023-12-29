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

import "hcm/pkg/kit"

// Pager 同步分页器，用于。
type Pager[ParamType any] interface {
	// BuildParam 构建同步请求参数
	BuildParam(uuids []string) (params ParamType)

	// SourcePager 数据源分页遍历
	SourcePager
	// TargetPager 目标源分页遍历
	TargetPager
}

// SourcePager 数据源分页
type SourcePager interface {
	// NextFromSource 返回下一批数据
	NextFromSource(kt *kit.Kit) (uuids []string, err error)
	// HasNextFromSource 是否还有下一页数据
	HasNextFromSource() (bool, error)
}

// TargetPager 目标源分页
type TargetPager interface {
	// NextFromTarget 返回下一批数据
	NextFromTarget(kt *kit.Kit) (uuidIDMap map[string]string, err error)
	// HasNextFromTarget 是否还有下一页数据
	HasNextFromTarget() (bool, error)
}
