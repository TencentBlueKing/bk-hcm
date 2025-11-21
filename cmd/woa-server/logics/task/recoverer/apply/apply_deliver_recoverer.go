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

	"hcm/cmd/woa-server/logics/task/scheduler/record"
	model "hcm/cmd/woa-server/model/task"
	"hcm/cmd/woa-server/storage/driver/mongodb"
	recovertask "hcm/cmd/woa-server/types/task"
	types "hcm/cmd/woa-server/types/task"
	"hcm/pkg/criteria/mapstr"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// recoverDeliverOrder 恢复当前正在转移的订单
func (r *applyRecoverer) recoverDeliverOrder(kt *kit.Kit, generateRecord *types.GenerateRecord,
	order *types.ApplyOrder) error {

	devices, err := r.getDeviceByStatus(kt, order.SubOrderId, true)
	if err != nil {
		logs.Errorf("failed to get devices by status, err: %v, subOrderId: %s, rid: %s", err, order.SubOrderId, kt.Rid)
		return err
	}

	for _, device := range devices {
		deliverRecord, err := record.GetDeliverRecord(kt, order.SubOrderId, device.Ip, device.AssetId)
		// 若没有deliverRecord,执行主机转移；否则根据deliverRecord状态判断是否需要转移主机
		if err != nil {
			if !mongodb.Client().IsNotFoundError(err) {
				logs.Errorf("failed to get deliver record by ip and assetId, err: %v, ip: %s, assetId: %s, rid: %s",
					err, device.Ip, device.AssetId, kt.Rid)
				continue
			}

			if err = r.schedulerIf.DeliverDevice(device, order); err != nil {
				logs.Errorf("failed to deliver device, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId,
					device.Ip, err, kt.Rid)
			}
			continue
		}
		if deliverRecord.Status != types.DeliverStatusHandling {
			continue
		}

		if err = r.recoverDelivering(kt, order, device); err != nil {
			logs.Errorf("failed to recover deliver order, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId,
				device.Ip, err, kt.Rid)
		}

	}
	// update deliver step
	if err := record.UpdateDeliverStep(order.SubOrderId, order.TotalNum); err != nil {
		logs.Errorf("failed to update deliverStep step, subOrderId: %s, err: %v, rid: %s", order.SubOrderId, err,
			kt.Rid)
		return err
	}

	return r.schedulerIf.FinalApplyStep(kt, generateRecord, order)
}

// recoverDeliverStep 恢复deliverStep为initing及handling状态订单
func (r *applyRecoverer) recoverDeliverStep(kt *kit.Kit, order *types.ApplyOrder) error {
	generateRecords, err := r.schedulerIf.GetGenerateRecords(kt, order.SubOrderId)
	if err != nil {
		logs.Errorf("failed to get generate records by subOrderId, subOrderId: %s, err: %v, rid: %s", order.SubOrderId,
			err, kt.Rid)
		return err
	}

	for _, generateRecord := range generateRecords {
		if generateRecord.Status == types.GenerateStatusSuccess && !generateRecord.IsMatched {
			err := r.recoverDeliverOrder(kt, generateRecord, order)
			if err != nil {
				logs.Errorf("failed to recover deliver order, subOrderId: %s, generateId: %d, err: %v, rid: %s",
					order.SubOrderId, generateRecord.GenerateId, err, kt.Rid)
				continue
			}
		}
	}

	return nil
}

// recoverDelivering 恢复deliverRecord为handling订单
func (r *applyRecoverer) recoverDelivering(kt *kit.Kit, order *types.ApplyOrder, device *types.DeviceInfo) error {
	ip := device.Ip
	bkBizID, err := r.getBizID(kt, ip)
	if err != nil {
		logs.Errorf("failed to get bk biz id, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId, ip, err,
			kt.Rid)
		return err
	}
	// 只转移在931业务下主机，避免操作仍在使用主机
	if bkBizID == recovertask.ResourceOperationService && order.BkBizId != recovertask.ResourceOperationService {
		if err = r.schedulerIf.DeliverDevice(device, order); err != nil {
			logs.Errorf("failed to deliver device, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId, ip, err,
				kt.Rid)
			return fmt.Errorf("failed to deliver device, subOrderId: %s, ip: %s, err: %v", order.SubOrderId, ip, err)
		}

		return nil
	}
	// 机器被转移,不在931业务下，也不再目标业务下
	if bkBizID != order.BkBizId {
		logs.Infof("recover: apply order host is not in original biz or target biz, subOrderId: %s, ip: %s, "+
			"order.bkBizId: %d, bkBizID: %d, rid: %s", order.SubOrderId, ip, order.BkBizId, bkBizID, kt.Rid)
		return fmt.Errorf("recover: apply order host is not in original biz or target biz, subOrderId: %s, ip: %s, "+
			"order.bkBizId: %d", order.SubOrderId, ip, order.BkBizId)
	}

	hostInfo, err := r.cmdbCli.GetHostInfoByIP(kt, ip, 0)
	if err != nil {
		logs.Errorf("recover: get host info by host ip failed, subOrderId: %s, ip: %s, err: %v, rid: %s",
			order.SubOrderId, ip, err, kt.Rid)
		return err
	}

	if err = r.schedulerIf.UpdateHostOperator(device, hostInfo.BkHostID, order.User); err != nil {
		logs.Errorf("failed to update host operator, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId,
			device.Ip, err, kt.Rid)
		return err
	}

	// 2. update device status
	if err = r.schedulerIf.SetDeviceDelivered(device); err != nil {
		logs.Errorf("failed to set device delivered, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId,
			device.Ip, err, kt.Rid)
		return fmt.Errorf("failed to set device delivered, subOrderId: %s, ip: %s, err: %v", order.SubOrderId,
			device.Ip, err)
	}

	// update deliver record
	if err = record.UpdateDeliverRecord(device, "success", types.DeliverStatusSuccess); err != nil {
		logs.Errorf("failed to deliver device, subOrderId: %s, ip: %s, err: %v, rid: %s", order.SubOrderId, device.Ip,
			err, kt.Rid)
		return fmt.Errorf("failed to deliver device, subOrderId: %s, ip: %s, err: %v", order.SubOrderId, device.Ip, err)
	}

	return nil
}

func (r *applyRecoverer) getDeviceByStatus(kt *kit.Kit, subOrderId string, isInited bool) ([]*types.DeviceInfo,
	error) {

	filter := &mapstr.MapStr{
		"suborder_id": subOrderId,
		"is_inited":   isInited,
	}
	devices, err := model.Operation().DeviceInfo().GetDeviceInfo(kt.Ctx, filter)
	if err != nil {
		logs.Errorf("failed to get device by isInited, subOrderId: %s, isInited: %v, err: %v, rid: %s", subOrderId,
			isInited, err, kt.Rid)
		return nil, err
	}

	return devices, nil
}
