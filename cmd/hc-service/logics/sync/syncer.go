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

import (
	"errors"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/maps"
)

// Interface 同步提供的函数能力。
type Interface[BatchSyncParamType any, SourceDataType SourceData, TargetDataType TargetData] interface {
	// AllPages 分页全量同步，将数据源分页调用批量同步进行同步。
	AllPages(kt *kit.Kit) (result *Result, err error)
	// BatchOrAll 批量/全量同步，取决于用户传的查询参数查询的是全量数据还是批量数据。
	BatchOrAll(kt *kit.Kit, params BatchSyncParamType) (result *Result, err error)
	// RemoveDeletedFromSource 移除已经从数据源删除的数据。
	RemoveDeletedFromSource(kt *kit.Kit) (ids []string, err error)
}

// Syncer 同步器
type Syncer[BatchSyncParamType any, SourceDataType SourceData, TargetDataType TargetData] struct {
	// Pager 分页同步处理器
	Pager Pager[BatchSyncParamType]

	// Handler 批量同步处理器
	Handler Handler[BatchSyncParamType, SourceDataType, TargetDataType]
}

// RemoveDeletedFromSource 移除已经从数据源删除的数据。
func (sync *Syncer[BatchSyncParamType, SourceDataType, TargetDataType]) RemoveDeletedFromSource(kt *kit.Kit) (
	ids []string, err error) {

	for {
		// 从目标源查询一批数据，判断这批数据是否有已经从数据源删除的数据
		uuidIDMapFromTarget, err := sync.Pager.NextFromTarget(kt)
		if err != nil {
			logs.Errorf("[%s] get next from target failed, err: %v, rid: %s", sync.Handler.Name(), err, kt.Rid)
			return nil, err
		}

		if len(uuidIDMapFromTarget) != 0 {
			// 从数据源查询数据
			params := sync.Pager.BuildParam(maps.Keys(uuidIDMapFromTarget))
			sourceData, err := sync.Handler.QueryFromSource(kt, params)
			if err != nil {
				logs.Errorf("[%s] query from source failed, err: %v, rid: %s", sync.Handler.Name(), err, kt.Rid)
				return nil, err
			}

			// 如果查询数据和返回数据数量不同，则证明目标源中有数据要被删除
			if len(uuidIDMapFromTarget) != len(sourceData) {
				for _, one := range sourceData {
					delete(uuidIDMapFromTarget, one.GetUUID())
				}

				delIDs := maps.Values(uuidIDMapFromTarget)
				if err = sync.Handler.DeleteTargetData(kt, params, delIDs); err != nil {
					logs.Errorf("[%s] delete target data failed, err: %v, rid: %s", sync.Handler.Name(), err, kt.Rid)
					return nil, err
				}

				ids = append(ids, delIDs...)
			}

			// 判断是否还有下一页资源需要同步
			hasNext, err := sync.Pager.HasNextFromTarget()
			if err != nil {
				logs.Errorf("[%s] exec has next from target failed, err: %v, rid: %s", sync.Handler.Name(), err, kt.Rid)
				return ids, err
			}

			if !hasNext {
				break
			}
		}
	}

	return ids, nil
}

// AllPages 分页全量同步，将数据源分页调用批量同步进行同步。
func (sync *Syncer[BatchSyncParamType, SourceDataType, TargetDataType]) AllPages(kt *kit.Kit) (
	result *Result, err error) {

	if sync.Handler == nil {
		return nil, errors.New("page sync handler is required")
	}

	delIDs, err := sync.RemoveDeletedFromSource(kt)
	if err != nil {
		logs.Errorf("[%s] remove deleted from source failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	result.DeleteIDs = append(result.DeleteIDs, delIDs...)

	for {
		// 获取下一页要同步资源的唯一ID列表
		uuids, err := sync.Pager.NextFromSource(kt)
		if err != nil {
			logs.Errorf("[%s] get next from source failed, err: %v, rid: %s", sync.Handler.Name(), err, kt.Rid)
			return nil, err
		}

		// 执行批量同步，同步这一页的资源
		if len(uuids) != 0 {
			params := sync.Pager.BuildParam(uuids)
			batchSyncResult, err := sync.BatchOrAll(kt, params)
			if err != nil {
				logs.Errorf("[%s] batch sync failed, err: %v, uuids: %v, rid: %s", sync.Handler.Name(),
					err, uuids, kt.Rid)
				return nil, err
			}

			result.DeleteIDs = append(result.DeleteIDs, batchSyncResult.DeleteIDs...)
			result.CreateIDs = append(result.CreateIDs, batchSyncResult.CreateIDs...)
			result.UpdateIDs = append(result.UpdateIDs, batchSyncResult.UpdateIDs...)
		}

		// 判断是否还有下一页资源需要同步
		hasNext, err := sync.Pager.HasNextFromSource()
		if err != nil {
			logs.Errorf("[%s] exec has next from source failed, err: %v, rid: %s", sync.Handler.Name(), err, kt.Rid)
			return nil, err
		}

		if !hasNext {
			break
		}
	}

	return result, nil
}

// BatchOrAll 批量/全量同步，取决于用户传的查询参数查询的是全量数据还是批量数据。
func (sync *Syncer[BatchSyncParamType, SourceDataType, TargetDataType]) BatchOrAll(
	kt *kit.Kit, params BatchSyncParamType) (result *Result, err error) {

	if sync.Handler == nil {
		return nil, errors.New("batch sync handler is required")
	}

	// 从数据源查询数据
	sourceData, err := sync.Handler.QueryFromSource(kt, params)
	if err != nil {
		logs.Errorf("[%s] query from source failed, err: %v, params: %+v, rid: %s", sync.Handler.Name(),
			err, params, kt.Rid)
		return nil, err
	}

	// 从目标源查询数据
	targetData, err := sync.Handler.QueryFromTarget(kt, params)
	if err != nil {
		logs.Errorf("[%s] query from target failed, err: %v, params: %+v, rid: %s", sync.Handler.Name(),
			err, params, kt.Rid)
		return nil, err
	}

	// 没有数据需要同步
	if len(sourceData) == 0 && len(targetData) == 0 {
		return new(Result), nil
	}

	// 对比数据源和目标源数据，对增/删/改数据进行分类
	createData, idUpdateDataMap, delIDs := Diff(sourceData, targetData, sync.Handler.DiffFunc)

	// TODO: 添加日志和metrics数量统计，和失败请求统计
	// 删除目标源中多余的数据
	if len(delIDs) > 0 {
		if err = sync.Handler.DeleteTargetData(kt, params, delIDs); err != nil {
			logs.Errorf("[%s] delete target data failed, err: %v, params: %+v, delIDs: %+v, rid: %s",
				sync.Handler.Name(), err, params, delIDs, kt.Rid)
			return nil, err
		}
	}

	// 更新源数据更新，但目标源没更新的数据
	if len(idUpdateDataMap) > 0 {
		if err = sync.Handler.UpdateTargetData(kt, params, idUpdateDataMap); err != nil {
			logs.Errorf("[%s] update target data failed, err: %v, params: %+v, updateMap: %+v, rid: %s",
				sync.Handler.Name(), err, params, idUpdateDataMap, kt.Rid)
			return nil, err
		}
	}

	var createIDs []string
	// 添加数据源多出的数据
	if len(createData) > 0 {
		createIDs, err = sync.Handler.CreateTargetData(kt, params, createData)
		if err != nil {
			logs.Errorf("[%s] create target data failed, err: %v, params: %+v, createData: %+v, rid: %s",
				sync.Handler.Name(), err, params, createData, kt.Rid)
			return nil, err
		}
	}

	// 聚合处理结果
	result = &Result{
		DeleteIDs: delIDs,
		CreateIDs: createIDs,
		UpdateIDs: maps.Keys(idUpdateDataMap),
	}

	return result, nil
}
