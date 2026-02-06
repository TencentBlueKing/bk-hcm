/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package dispatcher implements the dispatcher of recycle task
package dispatcher

import (
	"fmt"

	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/api/core"
	"hcm/pkg/logs"
)

// DetectFailedState the action to be executed in detect failed state
type DetectFailedState struct{}

// Name return the name of detect failed state
func (ds *DetectFailedState) Name() table.RecycleStatus {
	return table.RecycleStatusDetectFailed
}

// Execute executes action in detect failed state
func (ds *DetectFailedState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context in detect failed state")
		return fmt.Errorf("failed to convert to common context in detect failed state")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", ds.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", ds.Name())
	}

	// DETECT_FAILED 状态后触发回收失败通知检查
	if taskCtx.Dispatcher != nil && taskCtx.Dispatcher.notifier != nil {
		kt := core.NewBackendKit()
		kt.Ctx = taskCtx.Dispatcher.ctx
		if _, err := taskCtx.Dispatcher.notifier.SendDetectFailedNotice(kt, taskCtx.Order.OrderID); err != nil {
			logs.Errorf("failed to send detect failed notice, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}

	return nil
}

// UpdateState update next state
func (ds *DetectFailedState) UpdateState(ctx EventContext, ev *event.Event) error {
	return nil
}
