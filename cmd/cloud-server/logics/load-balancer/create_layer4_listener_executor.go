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
	"strings"

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
	"hcm/pkg/tools/classifier"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

var _ ImportExecutor = (*CreateLayer4ListenerExecutor)(nil)

func newCreateLayer4ListenerExecutor(cli *dataservice.Client, taskCli *taskserver.Client,
	vendor enumor.Vendor, bkBizID int64, accountID string, regionIDs []string) *CreateLayer4ListenerExecutor {

	return &CreateLayer4ListenerExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// CreateLayer4ListenerExecutor implements the excel import executor for creating layer-4 listeners.
// It handles the process of validating input, building and executing tasks to create listeners.
type CreateLayer4ListenerExecutor struct {
	*basePreviewExecutor

	taskCli     *taskserver.Client
	details     []*CreateLayer4ListenerDetail
	taskDetails []*createLayer4ListenerTaskDetail

	// existingDetails stores details of listeners that already exist, used for creating task management entries.
	existingDetails []*CreateLayer4ListenerDetail
}

// createLayer4ListenerTaskDetail links a CreateLayer4ListenerDetail with its corresponding
// async task flow and action IDs. This helps in tracking the relationship between the input detail,
// the asynchronous task, and the overall task management.
type createLayer4ListenerTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*CreateLayer4ListenerDetail
}

// Execute is the main entry point for the CreateLayer4ListenerExecutor.
// It orchestrates the entire process of creating layer-4 listeners based on the provided raw details.
// The process includes: unmarshalling data, validation, filtering, building task management entries,
// creating asynchronous task flows, and updating task management details.
func (c *CreateLayer4ListenerExecutor) Execute(kt *kit.Kit, source enumor.TaskManagementSource,
	rawDetails json.RawMessage) (taskID string, err error) {

	err = c.unmarshalData(rawDetails)
	if err != nil {
		return "", err
	}

	err = c.validate(kt)
	if err != nil {
		return "", err
	}
	c.filter()

	taskID, err = c.buildTaskManagementAndDetails(kt, source)
	if err != nil {
		return "", err
	}
	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		return "", err
	}
	err = c.updateTaskManagementAndDetails(kt, flowIDs, taskID)
	if err != nil {
		logs.Errorf("update task management and details failed, taskID: %s, flowIDs: %v, err: %v, rid: %s",
			taskID, flowIDs, err, kt.Rid)
		return "", err
	}
	return taskID, nil
}

// unmarshalData parses the raw JSON input into a slice of CreateLayer4ListenerDetail structs.
func (c *CreateLayer4ListenerExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := json.Unmarshal(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

// validate checks the input details for correctness and executability.
// It uses CreateLayer4ListenerPreviewExecutor for the actual validation logic.
// If any detail is found to be not executable, an error is returned.
func (c *CreateLayer4ListenerExecutor) validate(kt *kit.Kit) error {
	executor := &CreateLayer4ListenerPreviewExecutor{
		basePreviewExecutor: c.basePreviewExecutor,
		details:             c.details,
	}
	err := executor.validate(kt)
	if err != nil {
		return err
	}

	for _, detail := range c.details {
		if detail.Status == NotExecutable {
			return fmt.Errorf("create layer4 listener failed, record is not executable: %+v", detail)
		}
	}

	return nil
}

// filter removes non-executable details from the list and collects existing details.
// Only details with status Executable are kept for processing.
// Details with status Existing are moved to the existingDetails slice.
func (c *CreateLayer4ListenerExecutor) filter() {
	c.details = slice.Filter[*CreateLayer4ListenerDetail](c.details, func(detail *CreateLayer4ListenerDetail) bool {
		switch detail.Status {
		case Executable:
			return true
		case Existing:
			// 已存在的也创建对应的detail数据
			c.existingDetails = append(c.existingDetails, detail)
		}
		return false
	})
}

// buildFlows creates asynchronous task flows for creating listeners.
// It groups listener creation details by their cloud load balancer ID and builds a separate flow for each.
// If building a flow for a specific CLB fails, the corresponding task details are marked as failed.
func (c *CreateLayer4ListenerExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// Group listener creation details by their cloud load balancer ID.
	clbToDetails := make(map[string][]*createLayer4ListenerTaskDetail)
	for _, detail := range c.taskDetails {
		clbToDetails[detail.CloudClbID] = append(clbToDetails[detail.CloudClbID], detail)
	}
	lbMap, err := getLoadBalancersMapByCloudID(kt, c.dataServiceCli, c.vendor, c.accountID, c.bkBizID,
		converter.MapKeyToSlice(clbToDetails))
	if err != nil {
		return nil, err
	}

	flowIDs := make([]string, 0, len(clbToDetails))
	for clbCloudID, details := range clbToDetails {
		lb := lbMap[clbCloudID]
		flowID, err := c.buildFlow(kt, lb.ID, lb.CloudID, lb.Region, details)
		if err != nil {
			logs.Errorf("build flow for clb(%s) failed, err: %v, rid: %s", clbCloudID, err, kt.Rid)
			ids := make([]string, 0, len(details))
			for _, detail := range details {
				ids = append(ids, detail.taskDetailID)
			}
			err := updateTaskDetailState(kt, c.dataServiceCli, enumor.TaskDetailFailed, ids, err.Error())
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

// buildFlow constructs a single asynchronous task flow for a specific load balancer.
// It involves building the individual tasks within the flow, creating the flow itself,
// and locking the load balancer resource to prevent concurrent modifications.
func (c *CreateLayer4ListenerExecutor) buildFlow(kt *kit.Kit, lbID, lbCloudID, region string,
	details []*createLayer4ListenerTaskDetail) (string, error) {

	flowTasks, err := c.buildFlowTask(lbID, lbCloudID, region, details)
	if err != nil {
		logs.Errorf("build flow task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	_, err = checkResFlowRel(kt, c.dataServiceCli, lbID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, lbID: %s, err: %v, rid: %s", lbID, err, kt.Rid)
		return "", err
	}
	flowID, err := c.createFlowTask(kt, lbID, flowTasks)
	if err != nil {
		return "", err
	}
	err = lockResFlowStatus(kt, c.dataServiceCli, c.taskCli, lbID,
		enumor.LoadBalancerCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, lbID: %s, err: %v, rid: %s", lbID, err, kt.Rid)
		return "", err
	}

	for _, detail := range details {
		detail.flowID = flowID
	}
	return flowID, nil
}

// createFlowTask creates the main custom flow for listener creation and a secondary watch flow.
// The main flow contains the actual listener creation tasks.
// The watch flow monitors the status of the main flow and updates the resource status accordingly.
func (c *CreateLayer4ListenerExecutor) createFlowTask(kt *kit.Kit, lbID string,
	flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowLoadBalancerCreateListener,
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
				TaskType: enumor.CreateListenerTaskType,
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

func (c *CreateLayer4ListenerExecutor) buildFlowTask(lbID, lbCloudID, region string,
	details []*createLayer4ListenerTaskDetail) ([]ts.CustomFlowTask, error) {

	switch c.vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(lbID, lbCloudID, region, details), nil
	default:
		return nil, fmt.Errorf("vendor %s not supported", c.vendor)
	}
}

func (c *CreateLayer4ListenerExecutor) buildTCloudFlowTask(lbID, lbCloudID, region string,
	details []*createLayer4ListenerTaskDetail) []ts.CustomFlowTask {

	actionIDGenerator := counter.NewNumberCounterWithPrev(1, 10)
	result := []ts.CustomFlowTask{buildSyncClbFlowTask(c.vendor, lbCloudID, c.accountID, region, actionIDGenerator)}
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := actionIDGenerator()

		managementDetailIDs := make([]string, 0, len(taskDetails))
		listeners := make([]*hclb.TCloudListenerCreateReq, 0, len(taskDetails))
		for _, detail := range taskDetails {
			// 监听器名称
			listenerName := fmt.Sprintf("%s-%d", detail.Protocol, detail.ListenerPorts[0])
			if len(detail.Name) > 0 {
				listenerName = detail.Name
			}

			req := &hclb.TCloudListenerCreateReq{
				Name:          listenerName,
				BkBizID:       c.bkBizID,
				LbID:          lbID,
				Protocol:      detail.Protocol,
				Port:          int64(detail.ListenerPorts[0]),
				Scheduler:     string(detail.Scheduler),
				SessionExpire: int64(detail.Session),
				HealthCheck:   &corelb.TCloudHealthCheckInfo{},
			}
			if detail.HealthCheck {
				req.HealthCheck.HealthSwitch = converter.ValToPtr(int64(1))
			} else {
				req.HealthCheck.HealthSwitch = converter.ValToPtr(int64(0))
			}

			if len(detail.ListenerPorts) > 1 {
				req.EndPort = converter.ValToPtr(int64(detail.ListenerPorts[1]))
			}
			listeners = append(listeners, req)
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		}

		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudCreateListener,
			Params: &actionlb.BatchTaskTCloudCreateListenerOption{
				Vendor:              c.vendor,
				ManagementDetailIDs: managementDetailIDs,
				Listeners:           listeners,
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
	result = append(result, buildSyncClbFlowTask(c.vendor, lbCloudID, c.accountID, region, actionIDGenerator))
	return result
}

// createTaskDetails creates entries in the task_detail table for each listener to be created.
// These details are linked to the main task management entry (taskID).
func (c *CreateLayer4ListenerExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {

	if len(c.details) == 0 {
		return nil
	}
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskCreateLayer4Listener,
			State:            enumor.TaskDetailInit,
			Param:            detail,
		})
	}

	result, err := c.dataServiceCli.Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		return err
	}
	if len(result.IDs) != len(c.details) {
		return fmt.Errorf("create task details failed, expect created %d task details, but got %d",
			len(c.details), len(result.IDs))
	}

	for i := range result.IDs {
		taskDetail := &createLayer4ListenerTaskDetail{
			taskDetailID:               result.IDs[i],
			CreateLayer4ListenerDetail: c.details[i],
		}
		c.taskDetails = append(c.taskDetails, taskDetail)
	}
	return nil
}

// createExistingTaskDetails creates entries in the task_detail table for listeners that already exist.
// These are marked as successful and linked to the main task management entry.
func (c *CreateLayer4ListenerExecutor) createExistingTaskDetails(kt *kit.Kit, taskID string) error {
	if len(c.existingDetails) == 0 {
		return nil
	}
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.existingDetails {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskCreateLayer4Listener,
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

// buildTaskManagementAndDetails creates the main task management entry and its associated detail entries.
// This includes details for new listeners to be created and for listeners that already exist.
func (c *CreateLayer4ListenerExecutor) buildTaskManagementAndDetails(kt *kit.Kit, source enumor.TaskManagementSource) (
	string, error) {

	taskID, err := createTaskManagement(kt, c.dataServiceCli, c.bkBizID, c.vendor, c.accountID,
		converter.MapKeyToSlice(c.regionIDMap), source, enumor.TaskCreateLayer4Listener)
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

// updateTaskManagementAndDetails updates the main task management entry with the generated flow IDs
// and updates the individual task detail entries with their respective flow and action IDs.
func (c *CreateLayer4ListenerExecutor) updateTaskManagementAndDetails(kt *kit.Kit,
	flowIDs []string, taskID string) error {

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

// updateTaskDetails 更新task_detail的flow_id和task_action_id
func (c *CreateLayer4ListenerExecutor) updateTaskDetails(kt *kit.Kit) error {
	if len(c.taskDetails) == 0 {
		return nil
	}
	// group by flowID and actionID
	classifySlice := classifier.ClassifySlice(c.taskDetails, func(detail *createLayer4ListenerTaskDetail) string {
		return fmt.Sprintf("%s/%s", detail.flowID, detail.actionID)
	})
	for key, details := range classifySlice {
		split := strings.Split(key, "/")
		if len(split) != 2 {
			return fmt.Errorf("invalid key: %s", key)
		}
		flowID, actionID := split[0], split[1]
		for _, batch := range slice.Split(details, constant.BatchOperationMaxLimit) {
			ids := slice.Map(batch, func(detail *createLayer4ListenerTaskDetail) string {
				return detail.taskDetailID
			})
			updateDetailsReq := &task.BatchUpdateTaskDetailReq{
				IDs:           ids,
				FlowID:        flowID,
				TaskActionIDs: []string{actionID},
			}
			err := c.dataServiceCli.Global.TaskDetail.BatchUpdate(kt, updateDetailsReq)
			if err != nil {
				logs.Errorf("update task details failed, err: %v, req: %+v, rid: %s", err, updateDetailsReq, kt.Rid)
				return err
			}
		}
	}

	return nil
}
