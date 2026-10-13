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

// Package task 定义 task-server 的定时任务。
package task

import (
	"time"

	"hcm/cmd/task-server/logics/asyncflowcleanup"
	"hcm/pkg/cc"
	"hcm/pkg/criteria/enumor"
	croncore "hcm/pkg/cron/core"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/serviced"
	cvt "hcm/pkg/tools/converter"
)

// AsyncFlowAndTaskCleanupTask is the cron task for async flow and task history data cleanup.
type AsyncFlowAndTaskCleanupTask struct {
	logics *asyncflowcleanup.Logics
	sd     serviced.State
}

// NewAsyncFlowAndTaskCleanupTask creates a new async flow and task cleanup task.
func NewAsyncFlowAndTaskCleanupTask(logics *asyncflowcleanup.Logics, sd serviced.State) (croncore.Task, error) {
	return &AsyncFlowAndTaskCleanupTask{
		logics: logics,
		sd:     sd,
	}, nil
}

// Name returns the task name.
func (t *AsyncFlowAndTaskCleanupTask) Name() string {
	return string(enumor.CronTaskAsyncFlowAndTaskCleanup)
}

// Next returns the next execution time.
func (t *AsyncFlowAndTaskCleanupTask) Next() (time.Time, error) {
	interval := cvt.PtrToVal(cc.TaskServer().AsyncFlowAndTaskCleanup.IntervalMin)

	return time.Now().Add(time.Duration(interval) * time.Minute), nil
}

// GetURL returns the URL used to trigger the task externally.
func (t *AsyncFlowAndTaskCleanupTask) GetURL() string {
	return "/async_flow_and_task/cleanup"
}

// Do executes the task, only the master node does the real cleanup.
func (t *AsyncFlowAndTaskCleanupTask) Do(kt *kit.Kit) error {
	if t.sd == nil || !t.sd.IsMaster() {
		logs.Infof("current node is not master, skip async flow and task cleanup, rid: %s", kt.Rid)
		return nil
	}

	logs.Infof("master node executing async flow and task cleanup, rid: %s", kt.Rid)

	result, err := t.logics.Cleanup(kt)
	if err != nil {
		// 开关关闭或上一轮仍在执行，属于预期内的跳过，降级为 info 避免 cron 框架把它记为执行失败
		if asyncflowcleanup.IsSkipped(err) {
			logs.Infof("async flow and task cleanup skipped, reason: %s, rid: %s", err.Error(), kt.Rid)
			return nil
		}

		logs.Errorf("async flow and task cleanup failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	logs.Infof("async flow and task cleanup task done, flow: %d, task: %d, durationMs: %d, interrupted: %v, rid: %s",
		result.DeletedFlowCount, result.DeletedTaskCount, result.DurationMs, result.Interrupted, kt.Rid)

	return nil
}
