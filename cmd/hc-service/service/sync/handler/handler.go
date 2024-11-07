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

// Package handler ...
package handler

import (
	"sync"
	"time"

	"hcm/cmd/hc-service/logics/res-sync/common"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
)

// Handler 定义了全量同步操作函数。
type Handler interface {
	// Prepare 解析请求体，构建同步所需客户端。
	Prepare(cts *rest.Contexts) error
	// Next 去云上分页查询资源云ID，用于同步，每次分页查询 constant.CloudResourceSyncMaxLimit 条数据。
	Next(kt *kit.Kit) ([]string, error)
	// Sync 同步传入的 cloudIDs 的资源数据。
	Sync(kt *kit.Kit, cloudIDs []string) error
	// RemoveDeleteFromCloud 进行db和云上数据的全量对比，删除已经从云上删除的数据。
	RemoveDeleteFromCloud(kt *kit.Kit) error

	Name() enumor.CloudResourceType
}

// ResourceSync 资源同步流程。
func ResourceSync(cts *rest.Contexts, handler Handler) error {
	kt := cts.Kit

	// 解析请求参数到handler实现中，构建同步需要的客户端
	if err := handler.Prepare(cts); err != nil {
		logs.Errorf("%s sync handler to prepare failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
		return err
	}

	if err := handler.RemoveDeleteFromCloud(kt); err != nil {
		logs.Errorf("%s sync handler to removeDeleteFromCloud failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
		return err
	}

	for {
		cloudIDs, err := handler.Next(kt)
		if err != nil {
			logs.Errorf("%s sync handler to next failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
			return err
		}

		if len(cloudIDs) == 0 {
			break
		}

		if err = handler.Sync(kt, cloudIDs); err != nil {
			logs.Errorf("%s sync handler to sync failed, err: %v, rid: %s", handler.Name(), err, kt.Rid)
			return err
		}

		if len(cloudIDs) < constant.CloudResourceSyncMaxLimit {
			break
		}
	}

	return nil
}

// HandlerV2 实验性并发同步框架
type HandlerV2[T common.CloudResType] interface {

	// Prepare 解析请求体，构建同步所需客户端。
	Prepare(cts *rest.Contexts) error

	// Next 去云上分页查询资源云实例，用于同步，每次分页查询 constant.CloudResourceSyncMaxLimit 条数据。
	Next(kt *kit.Kit) ([]T, error)

	// Sync 同步传入的 云资源实例数据。
	Sync(kt *kit.Kit, instances []T) error

	// RemoveDeletedFromCloud 进行db和云上数据的全量对比，删除已经从云上删除的数据。
	RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error

	// Resource 返回支持的资源类型
	Resource() enumor.CloudResourceType

	// MaxConcurrent 支持的最大并发数量
	MaxConcurrent() uint

	// Describe 描述信息，用于日志输出
	Describe() string
}

const (
	syncQueueSize = 10
)

// ResourceSyncV2 资源同步，包含三个流程：1. 准备请求 2. 获取云上资源
func ResourceSyncV2[T common.CloudResType](cts *rest.Contexts, handler HandlerV2[T]) error {
	kt := cts.Kit

	// 解析请求参数到handler实现中，构建同步需要的客户端
	if err := handler.Prepare(cts); err != nil {
		logs.Errorf("%s sync handler to prepare failed, err: %v, rid: %s", handler.Describe(), err, kt.Rid)
		return err
	}
	logs.Infof("%s sync start with concorrent %d start, rit: %s", handler.Describe(), handler.MaxConcurrent(), kt.Rid)

	allCloudIDMap := make(map[string]struct{}, 1024)
	start := time.Now()
	total := 0
	syncInstCh := make(chan []T, syncQueueSize)
	syncWg := &sync.WaitGroup{}

	concurrent := int(max(handler.MaxConcurrent(), 1))
	for i := 0; i < concurrent; i++ {
		syncWg.Add(1)
		go syncConsumer(kt, handler, i, syncWg, syncInstCh)
	}
	var lastErr error
	for {
		startBatch := time.Now()
		instances, err := handler.Next(kt)
		if err != nil {
			logs.Errorf("%s sync handler to next failed, err: %v, rid: %s", handler.Describe(), err, kt.Rid)
			lastErr = err
			break
		}
		usedBatch := time.Since(startBatch)
		total += len(instances)
		logs.Infof("%s batch got: %d/%d, cost: %s, queue size: %d, rid: %s",
			handler.Describe(), len(instances), total, usedBatch, len(syncInstCh), kt.Rid)
		if len(instances) == 0 {
			break
		}

		for i := range instances {
			allCloudIDMap[instances[i].GetCloudID()] = struct{}{}
		}

		syncInstCh <- instances
		if len(instances) < constant.CloudResourceSyncMaxLimit {
			break
		}
	}
	close(syncInstCh)
	syncInstCh = nil
	logs.Infof("%s pull all cost: %s, count: %d, rid: %s", handler.Describe(), time.Since(start), total, kt.Rid)
	// 仅在拉全量数据没有失败的时候，删除云上已删除数据
	if lastErr == nil {
		if err := handler.RemoveDeletedFromCloud(kt, allCloudIDMap); err != nil {
			logs.Errorf("%s sync handler to remove deleted from cloud failed, err: %v, rid: %s",
				handler.Describe(), err, kt.Rid)
			lastErr = err
		}
		logs.Infof("%s remove deleted done, rid: %s", handler.Describe(), kt.Rid)
	}
	syncWg.Wait()
	logs.Infof("%s sync done, cost: %s, rid: %s", handler.Describe(), time.Since(start), kt.Rid)
	return lastErr
}

func syncConsumer[T common.CloudResType](kt *kit.Kit, handler HandlerV2[T], idx int, wg *sync.WaitGroup,
	syncInstCh chan []T) {

	defer func() {
		wg.Done()
		logs.Infof("[sync_consumer] %s %d exit, rid: %s", handler.Describe(), idx, kt.Rid)
	}()
	for {
		select {
		case instanceList, ok := <-syncInstCh:
			if !ok {
				return
			}
			logs.Infof("[sync_consumer] %s %d got: %d, rid: %s", handler.Describe(), idx, len(instanceList), kt.Rid)
			for _, instances := range slice.Split(instanceList, constant.CloudResourceSyncMaxLimit) {
				if err := handler.Sync(kt, instances); err != nil {
					logs.Errorf("%s handler to sync failed, err: %v, rid: %s", handler.Describe(), err, kt.Rid)
				}
			}
		case <-kt.Ctx.Done():
			return
		}
	}
}
