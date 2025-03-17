/*
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

// Package monthtask  在完成二级账号原始账单pull, split, summary 步骤后进行，
// 每个厂商支持多种任务，每个类型任务都需要实现pull, split, summary 三个步骤
package monthtask

import (
	"fmt"

	"hcm/cmd/task-server/logics/action/bill/monthtask"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dataservice "hcm/pkg/api/data-service"
	dsbillapi "hcm/pkg/api/data-service/bill"
	taskserver "hcm/pkg/api/task-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// MonthDescriberConstructor construct month task describer
type MonthDescriberConstructor func(accountID string) MonthTaskDescriber

// MonthTaskDescriberRegistry month task describe registry
var monthTaskDescriberRegistry = make(map[enumor.Vendor]MonthDescriberConstructor)

// MonthTaskDescriber month task describe interface
type MonthTaskDescriber interface {
	// GetMonthTaskTypes return all month task types supported by this vendor,
	// month task should be executed by given order.
	GetMonthTaskTypes() []enumor.MonthTaskType

	// GetTaskExtension extension passed to task server
	GetTaskExtension() (map[string]string, error)
}

// GetMonthTaskDescriber get describer by vendor
func GetMonthTaskDescriber(vendor enumor.Vendor, accountCloudID string) MonthTaskDescriber {
	constructor, ok := monthTaskDescriberRegistry[vendor]
	if ok {
		return constructor(accountCloudID)
	}
	return nil
}

// NewDefaultMonthTaskRunner ...
func NewDefaultMonthTaskRunner(kt *kit.Kit, vendor enumor.Vendor, rootAccountID, rootAccountCloudID string,
	cli *client.ClientSet) *DefaultMonthTaskRunner {

	return &DefaultMonthTaskRunner{
		rootAccountID:      rootAccountID,
		rootAccountCloudID: rootAccountCloudID,
		vendor:             vendor,
		client:             cli,
	}
}

// DefaultMonthTaskRunner ...
type DefaultMonthTaskRunner struct {
	rootAccountID      string
	rootAccountCloudID string
	vendor             enumor.Vendor
	client             *client.ClientSet

	// ext extension option for root account
	ext map[string]string
}

func (r *DefaultMonthTaskRunner) getBillSummary(kt *kit.Kit, billYear, billMonth int) (
	*billcore.SummaryRoot, error) {

	req := &dsbillapi.BillSummaryRootListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("root_account_id", r.rootAccountID),
			tools.RuleEqual("vendor", r.vendor),
			tools.RuleEqual("bill_year", billYear),
			tools.RuleEqual("bill_month", billMonth)),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	result, err := r.client.DataService().Global.Bill.ListBillSummaryRoot(kt, req)
	if err != nil {
		return nil, fmt.Errorf("get root account bill summary failed, err %s", err.Error())
	}
	if len(result.Details) == 0 {
		return nil, fmt.Errorf("root account bill summary not found")
	}
	return result.Details[0], nil
}

// EnsureMonthTask month task loop
func (r *DefaultMonthTaskRunner) EnsureMonthTask(kt *kit.Kit, billYear, billMonth int) error {
	rootSummary, err := r.getBillSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	monthDescriber := GetMonthTaskDescriber(r.vendor, r.rootAccountCloudID)
	if monthDescriber == nil {
		logs.Infof("no month pull task for root account (%s/%s(%s)), skip, rid: %s",
			r.vendor, r.rootAccountCloudID, r.rootAccountID, kt.Rid)
		return nil
	}
	r.ext, err = monthDescriber.GetTaskExtension()
	if err != nil {
		logs.Errorf("fail got generate month task extension for root account (%s, %s), err: %v, rid: %s",
			r.rootAccountID, r.vendor, err, kt.Rid)
		return err
	}
	logs.Infof("[%s] %s(%s) %d-%02d monthtask setting extension: %v, rid: %s",
		r.vendor, r.rootAccountCloudID, r.rootAccountID, billYear, billMonth, r.ext, kt.Rid)

	mainSummaryList, err := r.listAllMainSummary(kt, billYear, billMonth)
	if err != nil {
		return err
	}
	isAllAccounted := calculateAccountingState(mainSummaryList, rootSummary)
	if !isAllAccounted {
		logs.Infof("not all main account bill summary for [%s] root account %s(%s), %d-%02d were accounted, wait, "+
			"rid: %s", r.vendor, r.rootAccountCloudID, r.rootAccountID, billYear, billMonth, kt.Rid)
		return nil
	}
	// 进入 月度任务执行阶段
	return r.executeMonthTask(kt, monthDescriber, rootSummary)
}

func (r *DefaultMonthTaskRunner) executeMonthTask(kt *kit.Kit, monthDescriber MonthTaskDescriber,
	rootSummary *billcore.SummaryRoot) error {

	monthTaskTypeOrders := monthDescriber.GetMonthTaskTypes()
	monthTasks, err := r.listMonthPullTaskStub(kt, rootSummary.BillYear, rootSummary.BillMonth, monthTaskTypeOrders)
	if err != nil {
		logs.Errorf("list month task stub for ensuring month task failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	monthTaskTypeMap := make(map[enumor.MonthTaskType]*billcore.MonthTask, len(monthTasks))
	for _, task := range monthTasks {
		monthTaskTypeMap[task.Type] = task
	}

	for _, curType := range monthTaskTypeOrders {
		monthTask := monthTaskTypeMap[curType]
		monthTask, err = r.ensureMonthTaskStub(kt, monthTask, rootSummary, curType)
		if err != nil {
			logs.Errorf("fail to ensure month task stub for [%s] %s(%s), type: %s, err: %v, rid: %s",
				r.vendor, r.rootAccountCloudID, r.rootAccountID, curType, err, kt.Rid)
			return err
		}
		switch monthTask.State {
		case enumor.RootAccountMonthBillTaskStatePulling:
			if err := r.ensureMonthTaskPullStep(kt, monthTask); err != nil {
				return err
			}
			return nil
		case enumor.RootAccountMonthBillTaskStatePulled:
			if err := r.ensureMonthTaskSplitStep(kt, monthTask); err != nil {
				return err
			}
			return nil
		case enumor.RootAccountMonthBillTaskStateSplit:
			if err := r.ensureMonthTaskAccountedStep(kt, monthTask); err != nil {
				return err
			}
			return nil
		case enumor.RootAccountMonthBillTaskStateAccounted:
			// 进入下一个类型
			continue
		}
	}
	return nil
}

func (r *DefaultMonthTaskRunner) ensureMonthTaskStub(kt *kit.Kit, monthTask *billcore.MonthTask,
	rootSummary *billcore.SummaryRoot, curType enumor.MonthTaskType) (*billcore.MonthTask, error) {

	if monthTask != nil && monthTask.VersionID == rootSummary.CurrentVersion {
		return monthTask, nil
	}
	// version id 不一致，删除旧版本
	if monthTask != nil && monthTask.VersionID != rootSummary.CurrentVersion {
		err := r.deleteMonthPullTaskStub(kt, rootSummary.BillYear, rootSummary.BillMonth, curType)
		if err != nil {
			logs.Errorf("fail to delete old [%s] month task stub, type: %s, id: %s, err: %v, rid: %s",
				r.vendor, curType, monthTask.ID, err, kt.Rid)
			return nil, err
		}
	}
	if err := r.createMonthPullTaskStub(kt, rootSummary, curType); err != nil {
		logs.Errorf("fail to create [%s] month task, type: %s, err: %v, rid: %s",
			r.vendor, curType, err, kt.Rid)
		return nil, err
	}
	// get after create
	monthTasks, err := r.listMonthPullTaskStub(kt, rootSummary.BillYear, rootSummary.BillMonth,
		[]enumor.MonthTaskType{curType})
	if err != nil {
		logs.Errorf("fail to list %s month task stub after create, type: %s, err: %v, rid: %s",
			r.vendor, curType, err, kt.Rid)
		return nil, err
	}
	if len(monthTasks) != 1 {
		logs.Errorf("month task stub after create count is not 1, vendor: %s, type: %s, count: %d, rid: %s",
			r.vendor, curType, len(monthTasks), kt.Rid)
	}
	return monthTasks[0], nil
}

// return true if all main account state of current root version is in `accounted` or `wait_month_task` state
func calculateAccountingState(mainSummaryList []*dsbillapi.BillSummaryMain, rootSummary *billcore.SummaryRoot) (
	isAllAccounted bool) {

	isAllAccounted = true

	for _, mainSummary := range mainSummaryList {
		if mainSummary.CurrentVersion != rootSummary.CurrentVersion {
			isAllAccounted = false
			return
		}
		if mainSummary.State != enumor.MainAccountBillSummaryStateAccounted &&
			mainSummary.State != enumor.MainAccountBillSummaryStateWaitMonthTask {
			isAllAccounted = false
			return
		}
	}
	return isAllAccounted
}

func (r *DefaultMonthTaskRunner) listAllMainSummary(kt *kit.Kit, billYear, billMonth int) (
	[]*dsbillapi.BillSummaryMain, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", r.rootAccountID),
		tools.RuleEqual("vendor", r.vendor),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	result, err := r.client.DataService().Global.Bill.ListBillSummaryMain(
		kt, &dsbillapi.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Count: true,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("list main account bill summary of %+v failed, err %s", r, err.Error())
	}
	if result.Count == 0 {
		return nil, fmt.Errorf("empty count in result %+v", result)
	}
	logs.Infof("found %d main account summary for %+v, rid: %s", result.Count, r, kt.Rid)
	var mainSummaryList []*dsbillapi.BillSummaryMain
	for offset := uint64(0); offset < result.Count; offset = offset + uint64(core.DefaultMaxPageLimit) {
		result, err = r.client.DataService().Global.Bill.ListBillSummaryMain(
			kt, &dsbillapi.BillSummaryMainListReq{
				Filter: tools.ExpressionAnd(expressions...),
				Page: &core.BasePage{
					Start: 0,
					Limit: core.DefaultMaxPageLimit,
				},
			})
		if err != nil {
			return nil, fmt.Errorf("list main account bill summary of %+v failed, err %s", r, err.Error())
		}
		mainSummaryList = append(mainSummaryList, result.Details...)
	}
	return mainSummaryList, nil
}

func (r *DefaultMonthTaskRunner) ensureMonthTaskPullStep(kt *kit.Kit, task *billcore.MonthTask) error {

	if len(task.PullFlowID) == 0 {
		result, err := r.createMonthTaskFlow(kt, task, enumor.MonthTaskStepPull)
		if err != nil {
			logs.Errorf("fail to create month task pull flow of %s, err: %v, rid: %s",
				task.String(), err, kt.Rid)
			return err
		}
		logs.Infof("successfully create month task flow of %s, flow id: %s, rid: %s",
			task.String(), result.ID, kt.Rid)
		return nil
	}
	flow, err := r.client.TaskServer().GetFlow(kt, task.PullFlowID)
	if err != nil {
		logs.Errorf("get flow by id %s failed, err: %v, rid: %s", task.PullFlowID, err, kt.Rid)
		return err
	}
	// 如果任务失败，则重新创建
	if flow.State == enumor.FlowFailed {
		result, err := r.createMonthTaskFlow(kt, task, enumor.MonthTaskStepPull)
		if err != nil {
			logs.Errorf("fail to recreate month task pull flow of %s, err: %v, rid: %s",
				task.String(), err, kt.Rid)
			return err
		}
		logs.Infof("successfully recreate month task pull flow of %s, flow id: %s, rid: %s",
			task.String(), result.ID, kt.Rid)
		return nil
	}
	// 其它情况等待flow中更新task状态
	return nil
}

func (r *DefaultMonthTaskRunner) ensureMonthTaskSplitStep(kt *kit.Kit, task *billcore.MonthTask) error {
	if len(task.SplitFlowID) == 0 {
		result, err := r.createMonthTaskFlow(kt, task, enumor.MonthTaskStepSplit)
		if err != nil {
			logs.Errorf("failed to create month task split flow: %s, err: %v, rid: %s", task.String(), err, kt.Rid)
			return err
		}
		logs.Infof("successfully create month task split flow: %s, flow id: %s, rid: %s",
			task.String(), result.ID, kt.Rid)
		return nil
	}
	flow, err := r.client.TaskServer().GetFlow(kt, task.SplitFlowID)
	if err != nil {
		logs.Errorf("get flow by id %s failed, err:%v, rid: %s", task.SplitFlowID, err, kt.Rid)
		return err
	}
	// 如果任务失败，则重新创建
	if flow.State == enumor.FlowFailed {
		result, err := r.createMonthTaskFlow(kt, task, enumor.MonthTaskStepSplit)
		if err != nil {
			logs.Errorf("failed to recreate month task split flow: %s, err: %v,, rid: %s", task.String(), err, kt.Rid)
			return err
		}
		logs.Infof("successfully recreate month task spli flow: %s, flow id: %s, rid: %s",
			task.String(), result.ID, kt.Rid)
		return nil
	}
	// 其它情况等待flow中更新task状态
	return nil
}

func (r *DefaultMonthTaskRunner) ensureMonthTaskAccountedStep(kt *kit.Kit, task *billcore.MonthTask) error {

	if len(task.SummaryFlowID) == 0 {
		result, err := r.createMonthTaskFlow(kt, task, enumor.MonthTaskStepSummary)
		if err != nil {
			logs.Errorf("failed to create month task summary flow: %s, err: %v, rid: %s", task.String(), err, kt.Rid)
			return err
		}
		logs.Infof("successfully create month task summary flow: %s, flow id: %s, rid: %s", task.String(), result.ID,
			kt.Rid)
		return nil
	}
	flow, err := r.client.TaskServer().GetFlow(kt, task.SummaryFlowID)
	if err != nil {
		logs.Errorf("get flow by id %s failed, err %s, rid: %s", task.SummaryFlowID, err, kt.Rid)
		return err
	}
	// 如果任务失败，则重新创建
	if flow.State == enumor.FlowFailed {
		result, err := r.createMonthTaskFlow(kt, task, enumor.MonthTaskStepSummary)
		if err != nil {
			logs.Errorf("failed to recreate month task summary flow: %s, err: %v, rid: %s", task.String(), err, kt.Rid)
			return err
		}
		logs.Infof("successfully recreate month task summary flow: %s, flow id: %s, rid: %s", task.String(), result.ID,
			kt.Rid)
		return nil
	}
	// 其它情况等待flow中更新task状态
	return nil
}

func (r *DefaultMonthTaskRunner) createMonthTaskFlow(kt *kit.Kit, task *billcore.MonthTask,
	step enumor.MonthTaskStep) (*core.CreateResult, error) {

	memo := fmt.Sprintf("%s:%s %s(%s) %d-%02dv%d", task.Type, step, r.rootAccountCloudID, r.rootAccountID,
		task.BillYear, task.BillMonth, task.VersionID)

	flowReq := &taskserver.AddCustomFlowReq{
		Name: enumor.FlowBillMonthTask,
		Memo: memo,
		Tasks: []taskserver.CustomFlowTask{
			monthtask.BuildMonthTask(task.Type, step, r.rootAccountID, r.vendor, task.BillYear, task.BillMonth, r.ext),
		},
	}
	result, err := r.client.TaskServer().CreateCustomFlow(kt, flowReq)
	if err != nil {
		logs.Errorf("failed to create month task %s, err: %v, rid: %s", task.String(), err, kt.Rid)
		return nil, err
	}

	updateReq := &dsbillapi.BillMonthTaskUpdateReq{ID: task.ID}
	switch step {
	case enumor.MonthTaskStepPull:
		updateReq.PullFlowID = result.ID
	case enumor.MonthTaskStepSplit:
		updateReq.SplitFlowID = result.ID
	case enumor.MonthTaskStepSummary:
		updateReq.SummaryFlowID = result.ID
	default:
		return nil, fmt.Errorf("unsupported month task step %s", step)
	}
	if err := r.client.DataService().Global.Bill.UpdateBillMonthTask(kt, updateReq); err != nil {
		logs.Errorf("failed to update month pull task %s flow id %s, err: %v, rid: %s",
			task.String(), result.ID, err, kt.Rid)
		return nil, err
	}
	return result, nil
}

func (r *DefaultMonthTaskRunner) createMonthPullTaskStub(kt *kit.Kit, rootSummary *billcore.SummaryRoot,
	monthTaskType enumor.MonthTaskType) error {

	createReq := &dsbillapi.BillMonthTaskCreateReq{
		RootAccountID:      r.rootAccountID,
		RootAccountCloudID: r.rootAccountCloudID,
		Vendor:             r.vendor,
		Type:               monthTaskType,
		BillYear:           rootSummary.BillYear,
		BillMonth:          rootSummary.BillMonth,
		VersionID:          rootSummary.CurrentVersion,
		State:              enumor.RootAccountMonthBillTaskStatePulling,
	}
	taskResult, err := r.client.DataService().Global.Bill.CreateBillMonthTask(kt, createReq)
	if err != nil {
		logs.Infof("create month pull task failed, createReq:%+v, err: %v, rid: %s", createReq, err, kt.Rid)
		return err
	}
	logs.Infof("create [%s] %s(%s) month task %s stub success, period: %d-%02d taskID: %s, rid: %s",
		r.vendor, r.rootAccountCloudID, r.rootAccountCloudID, monthTaskType, rootSummary.BillYear,
		rootSummary.BillMonth, taskResult.ID, kt.Rid)
	return nil
}

func (r *DefaultMonthTaskRunner) listMonthPullTaskStub(kt *kit.Kit, billYear, billMonth int,
	types []enumor.MonthTaskType) ([]*billcore.MonthTask, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", r.rootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
		tools.RuleIn("type", types),
	}
	req := &dsbillapi.BillMonthTaskListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: core.DefaultMaxPageLimit,
			Sort:  "type",
			Order: core.Ascending,
		},
	}
	result, err := r.client.DataService().Global.Bill.ListBillMonthTask(kt, req)
	if err != nil {
		logs.Errorf("list month task stub failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("list month task stub failed, err: %v", err)
	}

	return result.Details, nil
}

func (r *DefaultMonthTaskRunner) deleteMonthPullTaskStub(kt *kit.Kit, billYear, billMonth int,
	curType enumor.MonthTaskType) error {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", r.rootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
		tools.RuleEqual("type", curType),
	}
	req := &dataservice.BatchDeleteReq{
		Filter: tools.ExpressionAnd(expressions...),
	}
	return r.client.DataService().Global.Bill.DeleteBillMonthTask(kt, req)
}
