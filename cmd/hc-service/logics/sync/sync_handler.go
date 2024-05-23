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

// Handler 定义批量同步所需函数。
type Handler[ParamType any, SourceDataType SourceData, TargetDataType TargetData] interface {
	// Name 返回处理器名称
	Name() HandlerName

	// QueryFromSource 从数据源查询数据
	QueryFromSource(kt *kit.Kit, params ParamType) (sourceData []SourceDataType, err error)
	// QueryFromTarget 从目标源查询数据。
	QueryFromTarget(kt *kit.Kit, params ParamType) (targetData []TargetDataType, err error)
	// DiffFunc 对比源数据和目标数据是否发生改变
	DiffFunc(sourceData SourceDataType, targetData TargetDataType) bool
	// DeleteTargetData 删除目标源中的数据。这部分数据是从数据源中已经删除了的，但目标源中还存在的数据。
	DeleteTargetData(kt *kit.Kit, params ParamType, delIDs []string) error
	// CreateTargetData 添加目标源中的数据。这部分数据是数据源中已经创建了的数据，但目标源还没有的数据。
	CreateTargetData(kt *kit.Kit, params ParamType, createData []SourceDataType) (ids []string, err error)
	// UpdateTargetData 更新源数据和目标源中存在字段发生改变的数据。
	UpdateTargetData(kt *kit.Kit, params ParamType, idUpdateDataMap map[string]SourceDataType) error
}
