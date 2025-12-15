/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package returner ...
package returner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"hcm/cmd/woa-server/dal/task/dao"
	"hcm/cmd/woa-server/dal/task/table"
	"hcm/cmd/woa-server/logics/task/recycler/event"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
)

func (r *Returner) returnCvm(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) (string, error) {
	instIds := make([]string, 0)
	isResourcePool := false
	for _, host := range hosts {
		if host.InstID == "" {
			logs.Warnf("invalid host %s with empty cvm instance id, rid: %s", host.IP, kt.Rid)
			continue
		}
		// 如果是滚服订单，并且退还方式是“资源池”的话，不做退还到CRP的处理
		if task.RecycleType == table.RecycleTypeRollServer && host.ReturnedWay == enumor.ResourcePoolReturnedWay {
			isResourcePool = true
			logs.Infof("return cvm host is rolling server need skip, subOrderID: %s, task: %+v, host: %+v, rid: %s",
				task.SuborderID, cvt.PtrToVal(task), cvt.PtrToVal(host), kt.Rid)
			continue
		}
		instIds = append(instIds, host.InstID)
	}

	// 如果这批主机Host都需要转移到资源池的话，则不需要调用crp接口
	if len(instIds) == 0 && isResourcePool {
		logs.Infof("recycler:logics:cvm:returnCvm:SKIP, not call cvm api, subOrderId: %s, task: %+v, hosts: %+v, "+
			"rid: %s", task.SuborderID, cvt.PtrToVal(task), cvt.PtrToSlice(hosts), kt.Rid)
		return enumor.RollingServerResourcePoolTask, nil
	}

	if len(instIds) == 0 {
		logs.Errorf("failed to create cvm return order, for instance id list is empty, subOrderID: %s, task: %+v, "+
			"rid: %s", task.SuborderID, cvt.PtrToVal(task), kt.Rid)
		return "", fmt.Errorf("failed to create cvm return order, for instance id list is empty")
	}

	req := r.createReturnReq(instIds, task)
	// call cvm return api
	maxRetry := 3
	var err error = nil
	resp := new(cvmapi.OrderCreateResp)
	for try := 0; try < maxRetry; try++ {
		resp, err = r.cvm.CreateCvmReturnOrder(nil, nil, req)
		if err != nil {
			logs.Errorf("recycler:logics:cvm:returnCvm:failed, failed to create cvm return order, subOrderID: %s, "+
				"err: %v, rid: %s", task.SuborderID, err, kt.Rid)
			// retry after 30 seconds
			time.Sleep(30 * time.Second)
			continue
		}

		if resp.Error.Code != 0 {
			logs.Errorf("recycler:logics:cvm:returnCvm:failed, failed to create cvm return order, subOrderID: %s, "+
				"code: %d, msg: %s, crpTraceID: %s, rid: %s", task.SuborderID, resp.Error.Code, resp.Error.Message,
				resp.TraceId, kt.Rid)
			// retry after 30 seconds
			time.Sleep(30 * time.Second)
			continue
		}

		break
	}

	if err != nil {
		logs.Errorf("recycler:logics:cvm:returnCvm:failed, failed to create cvm return order, subOrderId: %s, "+
			"err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		return "", err
	}

	respStr := ""
	if b, err := json.Marshal(resp); err == nil {
		respStr = string(b)
	}

	if resp.Error.Code != 0 {
		return "", fmt.Errorf("cvm return task failed, subOrderId: %s, code: %d, msg: %s, crpTraceID: %s",
			task.SuborderID, resp.Error.Code, resp.Error.Message, resp.TraceId)
	}

	if resp.Result.OrderId == "" {
		return "", fmt.Errorf("cvm return task return empty crp order id, subOrderId: %s", task.SuborderID)
	}
	logs.Infof("recycler:logics:cvm:returnCvm:success, return cvm resp: %s, subOrderId: %s, rid: %s",
		respStr, task.SuborderID, kt.Rid)
	return resp.Result.OrderId, nil
}

func (r *Returner) createReturnReq(instIds []string, task *table.ReturnTask) *cvmapi.ReturnReq {
	req := &cvmapi.ReturnReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmReturnMethod,
		},
		Params: &cvmapi.ReturnParam{
			// wait 7 days by default
			IsReturnNow:  0,
			InstanceList: instIds,
			// destroy data disk by default
			IsWithDataDisks: 1,
			// direct return by default
			ReturnType: 0,
			Reason:     "",
			// default "常规项目"
			ObsProject:      task.RecycleType.ToObsProject(),
			Force:           false,
			AcceptCostShare: true,
			ReturnForecast:  2,
		},
	}
	// TODO 返还预测全部返还至中转产品，业务需要预测的场景使用其他方案解决，不直接返还预测给业务
	//if task.ReturnForecast {
	//	req.Params.ReturnForecast = 1
	//	req.Params.ReturnForecastTime = task.ReturnForecastTime
	//}

	if task.ReturnPlan == table.RetPlanImmediate {
		req.Params.IsReturnNow = 1
	}

	// 记录日志，方便排查问题
	reqJson, err := json.Marshal(req)
	if err != nil {
		logs.Errorf("recycler:logics:cvm:returnCvm:jsonMarshal:failed, err: %+v, req: %+v, task: %+v", err, req, task)
		return nil
	}
	logs.Infof("recycler:logics:cvm:returnCvm:success, recycleType: %s, reqJson: %s, task: %+v, instIDs: %v",
		task.RecycleType, reqJson, cvt.PtrToVal(task), instIds)

	return req
}

// RecoverReturnCvm 恢复未回退完成状态为init的CVM单据
func (r *Returner) RecoverReturnCvm(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {
	// 1. 查询云梯中未完成return的实例
	cvms, err := r.getCvmInfo(kt, hosts)
	if err != nil {
		logs.Errorf("failed to get cvm info, err: %v, subOrderId: %s, rid: %s", task.SuborderID, err, kt.Rid)
		return &event.Event{
			Type:  event.ReturnFailed,
			Error: fmt.Errorf("failed to get cvm info, err: %v, subOrderId: %s", task.SuborderID, err),
		}
	}
	// 销毁成功，查不到实例
	if len(cvms) == 0 {
		return r.updateReturnSuccess(kt, task, hosts)
	}
	// 2. 获得待回退主机中未回退的实例，重新return
	instIds := make([]string, 0)
	for _, cvm := range cvms {
		instIds = append(instIds, cvm.InstanceId)
	}
	req := r.createReturnReq(instIds, task)
	resp := new(cvmapi.OrderCreateResp)

	var cvmNum = len(cvms)
	maxRetry := 3
	for try := 0; try < maxRetry; try++ {
		resp, err = r.cvm.CreateCvmReturnOrder(kt.Ctx, kt.Header(), req)
		if err != nil {
			logs.Errorf("recycler: failed to create cvm return order, subOrderId: %s, err: %v, rid: %s",
				task.SuborderID, err, kt.Rid)
			// retry after 30 seconds
			time.Sleep(30 * time.Second)
			continue
		}

		if resp.Error.Code != 0 {
			logs.Errorf("recycler:logics:cvm:returnCvm:failed, failed to create cvm return order, subOrderId: %s, "+
				"code: %d, msg: %s, rid: %s", task.SuborderID, resp.Error.Code, resp.Error.Message, kt.Rid)
			// retry after 30 seconds
			time.Sleep(30 * time.Second)
			continue
		}
		break
	}

	// 故障前未创建return单，恢复后成功创建return单
	if err == nil && resp.Error.Code == 0 {
		// 成功调用cvm回退接口
		logs.Infof("success to call cvm return api, num: %d, total: %d, subOrderId: %s, result: %v, traceID: %s, "+
			"rid: %s", cvmNum, len(hosts), task.SuborderID, resp.Result, resp.TraceId, kt.Rid)
		return r.updateReturnState(kt, err, resp.Result.OrderId, task, hosts)
	}

	logs.Errorf("return task is running, failed to return cvm, failedNum: %d, total: %d, subOrderId: %s, err: %v, "+
		"resp: %+v, rid: %s", cvmNum, len(hosts), task.SuborderID, err, cvt.PtrToVal(resp), kt.Rid)
	msg := fmt.Sprintf("%d hosts return failed, return is running or no exit cvms, subOrderId: %s", cvmNum,
		task.SuborderID)
	if err := r.UpdateReturnTaskStatus(kt, task.SuborderID, table.ReturnStatusFailed, msg); err != nil {
		logs.Errorf("failed to update return task info, subOrderId: %s, err: %v, rid: %s", task.SuborderID, err,
			kt.Rid)
		return &event.Event{Type: event.ReturnFailed, Error: err}
	}

	if err = r.UpdateOrderInfo(kt.Ctx, task.SuborderID, "AUTO", uint(len(hosts)-cvmNum), uint(cvmNum),
		0, msg); err != nil {
		// ignore update error and continue to deal status
		logs.Errorf("failed to update recycle order info, subOrderId: %s, err: %v, rid: %s", task.SuborderID, err,
			kt.Rid)
	}
	// 未知是否因拒绝失败，若return被拒绝，无法回滚，人工处理
	return &event.Event{
		Type: event.ReturnFailed,
		Error: fmt.Errorf("failed to return cvm, can not rollback transifer host if is rejected, num: %d, total: %d,"+
			" subOrderId: %s, err: %v", cvmNum, len(hosts), task.SuborderID, err),
	}
}

func (r *Returner) updateReturnSuccess(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {

	if err := r.UpdateHostState(table.RecycleStageDone, table.RecycleStatusDone, task.SuborderID); err != nil {
		logs.Errorf("failed to update host state, subOrderId: %s, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		return &event.Event{Type: event.ReturnFailed, Error: err}
	}

	if err := r.UpdateReturnTaskInfo(kt.Ctx, task, table.ReturnStatusSuccess, "success"); err != nil {
		logs.Errorf("failed to update return task info, subOrderId: %s, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		return &event.Event{Type: event.ReturnFailed, Error: err}
	}

	if err := r.UpdateOrderInfo(context.Background(), task.SuborderID, "AUTO", uint(len(hosts)), 0, 0,
		"success"); err != nil {
		logs.Errorf("failed to update recycle order, subOrderId: %s, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		// ignore update error and continue to query
	}
	return &event.Event{Type: event.ReturnSuccess}
}

func (r *Returner) getCvmInfo(kt *kit.Kit, hosts []*table.RecycleHost) ([]*cvmapi.InstanceItem, error) {
	// create job
	instanceIds := make([]string, 0)
	for _, host := range hosts {
		instanceIds = append(instanceIds, host.InstID)
	}

	req := &cvmapi.InstanceQueryReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmInstanceStatusMethod,
		},
		Params: &cvmapi.InstanceQueryParam{
			InstanceId: instanceIds,
		},
	}

	resp, err := r.cvm.QueryCvmInstances(nil, nil, req)
	if err != nil {
		logs.Errorf("failed to query cvm instance, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if resp.Error.Code != 0 {
		logs.Errorf("failed to query cvm instance, code: %d, msg: %s, crpTraceID: %s, rid: %s", resp.Error.Code,
			resp.Error.Message, resp.TraceId, kt.Rid)
		return nil, fmt.Errorf("failed to query cvm instance, code: %d, msg: %s, crpTraceID: %s", resp.Error.Code,
			resp.Error.Message, resp.TraceId)
	}

	if resp.Result == nil {
		logs.Errorf("failed to query cvm instance, for result is nil, rid: %s", kt.Rid)
		return nil, fmt.Errorf("failed to query cvm instance, for result is nil")
	}

	// 记录查询到的CVM信息，方便排查问题
	jsonRespData, err := json.Marshal(resp.Result.Data)
	if err != nil {
		logs.Warnf("query crp cvm instances failed to marshal resp, err: %v, crpTraceID: %s, rid: %s",
			err, resp.TraceId, kt.Rid)
	}
	logs.Infof("query crp cvm instances result, respDataJson: %s, crpTraceID: %s, rid: %s",
		jsonRespData, resp.TraceId, kt.Rid)

	// 只有状态为running的主机实例，才认为是需要回退的实例
	existCvms := make([]*cvmapi.InstanceItem, 0)
	for _, item := range resp.Result.Data {
		if item.InstanceStatus == enumor.CvmInstanceStatusRunning {
			existCvms = append(existCvms, item)
		}
	}
	return existCvms, nil
}

func (r *Returner) queryCvmOrder(kt *kit.Kit, task *table.ReturnTask, hosts []*table.RecycleHost) *event.Event {
	req := &cvmapi.ReturnDetailReq{
		ReqMeta: cvmapi.ReqMeta{
			Id:      cvmapi.CvmId,
			JsonRpc: cvmapi.CvmJsonRpc,
			Method:  cvmapi.CvmReturnDetailMethod,
		},
		Params: &cvmapi.ReturnDetailParam{
			OrderId: task.TaskID,
			Page: &cvmapi.Page{
				Start: 0,
				// max size 500
				Size: 500,
			},
		},
	}
	resp, err := r.cvm.QueryCvmReturnDetail(kt.Ctx, kt.Header(), req)
	if err != nil {
		// keep loop query when error occurs until timeout
		logs.Warnf("failed to query cvm return detail, err: %v, rid: %s", err, kt.Rid)
		return &event.Event{Type: event.ReturnHandling, Error: err}
	}

	respStr := ""
	if b, err := json.Marshal(resp); err == nil {
		respStr = string(b)
	}
	logs.Infof("query cvm return detail, subOrderID: %s, taskID: %s, hostNum: %d, resp: %s, rid: %s", task.SuborderID,
		task.TaskID, len(hosts), respStr, kt.Rid)
	if resp.Error.Code != 0 {
		// keep loop query when error occurs until timeout
		logs.Warnf("failed to query cvm return detail, code: %d, msg: %s, crpTraceID: %s, rid: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId, kt.Rid)
		ev := &event.Event{
			Type: event.ReturnHandling,
			Error: fmt.Errorf("failed to query cvm return detail, code: %d, msg: %s, crpTraceID: %s", resp.Error.Code,
				resp.Error.Message, resp.TraceId),
		}
		return ev
	}

	successCnt, failedCnt, runningCnt, isRejected := r.parseCvmReturnDetail(kt, hosts, resp.Result.Data)
	if runningCnt > 0 {
		_ = r.updateOrderIfChanged(kt, task, successCnt, failedCnt, runningCnt)
		// ignore update error and continue to query
		return &event.Event{Type: event.ReturnHandling, Error: nil}
	}
	if failedCnt > 0 {
		msg := fmt.Sprintf("%d hosts return failed", failedCnt)

		// transfer hosts back to recycle module if return order is rejected
		if isRejected {
			msg = "return order is rejected, hosts are transited back to recycle module"
			r.rollbackTransit(kt, hosts)
		}

		if err = r.UpdateReturnTaskStatus(kt, task.SuborderID, table.ReturnStatusFailed, msg); err != nil {
			logs.Errorf("failed to update return task info, order id: %s, err: %v, rid: %s",
				task.SuborderID, err, kt.Rid)
			return &event.Event{Type: event.ReturnFailed, Error: err}
		}
		err = r.UpdateOrderInfo(kt.Ctx, task.SuborderID, "AUTO", successCnt, failedCnt, runningCnt, msg)
		if err != nil {
			logs.Warnf("failed to update recycle order %s info, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
			// ignore update error and continue to query
		}
		return &event.Event{Type: event.ReturnFailed, Error: nil}
	}
	if err = r.UpdateReturnTaskStatus(kt, task.SuborderID, table.ReturnStatusSuccess, "success"); err != nil {
		logs.Errorf("failed to update return task info, order id: %s, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		return &event.Event{Type: event.ReturnFailed, Error: err}
	}
	if err = r.UpdateOrderInfo(kt.Ctx, task.SuborderID, "AUTO", successCnt, failedCnt, runningCnt,
		"success"); err != nil {
		logs.Warnf("failed to update recycle order %s info, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		// ignore update error and continue to query
	}
	return &event.Event{Type: event.ReturnSuccess}
}

// 仅在数量变化时更新子单信息，以免干扰上层判断状态变化
func (r *Returner) updateOrderIfChanged(kt *kit.Kit, task *table.ReturnTask, successCnt uint, failedCnt uint,
	runningCnt uint) error {

	cond := &mapstr.MapStr{
		"suborder_id": task.SuborderID,
	}
	order, err := dao.Set().RecycleOrder().GetRecycleOrder(kt.Ctx, cond)
	if err != nil {
		logs.Errorf("failed to get recycle order for update, subOrderId: %s, err: %v, rid: %s",
			task.SuborderID, err, kt.Rid)
		return err
	}

	if order.SuccessNum == successCnt && order.FailedNum == failedCnt && order.PendingNum == runningCnt {
		// no change, skip update
		return nil
	}

	err = r.UpdateOrderInfo(kt.Ctx, task.SuborderID, "AUTO", successCnt, failedCnt, runningCnt, "")
	if err != nil {
		logs.Warnf("failed to update recycle order %s info, err: %v, rid: %s", task.SuborderID, err, kt.Rid)
		// ignore update error and continue to query
	}

	return err
}

func (r *Returner) parseCvmReturnDetail(kt *kit.Kit, hosts []*table.RecycleHost, details []*cvmapi.ReturnDetail) (
	uint, uint, uint, bool) {

	mapInst2Detail := make(map[string]*cvmapi.ReturnDetail)
	for _, detail := range details {
		mapInst2Detail[detail.InstanceId] = detail
	}

	runningCnt := uint(0)
	failedCnt := uint(0)
	successCnt := uint(0)
	isRejected := false
	for _, host := range hosts {
		switch host.Status {
		case table.RecycleStatusDone:
			successCnt++
		case table.RecycleStatusReturnFailed:
			failedCnt++
		case table.RecycleStatusReturning:
			detail, ok := mapInst2Detail[host.InstID]
			if !ok {
				runningCnt++
				continue
			}
			// 20: 销毁完成, 127: 审批驳回, 128: 异常终止
			if detail.Status != 20 && detail.Status != 127 && detail.Status != 128 {
				runningCnt++
			}
			if err := r.updateCvmHostInfo(host, detail); err != nil {
				logs.Warnf("failed to update recycle host info, err: %v, rid: %s", err, kt.Rid)
			}

			switch detail.Status {
			case 20:
				successCnt++
			case 128:
				failedCnt++
			case 127:
				isRejected = true
				failedCnt++
			}

		default:
			logs.Warnf("%s query cvm return detail failed, for invalid recycle host status %s, rid: %s",
				host.IP, host.Status, kt.Rid)
			failedCnt++
		}
	}

	return successCnt, failedCnt, runningCnt, isRejected
}

func (r *Returner) updateCvmHostInfo(host *table.RecycleHost, detail *cvmapi.ReturnDetail) error {
	filter := mapstr.MapStr{
		"suborder_id": host.SuborderID,
		"ip":          host.IP,
	}

	now := time.Now()
	update := mapstr.MapStr{
		"return_tag":       detail.Tag,
		"return_cost_rate": detail.Partition,
		"return_plan_msg":  detail.RetPlanMsg,
		"return_time":      detail.FinishTime,
		"update_at":        now,
	}

	if detail.Status == 20 {
		update["stage"] = table.RecycleStageDone
		update["status"] = table.RecycleStatusDone
	} else if detail.Status == 127 || detail.Status == 128 {
		update["status"] = table.RecycleStatusReturnFailed
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, ip: %s, err: %v", host.IP, err)
		return err
	}

	return nil
}

// UpdateHostState 更新host状态及时间
func (r *Returner) UpdateHostState(recycleStage table.RecycleStage, recycleStatus table.RecycleStatus,
	subOrderId string) error {

	filter := mapstr.MapStr{
		"suborder_id": subOrderId,
	}
	update := mapstr.MapStr{
		"stage":     recycleStage,
		"status":    recycleStatus,
		"update_at": time.Now(),
	}

	if err := dao.Set().RecycleHost().UpdateRecycleHost(context.Background(), &filter, &update); err != nil {
		logs.Errorf("failed to update recycle host, suborderId: %s, err: %v", subOrderId, err)
		return err
	}
	return nil
}
