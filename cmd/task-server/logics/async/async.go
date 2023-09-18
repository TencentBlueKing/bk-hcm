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

package async

import (
	"fmt"
	"time"

	"hcm/cmd/task-server/logics/async/backends/iface"
	"hcm/cmd/task-server/logics/async/closer"
	"hcm/cmd/task-server/logics/async/commander"
	"hcm/cmd/task-server/logics/async/consumer"
	"hcm/cmd/task-server/logics/async/leader"
	"hcm/cmd/task-server/logics/async/watchdog"
	"hcm/pkg/cc"
	"hcm/pkg/logs"
)

// 所有组件的关闭操作
var closers []closer.Closer

// InitialOption 异步任务框架的启动参数
type InitialOption struct {
	Leader  leader.Leader
	Backend iface.Backend

	// NormalIntervalSec default 10s
	NormalIntervalSec time.Duration
	// ExecutorWorkerCnt default 10
	ExecutorWorkersCnt int
	// ParserWorkersCnt default 5
	ParserWorkersCnt int
	// FlowScheduleTimeout default 15s
	FlowScheduleTimeout time.Duration
}

// Validate 异步任务框架的启动参数校验和设置默认值操作
func (opt *InitialOption) Validate() error {
	if opt.Leader == nil {
		return fmt.Errorf("leader cannot be nil")
	}
	if opt.Backend == nil {
		return fmt.Errorf("backend cannot be nil")
	}

	opt.NormalIntervalSec = time.Duration(cc.TaskServer().Async.NormalIntervalSec) * time.Second
	opt.ExecutorWorkersCnt = cc.TaskServer().Async.ExecutorWorkerCnt
	opt.ParserWorkersCnt = cc.TaskServer().Async.ParserWorkersCnt
	opt.FlowScheduleTimeout = time.Duration(cc.TaskServer().Async.FlowScheduleTimeout) * time.Second

	return nil
}

// Start 嵌入到带有优雅退出的其它框架中使用 (注意需要显示调用Close)
func Start(opt *InitialOption) error {
	if err := Init(opt); err != nil {
		logs.Errorf("[async] [module-async] init err: %v", err)
		return err
	}

	return nil
}

// Init 异步任务框架所有组件初始化并启动
func Init(opt *InitialOption) error {
	if err := opt.Validate(); err != nil {
		logs.Errorf("[async] [module-async] opt validate err: %v", err)
		return err
	}

	// 主备架构，只有主节点工作，备用节点阻塞直到其切换为主节点
	for {
		time.Sleep(opt.NormalIntervalSec)

		if opt.Leader.IsLeader() {
			break
		}
	}

	// 初始化所有组件并设置关闭函数
	initCommonComponent(opt)

	return nil
}

// initCommonComponent 初始化所有组件并启动同时设置关闭函数
func initCommonComponent(opt *InitialOption) {
	// 设置leader和backend
	leader.SetLeader(opt.Leader)
	iface.SetBackend(opt.Backend)

	// 设置执行器
	exe := consumer.NewExecutor(opt.ExecutorWorkersCnt, opt.NormalIntervalSec)
	consumer.SetExecutor(exe)

	// 设置解析器
	p := consumer.NewParser(opt.ParserWorkersCnt, opt.NormalIntervalSec)
	consumer.SetParser(p)

	// 初始化执行器并启动同时设置关闭函数
	exe.Init()
	closers = append(closers, exe)
	// 初始化解析器并启动同时设置关闭函数
	p.Init()
	closers = append(closers, p)

	// 设置命令工具
	comm := &commander.AsyncCommander{}
	commander.SetCommander(comm)

	// 设置backend的关闭函数
	closers = append(closers, opt.Backend)
	// 设置leader的关闭函数
	closers = append(closers, opt.Leader)

	// 初始化watchdog并启动同时设置关闭函数
	wd := watchdog.NewWatchDog(opt.FlowScheduleTimeout, opt.NormalIntervalSec)
	wd.Init()
	closers = append(closers, wd)
}

// Close 执行异步任务框架所有组件的关闭函数
func Close() {
	logs.V(3).Infof("[async] [module-async] run closer begin")

	for i := range closers {
		closers[i].Close()
	}

	logs.V(3).Infof("[async] [module-async] run closer end")
}
