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

	"hcm/pkg/async/backend"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/logs"
)

// NewLeaderChangeHandler new leader change handler.
func NewLeaderChangeHandler(bd backend.Backend, ld leader.Leader, opt *Option) *LeaderChangeHandler {
	return &LeaderChangeHandler{
		opt:     opt,
		ld:      ld,
		bd:      bd,
		closeCh: make(chan struct{}),
		closers: make([]compctrl.Closer, 0),
		wg:      sync.WaitGroup{},
	}
}

// LeaderChangeHandler 主从切换处理器，负责主从切换后自动启动或者关闭这部分组件。
type LeaderChangeHandler struct {
	opt *Option

	ld leader.Leader
	bd backend.Backend

	dispatcher *Dispatcher
	watchDog   WatchDog

	closeCh chan struct{}

	closers []compctrl.Closer
	wg      sync.WaitGroup
}

// Start 负责处理主从切换后，处理逻辑。
func (handler *LeaderChangeHandler) Start() {
	handler.wg.Add(1)
	go handler.Do()
}

// Do 负责主节点组件的开启和关闭，在切主/切从的时候。
func (handler *LeaderChangeHandler) Do() {
	for {
		time.Sleep(time.Second)

		// 如果被关闭，退出循环
		select {
		case <-handler.closeCh:
			handler.closeLeaderComponent()
			break
		default:
		}

		// 如果是从节点，且主节点组件处于关闭状态，直接跳过即可
		if !handler.ld.IsLeader() && len(handler.closers) == 0 {
			continue
		}

		// 如果是主切从（从节点，但主节点组件处于开启状态），需要关闭主节点组件
		if !handler.ld.IsLeader() && len(handler.closers) != 0 {
			logs.Infof("the current node changes from the master node to the slave node, " +
				"and start to stop handleRunningFlow async tasks")

			handler.closeLeaderComponent()
			continue
		}

		// 如果是从切主，需要开启主节点组件
		if handler.ld.IsLeader() && len(handler.closers) == 0 {
			logs.Infof("the current node is master, start leader component...")
			handler.startLeaderComponent()
			logs.Infof("the current node is master, start leader success")
			continue
		}
	}

	handler.wg.Done()
}

func (handler *LeaderChangeHandler) startLeaderComponent() {
	dis := NewDispatcher(handler.bd, handler.ld, handler.opt.Dispatcher)
	dis.Start()
	handler.closers = append(handler.closers, dis)
	handler.dispatcher = dis

	// 初始化watchdog并启动同时设置关闭函数
	wd := NewWatchDog(handler.bd, handler.ld, handler.opt.WatchDog)
	wd.Start()
	handler.closers = append(handler.closers, wd)
	handler.watchDog = wd
}

// Close 主从切换处理器
func (handler *LeaderChangeHandler) Close() {

	logs.Infof("LeaderChangeHandler receive close cmd, start to close")

	close(handler.closeCh)
	handler.closeLeaderComponent()
	handler.wg.Wait()

	logs.Infof("LeaderChangeHandler close success")

}

// closeLeaderComponent 关闭主节点组件
func (handler *LeaderChangeHandler) closeLeaderComponent() {

	for i := range handler.closers {
		handler.closers[i].Close()
	}
	handler.closers = make([]compctrl.Closer, 0)

}
