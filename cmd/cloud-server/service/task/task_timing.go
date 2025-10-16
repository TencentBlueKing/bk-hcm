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
			Fields: []string{"id", "state", "flow_ids"},
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

	var sum, success, failed, cancel int
	req := &core.ListReq{
		Filter: tools.EqualExpression("task_management_id", data.ID),
		Fields: []string{"state"},
		Page:   core.NewDefaultBasePage(),
	}
	for {
		list, err := c.DataService().Global.TaskDetail.List(kt, req)
		if err != nil {
			logs.Errorf("list task detail failed, err: %v, req: %+v, rid: %s", err, req, kt.Rid)
			return "", err
		}
		sum += len(list.Details)

		for _, detail := range list.Details {
			switch detail.State {
			case enumor.TaskDetailSuccess:
				success++
			case enumor.TaskDetailFailed:
				failed++
			case enumor.TaskDetailCancel:
				cancel++
			}
		}
		if len(list.Details) < int(core.DefaultMaxPageLimit) {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	// 任务详情里的数据结果总和不等于任务详情终态的总和，或者先创建任务管理数据，后创建任务详情数据，出现时间差，导致任务详情为空的情况，
	// 那么此时任务管理状态需保持running
	if success+failed+cancel != sum || sum == 0 {
		return enumor.TaskManagementRunning, nil
	}

	var finalState enumor.TaskManagementState
	if success != 0 {
		finalState = enumor.TaskManagementSuccess

		if failed != 0 {
			finalState = enumor.TaskManagementDeliverPartial
		}
	}
	if failed == sum {
		finalState = enumor.TaskManagementFailed
	}
	if cancel != 0 {
		finalState = enumor.TaskManagementCancel
	}
	if finalState == "" {
		return data.State, nil
	}

	updateReq := &datatask.UpdateManagementReq{
		Items: []datatask.UpdateTaskManagementField{{ID: data.ID, State: finalState}},
	}
	if err := c.DataService().Global.TaskManagement.Update(kt, updateReq); err != nil {
		logs.Errorf("update task management failed, err: %v, req: %+v, rid: %s", err, updateReq, kt.Rid)
		return "", err
	}

	return finalState, nil
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
