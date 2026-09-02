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
	"fmt"
	"strconv"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	coretask "hcm/pkg/api/core/task"
	"hcm/pkg/api/data-service/task"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/producer"
	"hcm/pkg/cc"
	"hcm/pkg/client"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

// CreateCondSyncTask 创建负载均衡条件同步的任务管理与任务流
func CreateCondSyncTask(kt *kit.Kit, cliSet *client.ClientSet, opt *CondSyncLoadBalancerOption,
	regionDiffs []LoadBalancerRegionDiff) (string, error) {

	executor := newCondSyncLoadBalancerExecutor(cliSet.DataService(), cliSet.TaskServer(), opt, regionDiffs)
	return executor.Run(kt)
}

func newCondSyncLoadBalancerExecutor(cli *dataservice.Client, taskCli *taskserver.Client, opt *CondSyncLoadBalancerOption,
	regionDiffs []LoadBalancerRegionDiff) *CondSyncLoadBalancerExecutor {

	executor := &CondSyncLoadBalancerExecutor{
		operationType:  enumor.TaskSyncLoadBalancer,
		dataServiceCli: cli,
		taskCli:        taskCli,
		opt:            opt,
		bkBizID:        opt.bkBizID(),
	}
	executor.batches = executor.buildBatches(regionDiffs)

	return executor
}

// CondSyncLoadBalancerExecutor 负载均衡条件同步执行器
type CondSyncLoadBalancerExecutor struct {
	operationType  enumor.TaskOperation
	dataServiceCli *dataservice.Client
	taskCli        *taskserver.Client

	opt *CondSyncLoadBalancerOption
	// bkBizID 任务记录落库用的业务ID，未分配业务时为 constant.UnassignedBiz
	bkBizID int64
	// batches 已按先清理后同步排好序的任务批次，一个批次对应一个任务动作
	batches []*condSyncLoadBalancerBatch
}

// condSyncLoadBalancerBatch 一个批次对应一个任务动作，批次内的负载均衡同地域、同操作类型。
// 云上清理与同步都按地域调用下游接口，因此批次不跨地域。
type condSyncLoadBalancerBatch struct {
	region string
	// isDelete 为真表示清理云上已不存在的负载均衡，否则表示同步云上仍存在的负载均衡
	isDelete bool
	// actionID 该批次对应的任务动作ID，构建任务动作时回填
	actionID string
	details  []*condSyncTaskDetail
}

// condSyncTaskDetail 一台负载均衡一条任务详情，taskDetailID 在创建任务详情后回填
type condSyncTaskDetail struct {
	taskDetailID string
	param        *condSyncTaskDetailParam
}

// condSyncTaskDetailParam is stored in task_detail.param for displaying a single CLB sync item.
// Field names follow existing CLB task details so the frontend can reuse CLB VIP/ID columns.
type condSyncTaskDetailParam struct {
	// Op 本条详情对应的操作，枚举值：create/update/delete，对应页面「类别」列
	Op enumor.CondSyncLbOperation `json:"op"`
	// AccountID 账号ID
	AccountID string `json:"account_id"`
	// CloudLbID 云上负载均衡ID，对应页面「CLB ID」列
	CloudLbID string `json:"cloud_lb_id"`
	// ClbVipDomain 负载均衡 VIP，对应页面「CLB VIP/域名」列，只存单个地址，不与域名拼接
	ClbVipDomain string `json:"clb_vip_domain"`
	// Domain 负载均衡域名
	Domain string `json:"domain"`
	// Region 地域
	Region string `json:"region"`
}

// bkBizID 任务记录落库用的业务ID。
func (opt *CondSyncLoadBalancerOption) bkBizID() int64 {
	if opt == nil || opt.BkBizID == 0 {
		return constant.UnassignedBiz
	}

	return opt.BkBizID
}

// Run 执行器执行入口
func (c *CondSyncLoadBalancerExecutor) Run(kt *kit.Kit) (string, error) {
	// 创建异步管理任务、任务详情列表
	taskID, err := c.buildTaskManagementAndDetails(kt)
	if err != nil {
		logs.Errorf("create task management and details failed, err: %v, account: %s, rid: %s",
			err, c.opt.AccountID, kt.Rid)
		return "", err
	}

	// 创建Flow
	flowID, err := c.buildFlow(kt)
	if err != nil {
		logs.Errorf("build conditional sync load balancer flow failed, err: %v, account: %s, taskID: %s, rid: %s",
			err, c.opt.AccountID, taskID, kt.Rid)
		return "", err
	}

	// 把Flow跟异步管理任务进行关联
	if err = c.updateTaskManagementAndDetails(kt, taskID, flowID); err != nil {
		logs.Errorf("update task management and details failed, err: %v, taskID: %s, flowID: %s, rid: %s",
			err, taskID, flowID, kt.Rid)
		return "", err
	}

	if err = c.startFlow(kt, flowID); err != nil {
		logs.Errorf("start conditional sync load balancer flow failed, err: %v, taskID: %s, flowID: %s, rid: %s",
			err, taskID, flowID, kt.Rid)
		return "", err
	}

	logs.Infof("create conditional sync load balancer task success, vendor: %s, account: %s, bk_biz_id: %d, "+
		"taskID: %s, flowID: %s, batchCount: %d, rid: %s", c.opt.Vendor, c.opt.AccountID, c.bkBizID,
		taskID, flowID, len(c.batches), kt.Rid)

	return taskID, nil
}

// buildBatches 把各地域差异按先清理后同步的顺序拆成任务批次。
// 批大小在此处读取配置，不作为编排参数逐层透传。
func (c *CondSyncLoadBalancerExecutor) buildBatches(
	regionDiffs []LoadBalancerRegionDiff) []*condSyncLoadBalancerBatch {

	clbCondSync := cc.CloudServer().ConcurrentConfig.ClbCondSync
	deleteBatchSize := cvt.PtrToVal(clbCondSync.DeleteBatchSize)

	// 待同步的负载均衡数量超过阈值时放大批次，避免任务动作过多导致下游同步 worker 空转开销累积
	upsertCount := 0
	for _, diff := range regionDiffs {
		upsertCount += len(diff.Create) + len(diff.Update)
	}
	upsertBatchSize := cvt.PtrToVal(clbCondSync.UpsertBatchSize)
	if upsertCount > cvt.PtrToVal(clbCondSync.LargeUpsertThreshold) {
		upsertBatchSize = cvt.PtrToVal(clbCondSync.LargeUpsertBatchSize)
	}

	// 所有地域的清理批次都排在同步批次之前
	batches := make([]*condSyncLoadBalancerBatch, 0)
	for _, diff := range regionDiffs {
		deletes := c.newTaskDetails(enumor.CondSyncLbOpDelete, diff.Delete)
		batches = append(batches, splitCondSyncBatch(diff.Region, true, deletes, deleteBatchSize)...)
	}
	for _, diff := range regionDiffs {
		upserts := c.newTaskDetails(enumor.CondSyncLbOpCreate, diff.Create)
		upserts = append(upserts, c.newTaskDetails(enumor.CondSyncLbOpUpdate, diff.Update)...)
		batches = append(batches, splitCondSyncBatch(diff.Region, false, upserts, upsertBatchSize)...)
	}

	return batches
}

// newTaskDetails 把差异转成任务详情，delete 的展示信息来自DB，create/update 的展示信息来自云上
func (c *CondSyncLoadBalancerExecutor) newTaskDetails(op enumor.CondSyncLbOperation,
	briefs []corelb.LoadBalancerBrief) []*condSyncTaskDetail {

	details := make([]*condSyncTaskDetail, 0, len(briefs))
	for _, brief := range briefs {
		clbVipDomain := brief.Address
		if clbVipDomain == "" {
			clbVipDomain = brief.AddressIPv6
		}
		details = append(details, &condSyncTaskDetail{
			param: &condSyncTaskDetailParam{
				Op:           op,
				AccountID:    c.opt.AccountID,
				CloudLbID:    brief.CloudID,
				ClbVipDomain: clbVipDomain,
				Domain:       brief.Domain,
				Region:       brief.Region,
			},
		})
	}

	return details
}

func splitCondSyncBatch(region string, isDelete bool, details []*condSyncTaskDetail,
	batchSize int) []*condSyncLoadBalancerBatch {

	batches := make([]*condSyncLoadBalancerBatch, 0)
	for _, partDetails := range slice.Split(details, batchSize) {
		batches = append(batches, &condSyncLoadBalancerBatch{
			region:   region,
			isDelete: isDelete,
			details:  partDetails,
		})
	}

	return batches
}

// buildTaskManagementAndDetails 构建任务管理和详情
func (c *CondSyncLoadBalancerExecutor) buildTaskManagementAndDetails(kt *kit.Kit) (string, error) {
	taskID, err := c.createTaskManagement(kt)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, account: %s, rid: %s", err, c.opt.AccountID, kt.Rid)
		return "", err
	}

	if err = c.createTaskDetails(kt, taskID); err != nil {
		logs.Errorf("create task details failed, err: %v, taskID: %s, rid: %s", err, taskID, kt.Rid)
		return "", err
	}

	return taskID, nil
}

// createTaskManagement 创建任务管理记录
func (c *CondSyncLoadBalancerExecutor) createTaskManagement(kt *kit.Kit) (string, error) {
	createReq := &task.CreateManagementReq{
		Items: []task.CreateManagementField{
			{
				BkBizID:    c.bkBizID,
				Source:     enumor.TaskManagementSourceAPI.RefineByRequestSource(kt.RequestSource),
				Vendors:    []enumor.Vendor{c.opt.Vendor},
				AccountIDs: []string{c.opt.AccountID},
				Resource:   enumor.TaskManagementResClb,
				State:      enumor.TaskManagementRunning, // 默认:执行中
				Operations: []enumor.TaskOperation{c.operationType},
				Extension:  &coretask.ManagementExt{RegionIDs: c.opt.Regions},
			},
		},
	}

	result, err := c.dataServiceCli.Global.TaskManagement.Create(kt, createReq)
	if err != nil {
		logs.Errorf("create dataservice task management failed, err: %v, account: %s, rid: %s",
			err, c.opt.AccountID, kt.Rid)
		return "", err
	}

	if len(result.IDs) == 0 {
		return "", fmt.Errorf("create task management get new task ids failed")
	}

	return result.IDs[0], nil
}

// createTaskDetails 创建任务详情列表，一台负载均衡一条详情
func (c *CondSyncLoadBalancerExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {
	details := make([]*condSyncTaskDetail, 0)
	for _, batch := range c.batches {
		details = append(details, batch.details...)
	}

	items := make([]task.CreateDetailField, 0, len(details))
	for _, detail := range details {
		items = append(items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        c.operationType,
			State:            enumor.TaskDetailInit,
			Param:            detail.param,
		})
	}

	detailIDs := make([]string, 0, len(items))
	for _, partItems := range slice.Split(items, int(core.DefaultMaxPageLimit)) {
		result, err := c.dataServiceCli.Global.TaskDetail.Create(kt, &task.CreateDetailReq{Items: partItems})
		if err != nil {
			logs.Errorf("create dataservice task detail failed, err: %v, taskID: %s, rid: %s", err, taskID, kt.Rid)
			return err
		}
		detailIDs = append(detailIDs, result.IDs...)
	}

	if len(detailIDs) != len(items) {
		return fmt.Errorf("create task details failed, operation: %s, expect created[%d] task details, but got [%d]",
			c.operationType, len(items), len(detailIDs))
	}

	for i, detail := range details {
		detail.taskDetailID = detailIDs[i]
	}

	return nil
}

func (c *CondSyncLoadBalancerExecutor) buildFlow(kt *kit.Kit) (string, error) {
	flowTasks := c.buildFlowTasks()
	if len(flowTasks) == 0 {
		return "", fmt.Errorf("build conditional sync load balancer flow failed, no load balancer need to be synced")
	}

	addReq := &ts.AddCustomFlowReq{
		Name:        enumor.FlowBatchTaskSyncLoadBalancer,
		ShareData:   NewSubmitFlowShareData(c.bkBizID, c.opt.Vendor, OperationType(c.operationType), nil),
		Tasks:       flowTasks,
		IsInitState: true,
	}
	result, err := c.taskCli.CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create conditional sync load balancer custom flow failed, err: %v, "+
			"account: %s, rid: %s", err, c.opt.AccountID, kt.Rid)
		return "", err
	}

	return result.ID, nil
}

func (c *CondSyncLoadBalancerExecutor) startFlow(kt *kit.Kit, flowID string) error {
	req := &producer.UpdateCustomFlowStateOption{
		FlowInfos: []backend.UpdateFlowInfo{{
			ID:     flowID,
			Source: enumor.FlowInit,
			Target: enumor.FlowPending,
		}},
	}

	return c.taskCli.UpdateCustomFlowState(kt, req)
}

// buildFlowTasks 一个批次一个任务动作，动作之间用 depends_on 按批次顺序串行，
// 前一个动作失败时后续动作不再执行。批次已排好序，因此最后一个清理动作也是第一个同步动作的前置动作。
func (c *CondSyncLoadBalancerExecutor) buildFlowTasks() []ts.CustomFlowTask {
	flowTasks := make([]ts.CustomFlowTask, 0, len(c.batches))
	for i, batch := range c.batches {
		batch.actionID = strconv.Itoa(i + 1)

		var flowTask ts.CustomFlowTask
		if batch.isDelete {
			flowTask = c.buildDeleteFlowTask(batch)
		} else {
			flowTask = c.buildUpsertFlowTask(batch)
		}
		if i > 0 {
			flowTask.DependOn = []action.ActIDType{action.ActIDType(c.batches[i-1].actionID)}
		}

		flowTasks = append(flowTasks, flowTask)
	}

	return flowTasks
}

func (c *CondSyncLoadBalancerExecutor) buildDeleteFlowTask(batch *condSyncLoadBalancerBatch) ts.CustomFlowTask {
	managementDetailIDs, cloudIDs := condSyncBatchIDs(batch.details)

	return ts.CustomFlowTask{
		ActionID:   action.ActIDType(batch.actionID),
		ActionName: enumor.ActionBatchTaskSyncLbDelete,
		Params: &actionlb.BatchTaskSyncLbDeleteOption{
			Vendor:              c.opt.Vendor,
			AccountID:           c.opt.AccountID,
			Region:              batch.region,
			ManagementDetailIDs: managementDetailIDs,
			CloudIDs:            cloudIDs,
		},
	}
}

func (c *CondSyncLoadBalancerExecutor) buildUpsertFlowTask(batch *condSyncLoadBalancerBatch) ts.CustomFlowTask {
	managementDetailIDs, cloudIDs := condSyncBatchIDs(batch.details)

	return ts.CustomFlowTask{
		ActionID:   action.ActIDType(batch.actionID),
		ActionName: enumor.ActionBatchTaskSyncLbUpsert,
		Params: &actionlb.BatchTaskSyncLbUpsertOption{
			Vendor:              c.opt.Vendor,
			AccountID:           c.opt.AccountID,
			Region:              batch.region,
			ManagementDetailIDs: managementDetailIDs,
			CloudIDs:            cloudIDs,
			TagFilters:          c.opt.TagFilters,
		},
	}
}

func condSyncBatchIDs(details []*condSyncTaskDetail) (managementDetailIDs, cloudIDs []string) {
	managementDetailIDs = make([]string, 0, len(details))
	cloudIDs = make([]string, 0, len(details))
	for _, detail := range details {
		managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		cloudIDs = append(cloudIDs, detail.param.CloudLbID)
	}

	return managementDetailIDs, cloudIDs
}

func (c *CondSyncLoadBalancerExecutor) updateTaskManagementAndDetails(kt *kit.Kit, taskID string, flowID string) error {

	if err := c.updateTaskManagement(kt, taskID, []string{flowID}); err != nil {
		logs.Errorf("update task management failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, []string{flowID}, kt.Rid)
		return err
	}

	if err := c.updateTaskDetails(kt, flowID); err != nil {
		logs.Errorf("update task details failed, err: %v, taskID: %s, flowID: %s, rid: %s",
			err, taskID, flowID, kt.Rid)
		return err
	}

	return nil
}

func (c *CondSyncLoadBalancerExecutor) updateTaskManagement(kt *kit.Kit, taskID string, flowIDs []string) error {
	updateReq := &task.UpdateManagementReq{
		Items: []task.UpdateTaskManagementField{{
			ID:      taskID,
			FlowIDs: flowIDs,
		}},
	}
	if err := c.dataServiceCli.Global.TaskManagement.Update(kt, updateReq); err != nil {
		logs.Errorf("update task management failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return err
	}

	return nil
}

// updateTaskDetails 更新task_detail的flow_id和task_action_id
func (c *CondSyncLoadBalancerExecutor) updateTaskDetails(kt *kit.Kit, flowID string) error {
	// 同一批次内的详情共用flow与动作ID，按ID集合批量更新，避免逐条更新拖慢任务创建
	for _, batch := range c.batches {
		for _, partDetails := range slice.Split(batch.details, constant.BatchOperationMaxLimit) {
			ids := slice.Map(partDetails, func(detail *condSyncTaskDetail) string { return detail.taskDetailID })
			updateReq := &task.BatchUpdateTaskDetailReq{
				IDs:           ids,
				FlowID:        flowID,
				TaskActionIDs: []string{batch.actionID},
			}
			if err := c.dataServiceCli.Global.TaskDetail.BatchUpdate(kt, updateReq); err != nil {
				logs.Errorf("update task details failed, err: %v, flowID: %s, actionID: %s, count: %d, rid: %s",
					err, flowID, batch.actionID, len(ids), kt.Rid)
				return err
			}
		}
	}

	return nil
}

// taskManagementLister 查询 task_management 的接口。
type taskManagementLister interface {
	List(kt *kit.Kit, req *core.ListReq) (*task.ListManagementResult, error)
}

// flowLister 查询 flow 的接口。
type flowLister interface {
	ListFlow(kt *kit.Kit, req *core.ListReq) (*ts.ListFlowResult, error)
}

// CheckRunningCondSyncTask rejects duplicate CLB conditional sync tasks before doing diff.
// 按账号维度互斥。
func CheckRunningCondSyncTask(kt *kit.Kit, mgmtCli taskManagementLister, flowCli flowLister,
	opt *CondSyncLoadBalancerOption) error {

	expr := tools.ExpressionAnd(
		tools.RuleEqual("state", enumor.TaskManagementRunning),
		tools.RuleEqual("resource", enumor.TaskManagementResClb),
		tools.RuleJsonOverlaps("operations", []enumor.TaskOperation{enumor.TaskSyncLoadBalancer}),
		tools.RuleJsonOverlaps("vendors", []enumor.Vendor{opt.Vendor}),
		tools.RuleJsonOverlaps("account_ids", []string{opt.AccountID}),
	)

	req := &core.ListReq{
		Filter: expr,
		Fields: []string{"id", "flow_ids"},
		Page:   &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit},
	}
	flowIDs := make([]string, 0)
	for {
		resp, err := mgmtCli.List(kt, req)
		if err != nil {
			logs.Errorf("list conditional sync load balancer task management failed, err: %v, req: %+v, rid: %s",
				err, req, kt.Rid)
			return err
		}

		for _, management := range resp.Details {
			flowIDs = append(flowIDs, management.FlowIDs...)
		}

		if uint(len(resp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		req.Page.Start += uint32(core.DefaultMaxPageLimit)
	}

	runningFlowID, err := findRunningCondSyncFlow(kt, flowCli, slice.Unique(flowIDs))
	if err != nil {
		return err
	}
	if runningFlowID == "" {
		return nil
	}

	logs.Infof("conditional sync load balancer task is running, vendor: %s, account: %s, bk_biz_id: %d, "+
		"flow: %s, rid: %s", opt.Vendor, opt.AccountID, opt.bkBizID(), runningFlowID, kt.Rid)

	return errf.Newf(errf.TooManyRequest, "load balancer conditional sync task is running, vendor: %s, "+
		"account: %s, bk_biz_id: %d", opt.Vendor, opt.AccountID, opt.bkBizID())
}

// findRunningCondSyncFlow 返回首个未进入终态的flow id，全部终态时返回空串
func findRunningCondSyncFlow(kt *kit.Kit, flowCli flowLister, flowIDs []string) (string, error) {
	if len(flowIDs) == 0 {
		return "", nil
	}

	for _, batch := range slice.Split(flowIDs, int(core.DefaultMaxPageLimit)) {
		req := &core.ListReq{
			Filter: tools.ContainersExpression("id", batch),
			Fields: []string{"id", "state"},
			Page:   core.NewDefaultBasePage(),
		}
		resp, err := flowCli.ListFlow(kt, req)
		if err != nil {
			logs.Errorf("list conditional sync load balancer flow failed, err: %v, req: %+v, rid: %s",
				err, req, kt.Rid)
			return "", err
		}

		for _, flow := range resp.Details {
			if flow.State != enumor.FlowCancel && flow.State != enumor.FlowSuccess &&
				flow.State != enumor.FlowFailed {
				return flow.ID, nil
			}
		}
	}

	return "", nil
}
