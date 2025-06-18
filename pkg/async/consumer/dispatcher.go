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

package consumer

import (
	"errors"
	"sync"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// NewDispatcher new dispatcher.
func NewDispatcher(bd backend.Backend, ld leader.Leader, opt *DispatcherOption) *Dispatcher {
	return &Dispatcher{
		watchIntervalSec:              time.Duration(opt.WatchIntervalSec) * time.Second,
		bd:                            bd,
		ld:                            ld,
		closeCh:                       make(chan struct{}),
		wg:                            new(sync.WaitGroup),
		pendingFlowFetcherConcurrency: opt.PendingFlowFetcherConcurrency,
	}
}

// Dispatcher 派发器，负责将Pending状态的任务流，派发到指定节点去执行，并将Flow状态改为Scheduled。。
type Dispatcher struct {
	watchIntervalSec time.Duration

	bd backend.Backend
	ld leader.Leader

	wg      *sync.WaitGroup
	closeCh chan struct{}

	pendingFlowFetcherConcurrency uint
}

// Start dispatcher.
func (d *Dispatcher) Start() {
	d.wg.Add(1)
	go d.WatchPendingFlow()
}

// WatchPendingFlow 监听处于Pending状态的流，并派发到指定节点。
func (d *Dispatcher) WatchPendingFlow() {
	// 初始化协程池
	pool := newTenantWorkerPool(d.pendingFlowFetcherConcurrency,
		func(tenantID string) {
			kt := NewKit()
			kt.TenantID = tenantID
			if err := d.Do(kt); err != nil {
				logs.Errorf("%s: dispatcher do failed for tenant %s, err: %v, rid: %s",
					constant.AsyncTaskWarnSign, tenantID, err, kt.Rid)
			}
		})

	// 主任务分发循环
	for {
		select {
		case <-d.closeCh:
			pool.shutdownPoolGracefully()
			d.wg.Done()
			logs.Infof("received stop signal, stop watch pending flow job success.")
			return
		default:
		}

		err := pool.executeWithTenant()
		if err != nil {
			logs.Errorf("WatchPendingFlow failed to executeWithTenant, err: %v", err)
			time.Sleep(d.watchIntervalSec)
			continue
		}

		time.Sleep(d.watchIntervalSec)
	}
}

// Do 监听处于Pending状态的流，并派发到指定节点。
func (d *Dispatcher) Do(kt *kit.Kit) error {
	input := &backend.ListInput{
		Filter: tools.ExpressionAnd(
			// 走worker,state 索引
			tools.RuleEqual("worker", ""),
			tools.RuleEqual("state", enumor.FlowPending),
		),
		Page: core.NewDefaultBasePage(),
	}
	flows, err := d.bd.ListFlow(kt, input)
	if err != nil {
		logs.Errorf("list flow failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if len(flows) == 0 {
		logs.V(3).Infof("currently no task flows to assign, skip handleRunningFlow, rid: %s", kt.Rid)
		return nil
	}

	nodes, err := d.ld.AliveNodes()
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("alive nodes not found")
	}

	infos := make([]backend.UpdateFlowInfo, 0, len(flows))
	for index, one := range flows {
		infos = append(infos, backend.UpdateFlowInfo{
			ID:     one.ID,
			Source: enumor.FlowPending,
			Target: enumor.FlowScheduled,
			Worker: cvt.ValToPtr(nodes[index%len(nodes)]), // 任务分发算法，后续看是否优化
		})
	}

	if err = d.bd.BatchUpdateFlowStateByCAS(kt, infos); err != nil {
		logs.Errorf("batch update flow failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// Close dispatcher
func (d *Dispatcher) Close() {

	logs.Infof("dispatcher receive close cmd, start to close")

	close(d.closeCh)
	d.wg.Wait()

	logs.Infof("dispatcher close success")

}
