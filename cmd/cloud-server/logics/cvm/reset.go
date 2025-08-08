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

// Package cvm ...
package cvm

import (
	"fmt"

	actionflow "hcm/cmd/task-server/logics/flow"
	cscvm "hcm/pkg/api/cloud-server/cvm"
	"hcm/pkg/api/data-service/task"
	protocvm "hcm/pkg/api/hc-service/cvm"
	ts "hcm/pkg/api/task-server"
	tscvm "hcm/pkg/api/task-server/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/counter"
	"hcm/pkg/tools/slice"
)

// TaskManageBaseReq task manage base req
type TaskManageBaseReq struct {
	Vendors       []enumor.Vendor
	AccountIDs    []string
	BkBizID       int64
	Source        enumor.TaskManagementSource
	Resource      enumor.TaskManagementResource
	TaskOperation enumor.TaskOperation
	TaskType      enumor.TaskType
	Details       []*CvmResetTaskDetailReq
	taskDetails   []*BatchCvmResetTaskDetail
	UniqueID      string
}

// CvmResetTaskDetailReq cvm reset task detail req.
type CvmResetTaskDetailReq struct {
	cscvm.CvmBatchOperateHostInfo `json:",inline"`
	ImageNameOld                  string `json:"image_name_old" validate:"required"`
	CloudImageID                  string `json:"cloud_image_id" validate:"required"`
	ImageName                     string `json:"image_name" validate:"required"`
	Pwd                           string `json:"pwd" validate:"required"`
}

// BatchCvmResetTaskDetail 用于记录 detail - 异步任务flow&task - 任务管理 之间的关系
type BatchCvmResetTaskDetail struct {
	taskDetailID string
	flowID       string
	actionID     string
	*CvmResetTaskDetailReq
}

// CvmResetSystem cvm reset system
func (c *cvm) CvmResetSystem(kt *kit.Kit, params *TaskManageBaseReq) (string, error) {
	// 创建异步管理任务、任务详情列表
	taskID, err := c.buildTaskManagementAndDetails(kt, params)
	if err != nil {
		logs.Errorf("create task management and details failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	// 创建Flow
	flowIDs, err := c.buildFlows(kt, params)
	if err != nil {
		logs.Errorf("build cvm reset async flows failed, err: %v, source: %s, rid: %s",
			err, params.Source, kt.Rid)
		return "", err
	}

	// 把Flow跟异步管理任务进行关联
	err = c.updateTaskManagementAndDetails(kt, flowIDs, taskID, params)
	if err != nil {
		logs.Errorf("update task management and details failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return "", err
	}

	return taskID, nil
}

func (c *cvm) buildTaskManagementAndDetails(kt *kit.Kit, params *TaskManageBaseReq) (string, error) {
	taskID, err := c.createTaskManagement(kt, params.BkBizID, params.Vendors, params.AccountIDs, params.Source,
		params.TaskOperation, params.Resource)
	if err != nil {
		logs.Errorf("create task management failed, err: %v, params: %+v, rid: %s", err, cvt.PtrToVal(params), kt.Rid)
		return "", err
	}

	err = c.createTaskDetails(kt, taskID, params)
	if err != nil {
		logs.Errorf("create task details failed, err: %v, params: %+v, rid: %s", err, cvt.PtrToVal(params), kt.Rid)
		return "", err
	}

	return taskID, nil
}

// createTaskManagement 创建任务管理记录
func (c *cvm) createTaskManagement(kt *kit.Kit, bkBizID int64, vendors []enumor.Vendor, accountIDs []string,
	source enumor.TaskManagementSource, operation enumor.TaskOperation, resource enumor.TaskManagementResource) (
	string, error) {

	taskManagementCreateReq := &task.CreateManagementReq{
		Items: []task.CreateManagementField{
			{
				BkBizID:    bkBizID,
				Source:     source,
				Vendors:    vendors,
				AccountIDs: accountIDs,
				Resource:   resource,
				State:      enumor.TaskManagementRunning, // 默认:执行中
				Operations: []enumor.TaskOperation{operation},
			},
		},
	}

	result, err := c.client.DataService().Global.TaskManagement.Create(kt, taskManagementCreateReq)
	if err != nil {
		logs.Errorf("create dataservice task management failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	if len(result.IDs) == 0 {
		return "", fmt.Errorf("create task management get new task ids failed")
	}

	return result.IDs[0], nil
}

// createTaskDetails 创建任务详情列表
func (c *cvm) createTaskDetails(kt *kit.Kit, taskID string, param *TaskManageBaseReq) error {
	taskDetailsCreateReq := &task.CreateDetailReq{}
	for _, detail := range param.Details {
		detailParams := &CvmResetTaskDetailReq{
			CvmBatchOperateHostInfo: detail.CvmBatchOperateHostInfo,
			ImageNameOld:            detail.ImageNameOld,
			CloudImageID:            detail.CloudImageID,
			ImageName:               detail.ImageName,
		}
		taskDetailsCreateReq.Items = append(taskDetailsCreateReq.Items, task.CreateDetailField{
			BkBizID:          param.BkBizID,
			TaskManagementID: taskID,
			Operation:        param.TaskOperation,
			State:            enumor.TaskDetailInit,
			Param:            detailParams,
		})
	}

	result, err := c.client.DataService().Global.TaskDetail.Create(kt, taskDetailsCreateReq)
	if err != nil {
		logs.Errorf("create dataservice task detail failed, err: %v, taskID: %s, rid: %s", err, taskID, kt.Rid)
		return err
	}

	if len(result.IDs) != len(param.Details) {
		return fmt.Errorf("create task details failed, expect created[%d] task details, but got [%d]",
			len(param.Details), len(result.IDs))
	}

	for i := range result.IDs {
		taskDetail := &BatchCvmResetTaskDetail{
			taskDetailID:          result.IDs[i],
			CvmResetTaskDetailReq: param.Details[i],
		}
		param.taskDetails = append(param.taskDetails, taskDetail)
	}

	return nil
}

func (c *cvm) buildFlows(kt *kit.Kit, params *TaskManageBaseReq) ([]string, error) {
	// 按负载均衡ID进行分组
	vendorToDetails := make(map[enumor.Vendor][]*BatchCvmResetTaskDetail)
	for _, detail := range params.taskDetails {
		vendorToDetails[detail.Vendor] = append(vendorToDetails[detail.Vendor], detail)
	}

	flowIDs := make([]string, 0, len(vendorToDetails))
	for vendor, details := range vendorToDetails {
		flowID, err := c.buildFlow(kt, vendor, details, params.TaskType, params.UniqueID)
		if err != nil {
			logs.Errorf("build flow for cvm reset failed, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
			detailIDs := make([]string, 0, len(details))
			for _, detail := range details {
				detailIDs = append(detailIDs, detail.taskDetailID)
			}
			err = c.updateTaskDetailsState(kt, enumor.TaskDetailFailed, detailIDs, err.Error())
			if err != nil {
				logs.Errorf("update task details status failed, vendor: %s, err: %v, rid: %s", vendor, err, kt.Rid)
				return nil, err
			}
			continue
		}
		flowIDs = append(flowIDs, flowID)
	}

	if len(flowIDs) == 0 {
		logs.Errorf("build cvm reset flow failed, no cvm need reset, vendorToDetails: %+v, params: %+v, rid: %s",
			vendorToDetails, cvt.PtrToVal(params), kt.Rid)
		return nil, fmt.Errorf("build cvm reset failed, no cvm need to be reset")
	}
	return flowIDs, nil
}

func (c *cvm) buildFlow(kt *kit.Kit, vendor enumor.Vendor, details []*BatchCvmResetTaskDetail,
	taskType enumor.TaskType, uniqueID string) (string, error) {

	// 预检测
	lockRel, err := checkResFlowRel(kt, c.client.DataService(), uniqueID, enumor.CvmCloudResType)
	if err != nil {
		logs.Errorf("check resource flow relation failed, err: %v, uniqueID: %s, lockRel: %+v, rid: %s",
			err, uniqueID, cvt.PtrToVal(lockRel), kt.Rid)
		return "", err
	}

	flowTasks, err := c.buildFlowTask(vendor, details)
	if err != nil {
		logs.Errorf("build flow task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	flowID, err := c.createFlowTask(kt, flowTasks, taskType, enumor.FlowResetCvm, uniqueID)
	if err != nil {
		logs.Errorf("create flow task failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	err = lockResFlowStatus(kt, c.client.DataService(), c.client.TaskServer(), uniqueID,
		enumor.CvmCloudResType, flowID, taskType)
	if err != nil {
		logs.Errorf("lock resource flow status failed, err: %v, uniqueID: %s, rid: %s", err, uniqueID, kt.Rid)
		return "", err
	}

	for _, detail := range details {
		detail.flowID = flowID
	}

	return flowID, nil
}

func (c *cvm) buildFlowTask(vendor enumor.Vendor, details []*BatchCvmResetTaskDetail) ([]ts.CustomFlowTask, error) {
	switch vendor {
	case enumor.TCloud:
		return c.buildTCloudFlowTask(vendor, details)
	default:
		return nil, fmt.Errorf("build flow task failed, vendor: %s not supported", vendor)
	}
}

func (c *cvm) buildTCloudFlowTask(vendor enumor.Vendor, details []*BatchCvmResetTaskDetail) (
	[]ts.CustomFlowTask, error) {

	result := make([]ts.CustomFlowTask, 0)
	actionIDGenerator := counter.NewNumberCounterWithPrev(1, 10)
	for _, taskDetails := range slice.Split(details, 1) {
		cur, _ := actionIDGenerator()

		cvmResetList := make([]*protocvm.TCloudBatchResetCvmReq, 0)
		managementDetailIDs := make([]string, 0, len(taskDetails))
		for _, detail := range taskDetails {
			managementDetailIDs = append(managementDetailIDs, detail.taskDetailID)
			cvmResetList = append(cvmResetList, &protocvm.TCloudBatchResetCvmReq{
				Vendor:    detail.Vendor,
				AccountID: detail.AccountID,
				Region:    detail.Region,
				CloudIDs:  []string{detail.CloudID},
				ImageID:   detail.CloudImageID,
				ImageName: detail.ImageName,
				Password:  detail.Pwd,
				IPs:       detail.PrivateIPv4Addresses,
			})
		}

		tmpTask := ts.CustomFlowTask{
			ActionID:   action.ActIDType(cur),
			ActionName: enumor.ActionResetCvm,
			Params: &tscvm.BatchTaskCvmResetOption{
				Vendor:              vendor,
				ManagementDetailIDs: managementDetailIDs,
				CvmResetList:        cvmResetList,
			},
		}
		result = append(result, tmpTask)

		for _, detail := range taskDetails {
			detail.actionID = cur
		}
	}

	return result, nil
}

func (c *cvm) createFlowTask(kt *kit.Kit, flowTasks []ts.CustomFlowTask, taskType enumor.TaskType,
	flowName enumor.FlowName, uniqueID string) (string, error) {

	addReq := &ts.AddCustomFlowReq{
		Name: flowName,
		ShareData: tableasync.NewShareData(map[string]string{
			"unique_id": uniqueID,
		}),
		Tasks:       flowTasks,
		IsInitState: true,
	}
	result, err := c.client.TaskServer().CreateCustomFlow(kt, addReq)
	if err != nil {
		logs.Errorf("call taskserver to batch delete listener custom flow failed, err: %v, flowTasks: %+v, "+
			"rid: %s", err, flowTasks, kt.Rid)
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
				ResID:    uniqueID,
				ResType:  enumor.CvmCloudResType,
				TaskType: taskType,
			},
		}},
	}
	_, err = c.client.TaskServer().CreateTemplateFlow(kt, flowWatchReq)
	if err != nil {
		logs.Errorf("call taskserver to create res flow status watch task failed, err: %v, flowID: %s, rid: %s",
			err, flowID, kt.Rid)
		return "", err
	}

	return flowID, nil
}

func (c *cvm) updateTaskManagementAndDetails(kt *kit.Kit, flowIDs []string, taskID string,
	params *TaskManageBaseReq) error {

	if err := c.updateTaskManagement(kt, taskID, flowIDs); err != nil {
		logs.Errorf("update task management failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	if err := c.updateTaskDetails(kt, params); err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func (c *cvm) updateTaskManagement(kt *kit.Kit, taskID string, flowIDs []string) error {
	updateItem := task.UpdateTaskManagementField{
		ID:      taskID,
		FlowIDs: flowIDs,
	}
	updateReq := &task.UpdateManagementReq{
		Items: []task.UpdateTaskManagementField{updateItem},
	}
	err := c.client.DataService().Global.TaskManagement.Update(kt, updateReq)
	if err != nil {
		logs.Errorf("update task management failed, err: %v, taskID: %s, flowIDs: %v, rid: %s",
			err, taskID, flowIDs, kt.Rid)
		return err
	}

	return nil
}

func (c *cvm) updateTaskDetails(kt *kit.Kit, params *TaskManageBaseReq) error {
	updateItems := make([]task.UpdateTaskDetailField, 0, len(params.taskDetails))
	for _, detail := range params.taskDetails {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:            detail.taskDetailID,
			FlowID:        detail.flowID,
			TaskActionIDs: []string{detail.actionID},
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := c.client.DataService().Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

func (c *cvm) updateTaskDetailsState(kt *kit.Kit, state enumor.TaskDetailState,
	taskDetailIDs []string, reason string) error {

	updateItems := make([]task.UpdateTaskDetailField, 0, len(taskDetailIDs))
	for _, id := range taskDetailIDs {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:     id,
			State:  state,
			Reason: reason,
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := c.client.DataService().Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task details state failed, err: %v, state: %s, updateDetailsReq: %+v, rid: %s",
			err, state, updateDetailsReq, kt.Rid)
		return err
	}
	return nil
}
