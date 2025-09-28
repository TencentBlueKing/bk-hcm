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
	coretask "hcm/pkg/api/core/task"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/api/data-service/task"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

func newBatchListenerUnbindRsExecutor(cli *dataservice.Client, taskCli *taskserver.Client,
	vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *BatchListenerUnbindRsExecutor {

	return &BatchListenerUnbindRsExecutor{
		taskType:            enumor.ListenerUnbindRsTaskType,
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// BatchListenerUnbindRsExecutor 监听器批量解绑RS执行器
type BatchListenerUnbindRsExecutor struct {
	*basePreviewExecutor

	taskType    enumor.TaskType
	taskCli     *taskserver.Client
	params      *dataproto.ListListenerWithTargetsReq
	details     []*dataproto.ListBatchListenerResult
	taskDetails []*batchListenerUnbindRsTaskDetail
}

// 用于记录 detail - 异步任务flow&task - 任务管理 之间的关系
type batchListenerUnbindRsTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*dataproto.ListBatchListenerResult
}

// Execute 导入执行器的唯一入口
func (c *BatchListenerUnbindRsExecutor) Execute(kt *kit.Kit, source enumor.TaskManagementSource,
	rawDetails json.RawMessage) (string, error) {

	var err error
	err = c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		logs.Errorf("validate listener unbind rs failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
		return "", err
	}

	// 过滤不符合的数据
	c.filter()

	// 获取符合条件的监听器列表
	lblResp, err := c.dataServiceCli.Global.LoadBalancer.ListLoadBalancerListenerWithTargets(kt, c.params)
	if err != nil {
		logs.Errorf("list batch listener by rsip failed, lblReq: %+v, err: %v, rid: %s", c.params, err, kt.Rid)
		return "", err
	}

	// 没查到符合的监听器，直接返回
	if len(lblResp.Details) == 0 {
		logs.Warnf("list batch listener by rsip is empty, lblReq: %+v, rid: %s", cvt.PtrToVal(c.params), kt.Rid)
		return enumor.NoMatchTaskManageResult, nil
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
func (c *BatchListenerUnbindRsExecutor) Run(kt *kit.Kit, source enumor.TaskManagementSource) (string, error) {
	// 创建异步管理任务、任务详情列表
	taskID, err := c.buildTaskManagementAndDetails(kt, source)
	if err != nil {
		logs.Errorf("create task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// 创建Flow
	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		logs.Errorf("build listener unbind rs async flows failed, err: %v, source: %s, rid: %s", err, source, kt.Rid)
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

func (c *BatchListenerUnbindRsExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := json.Unmarshal(rawDetail, &c.params)
	if err != nil {
		return err
	}
	c.params.Vendor = c.vendor
	c.params.AccountID = c.accountID
	c.params.BkBizID = c.bkBizID
	return nil
}

func (c *BatchListenerUnbindRsExecutor) validate(kt *kit.Kit) error {
	for cur, detail := range c.params.ListenerQueryList {
		if err := detail.Validate(); err != nil {
			logs.Errorf("detail[%d] validate failed, err: %v, item: %+v, rid: %s", cur, err, detail, kt.Rid)
			return fmt.Errorf("detail[%d] validate failed, item: %+v, err: %v", cur, detail, err)
		}
	}
	return nil
}

func (c *BatchListenerUnbindRsExecutor) filter() {
	return
}

func (c *BatchListenerUnbindRsExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// 按负载均衡ID进行分组
	clbToDetails := make(map[string][]*batchListenerUnbindRsTaskDetail)
	for _, detail := range c.taskDetails {
		clbToDetails[detail.CloudClbID] = append(clbToDetails[detail.CloudClbID], detail)
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
			logs.Errorf("build flow for unbind listener clb: %s failed, err: %v, rid: %s", cloudClbID, err, kt.Rid)
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
		logs.Errorf("build clb flow failed, no clb need modified, clbToDetails: %+v, rid: %s", clbToDetails, kt.Rid)
		return nil, fmt.Errorf("build clb flow failed, no clb need to be modified")
	}

	return flowIDs, nil
}

// buildTaskManagementAndDetails 构建任务管理和详情
func (c *BatchListenerUnbindRsExecutor) buildTaskManagementAndDetails(
	kt *kit.Kit, source enumor.TaskManagementSource) (string, error) {

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
func (c *BatchListenerUnbindRsExecutor) createTaskManagement(
	kt *kit.Kit, source enumor.TaskManagementSource) (string, error) {

	taskManagementCreateReq := &task.CreateManagementReq{
		Items: []task.CreateManagementField{
			{
				BkBizID:    c.bkBizID,
				Source:     source,
				Vendors:    []enumor.Vendor{c.vendor},
				AccountIDs: []string{c.accountID},
				Resource:   enumor.TaskManagementResClb,
				State:      enumor.TaskManagementRunning, // 默认:执行中
				Operations: []enumor.TaskOperation{enumor.TaskUnbindListenerRs},
				Extension: &coretask.ManagementExt{
					LblTargetsReq: c.params,
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
func (c *BatchListenerUnbindRsExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskUnbindListenerRs,
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
		return fmt.Errorf("create task details failed, operation: %s, expect created[%d] task details, but got [%d]",
			enumor.TaskUnbindListenerRs, len(c.details), len(result.IDs))
	}

	for i := range result.IDs {
		taskDetail := &batchListenerUnbindRsTaskDetail{
			taskDetailID:            result.IDs[i],
			ListBatchListenerResult: c.details[i],
		}
		c.taskDetails = append(c.taskDetails, taskDetail)
	}

	return nil
}

func (c *BatchListenerUnbindRsExecutor) buildFlow(kt *kit.Kit, lbID string,
	details []*batchListenerUnbindRsTaskDetail) (string, error) {

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
		enumor.LoadBalancerCloudResType, flowID, enumor.ListenerUnbindRsTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, lbID: %s, rid: %s", err, lbID, kt.Rid)
		return "", err
	}

	for _, detail := range details {
		detail.flowID = flowID
	}

	return flowID, nil
}

func (c *BatchListenerUnbindRsExecutor) buildFlowTask(lbID string,
	details []*batchListenerUnbindRsTaskDetail) ([]ts.CustomFlowTask, error) {

	switch c.vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(lbID, details)
	default:
		return nil, fmt.Errorf("build flow task failed, lbID: %s, vendor: %s not supported", lbID, c.vendor)
	}
}

func (c *BatchListenerUnbindRsExecutor) buildTCloudFlowTask(lbID string, details []*batchListenerUnbindRsTaskDetail) (
	[]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	actionIDGenerator := counter.NewNumberCounterWithPrev(1, 10)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := actionIDGenerator()

		lblRsList := make([]*hclb.TCloudBatchUnbindRsReq, 0, len(taskDetails))
		managementDetailIDs := make([]string, 0, len(taskDetails))
		for _, detail := range taskDetails {
			unbindRsReq := &hclb.TCloudBatchUnbindRsReq{
				AccountID:           c.accountID,
				Region:              detail.Region,
				Vendor:              c.vendor,
				LoadBalancerCloudId: detail.CloudClbID,
				Details:             make([]*dataproto.ListBatchListenerResult, 0),
			}
			unbindRsReq.Details = append(unbindRsReq.Details, detail.ListBatchListenerResult)
			lblRsList = append(lblRsList, unbindRsReq)
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		}

		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudUnBindTarget,
			Params: &actionlb.BatchTaskUnBindTargetOption{
				Vendor:              c.vendor,
				LoadBalancerID:      lbID,
				ManagementDetailIDs: managementDetailIDs,
				LblList:             lblRsList,
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

func (c *BatchListenerUnbindRsExecutor) createFlowTask(kt *kit.Kit, lbID string,
	flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowBatchTaskListenerUnBindTarget,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       flowTasks,
		IsInitState: true,
	}
	result, err := c.taskCli.CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch listener unbind rs custom flow failed, err: %v, lbID: %s, "+
			"flowTasks: %+v, rid: %s", err, lbID, flowTasks, kt.Rid)
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

func (c *BatchListenerUnbindRsExecutor) updateTaskManagementAndDetails(kt *kit.Kit,
	flowIDs []string, taskID string) error {

	if err := c.updateTaskManagement(kt, taskID, flowIDs); err != nil {
		logs.Errorf("update task management failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return err
	}

	if err := c.updateTaskDetails(kt); err != nil {
		logs.Errorf("update task details failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return err
	}
	return nil
}

// updateTaskManagement 更新task_management的flow_id
func (c *BatchListenerUnbindRsExecutor) updateTaskManagement(kt *kit.Kit, taskID string, flowIDs []string) error {
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
func (c *BatchListenerUnbindRsExecutor) updateTaskDetails(kt *kit.Kit) error {
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
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (c *BatchListenerUnbindRsExecutor) updateTaskDetailsState(kt *kit.Kit, state enumor.TaskDetailState,
	taskDetails []*batchListenerUnbindRsTaskDetail) error {

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
