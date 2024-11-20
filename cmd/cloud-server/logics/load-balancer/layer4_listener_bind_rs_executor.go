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

var _ ImportExecutor = (*Layer4ListenerBindRSExecutor)(nil)

func newLayer4ListenerBindRSExecutor(cli *dataservice.Client, taskCli *taskserver.Client, vendor enumor.Vendor,
	bkBizID int64, accountID string, regionIDs []string) *Layer4ListenerBindRSExecutor {

	return &Layer4ListenerBindRSExecutor{
		taskCli:             taskCli,
		basePreviewExecutor: newBasePreviewExecutor(cli, vendor, bkBizID, accountID, regionIDs),
	}
}

// Layer4ListenerBindRSExecutor excel导入——创建四层监听器执行器
type Layer4ListenerBindRSExecutor struct {
	*basePreviewExecutor

	validator Layer4ListenerBindRSPreviewExecutor

	taskCli     *taskserver.Client
	details     []*Layer4ListenerBindRSDetail
	taskDetails []*layer4ListenerBindRSTaskDetail

	// detail.Status == Existing 的集合, 用于创建一条任务管理详情
	existingDetails []*Layer4ListenerBindRSDetail
}

type layer4ListenerBindRSTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*Layer4ListenerBindRSDetail
}

// Execute ...
func (c *Layer4ListenerBindRSExecutor) Execute(kt *kit.Kit, source enumor.TaskManagementSource,
	rawDetails json.RawMessage) (string, error) {

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
		return "", err
	}
	flowIDs, err := c.buildFlows(kt)
	if err != nil {
		return "", err
	}
	err = c.updateTaskManagementAndDetails(kt, flowIDs, taskID)
	if err != nil {
		logs.Errorf("update task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return taskID, nil
}

func (c *Layer4ListenerBindRSExecutor) unmarshalData(rawDetail json.RawMessage) error {
	err := json.Unmarshal(rawDetail, &c.details)
	if err != nil {
		return err
	}
	return nil
}

func (c *Layer4ListenerBindRSExecutor) validate(kt *kit.Kit) error {
	executor := &Layer4ListenerBindRSPreviewExecutor{
		basePreviewExecutor: c.basePreviewExecutor,
		details:             c.details,
	}
	err := executor.validate(kt)
	if err != nil {
		return err
	}

	for _, detail := range c.details {
		if detail.Status == NotExecutable {
			return fmt.Errorf("record(%v) is not executable", detail)
		}
	}
	return nil
}

func (c *Layer4ListenerBindRSExecutor) filter() {
	c.details = slice.Filter[*Layer4ListenerBindRSDetail](c.details, func(detail *Layer4ListenerBindRSDetail) bool {
		switch detail.Status {
		case Executable:
			return true
		case Existing:
			c.existingDetails = append(c.existingDetails, detail)
		}
		return false
	})
}

func (c *Layer4ListenerBindRSExecutor) buildFlows(kt *kit.Kit) ([]string, error) {
	// group by clb
	clbToDetails := make(map[string][]*layer4ListenerBindRSTaskDetail)
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

func (c *Layer4ListenerBindRSExecutor) buildFlow(kt *kit.Kit, lb corelb.BaseLoadBalancer, details []*layer4ListenerBindRSTaskDetail) (string, error) {

	// 将details根据targetGroupID进行分组，以targetGroupID的纬度创建flowTask
	tgToDetails, tgToListenerCloudIDs, err := c.createTaskDetailsGroupByTargetGroup(kt, lb.CloudID, details)
	if err != nil {
		logs.Errorf("create task details group by target group failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	actionIDGenerator := counter.NewNumberCounterWithPrev(1, 10)
	flowTasks := []ts.CustomFlowTask{buildSyncClbFlowTask(c.vendor, lb.CloudID, c.accountID, lb.Region, actionIDGenerator)}
	for targetGroupID, detailList := range tgToDetails {
		tmpTask, err := c.buildFlowTask(kt, lb, targetGroupID, detailList, actionIDGenerator, tgToListenerCloudIDs)
		if err != nil {
			return "", err
		}
		flowTasks = append(flowTasks, tmpTask...)
	}
	flowTasks = append(flowTasks, buildSyncClbFlowTask(c.vendor, lb.CloudID, c.accountID, lb.Region, actionIDGenerator))

	_, err = checkResFlowRel(kt, c.dataServiceCli, lb.ID, enumor.LoadBalancerCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, lbID: %s, err: %v, rid: %s", lb.ID, err, kt.Rid)
		return "", err
	}
	flowID, err := c.createFlowTask(kt, lb.ID, converter.MapKeyToSlice(tgToDetails), flowTasks)
	if err != nil {
		return "", err
	}
	err = lockResFlowStatus(kt, c.dataServiceCli, c.taskCli, lb.ID,
		enumor.LoadBalancerCloudResType, flowID, enumor.AddRSTaskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, lbID: %s, err: %v, rid: %s", lb.ID, err, kt.Rid)
		return "", err
	}

	for _, taskDetails := range tgToDetails {
		for _, detail := range taskDetails {
			detail.flowID = flowID
		}
	}
	return flowID, nil
}

func (c *Layer4ListenerBindRSExecutor) createTaskDetailsGroupByTargetGroup(kt *kit.Kit, lbCloudID string,
	details []*layer4ListenerBindRSTaskDetail) (map[string][]*layer4ListenerBindRSTaskDetail, map[string]string, error) {

	tgToDetails := make(map[string][]*layer4ListenerBindRSTaskDetail)
	tgToListenerCloudID := make(map[string]string)
	for _, detail := range details {
		listener, err := getListener(kt, c.dataServiceCli, c.accountID, lbCloudID, detail.Protocol,
			detail.ListenerPort[0], c.bkBizID, c.vendor)
		if err != nil {
			return nil, nil, err
		}
		if listener == nil {
			return nil, nil, fmt.Errorf("loadbalancer(%s) listener(%v) not found",
				detail.CloudClbID, detail.ListenerPort)
		}

		targetGroupID, err := getTargetGroupID(kt, c.dataServiceCli, listener.CloudID)
		if err != nil {
			return nil, nil, err
		}
		tgToListenerCloudID[targetGroupID] = listener.CloudID
		tgToDetails[targetGroupID] = append(tgToDetails[targetGroupID], detail)
	}
	return tgToDetails, tgToListenerCloudID, nil
}

func (c *Layer4ListenerBindRSExecutor) createFlowTask(kt *kit.Kit, lbID string, tgIDs []string,
	flowTasks []ts.CustomFlowTask) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: enumor.FlowTargetGroupAddRS,
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
				FlowID:     flowID,
				ResID:      lbID,
				ResType:    enumor.LoadBalancerCloudResType,
				SubResIDs:  tgIDs,
				SubResType: enumor.TargetGroupCloudResType,
				TaskType:   enumor.AddRSTaskType,
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

func (c *Layer4ListenerBindRSExecutor) buildFlowTask(kt *kit.Kit, lb corelb.BaseLoadBalancer,
	targetGroupID string, details []*layer4ListenerBindRSTaskDetail, generator func() (cur string, prev string),
	tgToListenerCloudIDs map[string]string) ([]ts.CustomFlowTask, error) {

	switch c.vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(kt, lb, targetGroupID, details, generator, tgToListenerCloudIDs)
	default:
		return nil, fmt.Errorf("not support vendor: %s", c.vendor)
	}
}

func (c *Layer4ListenerBindRSExecutor) buildTCloudFlowTask(kt *kit.Kit, lb corelb.BaseLoadBalancer,
	targetGroupID string, details []*layer4ListenerBindRSTaskDetail,
	generator func() (cur string, prev string), tgToListenerCloudIDs map[string]string) ([]ts.CustomFlowTask, error) {

	tCloudLB, err := getTCloudLoadBalancer(kt, c.dataServiceCli, lb.ID)
	if err != nil {
		return nil, err
	}
	result := make([]ts.CustomFlowTask, 0)
	for _, taskDetails := range slice.Split(details, constant.BatchTaskMaxLimit) {
		cur, prev := generator()

		targets := make([]*hclb.RegisterTarget, 0, len(taskDetails))
		managementDetailIDs := make([]string, 0, len(taskDetails))
		for _, detail := range taskDetails {
			target := &hclb.RegisterTarget{
				TargetType: detail.InstType,
				Port:       int64(detail.RsPort[0]),
				Weight:     int64(detail.Weight),
			}
			if detail.InstType == enumor.EniInstType {
				target.EniIp = detail.RsIp
			}

			if detail.InstType == enumor.CvmInstType && !converter.PtrToVal(tCloudLB.Extension.SnatPro) {
				// 跨域2.0不进行cvm校验

				cloudVpcIDs := []string{lb.CloudVpcID}
				if converter.PtrToVal(tCloudLB.Extension.Snat) {
					cloudVpcIDs = append(cloudVpcIDs, converter.PtrToVal(tCloudLB.Extension.TargetCloudVpcID))
				}

				cvm, err := getCvm(kt, c.dataServiceCli, detail.RsIp, c.vendor, c.bkBizID, c.accountID, cloudVpcIDs)
				if err != nil {
					logs.Errorf("call data-service to get cvm failed, ip: %s, err: %v, rid: %s", detail.RsIp, err, kt.Rid)
					return nil, err
				}
				if cvm == nil {
					return nil, fmt.Errorf("rs(%s) not found", detail.RsIp)
				}
				target.CloudInstID = cvm.CloudID
				target.InstName = cvm.Name
				target.PrivateIPAddress = cvm.PrivateIPv4Addresses
				target.PublicIPAddress = cvm.PublicIPv4Addresses
				target.Zone = cvm.Zone
			}
			targets = append(targets, target)
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
		}

		req := &hclb.BatchRegisterTCloudTargetReq{
			CloudListenerID: tgToListenerCloudIDs[targetGroupID],
			TargetGroupID:   targetGroupID,
			RuleType:        enumor.Layer4RuleType,
			Targets:         targets,
		}
		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionBatchTaskTCloudBindTarget,
			Params: &actionlb.BatchTaskBindTargetOption{
				Vendor:                       c.vendor,
				LoadBalancerID:               lb.ID,
				ManagementDetailIDs:          managementDetailIDs,
				BatchRegisterTCloudTargetReq: req,
			},
			Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
		}
		if prev != "" {
			tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
		}
		result = append(result, tmpTask)
		// update taskDetail.actionID
		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}

	return result, nil
}

func (c *Layer4ListenerBindRSExecutor) buildTaskManagementAndDetails(kt *kit.Kit, source enumor.TaskManagementSource) (string, error) {
	taskID, err := createTaskManagement(kt, c.dataServiceCli, c.bkBizID, c.vendor, c.accountID,
		converter.MapKeyToSlice(c.regionIDMap), source, enumor.TaskBindingLayer4RS)
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

func (c *Layer4ListenerBindRSExecutor) createTaskDetails(kt *kit.Kit, taskID string) error {
	if len(c.details) == 0 {
		return nil
	}
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.details {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskBindingLayer4RS,
			Param:            detail,
			State:            enumor.TaskDetailInit,
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
		taskDetail := &layer4ListenerBindRSTaskDetail{
			taskDetailID:               result.IDs[i],
			Layer4ListenerBindRSDetail: c.details[i],
		}
		c.taskDetails = append(c.taskDetails, taskDetail)
	}
	return nil
}

func (c *Layer4ListenerBindRSExecutor) updateTaskManagementAndDetails(kt *kit.Kit, flowIDs []string,
	taskID string) error {

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

func (c *Layer4ListenerBindRSExecutor) updateTaskDetails(kt *kit.Kit) error {
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

func (c *Layer4ListenerBindRSExecutor) createExistingTaskDetails(kt *kit.Kit, taskID string) error {
	if len(c.existingDetails) == 0 {
		return nil
	}
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range c.existingDetails {
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          c.bkBizID,
			TaskManagementID: taskID,
			Operation:        enumor.TaskBindingLayer4RS,
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
