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

package tcloud

import (
	"fmt"

	actioncvm "hcm/cmd/task-server/logics/action/cvm"
	ts "hcm/pkg/api/task-server"
	"hcm/pkg/async/action"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/logs"
)

// Deliver 执行资源交付
func (a *ApplicationOfCreateTCloudCvm) Deliver() (enumor.ApplicationStatus, map[string]interface{}, error) {
	req := a.toHcProtoTCloudBatchCreateReq(false)
	opt := &actioncvm.AssignCvmOption{BizID: a.req.BkBizID, BkCloudID: a.req.BkCloudID}
	tasks := actioncvm.BuildCreateCvmTasks(req.RequiredCount, constant.BatchCreateCvmFromCloudMaxLimit, opt,
		func(actionID action.ActIDType, count int64) ts.CustomFlowTask {
			req.RequiredCount = count
			return ts.CustomFlowTask{
				ActionID:   actionID,
				ActionName: enumor.ActionCreateCvm,
				Params: &actioncvm.CreateOption{
					Vendor:               enumor.TCloud,
					TCloudBatchCreateReq: *req,
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
