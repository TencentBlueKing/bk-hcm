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

// Package consumer 消费者
package consumer

import (
	"errors"
	"time"

	"hcm/pkg/async/backend"
	"hcm/pkg/async/closer"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/atomic"
)

// Consumer 定义异步任务消费接口。
type Consumer interface {
	closer.Closer
	// Start 启动消费者，开始消费异步任务。
	Start(optFunc ...Option) error
}

var _ Consumer = new(consumer)

// NewConsumer new consumer.
func NewConsumer(bd backend.Backend, ld leader.Leader, register prometheus.Registerer) (Consumer, error) {
	if bd == nil {
		return nil, errors.New("backend is required")
	}

	if ld == nil {
		return nil, errors.New("leader is required")
	}

	if register == nil {
		return nil, errors.New("metrics register is required")
	}

	return &consumer{
		backend: bd,
		leader:  ld,
		mc:      initMetric(register),
		closers: make([]closer.Closer, 0),
		enable:  new(atomic.Bool),
	}, nil
}

type consumer struct {
	backend backend.Backend
	leader  leader.Leader
	mc      *metric

	executor Executor
	parser   Parser
	watchDog WatchDog
	cmd      Commander

	enable *atomic.Bool

	// closers 所有组件的关闭操作
	closers []closer.Closer
}

// Start 开启消费者消费功能，注：只有主节点进行异步任务消费。
func (csm *consumer) Start(optFunc ...Option) error {
	if csm.enable.Load() {
		return errors.New("already started, cannot be started again")
	}

	opt := new(options)
	for index := range optFunc {
		optFunc[index](opt)
	}

	opt.tryDefaultValue()
	if err := opt.Validate(); err != nil {
		return err
	}

	if err := opt.Validate(); err != nil {
		logs.Errorf("[async] [module-async] opt validate err: %v", err)
		return err
	}

	csm.enable.Store(true)
	logs.Infof("consumer start handle async tasks")

	go func() {
		for {
			time.Sleep(time.Second)

			// 如果消费者被关闭了
			if !csm.enable.Load() {
				logs.Infof("consumer is closed, stop handle async tasks")
				csm.close()
				break
			}

			// 如果是从节点
			if !csm.leader.IsLeader() && len(csm.closers) == 0 {
				continue
			}

			// 如果是主切从
			if !csm.leader.IsLeader() && len(csm.closers) != 0 {
				logs.Infof("the current node changes from the master node to the slave node, " +
					"and start to stop handle async tasks")
				csm.close()
				continue
			}

			// 如果是从切主
			if csm.leader.IsLeader() && len(csm.closers) == 0 {
				logs.Infof("the current node is master, start init common component")
				// 初始化所有组件并设置关闭函数
				csm.initCommonComponent(opt)
				logs.Infof("the current node is master, init common component success")
				continue
			}
		}
	}()

	return nil
}

// initCommonComponent 初始化所有组件并启动同时设置关闭函数
func (csm *consumer) initCommonComponent(opt *options) {
	// 设置执行器
	csm.executor = NewExecutor(csm.backend, opt.executorWorkersCnt, opt.normalIntervalSec)

	// 设置解析器
	csm.parser = NewParser(csm.backend, csm.executor, opt.parserWorkersCnt, opt.normalIntervalSec)

	// 设置执行器获取解析器函数。
	csm.executor.SetGetParserFunc(func() Parser {
		return csm.parser
	})

	// 初始化执行器并启动同时设置关闭函数
	csm.executor.Start()
	csm.closers = append(csm.closers, csm.executor)

	// 初始化解析器并启动同时设置关闭函数
	csm.parser.Start()
	csm.closers = append(csm.closers, csm.parser)

	// 设置命令工具
	csm.cmd = NewCommander()

	// 初始化watchdog并启动同时设置关闭函数
	csm.watchDog = NewWatchDog(opt.flowScheduleTimeoutSec, opt.normalIntervalSec)
	csm.watchDog.Start()
	csm.closers = append(csm.closers, csm.watchDog)
}

// Close 执行异步任务框架所有组件的关闭函数
func (csm *consumer) Close() {
	csm.enable.Store(false)
	csm.close()
}

func (csm *consumer) close() {
	logs.V(3).Infof("[async] [module-async] run closer begin")

	for i := range csm.closers {
		csm.closers[i].Close()
	}

	logs.V(3).Infof("[async] [module-async] run closer end")
}
