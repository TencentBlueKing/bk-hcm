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
	"errors"
	"time"

	"hcm/cmd/cloud-server/logics/tenant"
	"hcm/pkg/api/core"
	coretask "hcm/pkg/api/core/task"
	datatask "hcm/pkg/api/data-service/task"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/metrics"
	"hcm/pkg/serviced"
	"hcm/pkg/tools/slice"

	"golang.org/x/sync/errgroup"
)

// TimingHandleTaskMgmtState 定时更新任务管理数据状态
func TimingHandleTaskMgmtState(c *client.ClientSet, sd serviced.State, interval time.Duration) {
	if cc.CloudServer().TaskManagement.Disable {
		logs.Warnf("task management state background update has been disabled")
		return
	}
	for {
		time.Sleep(interval)

		if !sd.IsMaster() {
			continue
		}

		kt := core.NewBackendKit()
		listReq := &core.ListReq{
			Filter: tools.EqualExpression("state", enumor.TaskManagementRunning),
			// Extra fields are required so that, when the management
			// transitions to a terminal state below, we can emit
			// hcm_async_task_manage_* metrics with the correct business
			// dimensions and end-to-end cost. created_at is used to
			// compute (now - created_at) as the cost.
			Fields: []string{"id", "state", "flow_ids", "bk_biz_id", "vendors", "operations", "created_at"},
			Page:   core.NewDefaultBasePage(),
		}

		tenantIDs, err := tenant.ListAllTenantID(kt, c.DataService())
		if err != nil {
			logs.Errorf("failed to list all tenant ids, err: %v, rid: %s", err, kt.Rid)
			continue
		}

		eg, _ := errgroup.WithContext(kt.Ctx)
		for _, id := range tenantIDs {
			tenantID := id
			eg.Go(func() error {
				tenantKt := kt.NewSubKitWithTenant(tenantID)
				list, subErr := c.DataService().Global.TaskManagement.List(tenantKt, listReq)
				if subErr != nil {
					logs.Errorf("list task management failed, err: %v, tenant: %s, req: %+v, rid: %s", subErr,
						tenantID, listReq, tenantKt.Rid)
					return subErr
				}

				for _, management := range list.Details {
					if _, subErr = refreshTaskMgmtState(tenantKt, c, management); subErr != nil {
						logs.Errorf("refresh task management state failed, err: %v, tenant: %s, data: %+v, rid: %s",
							subErr, tenantID, management, tenantKt.Rid)
						return subErr
					}
				}
				return nil
			})
		}

		if err := eg.Wait(); err != nil {
			logs.Errorf("handle task management state failed, err: %v, rid: %s", err, kt.Rid)
			continue
		}
	}
}

// refreshTaskMgmtState 刷新任务管理数据状态，按照下面步骤执行:
// 1. 先判断任务管理数据对应的状态，如果状态处于未完结状态（即处于running），则返回；
// 2. 如果有处于running的数据，根据flow id查询下面的flow是否都已经执行完了，未执行完则返回；
// 3. 如果都已经执行完了，判断任务详情里的数据结果，根据结果更新任务管理数据状态，返回结果。
func refreshTaskMgmtState(kt *kit.Kit, c *client.ClientSet, data coretask.Management) (enumor.TaskManagementState,
	error) {

	if data.State != enumor.TaskManagementRunning {
		return data.State, nil
	}

	isDone, err := isFlowDone(kt, c, data.FlowIDs)
	if err != nil {
		logs.Errorf("failed to determine whether flow ends, err: %v, flow ids: %v, rid: %s", err, data.FlowIDs, kt.Rid)
		return "", err
	}
	if !isDone {
		return enumor.TaskManagementRunning, nil
	}

	summary, err := listTaskDetailSummary(kt, c, data.ID)
	if err != nil {
		return "", err
	}
	if !summary.isTerminal() {
		return enumor.TaskManagementRunning, nil
	}

	finalState := summary.finalState()
	if finalState == "" {
		return data.State, nil
	}
	if err := updateTaskMgmtState(kt, c, data.ID, finalState); err != nil {
		return "", err
	}

	emitTaskProgressMetrics(kt, data, summary.details, finalState)
	return finalState, nil
}

type taskDetailSummary struct {
	sum     int
	success int
	failed  int
	cancel  int
	details []coretask.Detail
}

func listTaskDetailSummary(kt *kit.Kit, c *client.ClientSet, taskManagementID string) (*taskDetailSummary, error) {
	summary := &taskDetailSummary{details: make([]coretask.Detail, 0)}
	req := &core.ListReq{
		Filter: tools.EqualExpression("task_management_id", taskManagementID),
		Fields: []string{"id", "state", "bk_biz_id", "operation", "reason", "created_at", "updated_at"},
		Page:   core.NewDefaultBasePage(),
	}

	for {
		list, err := c.DataService().Global.TaskDetail.List(kt, req)
		if err != nil {
			logs.Errorf("list task detail failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return nil, err
		}

		summary.sum += len(list.Details)
		summary.details = append(summary.details, list.Details...)
		for _, detail := range list.Details {
			switch detail.State {
			case enumor.TaskDetailSuccess:
				summary.success++
			case enumor.TaskDetailFailed:
				summary.failed++
			case enumor.TaskDetailCancel:
				summary.cancel++
			}
		}

		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			return summary, nil
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}
}

func (s *taskDetailSummary) isTerminal() bool {
	return s.sum != 0 && s.success+s.failed+s.cancel == s.sum
}

func (s *taskDetailSummary) finalState() enumor.TaskManagementState {
	var finalState enumor.TaskManagementState
	if s.success != 0 {
		finalState = enumor.TaskManagementSuccess
		if s.failed != 0 {
			finalState = enumor.TaskManagementDeliverPartial
		}
	}
	if s.failed == s.sum {
		finalState = enumor.TaskManagementFailed
	}
	if s.cancel != 0 {
		finalState = enumor.TaskManagementCancel
	}
	return finalState
}

func updateTaskMgmtState(kt *kit.Kit, c *client.ClientSet, taskManagementID string,
	finalState enumor.TaskManagementState) error {

	updateReq := &datatask.UpdateManagementReq{
		Items: []datatask.UpdateTaskManagementField{{ID: taskManagementID, State: finalState}},
	}
	if err := c.DataService().Global.TaskManagement.Update(kt, updateReq); err != nil {
		logs.Errorf("update task management failed, err: %v, req: %+v, rid: %s", err, updateReq, kt.Rid)
		return err
	}
	return nil
}

// emitTaskProgressMetrics emits one hcm_async_task_manage_* sample for the
// management and one hcm_async_task_detail_* sample per detail row. Vendor
// is taken from data.Vendors[0] (CLB managements always carry exactly one
// vendor). Operation is taken from data.Operations[0] for the management
// and from each detail.Operation for the details.
func emitTaskProgressMetrics(kt *kit.Kit, data coretask.Management, details []coretask.Detail,
	finalState enumor.TaskManagementState) {

	vendor := ""
	if len(data.Vendors) > 0 {
		vendor = string(data.Vendors[0])
	}
	mgmtOp := ""
	if len(data.Operations) > 0 {
		mgmtOp = string(data.Operations[0])
	}
	if !validManagementMetricDims(data, vendor, mgmtOp) {
		logs.Warnf("skip task management progress metric for incomplete dimensions, taskID: %s, bkBizID: %d, "+
			"vendor: %s, operation: %s, createdAt: %s, rid: %s", data.ID, data.BkBizID, vendor, mgmtOp,
			data.CreatedAt, kt.Rid)
	} else if mgmtCost, ok := terminalCost(data.CreatedAt, ""); !ok {
		logs.Warnf("skip task management progress metric for invalid cost, taskID: %s, createdAt: %s, "+
			"rid: %s", data.ID, data.CreatedAt, kt.Rid)
	} else {
		mgmtErrType := classifyManagementState(finalState)
		metrics.ObserveTaskManagement(data.BkBizID, vendor, mgmtOp, string(finalState), mgmtCost, mgmtErrType)
	}

	for _, detail := range details {
		// detail rows whose state is not a recognized terminal value are
		// skipped to avoid polluting the histogram with mid-flight rows.
		switch detail.State {
		case enumor.TaskDetailSuccess, enumor.TaskDetailFailed, enumor.TaskDetailCancel:
		default:
			continue
		}
		if !validDetailMetricDims(detail) {
			logs.Warnf("skip task detail progress metric for incomplete dimensions, taskID: %s, detailID: %s, "+
				"bkBizID: %d, operation: %s, createdAt: %s, updatedAt: %s, rid: %s", data.ID, detail.ID,
				detail.BkBizID, detail.Operation, detail.CreatedAt, detail.UpdatedAt, kt.Rid)
			continue
		}
		cost, ok := terminalCost(detail.CreatedAt, detail.UpdatedAt)
		if !ok {
			logs.Warnf("skip task detail progress metric for invalid cost, taskID: %s, detailID: %s, "+
				"createdAt: %s, updatedAt: %s, rid: %s", data.ID, detail.ID, detail.CreatedAt, detail.UpdatedAt,
				kt.Rid)
			continue
		}
		errType := classifyDetailState(detail.State, detail.Reason)
		metrics.ObserveTaskDetail(detail.BkBizID, vendor, string(detail.Operation),
			string(detail.State), cost, errType)
	}
}

func validManagementMetricDims(data coretask.Management, vendor, operation string) bool {
	return data.BkBizID > 0 && vendor != "" && operation != "" && data.CreatedAt != ""
}

func validDetailMetricDims(detail coretask.Detail) bool {
	return detail.BkBizID > 0 && detail.Operation != "" && detail.CreatedAt != "" && detail.UpdatedAt != ""
}

// terminalCost returns the duration between two RFC3339 timestamp strings.
// If `end` is empty, time.Now() is used. Parse failures return ok=false so
// callers can skip invalid histogram observations.
func terminalCost(start, end string) (cost time.Duration, ok bool) {
	startT, err := time.Parse(time.RFC3339, start)
	if err != nil {
		return 0, false
	}
	endT := time.Now()
	if end != "" {
		t, err := time.Parse(time.RFC3339, end)
		if err != nil {
			return 0, false
		}
		endT = t
	}
	if !endT.After(startT) {
		return 0, false
	}
	return endT.Sub(startT), true
}

// classifyManagementState maps a task management terminal state into a
// normalized err_type. Success returns ErrTypeOK so callers know to skip
// fail_total emission.
func classifyManagementState(state enumor.TaskManagementState) metrics.ErrType {
	switch state {
	case enumor.TaskManagementSuccess:
		return metrics.ErrTypeOK
	case enumor.TaskManagementCancel:
		return metrics.ErrTypeCancel
	case enumor.TaskManagementFailed, enumor.TaskManagementDeliverPartial:
		return metrics.ErrTypeHCMError
	default:
		return metrics.ErrTypeUnknown
	}
}

// classifyDetailState maps a task detail terminal state plus its reason
// string into a normalized err_type. The reason text (when present) is fed
// through metrics.ClassifyError so timeout / network / cloud_error are
// surfaced where possible.
func classifyDetailState(state enumor.TaskDetailState, reason string) metrics.ErrType {
	switch state {
	case enumor.TaskDetailSuccess:
		return metrics.ErrTypeOK
	case enumor.TaskDetailCancel:
		return metrics.ErrTypeCancel
	case enumor.TaskDetailFailed:
		if reason == "" {
			return metrics.ErrTypeHCMError
		}
		if et := metrics.ClassifyError(errors.New(reason)); et != metrics.ErrTypeOK {
			return et
		}
		return metrics.ErrTypeHCMError
	default:
		return metrics.ErrTypeUnknown
	}
}

func isFlowDone(kt *kit.Kit, c *client.ClientSet, flowIDs []string) (bool, error) {
	for _, batch := range slice.Split(flowIDs, int(core.DefaultMaxPageLimit)) {
		flowReq := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Fields: []string{"id", "state"},
			Page:   core.NewDefaultBasePage(),
		}
		list, err := c.TaskServer().ListFlow(kt, flowReq)
		if err != nil {
			logs.Errorf("list flow failed, err: %v, req: %+v, rid: %s", err, flowReq, kt.Rid)
			return false, err
		}
		for _, flow := range list.Details {
			if flow.State != enumor.FlowCancel && flow.State != enumor.FlowSuccess && flow.State != enumor.FlowFailed {
				return false, nil
			}
		}

		lockReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleIn("owner", batch),
			),
			Page: core.NewCountPage(),
		}
		resp, err := c.DataService().Global.LoadBalancer.ListResFlowLock(kt, lockReq)
		if err != nil {
			logs.Errorf("count res flow lock failed, err: %v, flowIDs: %v, rid: %s", err, batch, kt.Rid)
			return false, err
		}
		if resp.Count != 0 {
			return false, nil
		}
	}

	return true, nil
}
