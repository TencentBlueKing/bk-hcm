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
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/async/backend"
	"hcm/pkg/async/producer"
	dataservice "hcm/pkg/client/data-service"
	taskserver "hcm/pkg/client/task-server"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

func lockResFlowStatus(kt *kit.Kit, dataCli *dataservice.Client, taskCli *taskserver.Client, resID string,
	resType enumor.CloudResourceType, flowID string, taskType enumor.TaskType) error {

	// 锁定资源跟Flow的状态
	opt := &dataproto.ResFlowLockReq{
		ResID:    resID,
		ResType:  resType,
		FlowID:   flowID,
		Status:   enumor.ExecutingResFlowStatus,
		TaskType: taskType,
	}
	err := dataCli.Global.LoadBalancer.ResFlowLock(kt, opt)
	if err != nil {
		logs.Errorf("call dataservice to lock res and flow failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
		return err
	}

	// 更新Flow状态为pending
	flowStateReq := &producer.UpdateCustomFlowStateOption{
		FlowInfos: []backend.UpdateFlowInfo{{
			ID:     flowID,
			Source: enumor.FlowInit,
			Target: enumor.FlowPending,
		}},
	}
	err = taskCli.UpdateCustomFlowState(kt, flowStateReq)
	if err != nil {
		logs.Errorf("call taskserver to update flow state failed, err: %v, flowID: %s, rid: %s", err, flowID, kt.Rid)
		return err
	}

	return nil
}

func checkResFlowRel(kt *kit.Kit, dataCli *dataservice.Client, resID string, resType enumor.CloudResourceType) (*corelb.BaseResFlowLock, error) {

	// 预检测-当前资源是否有锁定中的数据
	lockReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("res_type", resType),
		),
		Page: core.NewDefaultBasePage(),
	}
	lockRet, err := dataCli.Global.LoadBalancer.ListResFlowLock(kt, lockReq)
	if err != nil {
		logs.Errorf("list res flow lock failed, err: %v, resID: %s, resType: %s, rid: %s", err, resID, resType,
			kt.Rid)
		return nil, err
	}
	if len(lockRet.Details) > 0 {
		return &lockRet.Details[0], errf.Newf(errf.LoadBalancerTaskExecuting, "resID: %s is processing", resID)
	}

	// 预检测-当前资源是否有未终态的状态
	flowRelReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", resID),
			tools.RuleEqual("res_type", resType),
			tools.RuleEqual("status", enumor.ExecutingResFlowStatus),
		),
		Page: core.NewDefaultBasePage(),
	}
	flowRelRet, err := dataCli.Global.LoadBalancer.ListResFlowRel(kt, flowRelReq)
	if err != nil {
		logs.Errorf("list res flow rel failed, err: %v, resID: %s, resType: %s, rid: %s", err, resID, resType, kt.Rid)
		return nil, err
	}
	if len(flowRelRet.Details) > 0 {
		return &corelb.BaseResFlowLock{
				ResID:   resID,
				ResType: resType,
				Owner:   flowRelRet.Details[0].FlowID,
			},
			errf.Newf(errf.LoadBalancerTaskExecuting, "%s of resID: %s is processing", resType, resID)
	}

	return nil, nil
}
