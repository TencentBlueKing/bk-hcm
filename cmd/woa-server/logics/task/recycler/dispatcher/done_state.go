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

// DoneState the action to be executed in done state
type DoneState struct{}

// Name return the name of done state
func (ds *DoneState) Name() table.RecycleStatus {
	return table.RecycleStatusDone
}

// Execute executes action in done state
func (ds *DoneState) Execute(ctx EventContext) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context in done state")
		return fmt.Errorf("failed to convert to common context in done state")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil", ds.Name())
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", ds.Name())
	}

	// DONE 状态后触发回收成功通知检查
	if taskCtx.Dispatcher != nil && taskCtx.Dispatcher.notifier != nil {
		kt := core.NewBackendKit()
		kt.Ctx = taskCtx.Dispatcher.ctx
		if _, err := taskCtx.Dispatcher.notifier.SendSuccessNotice(kt, taskCtx.Order.OrderID); err != nil {
			logs.Errorf("state %s failed to execute, for notifier send success notice failed, err: %v, rid: %s",
				ds.Name(), err, kt.Rid)
			return err
		}
	}

	return nil
}

// UpdateState update next state
func (ds *DoneState) UpdateState(ctx EventContext, ev *event.Event) error {
	return nil
}
