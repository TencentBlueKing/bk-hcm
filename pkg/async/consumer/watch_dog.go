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
	"sync"
	"time"

	"hcm/pkg/async/closer"
	"hcm/pkg/logs"
)

// WatchDog define watch dog interface.
type WatchDog interface {
	closer.Closer
	// Start 启动watch dog，修复异常的异步任务流程。
	Start()
}

// watchDog 任务流、任务纠正策略
type watchDog struct {
	flowScheduledTimeout time.Duration
	normalIntervalSec    time.Duration

	wg      sync.WaitGroup
	closeCh chan struct{}
}

// NewWatchDog 创建一个watchdog
func NewWatchDog(flowScheduledTimeout time.Duration, normalIntervalSec time.Duration) *watchDog {

	return &watchDog{
		flowScheduledTimeout: flowScheduledTimeout,
		normalIntervalSec:    normalIntervalSec,
		closeCh:              make(chan struct{}),
	}
}

// Start 启动定义的WatchDog
func (wd *watchDog) Start() {
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleExpiredTasks)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleLongTimePendingFlows)
	wd.wg.Add(1)
	go wd.watchWrapper(wd.handleLongTimedRunningFlows)
}

// 定期处理异常任务流或任务
func (wd *watchDog) watchWrapper(do func() error) {
	ticker := time.NewTicker(wd.normalIntervalSec)
	defer ticker.Stop()

	closed := false
	for !closed {
		select {
		case <-wd.closeCh:
			closed = true
		case <-ticker.C:
			if err := do(); err != nil {
				logs.Errorf("[async] [module-watchdog] do watch func error %v", err)
			}
		}
	}

	wd.wg.Done()
}

// Close 等待当前执行体执行完成后再关闭
func (wd *watchDog) Close() {
	close(wd.closeCh)
	wd.wg.Wait()
}

// TODO：处理过期并且状态非成功或者失败的任务
func (wd *watchDog) handleExpiredTasks() error {
	return nil
}

// TODO：处理长时间处于pending状态的flow
func (wd *watchDog) handleLongTimePendingFlows() error {
	return nil
}

// TODO：处理长时间处于running状态的flow
func (wd *watchDog) handleLongTimedRunningFlows() error {
	return nil
}
