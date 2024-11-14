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
	taskCore "hcm/pkg/api/core/task"
	"hcm/pkg/api/data-service/task"
	"hcm/pkg/api/hc-service/sync"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// ImportExecutor 导入执行器
type ImportExecutor interface {
	// Execute 导入执行器的唯一入口, 内部执行流程：unmarshalData -> validate -> filter
	// -> buildTaskManagementAndDetails -> buildFlows -> updateTaskManagementAndDetails
	Execute(*kit.Kit, enumor.TaskManagementSource, json.RawMessage) (taskID string, err error)

	// unmarshalData 传入json数据, 反序列化到预览数据结构中
	unmarshalData(json.RawMessage) error
	// validate 校验数据合法性, 主要调用 ImportPreviewExecutor的validate方法
	validate(*kit.Kit) error
	// filter filter existing record
	filter()
	// buildFlows 构建异步任务
	buildFlows(*kit.Kit) (flowIDs []string, err error)
	// buildTaskManagementAndDetails 创建任务管理&任务详情
	buildTaskManagementAndDetails(kt *kit.Kit, source enumor.TaskManagementSource) (taskManageID string, err error)
	// updateTaskManagementAndDetails 更新任务管理 flowID & 任务详情关于异步任务的详细数据
	updateTaskManagementAndDetails(kt *kit.Kit, flowIDs []string, taskID string) error
}

// NewImportExecutor ...
func NewImportExecutor(operationType OperationType, dataCli *dataservice.Client,
	taskCli *taskserver.Client, vendor enumor.Vendor, bkBizID int64,
	accountID string, regionIDs []string) (ImportExecutor, error) {

	switch operationType {
	case CreateLayer4Listener:
		return newCreateLayer4ListenerExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case CreateLayer7Listener:
		return newCreateLayer7ListenerExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case CreateUrlRule:
		return newCreateUrlRuleExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case Layer4ListenerBindRs:
		return newLayer4ListenerBindRSExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case Layer7ListenerBindRs:
		return newLayer7ListenerBindRSExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case LayerListenerUnbindRs:
		return newBatchListenerUnbindRsExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case LayerListenerRsWeight:
		return newBatchListenerModifyRsWeightExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	case ListenerDelete:
		return newBatchDeleteListenerExecutor(dataCli, taskCli, vendor, bkBizID, accountID, regionIDs), nil
	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operationType)
	}
}

func buildSyncClbFlowTask(vendor enumor.Vendor, lbCloudID, accountID, region string,
	generator func() (cur string, prev string)) ts.CustomFlowTask {

	cur, prev := generator()
	tmpTask := ts.CustomFlowTask{
		ActionID:   action.ActIDType(cur),
		ActionName: enumor.ActionSyncTCloudLoadBalancer,
		Params: &actionlb.SyncTCloudLoadBalancerOption{
			Vendor: vendor,
			TCloudSyncReq: &sync.TCloudSyncReq{
				AccountID: accountID,
				Region:    region,
				CloudIDs:  []string{lbCloudID},
			},
		},
		Retry: tableasync.NewRetryWithPolicy(3, 100, 200),
	}
	if prev != "" {
		tmpTask.DependOn = []action.ActIDType{action.ActIDType(prev)}
	}
	return tmpTask
}

func createTaskManagement(kt *kit.Kit, cli *dataservice.Client, bkBizID int64, vendor enumor.Vendor, accountID string,
	regionIDs []string, source enumor.TaskManagementSource, operation enumor.TaskOperation) (string, error) {

	taskManagementCreateReq := &task.CreateManagementReq{
		Items: []task.CreateManagementField{
			{
				BkBizID:    bkBizID,
				Source:     source,
				Vendors:    []enumor.Vendor{vendor},
				AccountIDs: []string{accountID},
				Resource:   enumor.TaskManagementResClb,
				State:      enumor.TaskManagementRunning,
				Operations: []enumor.TaskOperation{operation},
				Extension:  &taskCore.ManagementExt{RegionIDs: regionIDs},
			},
		},
	}

	result, err := cli.Global.TaskManagement.Create(kt, taskManagementCreateReq)
	if err != nil {
		logs.Errorf("create task management failed, req: %v, err: %v, rid: %s", taskManagementCreateReq, err, kt.Rid)
		return "", err
	}
	if len(result.IDs) == 0 {
		return "", fmt.Errorf("create task management failed")
	}
	return result.IDs[0], nil
}

func updateTaskManagement(kt *kit.Kit, cli *dataservice.Client, taskID string, flowIDs []string) error {

	if len(flowIDs) == 0 {
		return nil
	}
	updateItem := task.UpdateTaskManagementField{
		ID:      taskID,
		FlowIDs: flowIDs,
	}
	updateReq := &task.UpdateManagementReq{
		Items: []task.UpdateTaskManagementField{updateItem},
	}
	err := cli.Global.TaskManagement.Update(kt, updateReq)
	if err != nil {
		logs.Errorf("update task management failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	return nil
}

func updateTaskDetailState(kt *kit.Kit, cli *dataservice.Client, state enumor.TaskDetailState, ids []string,
	reason string) error {

	if len(ids) == 0 {
		return nil
	}
	updateItems := make([]task.UpdateTaskDetailField, 0, len(ids))
	for _, id := range ids {
		updateItems = append(updateItems, task.UpdateTaskDetailField{
			ID:     id,
			State:  state,
			Reason: reason,
		})
	}
	updateDetailsReq := &task.UpdateDetailReq{
		Items: updateItems,
	}
	err := cli.Global.TaskDetail.Update(kt, updateDetailsReq)
	if err != nil {
		logs.Errorf("update task detail state failed, req: %v, err: %v, rid: %s", updateDetailsReq, err, kt.Rid)
		return err
	}
	return nil
}
