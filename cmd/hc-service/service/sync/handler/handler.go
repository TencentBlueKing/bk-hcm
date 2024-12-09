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
	"fmt"
	"sync"
	"sync/atomic"
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

	// Next 去云上分页查询资源云实例，用于同步，每次分页查询至少 constant.CloudResourceSyncMaxLimit 条数据。
	Next(kt *kit.Kit) ([]T, error)

	// Sync 同步传入的 云资源实例数据。
	Sync(kt *kit.Kit, instances []T) error

	// RemoveDeletedFromCloud 进行db和云上数据的全量对比，删除已经从云上删除的数据。
	RemoveDeletedFromCloud(kt *kit.Kit, allCloudIDMap map[string]struct{}) error

	// Resource 返回支持的资源类型
	Resource() enumor.CloudResourceType

	// SyncConcurrent 支持的并发数
	SyncConcurrent() uint

	// Describe 描述信息，用于日志输出
	Describe() string
}

const (
	syncQueueSize              = 10
	lastNBatchToSplitMiniBatch = 2
)

// ResourceSyncV2 资源同步，包含三个流程：1. 准备请求 2. 获取云上实例列表 3. 清理云上已删除实例 4. 同步实例详情
func ResourceSyncV2[T common.CloudResType](cts *rest.Contexts, handler HandlerV2[T]) error {
	kt := cts.Kit

	// 1. 解析请求参数到handler实现中，构建同步需要的客户端
	if err := handler.Prepare(cts); err != nil {
		logs.Errorf("[ResourceSyncV2] %s sync handler to prepare failed, err: %v, rid: %s",
			handler.Describe(), err, kt.Rid)
		return err
	}
	// 2. 获取云上实例列表
	logs.Infof("[ResourceSyncV2] %s sync start with concurrent %d start, rid: %s",
		handler.Describe(), handler.SyncConcurrent(), kt.Rid)
	allCloudIDMap := make(map[string]struct{}, 1024)
	allInstanceList := make([][]T, 0)
	start := time.Now()
	total := 0
	for {
		startBatch := time.Now()
		instances, err := handler.Next(kt)
		if err != nil {
			logs.Errorf("[ResourceSyncV2] %s sync handler to next failed, err: %v, rid: %s",
				handler.Describe(), err, kt.Rid)
			return err
		}
		usedBatch := time.Since(startBatch)
		total += len(instances)
		logs.Infof("[ResourceSyncV2] %s batch got: %d/%d, cost: %s, rid: %s",
			handler.Describe(), len(instances), total, usedBatch, kt.Rid)
		if len(instances) == 0 {
			break
		}

		for i := range instances {
			allCloudIDMap[instances[i].GetCloudID()] = struct{}{}
		}

		allInstanceList = append(allInstanceList, instances)
		if len(instances) < constant.CloudResourceSyncMaxLimit {
			break
		}
	}
	logs.Infof("[ResourceSyncV2] %s pull all cost: %s, count: %d, rid: %s", handler.Describe(), time.Since(start),
		total, kt.Rid)

	// 3. 删除云上已删除数据
	if err := handler.RemoveDeletedFromCloud(kt, allCloudIDMap); err != nil {
		logs.Errorf("[ResourceSyncV2] %s sync handler to remove deleted from cloud failed, err: %v, rid: %s",
			handler.Describe(), err, kt.Rid)
		return err
	}
	logs.Infof("[ResourceSyncV2] %s remove deleted done, rid: %s", handler.Describe(), kt.Rid)

	// 并发同步资源实例
	syncWg := &sync.WaitGroup{}
	syncInstCh := make(chan []T, syncQueueSize)
	concurrent := int(max(handler.SyncConcurrent(), 1))
	var successCnt, failedCnt = &atomic.Uint64{}, &atomic.Uint64{}

	for i := 0; i < concurrent; i++ {
		syncWg.Add(1)
		go syncConsumer(kt, handler, i, syncWg, syncInstCh, successCnt, failedCnt)
	}
	left := total
	// 倒序同步，因为新建的往往按序插入，所以倒序同步可以保证新建的实例先同步。
	for i := len(allInstanceList) - 1; i >= 0; i-- {
		// 单次拉取数量可能大于 constant.CloudResourceSyncMaxLimit，取决于并发数
		batchs := slice.Split(allInstanceList[i], constant.CloudResourceSyncMaxLimit)
		for _, instBatch := range batchs {
			if i > 0 || concurrent < 2 {
				left -= len(instBatch)
				syncInstCh <- instBatch
				logs.Infof("[ResourceSyncV2] %s resource left %d, rid: %s", handler.Describe(), left, kt.Rid)
				continue
			}
			logs.Infof("[ResourceSyncV2] handling last batch, handler: %s, left: %d, rid: %s",
				handler.Describe(), left, kt.Rid)
			// 如果是最后一批，按并发数再拆开
			for _, miniBatch := range slice.Split(instBatch, max(len(instBatch)/(concurrent), 10)) {
				left -= len(miniBatch)
				syncInstCh <- miniBatch
				logs.Infof("[ResourceSyncV2] %s resource left %d, rid: %s", handler.Describe(), left, kt.Rid)
			}
		}
	}
	close(syncInstCh)
	syncWg.Wait()
	success, failed := successCnt.Load(), failedCnt.Load()
	cost := time.Since(start)
	logs.Infof("[ResourceSyncV2] %s sync done, total/success/failed: %d/%d/%d, avg: %.2f res/s, cost: %s, rid: %s",
		handler.Describe(), total, success, failed, float64(total)/cost.Seconds(), cost, kt.Rid)
	if failed != 0 {
		return fmt.Errorf("%d load balancer sync failed", failed)
	}
	return nil
}

func syncConsumer[T common.CloudResType](kt *kit.Kit, handler HandlerV2[T], idx int,
	wg *sync.WaitGroup, syncInstCh chan []T, globalSuccess, globalFailed *atomic.Uint64) {

	start := time.Now()
	var total, failed int
	defer func() {
		cost := time.Since(start)
		wg.Done()
		logs.Infof("[ResourceSyncV2] consumer[%d] %s exit, total(failed): %d(%d), avg: %.2f res/s, cost: %s, rid: %s",
			idx, handler.Describe(), total, failed, float64(total)/cost.Seconds(), cost, kt.Rid)
	}()

	for {
		select {
		case instanceList, ok := <-syncInstCh:
			if !ok {
				return
			}

			logs.Infof("[ResourceSyncV2] consumer[%d] %s got: %d, queue: %d, rid: %s",
				idx, handler.Describe(), len(instanceList), len(syncInstCh), kt.Rid)
			for _, instances := range slice.Split(instanceList, constant.CloudResourceSyncMaxLimit) {
				total += len(instances)
				if err := handler.Sync(kt, instances); err != nil {
					logs.Errorf("[ResourceSyncV2] consumer[%d] %s handler to sync failed, err: %v, rid: %s",
						idx, handler.Describe(), err, kt.Rid)
					failed += len(instances)
					globalFailed.Add(uint64(len(instances)))
					continue
				}
				globalSuccess.Add(uint64(len(instances)))
			}
		case <-kt.Ctx.Done():
			return
		}
	}
}
