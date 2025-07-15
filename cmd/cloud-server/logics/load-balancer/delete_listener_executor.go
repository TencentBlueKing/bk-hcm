/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package lblogic

import (
	"encoding/json"
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	coretask "hcm/pkg/api/core/task"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/task"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

func newBatchDeleteListenerExecutor(cli *dataservice.Client, taskCli *taskserver.Client,
	vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *BatchDeleteListenerExecutor {

	return &BatchDeleteListenerExecutor{
		taskType:            enumor.DeleteListenerTaskType,
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// BatchDeleteListenerExecutor 删除监听器执行器
type BatchDeleteListenerExecutor struct {
	*basePreviewExecutor

	taskType    enumor.TaskType
	taskCli     *taskserver.Client
	params      *dataproto.BatchDeleteListenerReq
	details     []*corelb.BaseListener
	taskDetails []*batchDeleteListenerTaskDetail
}

// 用于记录 detail - 异步任务flow&task - 任务管理 之间的关系
type batchDeleteListenerTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*corelb.BaseListener
}

// Execute 导入执行器的唯一入口
func (c *BatchDeleteListenerExecutor) Execute(kt *kit.Kit, source enumor.TaskManagementSource,
	rawDetails json.RawMessage) (string, error) {

	var err error
	err = c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		logs.Errorf("validate delete listener failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
		return "", err
	}

	// 过滤不符合的数据
	c.filter()

	// 获取符合条件的监听器列表
	lblResp, err := c.dataServiceCli.Global.LoadBalancer.ListBatchListeners(kt, c.params)
	if err != nil {
		// 如果没有符合条件的监听器，直接返回
		if ef := errf.Error(err); ef != nil && ef.Code == errf.RecordNotFound {
			logs.ErrorJson("list batch listener is empty, err: %+v, lblReq: %+v, rid: %s", err, c.params, kt.Rid)
			return enumor.NoMatchTaskManageResult, nil
		}
		logs.ErrorJson("list batch listener failed, err: %v, lblReq: %+v, rid: %s", err, c.params, kt.Rid)
		return "", err
	}

	// 没查到符合的监听器，直接返回
	if len(lblResp.Details) == 0 {
		logs.Infof("list batch listener is empty, lblReq: %+v, rid: %s", cvt.PtrToVal(c.params), kt.Rid)
		return enumor.NoMatchTaskManageResult, nil
	}

	lbIDs := make([]string, 0)
	lblIDs := make([]string, 0)
	for _, item := range lblResp.Details {
		lbIDs = append(lbIDs, item.LbID)
		lblIDs = append(lblIDs, item.ID)
	}

	// 检查监听器下是否还有绑定的RS
	if err = c.checkListenerBindTargets(kt, lbIDs, lblIDs); err != nil {
		logs.Errorf("listener has bind target group and rs list, err: %v, source: %s, lblIDs: %v, rid: %s",
			err, source, lblIDs, kt.Rid)
		return "", err
	}

	// 把符合条件的监听器列表赋值给details
	c.details = lblResp.Details

	taskID, err := c.Run(kt, source)
	if err != nil {
		return "", err
	}

	return taskID, nil
}

// Run 执行器执行入口
func (c *BatchDeleteListenerExecutor) Run(kt *kit.Kit, source enumor.TaskManagementSource) (string, error) {
	// 创建异步管理任务、任务详情列表
	taskID, err := c.buildTaskManagementAndDetails(kt, source)
	if err != nil {
		logs.Errorf("create task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// 创建Flow
	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		logs.Errorf("build delete listener async flows failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
		return "", err
	}

	// 把Flow跟异步管理任务进行关联
	err = c.updateTaskManagementAndDetails(kt, flowIDs, taskID)
	if err != nil {
		logs.Errorf("update task management and details failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return "", err
	}

	return taskID, nil
}

// checkListenerBindTargets 检查负载均衡ID、监听器ID，是否绑定目标组及RS
func (c *BatchDeleteListenerExecutor) checkListenerBindTargets(kt *kit.Kit, lbIDs []string, lblIDs []string) error {
	// IN查询有500条的限制
	tgLblRelAllList := make([]corelb.BaseTargetListenerRuleRel, 0)
	for _, lblPartIDs := range slice.Split(lblIDs, int(core.DefaultMaxPageLimit)) {
		tgLblRelReq := &core.ListReq{
			Fields: []string{"target_group_id"},
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", c.vendor),
				tools.RuleIn("lb_id", lbIDs),
				tools.RuleIn("lbl_id", lblPartIDs),
			),
			Page: core.NewDefaultBasePage(),
		}
		for {
			tgLblRelList, err := c.dataServiceCli.Global.LoadBalancer.ListTargetGroupListenerRel(kt, tgLblRelReq)
			if err != nil {
				return err
			}

			tgLblRelAllList = append(tgLblRelAllList, tgLblRelList.Details...)
			if len(tgLblRelList.Details) < int(core.DefaultMaxPageLimit) {
				break
			}
			tgLblRelReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
	}

	// 该监听器没有绑定目标组
	if len(tgLblRelAllList) == 0 {
		return nil
	}

	targetGroupIDs := make([]string, 0)
	for _, item := range tgLblRelAllList {
		targetGroupIDs = append(targetGroupIDs, item.TargetGroupID)
	}

	// 只需要查询目标组绑定的权重非0的RS数量即可
	targetNum := uint64(0)
	for _, tgPartIDs := range slice.Split(targetGroupIDs, int(core.DefaultMaxPageLimit)) {
		targetReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("account_id", c.accountID),
				tools.RuleIn("target_group_id", tgPartIDs),
				tools.RuleNotEqual("weight", 0),
			),
			Page: &core.BasePage{Count: true},
		}
		targetList, err := c.dataServiceCli.Global.LoadBalancer.ListTarget(kt, targetReq)
		if err != nil {
			return err
		}
		targetNum += targetList.Count
	}

	if targetNum > 0 {
		return fmt.Errorf("listener[%v] has bind non-zero weight target num: %d, tgGroupIDs: %v",
			lblIDs, targetNum, targetGroupIDs)
	}

	return nil
}

func (c *BatchDeleteListenerExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := json.Unmarshal(rawDetail, &c.params)
	if err != nil {
		return err
	}
	c.params.Vendor = c.vendor
	c.params.AccountID = c.accountID
	c.params.BkBizID = c.bkBizID
	return nil
}

func (c *BatchDeleteListenerExecutor) validate(kt *kit.Kit) error {
	for cur, detail := range c.params.ListenerQueryList {
		if err := detail.Validate(); err != nil {
			logs.Errorf("detail[%d] validate failed, err: %v, item: %+v, rid: %s", cur, err, detail, kt.Rid)
			return fmt.Errorf("detail[%d] validate failed, item: %+v, err: %v", cur, detail, err)
		}
	}
	return nil
}

func (c *BatchDeleteListenerExecutor) filter() {
	return
}

// buildTaskManagementAndDetails 构建任务管理和详情
func (c *BatchDeleteListenerExecutor) buildTaskManagementAndDetails(kt *kit.Kit,
	source enumor.TaskManagementSource) (string, error) {

	taskID, err := c.createTaskManagement(kt, source)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
		return "", err
	}

	err = c.createTaskDetails(kt, taskID)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
		return "", err
	}

	return taskID, nil
}

// createTaskManagement 创建任务管理记录
func (c *BatchDeleteListenerExecutor) createTaskManagement(kt *kit.Kit,
	source enumor.TaskManagementSource) (string, error) {

	taskManagementCreateReq := &task.CreateManagementReq{
		Items: []task.CreateManagementField{
			{
				BkBizID:    c.bkBizID,
				Source:     source,
				Vendors:    []enumor.Vendor{c.vendor},
				AccountIDs: []string{c.accountID},
				Resource:   enumor.TaskManagementResClb,
				State:      enumor.TaskManagementRunning, // 默认:执行中
				Operations: []enumor.TaskOperation{enumor.TaskDeleteListener},
				Extension: &coretask.ManagementExt{
					LblDeleteReq: c.params,
				},
			},
		},
	}

	result, err := c.dataServiceCli.Global.TaskManagement.Create(kt, taskManagementCreateReq)
	if err != nil {
		logs.Errorf("create dataservice task management failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
		return "", err
	}

	if len(result.IDs) == 0 {
		return "", fmt.Errorf("create task management get new task ids failed")
	}

	return result.IDs[0], nil
}

// createTaskDetails 创建任务详情列表
func (c *BatchDeleteListenerExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskDeleteListener,
			State:            enumor.TaskDetailInit,
			Param:            detail,
		})
	}

	result, err := c.dataServiceCli.Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		logs.Errorf("create dataservice task detail failed, err: %v, taskID: %s, rid: %s", err, taskID, kt.Rid)
		return err
	}

	if len(result.IDs) != len(c.details) {
		return fmt.Errorf("create task details failed, expect created[%d] task details, but got [%d]",
			len(c.details), len(result.IDs))
	}

	for i := range result.IDs {
		taskDetail := &batchDeleteListenerTaskDetail{
			taskDetailID: result.IDs[i],
			BaseListener: c.details[i],
		}
		c.taskDetails = append(c.taskDetails, taskDetail)
	}

	return nil
}

func (c *BatchDeleteListenerExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// 按负载均衡ID进行分组
	clbToDetails := make(map[string][]*batchDeleteListenerTaskDetail)
	for _, detail := range c.taskDetails {
		clbToDetails[detail.CloudLbID] = append(clbToDetails[detail.CloudLbID], detail)
	}

	// 批量获取负载均衡列表
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli, c.vendor, c.accountID, c.bkBizID,
		cvt.MapKeyToSlice(clbToDetails))
	if err != nil {
		return nil, err
	}

	flowIDs := make([]string, 0, len(clbToDetails))
	for cloudClbID, details := range clbToDetails {
		lbInfo := lbMap[cloudClbID]
		flowID, err := c.buildFlow(kt, lbInfo.ID, details)
		if err != nil {
			logs.Errorf("build flow for delete listener clb(%s) failed, err: %v, rid: %s", cloudClbID, err, kt.Rid)
			err = c.updateTaskDetailsState(kt, enumor.TaskDetailFailed, details)
			if err != nil {
				logs.Errorf("update task details status failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			continue
		}
		flowIDs = append(flowIDs, flowID)
	}

	if len(flowIDs) == 0 {
		logs.Errorf("build clb flow failed, no clb need delete, clbToDetails: %+v, err: %v, rid: %s",
			clbToDetails, err, kt.Rid)
		return nil, fmt.Errorf("build clb flow failed, no clb need to be delete, err: %v", err)
	}
	return flowIDs, nil
}

func (c *BatchDeleteListenerExecutor) buildFlow(kt *kit.Kit, lbID string,
	details []*batchDeleteListenerTaskDetail) (string, error) {

	// 预检测
	lockRel, err := checkResFlowRel(kt, c.dataServiceCli, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, lbID: %s, lockRel: %+v, rid: %s",
			err, lbID, cvt.PtrToVal(lockRel), kt.Rid)
		return "", err
	}

	flowTasks, err := c.buildFlowTask(lbID, details)
	if err != nil {
		logs.Errorf("build flow task failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
		return "", err
	}

	flowID, err := c.createFlowTask(kt, lbID, flowTasks)
	if err != nil {
		logs.Errorf("create flow task failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
		return "", err
	}

	err = lockResFlowStatus(kt, c.dataServiceCli, c.taskCli, lbID,
		enumor.LoadBalancerCloudResType, flowID, enumor.DeleteListenerTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
		return "", err
	}

	for _, detail := range details {
		detail.flowID = flowID
	}

	return flowID, nil
}

func (c *BatchDeleteListenerExecutor) buildFlowTask(lbID string,
	details []*batchDeleteListenerTaskDetail) ([]ts.CustomFlowTask, error) {

	switch c.vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(lbID, details)
	default:
		return nil, fmt.Errorf("build flow task failed, lbID: %s, vendor: %s not supported", lbID, c.vendor)
	}
}

func (c *BatchDeleteListenerExecutor) buildTCloudFlowTask(lbID string, details []*batchDeleteListenerTaskDetail) (
	[]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	actionIDGenerator := counter.NewNumberCounterWithPrev(1, 10)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := actionIDGenerator()

		lblIDs := make([]string, 0, len(taskDetails))
		managementDetailIDs := make([]string, 0, len(taskDetails))
		for _, detail := range taskDetails {
			lblIDs = append(lblIDs, detail.ID)
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		}

		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskDeleteListener,
			Params: &actionlb.BatchTaskDeleteListenerOption{
				Vendor:              c.vendor,
				LoadBalancerID:      lbID,
				ManagementDetailIDs: managementDetailIDs,
				BatchDeleteReq:      &core.BatchDeleteReq{IDs: lblIDs},
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		result = append(result, tmpTask)

		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}

	return result, nil
}

func (c *BatchDeleteListenerExecutor) createFlowTask(kt *kit.Kit, lbID string,
	flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowBatchTaskDeleteListener,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       flowTasks,
		IsInitState: true,
	}
	result, err := c.taskCli.CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch delete listener custom flow failed, err: %v, lbID: %s, flowTasks: %+v, "+
			"rid: %s", err, lbID, flowTasks, kt.Rid)
		return "", err
	}

	flowID := result.ID
	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:     flowID,
				ResID:      lbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  nil,
				SubResType: enumor.ListenerCloudResType,
				TaskType:   c.taskType,
			},
		}},
	}
	_, err = c.taskCli.CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, lbID: %s, "+
			"rid: %s", err, flowID, lbID, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (c *BatchDeleteListenerExecutor) updateTaskManagementAndDetails(kt *kit.Kit,
	flowIDs []string, taskID string) error {

	if err := c.updateTaskManagement(kt, taskID, flowIDs); err != nil {
		logs.Errorf("update task management failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := c.updateTaskDetails(kt); err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

// updateTaskManagement 更新task_management的flow_id
func (c *BatchDeleteListenerExecutor) updateTaskManagement(kt *kit.Kit, taskID string, flowIDs []string) error {
	updateItem := task.UpdateTaskManagementField{
		ID:      taskID,
		FlowIDs: flowIDs,
	}
	updateReq := &task.UpdateManagementReq{
		Items: []task.UpdateTaskManagementField{updateItem},
	}
	err := c.dataServiceCli.Global.TaskManagement.Update(kt, updateReq)
	if err != nil {
		logs.Errorf("update task management failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return err
	}

	return nil
}

// updateTaskDetails 更新task_detail的flow_id和task_action_id
func (c *BatchDeleteListenerExecutor) updateTaskDetails(kt *kit.Kit) error {
	updateItems := make([]task.UpdateTaskDetailField, 0, len(c.taskDetails))
	for _, detail := range c.taskDetails {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:            detail.taskDetailID,
			FlowID:        detail.flowID,
			TaskActionIDs: []string{detail.actionID},
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := c.dataServiceCli.Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details failed, err: %v, req: %+v, rid: %s", err, updateDetailsReq, kt.Rid)
		return err
	}

	return nil
}

func (c *BatchDeleteListenerExecutor) updateTaskDetailsState(kt *kit.Kit, state enumor.TaskDetailState,
	taskDetails []*batchDeleteListenerTaskDetail) error {

	updateItems := make([]task.UpdateTaskDetailField, 0, len(taskDetails))
	for _, detail := range taskDetails {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:    detail.taskDetailID,
			State: state,
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := c.dataServiceCli.Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details state failed, err: %v, state: %s, updateDetailsReq: %+v, rid: %s",
			err, state, updateDetailsReq, kt.Rid)
		return err
	}
	return nil
}
