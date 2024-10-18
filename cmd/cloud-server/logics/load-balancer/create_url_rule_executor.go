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

package lblogic

import (
	"encoding/json"
	"fmt"

	actionlb "hcm/cmd/task-server/logics/action/load-balancer"
	actionflow "hcm/cmd/task-server/logics/flow"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
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
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

var _ ImportExecutor = (*CreateUrlRuleExecutor)(nil)

func newCreateUrlRuleExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor,
	bkBizID int64, accountID string, regionIDs []string) *CreateUrlRuleExecutor {

	return &CreateUrlRuleExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateUrlRuleExecutor excel导入——创建四层监听器执行器
type CreateUrlRuleExecutor struct {
	*basePreviewExecutor

	taskCli     *taskserver.Client
	details     []*CreateUrlRuleDetail
	taskDetails []*createUrlRuleTaskDetail

	// detail.Status == Existing 的集合, 用于创建一条任务管理详情
	existingDetails []*CreateUrlRuleDetail
}

type createUrlRuleTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*CreateUrlRuleDetail
}

// Execute ...
func (c *CreateUrlRuleExecutor) Execute(kt *kit.Kit, source enumor.TaskManagementSource, rawDetails json.RawMessage) (string, error) {
	err := c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		return "", err
	}
	c.filter()

	taskID, err := c.buildTaskManagementAndDetails(kt, source)
	if err != nil {
		logs.Errorf("build task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		logs.Errorf("build flows failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	err = c.updateTaskManagementAndDetails(kt, flowIDs, taskID)
	if err != nil {
		logs.Errorf("update task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return taskID, nil
}

func (c *CreateUrlRuleExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := json.Unmarshal(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *CreateUrlRuleExecutor) validate(kt *kit.Kit) error {
	executor := &CreateUrlRulePreviewExecutor{
		basePreviewExecutor: c.basePreviewExecutor,
		details:             c.details,
	}
	err := executor.validate(kt)
	if err != nil {
		logs.Errorf("validate data failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	for _, detail := range c.details {
		if detail.Status == NotExecutable {
			return fmt.Errorf("record(%v) is not executable", detail)
		}
	}

	return nil
}
func (c *CreateUrlRuleExecutor) filter() {
	c.details = slice.Filter[*CreateUrlRuleDetail](c.details, func(detail *CreateUrlRuleDetail) bool {
		switch detail.Status {
		case Executable:
			return true
		case Existing:
			c.existingDetails = append(c.existingDetails, detail)
		}
		return false
	})
}

func (c *CreateUrlRuleExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// group by clb
	clbToDetails := make(map[string][]*createUrlRuleTaskDetail)
	for _, detail := range c.taskDetails {
		clbToDetails[detail.CloudClbID] = append(clbToDetails[detail.CloudClbID], detail)
	}
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli, c.vendor, c.accountID, c.bkBizID,
		converter.MapKeyToSlice(clbToDetails))
	if err != nil {
		logs.Errorf("get load balancers by cloud id failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	flowIDs := make([]string, 0, len(clbToDetails))
	for clbCloudID, details := range clbToDetails {
		lb := lbMap[clbCloudID]
		flowID, err := c.buildFlow(kt, lb, details)
		if err != nil {
			logs.Errorf("build flow for clb(%s) failed, err: %v, rid: %s", clbCloudID, err, kt.Rid)

			ids := make([]string, 0, len(details))
			for _, detail := range details {
				ids = append(ids, detail.taskDetailID)
			}
			err = updateTaskDetailState(kt, c.dataServiceCli, enumor.TaskDetailFailed, ids, err.Error())
			if err != nil {
				logs.Errorf("update task details status failed, err: %v, rid: %s", err, kt.Rid)
				return nil, err
			}
			continue
		}
		flowIDs = append(flowIDs, flowID)
	}

	return flowIDs, nil
}

func (c *CreateUrlRuleExecutor) buildFlow(kt *kit.Kit, lb corelb.BaseLoadBalancer,
	details []*createUrlRuleTaskDetail) (string, error) {

	listenerToDetails, err := c.mapByListener(kt, lb.CloudID, details)
	if err != nil {
		return "", err
	}

	actionIDGenerator := counter.NewNumberCounterWithPrev(1, 10)
	result := []ts.CustomFlowTask{buildSyncClbFlowTask(lb.CloudID, c.accountID, lb.Region, actionIDGenerator)}
	for lblID, taskDetails := range listenerToDetails {
		tasks, err := c.buildFlowTask(lb.ID, lblID, taskDetails, actionIDGenerator)
		if err != nil {
			logs.Errorf("build flow task failed, err: %v, rid: %s", err, kt.Rid)
			return "", err
		}
		result = append(result, tasks...)
	}
	result = append(result, buildSyncClbFlowTask(lb.CloudID, c.accountID, lb.Region, actionIDGenerator))

	_, err = checkResFlowRel(kt, c.dataServiceCli, lb.ID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	flowID, err := c.createFlowTask(kt, lb.ID, result)
	if err != nil {
		logs.Errorf("create flow task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	err = lockResFlowStatus(kt, c.dataServiceCli, c.taskCli, lb.ID,
		enumor.LoadBalancerCloudResType, flowID, enumor.CreateUrlRuleTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	for _, detail := range details {
		detail.flowID = flowID
	}
	return flowID, nil
}

func (c *CreateUrlRuleExecutor) mapByListener(kt *kit.Kit, lbCloudID string, details []*createUrlRuleTaskDetail) (
	map[string][]*createUrlRuleTaskDetail, error) {

	listenerToDetails := make(map[string][]*createUrlRuleTaskDetail)
	for _, detail := range details {
		listener, err := getListener(kt, c.dataServiceCli, c.accountID, lbCloudID, detail.Protocol,
			detail.ListenerPort[0], c.bkBizID, c.vendor)
		if err != nil {
			logs.Errorf("get listener failed, lb(%s), port(%v),err: %v, rid: %s",
				lbCloudID, detail.ListenerPort, err, kt.Rid)
			return nil, err
		}
		if listener == nil {
			return nil, fmt.Errorf("clb(%s) listener(%d) not found", lbCloudID, detail.ListenerPort[0])
		}
		listenerToDetails[listener.ID] = append(listenerToDetails[listener.ID], detail)
	}
	return listenerToDetails, nil
}

func (c *CreateUrlRuleExecutor) createFlowTask(kt *kit.Kit, lbID string,
	flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowLoadBalancerCreateUrlRule,
		ShareData: tableasync.NewShareData(map[string]string{
			"lb_id": lbID,
		}),
		Tasks:       flowTasks,
		IsInitState: true,
	}
	result, err := c.taskCli.CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch add rs custom flow failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flowID := result.ID
	// 从Flow，负责监听主Flow的状态
	flowWatchReq := &ts.AddTemplateFlowReq{
		Name: enumor.FlowLoadBalancerOperateWatch,
		Tasks: []ts.TemplateFlowTask{{
			ActionID: "1",
			Params: &actionflow.LoadBalancerOperateWatchOption{
				FlowID:   flowID,
				ResID:    lbID,
				ResType:  enumor.LoadBalancerCloudResType,
				TaskType: enumor.CreateUrlRuleTaskType,
			},
		}},
	}
	_, err = c.taskCli.CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (c *CreateUrlRuleExecutor) buildFlowTask(lbID, lblID string, details []*createUrlRuleTaskDetail,
	actionIDGenerator func() (cur string, prev string)) ([]ts.CustomFlowTask, error) {

	switch c.vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(lbID, lblID, details, actionIDGenerator)
	default:
		return nil, fmt.Errorf("vendor %s not supported", c.vendor)
	}
}

func (c *CreateUrlRuleExecutor) buildTCloudFlowTask(lbID, lblID string, details []*createUrlRuleTaskDetail,
	actionIDGenerator func() (cur string, prev string)) ([]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := actionIDGenerator()

		managementDetailIDs := make([]string, 0, len(taskDetails))
		rules := make([]hclb.TCloudRuleCreate, 0, len(taskDetails))
		for _, detail := range taskDetails {
			req := hclb.TCloudRuleCreate{
				Url:               detail.UrlPath,
				Domains:           []string{detail.Domain},
				SessionExpireTime: converter.ValToPtr(int64(detail.Session)),
				Scheduler:         converter.ValToPtr(string(detail.Scheduler)),
				DefaultServer:     converter.ValToPtr(detail.DefaultDomain),
				HealthCheck:       &corelb.TCloudHealthCheckInfo{},
			}
			if detail.HealthCheck {
				req.HealthCheck.HealthSwitch = converter.ValToPtr(int64(1))
			}

			rules = append(rules, req)
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		}

		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudCreateL7Rule,
			Params: &actionlb.BatchTaskTCloudCreateL7RuleOption{
				Vendor:                   c.vendor,
				LoadBalancerID:           lbID,
				ListenerID:               lblID,
				ManagementDetailIDs:      managementDetailIDs,
				TCloudRuleBatchCreateReq: &hclb.TCloudRuleBatchCreateReq{Rules: rules},
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

func (c *CreateUrlRuleExecutor) buildTaskManagementAndDetails(kt *kit.Kit, source enumor.TaskManagementSource) (
	string, error) {

	taskID, err := createTaskManagement(kt, c.dataServiceCli, c.bkBizID, c.vendor, c.accountID,
		converter.MapKeyToSlice(c.regionIDMap), source, enumor.TaskCreateLayer7Rule)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	err = c.createTaskDetails(kt, taskID)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	err = c.createExistingTaskDetails(kt, taskID)
	if err != nil {
		logs.Errorf("create existing task details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return taskID, nil
}

func (c *CreateUrlRuleExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {
	if len(c.details) == 0 {
		return nil
	}
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskCreateLayer7Rule,
			Param:            detail,
			State:            enumor.TaskDetailInit,
		})
	}

	result, err := c.dataServiceCli.Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	if len(result.IDs) != len(c.details) {
		return fmt.Errorf("create task details failed, expect created %d task details, but got %d",
			len(c.details), len(result.IDs))
	}

	for i := range result.IDs {
		taskDetail := &createUrlRuleTaskDetail{
			taskDetailID:        result.IDs[i],
			CreateUrlRuleDetail: c.details[i],
		}
		c.taskDetails = append(c.taskDetails, taskDetail)
	}
	return nil
}

func (c *CreateUrlRuleExecutor) updateTaskManagementAndDetails(kt *kit.Kit, flowIDs []string, taskID string) error {
	if err := updateTaskManagement(kt, c.dataServiceCli, taskID, flowIDs); err != nil {
		logs.Errorf("update task management failed, taskID(%s), err: %v, rid: %s", taskID, err, kt.Rid)
		return err
	}
	if err := c.updateTaskDetails(kt); err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (c *CreateUrlRuleExecutor) updateTaskDetails(kt *kit.Kit) error {
	if len(c.taskDetails) == 0 {
		return nil
	}
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

func (c *CreateUrlRuleExecutor) createExistingTaskDetails(kt *kit.Kit, taskID string) error {
	if len(c.existingDetails) == 0 {
		return nil
	}
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.existingDetails {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskCreateLayer7Rule,
			State:            enumor.TaskDetailSuccess,
			Param:            detail,
		})
	}

	_, err := c.dataServiceCli.Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		return err
	}
	return nil
}
