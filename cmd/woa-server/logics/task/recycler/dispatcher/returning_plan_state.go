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

// Package dispatcher implements state machine dispatcher
package dispatcher

import (
	"context"
	"errors"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// ReturningPlanState the action to be executed in returning plan state
type ReturningPlanState struct{}

// Name return the name of return failed state
func (rs *ReturningPlanState) Name() table.RecycleStatus {
	return table.RecycleStatusReturningPlan
}

// Execute executes action in return failed state
func (rs *ReturningPlanState) Execute(ctx EventContext) error {
	kt := core.NewBackendKit()
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to common context, rid: %s", kt.Rid)
		return errors.New("failed to convert to common context")
	}

	if taskCtx.Order == nil {
		logs.Errorf("state %s failed to execute, for invalid context order is nil, rid: %s", rs.Name(), kt.Rid)
		return fmt.Errorf("state %s failed to execute, for invalid context order is nil", rs.Name())
	}
	orderId := taskCtx.Order.SuborderID

	ev := &event.Event{
		Type: event.ReturnPlanSuccess,
	}
	logs.Infof("recycler:logics:cvm:ReturningPlanState:start, subOrderID: %s, ev: %+v, rid: %s",
		orderId, cvt.PtrToVal(ev), kt.Rid)
	if taskCtx.Order.ReturnForecast {
		ev = taskCtx.Dispatcher.returner.HandleReturnPlan(kt, taskCtx.Order)
	}
	logs.Infof("recycler:logics:cvm:ReturningPlanState:finish, subOrderID: %s, ev: %+v, rid: %s",
		orderId, cvt.PtrToVal(ev), kt.Rid)
	return rs.UpdateState(ctx, ev)
}

// UpdateState update next state
func (rs *ReturningPlanState) UpdateState(ctx EventContext, ev *event.Event) error {
	taskCtx, ok := ctx.(*CommonContext)
	if !ok {
		logs.Errorf("failed to convert to CommonContext, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			rs.Name())
		return fmt.Errorf("failed to convert toCommonContext, subOrderId: %s, state: %s", taskCtx.Order.SuborderID,
			rs.Name())
	}

	// set next state
	if errUpdate := rs.setNextState(taskCtx.Order, ev); errUpdate != nil {
		logs.Errorf("failed to update recycle order state, subOrderId: %s, err: %v", taskCtx.Order.SuborderID,
			errUpdate)
		return errUpdate
	}

	if taskCtx.Dispatcher == nil {
		logs.Errorf("recycler:logics:cvm:ReturningPlanState:failed, failed to add order to dispatch, "+
			"for dispatcher is nil, subOrderId: %s, state: %s", taskCtx.Order.SuborderID, rs.Name())
		return fmt.Errorf("failed to add order to dispatch, for dispatcher is nil, subOrderId: %s, state: %s",
			taskCtx.Order.SuborderID, rs.Name())
	}

	taskCtx.Dispatcher.Add(taskCtx.Order.SuborderID)

	// 记录日志
	logs.Infof("recycler: finish return plan state, subOrderId: %s, ev: %+v", taskCtx.Order.SuborderID,
		cvt.PtrToVal(ev))
	return nil
}

func (rs *ReturningPlanState) setNextState(order *table.RecycleOrder, ev *event.Event) error {
	filter := mapstr.MapStr{
		"suborder_id": order.SuborderID,
	}

	update := mapstr.MapStr{
		"update_at": time.Now(),
	}

	switch ev.Type {
	case event.ReturnPlanSuccess:
		update["stage"] = table.RecycleStageDone
		update["status"] = table.RecycleStatusDone
		update["handler"] = "AUTO"
	case event.ReturnPlanFailed:
		update["status"] = table.RecycleStatusReturnPlanFailed
		if ev.Error != nil {
			update["message"] = ev.Error.Error()
		}
	default:
		logs.Errorf("unknown event type: %s, subOrderId: %s, status: %s", ev.Type, order.SuborderID, order.Status)
		return fmt.Errorf("unknown event type: %s, subOrderId: %s, status: %s", ev.Type, order.SuborderID, order.Status)
	}

	if err := dao.Set().RecycleOrder().UpdateRecycleOrder(context.Background(), &filter, &update); err != nil {
		logs.Warnf("failed to update recycle order %s, err: %v", order.SuborderID, err)
		return err
	}

	return nil
}
