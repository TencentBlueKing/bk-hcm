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

package task

import (
	"testing"
	"time"

	"hcm/cmd/task-server/logics/asyncflowcleanup"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	cvt "hcm/pkg/tools/converter"

	"github.com/stretchr/testify/assert"
)

// taskServerSetting 供用例改写清理配置，InitRuntime 持有其指针，改字段即改 cc.TaskServer() 的返回值。
var taskServerSetting = new(cc.TaskServerSetting)

func TestMain(m *testing.M) {
	cc.InitRuntime(taskServerSetting)
	m.Run()
}

// stubState 模拟 serviced.State，用于控制当前节点是否为 master。
type stubState struct {
	master bool
}

// IsMaster 返回预设的 master 标记。
func (s *stubState) IsMaster() bool {
	return s.master
}

// DisableMasterSlave 测试桩不需要实现主从开关。
func (s *stubState) DisableMasterSlave(disable bool) {}

// TestAsyncFlowAndTaskCleanupTaskMeta 校验任务元信息：名称、URL 与执行周期。
func TestAsyncFlowAndTaskCleanupTaskMeta(t *testing.T) {
	task := &AsyncFlowAndTaskCleanupTask{}

	assert.Equal(t, "async_flow_and_task_cleanup", task.Name())
	assert.Equal(t, "/async_flow_and_task/cleanup", task.GetURL())

	taskServerSetting.AsyncFlowAndTaskCleanup = cc.AsyncFlowAndTaskCleanup{
		Enabled:     cvt.ValToPtr(true),
		IntervalMin: cvt.ValToPtr(constant.DefaultAsyncFlowCleanupIntervalMin),
	}
	next, err := task.Next()
	assert.NoError(t, err)
	assertNextAround(t, next, constant.DefaultAsyncFlowCleanupIntervalMin)

	taskServerSetting.AsyncFlowAndTaskCleanup.IntervalMin = cvt.ValToPtr(5)
	next, err = task.Next()
	assert.NoError(t, err)
	assertNextAround(t, next, 5)
}

// assertNextAround 断言下次执行时间落在「当前时间 + intervalMin」附近，容忍用例执行本身的耗时。
func assertNextAround(t *testing.T, next time.Time, intervalMin int) {
	expect := time.Now().Add(time.Duration(intervalMin) * time.Minute)
	diff := next.Sub(expect)
	assert.True(t, diff > -time.Minute && diff < time.Minute,
		"next %s should be around %s", next, expect)
}

// TestAsyncFlowAndTaskCleanupTaskDoOnSlave slave 节点直接跳过，不触发清理。
func TestAsyncFlowAndTaskCleanupTaskDoOnSlave(t *testing.T) {
	// 开关打开，确保跳过是 master 判定的结果而非配置关闭
	taskServerSetting.AsyncFlowAndTaskCleanup = cc.AsyncFlowAndTaskCleanup{
		Enabled:         cvt.ValToPtr(true),
		IntervalMin:     cvt.ValToPtr(constant.DefaultAsyncFlowCleanupIntervalMin),
		RetentionDays:   cvt.ValToPtr(constant.DefaultAsyncFlowCleanupRetentionDays),
		BatchIntervalMs: cvt.ValToPtr(constant.DefaultAsyncFlowCleanupBatchIntervalMs),
	}

	// logics 为 nil，一旦跳过逻辑失效就会 panic，可反证 slave 分支未进入清理主体
	task := &AsyncFlowAndTaskCleanupTask{logics: nil, sd: &stubState{master: false}}
	assert.NoError(t, task.Do(kit.New()))

	task = &AsyncFlowAndTaskCleanupTask{logics: nil, sd: nil}
	assert.NoError(t, task.Do(kit.New()))
}

// TestAsyncFlowAndTaskCleanupTaskDoDisabled master 节点在开关关闭时把防重入/关闭错误降级为跳过。
func TestAsyncFlowAndTaskCleanupTaskDoDisabled(t *testing.T) {
	taskServerSetting.AsyncFlowAndTaskCleanup = cc.AsyncFlowAndTaskCleanup{Enabled: cvt.ValToPtr(false)}

	task := &AsyncFlowAndTaskCleanupTask{
		logics: asyncflowcleanup.NewLogics(nil, nil),
		sd:     &stubState{master: true},
	}

	assert.NoError(t, task.Do(kit.New()))
}
