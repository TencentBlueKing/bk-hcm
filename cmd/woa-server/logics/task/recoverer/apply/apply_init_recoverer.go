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

package apply

import (
	"fmt"
	"strconv"
	"sync"

	"hcm/cmd/woa-server/logics/task/scheduler/record"
	recovertask "hcm/cmd/woa-server/types/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/sopsapi"
)

// recoverInitStep 恢复当前正在初始化的订单，initStep为handling或init
func (r *applyRecoverer) recoverInitStep(kt *kit.Kit, order *types.ApplyOrder) error {
	logs.Infof("start recover init step, subOrderId: %s, rid: %s", order.SubOrderId, kt.Rid)
	generateRecords, err := r.schedulerIf.GetGenerateRecords(kt, order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get generate records by subOrderId, err: %v, subOrderId: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		return err
	}

	successCount, failedCount := 0, 0
	for _, generateRecord := range generateRecords {
		if generateRecord.Status != types.GenerateStatusSuccess {
			logs.Warnf("generate record status is not success, can not init, subOrderId: %s, generateId: %d, rid: %s",
				order.SubOrderId, generateRecord.GenerateId, kt.Rid)
			continue
		}
		if err = r.recoverInitOrder(kt, generateRecord, order); err != nil {
			// 忽略某个生产单机器的初始化失败，继续执行其余机器初始化操作
			logs.Errorf("failed to recover apply init order, err: %v, subOrderId: %s, generateId: %d, rid: %s", err,
				order.SubOrderId, generateRecord.GenerateId, kt.Rid)
			failedCount++
			continue
		}
		successCount++
	}
	logs.Infof("success recover init step, subOrderId: %s, successNum: %d, failedNum: %d, rid: %s", order.SubOrderId,
		successCount, failedCount, kt.Rid)
	return nil
}

// recoverInitingRecord 根据initRecord状态恢复订单
func (r *applyRecoverer) recoverInitingRecord(kt *kit.Kit, subOrderId string, ip string) (*types.DeviceInfo, error) {
	bizId, err := r.getHostBizID(kt, ip)
	if err != nil {
		logs.Errorf("failed to get bizId, err: %v, subOrderId: %s, ip: %s, rid: %s", err, subOrderId, ip, kt.Rid)
		return nil, err
	}

	initRecord, err := record.GetInitRecord(kt, subOrderId, ip)
	if err != nil {
		logs.Errorf("failed to get init record, err: %v, subOrderId: %s, rid: %s", err, subOrderId, kt.Rid)
		return nil, err
	}

	device, err := r.getDeviceByIp(kt, subOrderId, ip)
	if err != nil {
		logs.Errorf("failed to get device by ip: %s, err: %v, subOrderId: %s, rid: %s", ip, err, subOrderId, kt.Rid)
		return nil, err
	}

	if initRecord == nil {
		err = r.schedulerIf.ProcessInitStep(device)
		if err != nil {
			logs.Errorf("failed to init device by ip: %s, subOrderId: %s, err: %v, rid: %s", ip, subOrderId, err,
				kt.Rid)
			return nil, err
		}
		return device, nil
	}

	switch initRecord.Status {
	case types.InitStatusInit:
		sopsTasks, err := r.getInitTask(kt, bizId, subOrderId, ip)
		if err != nil {
			logs.Errorf("failed to get sops task list, subOrderId: %s, err: %v, rid: %s", subOrderId, err, kt.Rid)
			return nil, err
		}

		if len(sopsTasks) == 0 {
			err = r.schedulerIf.ProcessInitStep(device)
			if err != nil {
				logs.Errorf("failed to init device, subOrderId: %s, ip: %s, err: %v, rid: %s", subOrderId, ip, err,
					kt.Rid)
				return nil, err
			}
			return device, nil
		}
		device, err := r.dealInitingTask(kt, sopsTasks, bizId, device)
		if err != nil {
			logs.Errorf("failed to recover apply initing task, err: %v, subOrderId: %s, rid: %s", err, subOrderId,
				kt.Rid)
			return nil, err
		}

		return device, nil

	case types.InitStatusHandling:
		if err = r.schedulerIf.CheckSopsUpdate(bizId, device, initRecord.TaskLink, initRecord.TaskId); err != nil {
			logs.Errorf("failed to check sops update, subOrderId: %s, err: %v, rid: %s", subOrderId, err, kt.Rid)
			return nil, err
		}
		return device, nil

	case types.InitStatusSuccess:
		return device, nil
	default:
		logs.Errorf("unkonwn init status, status: %d, subOrderId: %s, rid: %s,", initRecord.Status, subOrderId, kt.Rid)
		return nil, fmt.Errorf("unkonwn init status, status: %d, subOrderId: %s", initRecord.Status, subOrderId)
	}
}

// dealInitingTask 处理状态为initing的sops任务
func (r *applyRecoverer) dealInitingTask(kt *kit.Kit, sopsTasks []*sopsapi.GetTaskListRst, bizId int64,
	device *types.DeviceInfo) (*types.DeviceInfo, error) {

	// 检查多个sops任务是否有成功的
	taskDetail, err := r.getSopsResult(kt, sopsTasks, bizId, device)
	if err != nil {
		logs.Errorf("failed to get sops result, err: %v, subOrderId: %s, rid: %s", err, device.SubOrderId, kt.Rid)
		return nil, fmt.Errorf("failed to get sops result, subOrderId: %s, err: %v", device.SubOrderId, err)
	}

	taskId := strconv.FormatUint(taskDetail.Data.Id, 10)
	// 3. update device status
	device.InitTaskId = taskId
	device.InitTaskLink = taskDetail.Data.TaskUrl
	logs.Infof("sops task success, task id: %s, task url: %s, subOrderId: %s, rid: %s", taskId, taskDetail.Data.TaskUrl,
		device.SubOrderId, kt.Rid)
	return device, nil
}

// getSopsResult initRecord状态为init时，有多个sopsTasks任务时，检查获取成功执行task
func (r *applyRecoverer) getSopsResult(kt *kit.Kit, sopsTasks []*sopsapi.GetTaskListRst, bizId int64,
	device *types.DeviceInfo) (*sopsapi.GetTaskDetailDataResp, error) {

	var err error
	var taskResult *sopsapi.GetTaskDetailDataResp
	for _, sopsTask := range sopsTasks {
		taskDetail, err := r.sopsCli.GetTaskDetail(kt, kt.Header(), int64(sopsTask.ID),
			recovertask.ResourceOperationService)
		if err != nil {
			logs.Errorf("failed to get sops result, err: %v, subOrderId: %s, rid: %s", err, device.SubOrderId, kt.Rid)
			continue
		}
		taskId := strconv.FormatUint(sopsTask.ID, 10)
		taskUrl := taskDetail.Data.TaskUrl
		if err = r.schedulerIf.CheckSopsUpdate(bizId, device, taskUrl, taskId); err != nil {
			logs.Errorf("failed to check sops update, subOrderId: %s, err: %v, rid: %s", device.SubOrderId, err, kt.Rid)
			continue
		}
		taskResult = taskDetail
		break
	}

	if taskResult != nil {
		return taskResult, nil
	}
	logs.Errorf("failed to check sops update, subOrderId: %s, err: %v, rid: %s", device.SubOrderId, err, kt.Rid)
	return nil, err
}

// recoverInitOrder 执行恢复正在初始化的订单
func (r *applyRecoverer) recoverInitOrder(kt *kit.Kit, generateRecord *types.GenerateRecord,
	order *types.ApplyOrder) error {

	initRecords, err := record.GetInitRecords(kt, order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get init records, subOrderId: %s, err: %v, rid: %s", order.SubOrderId, err, kt.Rid)
		return err
	}
	// 更新update_time，触发监听器
	if len(initRecords) == 0 {
		if err = r.updateGenerateRecord(kt, generateRecord.GenerateId, types.GenerateStatusSuccess); err != nil {
			logs.Errorf("failed to update generate record, err: %v, generateId: %d, subOrderId: %s, rid: %s", err,
				generateRecord.GenerateId, generateRecord.SubOrderId, kt.Rid)
			return err
		}
		return nil
	}

	mutex := sync.Mutex{}
	observeDevices := make([]*types.DeviceInfo, 0)
	wg := sync.WaitGroup{}
	for _, ip := range generateRecord.SuccessList {
		wg.Add(1)
		go func(kt *kit.Kit, subOrderId string, ip string) {
			defer wg.Done()
			device, err := r.recoverInitingRecord(kt, subOrderId, ip)
			if err != nil {
				logs.Errorf("failed to recover init order, err: %v, subOrderId: %s, ip: %s, rid: %s", err,
					order.SubOrderId, ip, kt.Rid)
				return
			}
			// 成功初始化机器加入device列表，用于后续转移
			mutex.Lock()
			observeDevices = append(observeDevices, device)
			mutex.Unlock()
		}(kt, order.SubOrderId, ip)
	}
	wg.Wait()

	// update init step
	if err := record.UpdateInitStep(order.SubOrderId, order.TotalNum); err != nil {
		logs.Errorf("failed to update init step, err: %v, subOrderId: %s, rid: %s", err, order.SubOrderId, kt.Rid)
		return err
	}
	if len(observeDevices) != 0 {
		if err = r.deliverDevices(kt, order, observeDevices); err != nil {
			logs.Errorf("failed to update init deliver, err: %v, subOrderId: %s, rid: %s", err, order.SubOrderId,
				kt.Rid)
			return err
		}
	}
	// 执行订单最后的匹配数量，更新订单状态
	return r.schedulerIf.FinalApplyStep(kt, generateRecord, order)
}

// recoverInitingRecord 初始化后更新初始化步骤状态，恢复转移等过程
func (r *applyRecoverer) deliverDevices(kt *kit.Kit, order *types.ApplyOrder,
	observeDevices []*types.DeviceInfo) error {

	var err error
	if order.EnableDiskCheck {
		observeDevices, err = r.schedulerIf.RunDiskCheck(order, observeDevices)
		if err != nil {
			logs.Errorf("failed to run disk check task, err: %v, subOrderId: %s, rid: %s", err, order.SubOrderId,
				kt.Rid)
			return err
		}
	}
	if err = r.schedulerIf.DeliverDevices(kt, order, observeDevices); err != nil {
		logs.Errorf("failed to deliver devices, err: %v, subOrderId: %s, rid: %s", err, order.SubOrderId, kt.Rid)
		return err
	}
	return nil
}
