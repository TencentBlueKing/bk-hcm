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

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	cvt "hcm/pkg/tools/converter"
	// 注册Action和Template
	_ "hcm/pkg/async/action"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/compctrl"
	"hcm/pkg/async/consumer/leader"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/prometheus/client_golang/prometheus"
)

/*
Consumer 异步任务消费者。组件分为两类，公共组件、主节点组件。

主节点组件：（会根据主从判断，自动启动或者关闭这部分组件）
  - dispatcher（派发器）: 负责将Pending状态的任务流，派发到指定节点去执行，并将Flow状态改为Scheduled。
  - watchDog（看门狗）:
    1. 处理超时任务
    2. 处理处于Scheduled状态，但执行节点已经挂掉的任务流
    3. 处理处于Running状态，但执行节点正在Shutdown或者已经挂掉的任务流

公共组件：
  - scheduler（调度器）:
    1. 获取分配给当前节点的处于Scheduled状态的任务流，构建任务流树，将待执行任务推送到执行器执行。
    2. 分析执行器执行完的任务，判断任务流树状态，如果任务流处理完，更新状态，否则将子节点推送到执行器执行。
  - executor（执行器）: 准备任务执行所需要的超时控制，共享数据等工具，并执行任务。
  - commander（指挥者）:
    1. 强制关闭处于执行中的任务
*/
type Consumer interface {
	compctrl.Closer
	// Start 启动消费者，开始消费异步任务。
	Start() error
	CancelFlow(kit *kit.Kit, flowId string) error
}

var _ Consumer = new(consumer)

// NewConsumer new consumer.
func NewConsumer(bd backend.Backend, ld leader.Leader, register prometheus.Registerer, opt *Option) (Consumer, error) {
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
		opt:     opt,
		backend: bd,
		leader:  ld,
		mc:      initMetric(register),
		closers: make([]compctrl.Closer, 0),
	}, nil
}

type consumer struct {
	opt *Option

	backend backend.Backend
	leader  leader.Leader
	mc      *metric

	executor  Executor
	scheduler Scheduler
	watchDog  WatchDog
	cmd       Commander

	// closers 所有组件的关闭操作
	closers []compctrl.Closer
}

// Start 开启消费者消费功能，注：只有主节点进行异步任务消费。
func (csm *consumer) Start() error {
	// kit of consumer, with node uuid as rid
	kt := kit.New()
	kt.Rid = csm.leader.CurrNode()
	csm.initCommonComponent(kt, csm.opt)
	csm.initLeaderComponent(kt, csm.opt)

	return nil
}

// initLeaderComponent 初始化主节点私有组件并启动，同时设置关闭函数
func (csm *consumer) initLeaderComponent(kt *kit.Kit, opt *Option) {

	handler := NewLeaderChangeHandler(csm.backend, csm.leader, opt)
	handler.Start()
	csm.closers = append(csm.closers, handler)

}

// initCommonComponent 初始化主从节点公共组件并启动，同时设置关闭函数
func (csm *consumer) initCommonComponent(kt *kit.Kit, opt *Option) {
	// 设置执行器
	csm.executor = NewExecutor(kt, csm.backend, opt.Executor)

	// 设置调度器
	csm.scheduler = NewScheduler(csm.backend, csm.executor, csm.leader, opt.Scheduler)

	// 设置执行器获取调度器函数。
	csm.executor.SetGetSchedulerFunc(func() Scheduler {
		return csm.scheduler
	})

	// 初始化执行器并启动同时设置关闭函数
	csm.executor.Start()
	csm.closers = append(csm.closers, csm.executor)

	// 初始化调度器并启动同时设置关闭函数
	csm.scheduler.Start()
	csm.closers = append(csm.closers, csm.scheduler)

	// 设置命令工具
	csm.cmd = NewCommander(csm.executor)
}

// Close 执行异步任务框架所有组件的关闭函数
func (csm *consumer) Close() {

	logs.Infof("consumer receive close cmd, start to close")

	for i := range csm.closers {
		csm.closers[i].Close()
	}

	logs.Infof("consumer close success")

}

// CancelFlow 更新任务状态为 cancel
func (csm *consumer) CancelFlow(kt *kit.Kit, flowId string) error {

	flowList, err := csm.backend.ListFlow(kt, &backend.ListInput{
		Filter: tools.EqualExpression("id", flowId),
		Page:   core.NewDefaultBasePage(),
	})
	if err != nil {
		return err
	}
	if len(flowList) == 0 {
		return errors.New("flow not found " + flowId)
	}
	flow := flowList[0]

	if flow.State == enumor.FlowCancel {
		return errors.New("flow has already been canceled")
	}
	if flow.State == enumor.FlowSuccess {
		return errors.New("flow has already succeeded")
	}

	// 取消flow 需要执行该flow的worker执行，调用该方法的时候，对应flow 不一定在当前worker上，因此这里先
	// 更改flow状态为canceled，后续步骤由对应worker上的`canceledFlowWatcher`函数继续执行
	err = updateFlowStateAndReason(kt, csm.backend, flowId, flow.State,
		enumor.FlowCancel, "canceled from "+cvt.PtrToVal(flow.Worker))
	if err != nil {
		logs.Errorf("fail to update flow state to canceled, err: %v, flow_id: %s, rid: %s", err, flowId, kt.Rid)
		return err
	}

	return nil
}
