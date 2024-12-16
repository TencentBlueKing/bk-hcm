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

package huawei

import (
	"fmt"

	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	protocloud "hcm/pkg/api/data-service/cloud"
	protodisk "hcm/pkg/api/data-service/cloud/disk"
	protoni "hcm/pkg/api/data-service/cloud/network-interface"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
)

// Deliver 执行资源交付
func (a *ApplicationOfCreateHuaWeiCvm) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {

	req := a.toHcProtoHuaWeiBatchCreateReq(false)
	tasks := actioncvm.BuildCreateCvmTasks(int64(req.RequiredCount), a.req.BkBizID,
		constant.BatchCreateCvmFromCloudMaxLimit,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = int32(count)
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:               enumor.HuaWei,
					HuaWeiBatchCreateReq: *req,
				},
			}
		})
	addReq := &ts.AddCustomFlowReq{
		Name:  enumor.FlowCreateCvm,
		Tasks: tasks,
	}
	result, err := a.Client.TaskServer().CreateCustomFlow(a.Cts.Kit, addReq)
	if err != nil {
		logs.Errorf("call taskserver to create custom flow failed, err: %v, rid: %s", err, a.Cts.Kit.Rid)
		return enumor.DeliverError, map[string]interface{}{"error": fmt.Errorf("delivery task failed, err: %v",
			err)}, err
	}
	deliverDetail := map[string]interface{}{"flow_id": result.ID}

	return enumor.Delivering, deliverDetail, nil
}

func (a *ApplicationOfCreateHuaWeiCvm) assignToBiz(cloudCvmIDs []string) ([]string, error) {
	req := a.req
	// 云ID查询主机
	cvmInfo, err := a.ListCvm(a.Vendor(), req.AccountID, req.Region, cloudCvmIDs)
	if err != nil {
		return []string{}, err
	}
	cvmIDs := make([]string, 0, len(cvmInfo))
	for _, cvm := range cvmInfo {
		cvmIDs = append(cvmIDs, cvm.ID)
	}

	// 主机分配给业务
	err = a.Client.DataService().Global.Cvm.BatchUpdateCvmCommonInfo(
		a.Cts.Kit,
		&protocloud.CvmCommonInfoBatchUpdateReq{IDs: cvmIDs, BkBizID: req.BkBizID},
	)
	if err != nil {
		return cvmIDs, err
	}

	// create deliver audit
	err = a.Audit.ResDeliverAudit(a.Cts.Kit, enumor.CvmAuditResType, cvmIDs, int64(req.BkBizID))
	if err != nil {
		logs.Errorf("create deliver cvm audit failed, err: %v, rid: %s", err, a.Cts.Kit)
		return nil, err
	}

	// 主机关联资源硬盘分配给业务
	diskIDs, err := a.ListDiskIDByCvm(cvmIDs)
	if err != nil {
		return cvmIDs, err
	}
	if len(diskIDs) > 0 {
		_, err = a.Client.DataService().Global.BatchUpdateDisk(
			a.Cts.Kit,
			&protodisk.DiskBatchUpdateReq{
				IDs:     diskIDs,
				BkBizID: uint64(req.BkBizID),
			},
		)
		if err != nil {
			return cvmIDs, err
		}

		// create deliver audit
		err = a.Audit.ResDeliverAudit(a.Cts.Kit, enumor.DiskAuditResType, diskIDs, int64(req.BkBizID))
		if err != nil {
			logs.Errorf("create deliver disk audit failed, err: %v, rid: %s", err, a.Cts.Kit)
			return nil, err
		}
	}

	// 主机关联资源网络接口分配给业务
	niIDs, err := a.ListNIIDByCvm(cvmIDs)
	if err != nil {
		return cvmIDs, err
	}
	if len(niIDs) > 0 {
		err = a.Client.DataService().Global.NetworkInterface.BatchUpdateNICommonInfo(
			a.Cts.Kit,
			&protoni.NetworkInterfaceCommonInfoBatchUpdateReq{
				IDs:     niIDs,
				BkBizID: int64(req.BkBizID),
			},
		)
		if err != nil {
			return cvmIDs, err
		}

		// create deliver audit
		err = a.Audit.ResDeliverAudit(a.Cts.Kit, enumor.NetworkInterfaceAuditResType, niIDs, int64(req.BkBizID))
		if err != nil {
			logs.Errorf("create deliver ni audit failed, err: %v, rid: %s", err, a.Cts.Kit)
			return nil, err
		}
	}

	return cvmIDs, nil
}
