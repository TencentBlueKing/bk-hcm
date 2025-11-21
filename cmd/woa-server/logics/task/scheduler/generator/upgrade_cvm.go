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

package generator

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"hcm/cmd/woa-server/logics/task/scheduler/record"
	model "hcm/cmd/woa-server/model/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/thirdparty/cvmapi"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/utils"

	"go.mongodb.org/mongo-driver/mongo"
)

const (
	// crpUpgradeFailedMsg CRP升降配失败的错误描述
	crpUpgradeFailedMsg = "CRP升降配失败"
)

// UpgradeCVMSync 创建升降配CVM单据，同步操作，直接返回CRP单号
func (g *Generator) UpgradeCVMSync(kt *kit.Kit, order *types.ApplyOrder) (orderID string, err error) {
	// start generate step
	if err := record.StartStep(order.SubOrderId, types.StepNameGenerate); err != nil {
		logs.Errorf("failed to start generate step, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
			kt.Rid)
		return "", err
	}

	defer func() {
		if subErr := record.UpdateGenerateStep(order.SubOrderId, order.TotalNum, err); subErr != nil {
			logs.Errorf("failed to generate device, order id: %s, err: %v, rid: %s", order.SubOrderId, subErr,
				kt.Rid)
			return
		}
	}()

	// 1. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
			kt.Rid)
		return "", err
	}

	// cvm upgrade do not require separate campus
	_, orderID, err = g.batchUpgradeCvm(kt, order, order.TotalNum-uint(len(existDevices)))
	if err != nil {
		logs.Errorf("failed to upgrade cvm, suborder id: %s, rid: %s", order.SubOrderId, kt.Rid)
		return "", err
	}
	return orderID, nil
}

// UpgradeCVM upgrade cvm devices
func (g *Generator) UpgradeCVM(kt *kit.Kit, order *types.ApplyOrder) error {
	// 1. get history generated devices
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
			kt.Rid)
		return err
	}

	// check if need generate cvm
	existCount := uint(len(existDevices))
	if existCount >= order.TotalNum {
		logs.Infof("apply order %s has been scheduled %d cvm, rid: %s", order.SubOrderId, existCount, kt.Rid)
		// check if need retry match task
		if err := g.retryMatchDevice(existDevices); err != nil {
			logs.Warnf("failed to retry match device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
				kt.Rid)
		}
		return nil
	}

	logs.Infof("apply order %s existing device number: %d, rid: %s", order.SubOrderId, existCount, kt.Rid)

	// cvm upgrade do not require separate campus
	if err = g.generateCVMConcentrate(kt, order, existDevices, []string{order.Spec.Zone}); err != nil {
		logs.Errorf("failed to upgrade cvm, suborder id: %s, rid: %s", order.SubOrderId, kt.Rid)
		return err
	}
	return nil
}

// batchUpgradeCvm  batch upgrade cvm
func (g *Generator) batchUpgradeCvm(kt *kit.Kit, order *types.ApplyOrder, replicas uint) (uint64, string, error) {
	logs.Infof("start batch upgrade cvm, sub order id: %s, rid: %s", order.SubOrderId, kt.Rid)

	generateID, err := g.initGenerateRecord(kt.Ctx, order.ResourceType, order.SubOrderId, replicas, false)
	if err != nil {
		logs.Errorf("failed to upgrade cvm when init generate record, err: %v, sub order id: %s, rid: %s", err,
			order.SubOrderId, kt.Rid)
		return 0, "", fmt.Errorf("failed to upgrade cvm, sub order id: %s, err: %v", order.SubOrderId, err)
	}

	upgradeCvmReqParam, err := g.buildUpgradeCvmReq(kt, order, replicas)
	if err != nil {
		logs.Errorf("failed to upgrade cvm when build cvm request, err: %v, generateID: %d, sub order id: %s, "+
			"rid: %s", err, generateID, order.SubOrderId, kt.Rid)
		return 0, "", fmt.Errorf("failed to upgrade cvm, sub order id: %s, err: %v", order.SubOrderId, err)
	}

	var orderID string
	if orderID, err = g.upgradeCvmAndWatch(kt, order, upgradeCvmReqParam, generateID); err != nil {
		logs.Errorf("failed to upgrade cvm, err: %v, sub order id: %s, generateID: %d, rid: %s", err,
			order.SubOrderId, generateID, kt.Rid)
		return 0, "", err
	}
	logs.Infof("success to upgrade cvm, sub order id: %s, generate id: %d, crpOrderID: %s, rid: %s",
		order.SubOrderId, generateID, orderID, kt.Rid)

	return generateID, orderID, nil
}

// upgradeCvmAndWatch upgrade cvm
func (g *Generator) upgradeCvmAndWatch(kt *kit.Kit, order *types.ApplyOrder, reqParam *cvmapi.UpgradeParam,
	generateID uint64) (string, error) {

	crpOrderID, err := g.launchUpgradeCVM(kt, reqParam, order)
	if err != nil {
		logs.Errorf("failed to launch upgrade cvm, err: %v, order id: %s, req: %v, rid: %s", err,
			order.SubOrderId, reqParam, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateID,
			types.GenerateStatusFailed, err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to upgrade cvm when update generate record, order id: %s, crp id: %s, err: %v, rid: %s",
				order.SubOrderId, crpOrderID, errRecord, kt.Rid)
			return "", fmt.Errorf("failed to launch upgrade cvm, order id: %s, crp id: %s, err: %v", order.SubOrderId,
				crpOrderID, errRecord)
		}

		return "", fmt.Errorf("failed to launch upgrade cvm, order id: %s, err: %v", order.SubOrderId, err)
	}

	// update generate record status to Query
	if err = g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateID, types.GenerateStatusHandling,
		"handling", crpOrderID, nil); err != nil {
		logs.Errorf("failed to upgrade cvm when update generate record, order id: %s, crp id: %s, err: %v, rid: %s",
			order.SubOrderId, crpOrderID, err, kt.Rid)

		return "", fmt.Errorf("failed to launch upgrade cvm, order id: %s, crp id: %s, err: %v",
			order.SubOrderId, crpOrderID, err)
	}

	// TODO 目前的实现先返回CRP单据给用户，通过异步轮询结果；后续整个提单操作改为后台任务后这里可以改回同步
	backendKt := kt.NewSubKitWithCtx(context.TODO())
	// check cvm task result and update generate record
	go func() {
		defer func() {
			// 临时的异步实现，需在generate完成后更新step和order status
			logs.Infof("update generate step, order id: %s, err: %v, rid: %s", order.SubOrderId, err, backendKt.Rid)
			if subErr := record.UpdateGenerateStep(order.SubOrderId, order.TotalNum, err); subErr != nil {
				logs.Errorf("failed to generate device, order id: %s, err: %v, rid: %s", order.SubOrderId, subErr,
					backendKt.Rid)
				return
			}
			if err != nil {
				// check all generate records and update apply order status
				if subErr := g.UpdateOrderStatus(order.ResourceType, order.SubOrderId); subErr != nil {
					logs.Errorf("failed to update order status, subOrderId: %s, err: %v, rid: %s",
						order.SubOrderId, subErr, kt.Rid)
				}
			}
		}()

		err = g.AddUpgradeCvmDevices(backendKt, crpOrderID, generateID, order)
		if err != nil {
			logs.Errorf("failed to upgrade cvm when add upgrade cvm devices, order id: %s, crp id: %s, err: %v, rid: %s",
				order.SubOrderId, crpOrderID, err, backendKt.Rid)
		}
	}()

	return crpOrderID, nil
}

// buildUpgradeCvmReq construct a cvm upgrade request
func (g *Generator) buildUpgradeCvmReq(kt *kit.Kit, order *types.ApplyOrder, replicas uint) (
	*cvmapi.UpgradeParam, error) {

	// 获取已完成升配的cvm
	existDevices, err := g.getUnreleasedDevice(order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get unreleased device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
			kt.Rid)
		return nil, err
	}
	existsInstanceIDs := make(map[string]interface{})
	for _, device := range existDevices {
		existsInstanceIDs[device.InstanceID] = struct{}{}
	}

	// construct cvm upgrade req
	upgradeCVMReq := &cvmapi.UpgradeParam{
		Reason: order.Remark,
		Data:   make([]cvmapi.UpgradeParamInstance, 0, len(order.UpgradeCVMList)),
	}

	for _, item := range order.UpgradeCVMList {
		if _, ok := existsInstanceIDs[item.InstanceID]; ok {
			continue
		}

		upgradeCVMReq.Data = append(upgradeCVMReq.Data, cvmapi.UpgradeParamInstance{
			InstanceID:         item.InstanceID,
			TargetInstanceType: item.TargetInstanceType,
		})
	}

	// 理论上，之前步骤中计算出的待升配的replicas数量应等于request中的data数量
	if replicas != uint(len(upgradeCVMReq.Data)) {
		logs.Errorf("failed to build upgrade cvm req, request length: %d not eq replicas: %d, rid: %s",
			len(upgradeCVMReq.Data), replicas, kt.Rid)
		return nil, fmt.Errorf("failed to build upgrade cvm req, request length: %d not eq replicas: %d",
			len(upgradeCVMReq.Data), replicas)
	}

	return upgradeCVMReq, nil
}

// launchUpgradeCVM launch cvm upgrade request
func (g *Generator) launchUpgradeCVM(kt *kit.Kit, reqParam *cvmapi.UpgradeParam, order *types.ApplyOrder) (
	string, error) {

	req := cvmapi.NewCvmUpgradeOrderReq(reqParam)

	// 增加日志记录
	jsonReq, err := json.Marshal(req)
	if err != nil {
		logs.Warnf("failed to marshal upgrade cvm req, req: %v, err: %v, rid: %s", req, err, kt.Rid)
	}
	logs.Infof("launch upgrade cvm, subOrderID: %s, req: %s, rid: %s", order.SubOrderId, string(jsonReq), kt.Rid)

	// call cvm api to launchCvm cvm order
	maxRetry := 3
	resp := new(cvmapi.OrderCreateResp)
	for try := 0; try < maxRetry; try++ {
		// need not wait for the first try
		if try != 0 {
			// retry after 5 seconds
			// TODO 改为异步后可以适当延长sleep时间
			time.Sleep(5 * time.Second)
		}

		resp, err = g.cvm.CreateUpgradeOrder(kt, req)
		if err != nil {
			logs.Warnf("retry to create cvm upgrade order, subOrderID: %s, req: %s, err: %v, rid: %s",
				order.SubOrderId, string(jsonReq), err, kt.Rid)
			continue
		}

		if resp == nil {
			logs.Warnf("failed to create cvm upgrade order, subOrderID: %s, resp is nil, rid: %s",
				order.SubOrderId, kt.Rid)
			continue
		}

		if resp.Error.Code != 0 {
			logs.Warnf("failed to create cvm upgrade order, subOrderID: %s, code: %d, msg: %s, crpTraceID: %s, "+
				"rid: %s", order.SubOrderId, resp.Error.Code, resp.Error.Message, resp.TraceId, kt.Rid)
			// CRP有明确状态码的失败暂时不做重试
			break
		}

		break
	}

	if err != nil {
		logs.Errorf("failed to create cvm upgrade order, subOrderID: %s, req: %s, err: %v, rid: %s",
			order.SubOrderId, string(jsonReq), err, kt.Rid)
		return "", err
	}

	respStr := ""
	b, err := json.Marshal(resp)
	if err != nil {
		logs.Warnf("failed to marshal cvm upgrade order create resp, resp: %v, err: %v, rid: %s", resp, err,
			kt.Rid)
	}

	respStr = string(b)
	logs.Infof("create cvm upgrade order, subOrderID: %s, req: %s, resp: %s, rid: %s", order.SubOrderId,
		string(jsonReq), respStr, kt.Rid)

	if resp.Error.Code != 0 {
		return "", fmt.Errorf("failed to create upgrade order, code: %d, msg: %s, crpTraceID: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId)
	}
	if resp.Result.OrderId == "" {
		return "", fmt.Errorf("create upgrade order return empty order id, crpTraceID: %s", resp.TraceId)
	}

	return resp.Result.OrderId, nil
}

// AddUpgradeCvmDevices check generated device, create device infos and update generate record status
func (g *Generator) AddUpgradeCvmDevices(kt *kit.Kit, taskID string, generateID uint64, order *types.ApplyOrder) error {

	// 1. check cvm task result
	if err := g.CheckUpgradeCVM(kt, taskID, order.SubOrderId); err != nil {
		logs.Errorf("scheduler:logics:upgrade:cvm:failed, failed to upgrade cvm when check generate task, "+
			"order id: %s, task id: %s, err: %v, rid: %s", order.SubOrderId, taskID, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateID, types.GenerateStatusFailed,
			err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to upgrade cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskID, errRecord, kt.Rid)
			return fmt.Errorf("failed to upgrade cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
				taskID, errRecord)
		}
		return fmt.Errorf("failed to upgrade cvm, order id: %s, task id: %s, err: %v", order.SubOrderId,
			taskID, err)
	}

	// 2. get generated cvm instances
	hosts, err := g.listUpgradeCVM(kt, taskID)
	if err != nil {
		logs.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v, rid: %s",
			order.SubOrderId, taskID, err, kt.Rid)

		// update generate record status to Done
		if errRecord := g.UpdateGenerateRecord(kt.Ctx, order.ResourceType, generateID, types.GenerateStatusFailed,
			err.Error(), "", nil); errRecord != nil {
			logs.Errorf("failed to upgrade cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskID, errRecord, kt.Rid)
			return fmt.Errorf("failed to list upgraded cvm, order id: %s, task id: %s, err: %v",
				order.SubOrderId, taskID, errRecord)
		}

		return fmt.Errorf("failed to list created cvm, order id: %s, task id: %s, err: %v",
			order.SubOrderId, taskID, err)
	}
	// 3. create device infos
	return g.createUpgradeDeviceInfo(kt, order, generateID, hosts, taskID)
}

// CheckUpgradeCVM checks cvm upgrade task result
func (g *Generator) CheckUpgradeCVM(kt *kit.Kit, orderID, subOrderID string) error {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, fmt.Errorf("failed to query cvm order by id %s, err: %v", orderID, err)
		}

		if obj == nil {
			return false, fmt.Errorf("cvm order %s not found", orderID)
		}

		resp, ok := obj.(*cvmapi.UpgradeDetailResp)
		if !ok {
			return false, fmt.Errorf("object with order id %s is not a cvm upgrade order response: %+v", orderID, obj)
		}

		if resp.Error.Code != 0 {
			return false, fmt.Errorf("query cvm upgrade order failed, code: %d, msg: %s, crpTraceID: %s",
				resp.Error.Code, resp.Error.Message, resp.TraceId)
		}

		if resp.Result == nil {
			return false, fmt.Errorf("query cvm upgrade order failed, for result is null, crpTraceID: %s, resp: %+v",
				resp.TraceId, resp)
		}

		// 检查CRP订单是否超出处理时间并记录日志
		g.checkRecordCrpOrderTimeout(kt, resp.Result.OrderID, resp.Result.CreateTime, resp.TraceId,
			subOrderID, "")

		if resp.Result.Status != enumor.CrpUpgradeOrderFinish &&
			resp.Result.Status != enumor.CrpUpgradeOrderReject &&
			resp.Result.Status != enumor.CrpUpgradeOrderFailed {
			return false, fmt.Errorf("cvm order %s handling", orderID)
		}

		crpURL := cvmapi.CvmUpgradeLinkPrefix + resp.Result.OrderID
		if resp.Result.Status != enumor.CrpUpgradeOrderFinish {
			return true, fmt.Errorf("%s，详情可咨询2000(TEG技术支持)，CRP申请单链接: %s, 状态: %s", crpUpgradeFailedMsg,
				crpURL, resp.Result.Status.StatusName())
		}

		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := cvmapi.NewCvmUpgradeDetailReq(&cvmapi.UpgradeDetailParam{OrderID: orderID})
		resp, err := g.cvm.QueryCvmUpgradeDetail(kt, req)
		if err != nil {
			return nil, err
		}

		// call cvm api to query cvm order status
		return resp, nil
	}

	// TODO: get retry strategy from config
	_, err := utils.Retry(doFunc, checkFunc, uint64(7*types.OneDayDuration.Seconds()), 60)
	return err
}

// listUpgradeCVM lists upgrade cvm by order id
func (g *Generator) listUpgradeCVM(kt *kit.Kit, orderId string) ([]cvmapi.UpgradeDetailInstance, error) {
	checkFunc := func(obj interface{}, err error) (bool, error) {
		if err != nil {
			return false, err
		}
		return true, nil
	}

	doFunc := func() (interface{}, error) {
		// construct order status request
		req := &cvmapi.UpgradeDetailReq{
			ReqMeta: cvmapi.ReqMeta{
				Id:      cvmapi.CvmId,
				JsonRpc: cvmapi.CvmJsonRpc,
				Method:  cvmapi.CvmUpgradeDetailMethod,
			},
			Params: &cvmapi.UpgradeDetailParam{
				OrderID: orderId,
			},
		}
		return g.cvm.QueryCvmUpgradeDetail(kt, req)
	}

	// TODO: get retry strategy from config
	obj, err := utils.Retry(doFunc, checkFunc, 120, 5)

	if err != nil {
		return nil, err
	}
	resp, ok := obj.(*cvmapi.UpgradeDetailResp)
	if !ok {
		return nil, fmt.Errorf("object with order id %s is not a cvm instance response: %+v", orderId, obj)
	}

	logs.Infof("get cvm instance, crpOrderID: %s, resp: %+v， rid: %s", orderId, resp, kt.Rid)

	if resp.Error.Code != 0 {
		return nil, fmt.Errorf("list cvm instance failed, code: %d, msg: %s, crpTraceID: %s",
			resp.Error.Code, resp.Error.Message, resp.TraceId)
	}

	if resp.Result == nil {
		return nil, fmt.Errorf("list cvm upgrade instance failed, for result is null, crpTraceID: %s, resp: %+v",
			resp.TraceId, resp)
	}

	return resp.Result.DetailList, nil
}

func (g *Generator) createUpgradeDeviceInfo(kt *kit.Kit, order *types.ApplyOrder, generateID uint64,
	hosts []cvmapi.UpgradeDetailInstance, taskID string) error {

	deviceList := make([]*types.DeviceInfo, 0)
	successAssetID := make([]string, 0)
	for _, host := range hosts {
		// 失败的机器不记录
		if host.Status != enumor.CrpUpgradeCVMSuccess {
			continue
		}

		deviceList = append(deviceList, &types.DeviceInfo{
			// 升降配无IP信息
			Ip:               "",
			AssetId:          host.InstanceAssetID,
			InstanceID:       host.InstanceID,
			DeviceType:       host.TargetInstanceType,
			GenerateTaskId:   taskID,
			GenerateTaskLink: cvmapi.CvmUpgradeLinkPrefix + taskID,
			Deliverer:        "icr",
			CloudZone:        host.Zone, // 记录当前主机所在可用区
		})
		successAssetID = append(successAssetID, host.InstanceAssetID)
	}

	// NOTE: sleep 15 seconds to wait for CMDB host sync.
	time.Sleep(15 * time.Second)

	txnErr := dal.RunTransaction(kt, func(sc mongo.SessionContext) error {
		// 1. save generated cvm instances info
		sessionKit := kt.NewSubKitWithCtx(sc)
		if err := g.createUpgradeDeviceInfos(sessionKit, order, generateID, deviceList); err != nil {
			logs.Errorf("failed to update generated device, order id: %s, err: %v, rid: %s", order.SubOrderId, err,
				kt.Rid)
			// update generate record status to Done
			// 不参与回滚
			if err := g.UpdateGenerateRecord(context.Background(), order.ResourceType, generateID,
				types.GenerateStatusFailed, err.Error(), "", nil); err != nil {
				logs.Errorf("failed to update generate record, generate id: %d, err: %v, rid: %s", generateID, err,
					kt.Rid)
				return err
			}

			return fmt.Errorf("failed to update generated device, order id: %s, err: %v", order.SubOrderId, err)
		}

		// 2. update generate record status to success
		if err := g.UpdateGenerateRecord(sc, order.ResourceType, generateID, types.GenerateStatusSuccess, "success",
			"", successAssetID); err != nil {
			logs.Errorf("failed to update cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
				order.SubOrderId, taskID, err, kt.Rid)
			return fmt.Errorf("failed to update cvm, order id: %s, task id: %s, err: %v", order.SubOrderId, taskID, err)
		}

		return nil
	})

	if txnErr != nil {
		logs.Errorf("failed to update cvm when update generate record, order id: %s, task id: %s, err: %v, rid: %s",
			order.SubOrderId, taskID, txnErr, kt.Rid)
		return fmt.Errorf("failed to update cvm when update generate record, order id: %s, task id: %s, "+
			"err: %v", order.SubOrderId, taskID, txnErr)
	}
	return nil
}

func (g *Generator) createUpgradeDeviceInfos(kt *kit.Kit, order *types.ApplyOrder, generateID uint64,
	items []*types.DeviceInfo) error {

	assetIDs := make([]string, 0)
	zoneIDs := make([]string, 0)
	for _, item := range items {
		assetIDs = append(assetIDs, item.AssetId)
		zoneIDs = append(zoneIDs, item.CloudZone)
	}

	regionList, err := g.getRegionList(kt, zoneIDs)
	if err != nil {
		logs.Errorf("failed to get region list, order id: %s, generateId: %d, err: %v, rid: %s", order.SubOrderId,
			generateID, err, kt.Rid)
		return err
	}
	zoneRegionMap := make(map[string]string)
	for _, item := range regionList {
		zoneRegionMap[item.Zone] = item.Region
	}

	mapAssetIDToHost, err := g.syncHostToCMDB(kt, order, generateID, []string{}, assetIDs)
	if err != nil {
		logs.Errorf("failed to syn to cmdb, order id: %s, generateId: %d, err: %v, rid: %s", order.SubOrderId,
			generateID, err, kt.Rid)
		return err
	}

	devices := g.buildUpgradeDevicesInfo(kt, items, order, generateID, mapAssetIDToHost, zoneRegionMap)
	if err = model.Operation().DeviceInfo().CreateDeviceInfos(kt.Ctx, devices); err != nil {
		logs.Errorf("failed to save device info to db, order id: %s, generateId: %d, err: %v, devicesNum: %d, "+
			"devices: %+v, rid: %s", order.SubOrderId, generateID, err, len(devices), cvt.PtrToSlice(devices), kt.Rid)
		return err
	}

	logs.Infof("successfully sync device info to cc, orderId: %s, generateId: %d, assets: %+v, "+
		"devices: %+v, rid: %s", order.SubOrderId, generateID, assetIDs, cvt.PtrToSlice(devices), kt.Rid)

	return nil
}

func (g *Generator) buildUpgradeDevicesInfo(kt *kit.Kit, items []*types.DeviceInfo, order *types.ApplyOrder, generateID uint64,
	mapAssetIDToHost map[string]*cmdb.Host, zoneRegionMap map[string]string) []*types.DeviceInfo {

	// save device info to db
	now := time.Now()
	var devices []*types.DeviceInfo

	for _, item := range items {
		if isDup, _ := g.isDuplicateHost(order.SubOrderId, item.AssetId); isDup {
			logs.Warnf("duplicate host for order id: %s, ip: %s, assetId: %s", order.SubOrderId, item.Ip, item.AssetId)
			continue
		}
		device := &types.DeviceInfo{
			OrderId:      order.OrderId,
			SubOrderId:   order.SubOrderId,
			GenerateId:   generateID,
			BkBizId:      int(order.BkBizId),
			User:         order.User,
			InstanceID:   item.InstanceID,
			AssetId:      item.AssetId,
			Ip:           item.Ip,
			RequireType:  order.RequireType,
			ResourceType: order.ResourceType,
			// set device type according to order specification by default
			DeviceType:  item.DeviceType,
			Description: order.Description,
			Remark:      order.Remark,
			// 升降配机器目前不需要init和deliver，直接设置为true
			IsInited:          true,
			IsDelivered:       true,
			IsChecked:         false,
			IsMatched:         false,
			IsDiskChecked:     false,
			GenerateTaskId:    item.GenerateTaskId,
			GenerateTaskLink:  item.GenerateTaskLink,
			InitTaskId:        item.InitTaskId,
			InitTaskLink:      item.InitTaskLink,
			DiskCheckTaskId:   item.DiskCheckTaskId,
			DiskCheckTaskLink: item.DiskCheckTaskLink,
			Deliverer:         item.Deliverer,
			IsManualMatched:   item.IsManualMatched,
			CloudZone:         item.CloudZone,
			CloudRegion:       zoneRegionMap[item.CloudZone],
			CreateAt:          now,
			UpdateAt:          now,
		}
		// add device detail info from cc
		if host, ok := mapAssetIDToHost[item.AssetId]; !ok {
			logs.Warnf("failed to get host detail info in cc, subOrderID: %s, assetID: %s", order.SubOrderId,
				item.AssetId)
		} else {
			// update device type from cc
			// device.DeviceType = host.SvrDeviceClass
			device.ZoneName = host.SubZone
			zoneId, err := strconv.Atoi(host.SubZoneId)
			if err != nil {
				logs.Warnf("failed to convert sub zone id %s to int", host.SubZoneId)
				device.ZoneID = 0
			} else {
				device.ZoneID = zoneId
			}
			device.ModuleName = host.ModuleName
			device.Equipment = host.RackId
		}

		// 获取母机IP
		ownerIP, err := g.getOwnerIP(kt, item.AssetId)
		if err != nil {
			logs.Warnf("failed to get owner IP for assetID: %s, subOrderID: %s, err: %v, rid: %s",
				item.AssetId, order.SubOrderId, err, kt.Rid)
		} else {
			device.OwnerIP = ownerIP
		}

		devices = append(devices, device)
	}
	return devices
}
