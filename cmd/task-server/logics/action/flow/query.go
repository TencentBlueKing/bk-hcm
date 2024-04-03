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

package actionflow

import (
	"fmt"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	dataproto "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/logs"
)

var _ action.Action = new(FlowWatchAction)
var _ action.ParameterAction = new(FlowWatchAction)

// FlowWatchAction define flow watch.
type FlowWatchAction struct{}

// FlowWatchOption define flow watch option.
type FlowWatchOption struct {
	FlowID   string                   `json:"flow_id" validate:"required"`
	ResID    string                   `json:"res_id" validate:"required"`
	ResType  enumor.CloudResourceType `json:"res_type" validate:"required"`
	TaskType enumor.TaskType          `json:"task_type" validate:"required"`
}

// Validate FlowWatchOption.
func (opt FlowWatchOption) Validate() error {
	return opt.Validate()
}

// ParameterNew return request params.
func (act FlowWatchAction) ParameterNew() (params interface{}) {
	return new(FlowWatchOption)
}

// Name return action name
func (act FlowWatchAction) Name() enumor.ActionName {
	return enumor.ActionFlowWatch
}

// Run flow watch.
func (act FlowWatchAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*FlowWatchOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	end := time.Now().Add(5 * time.Minute)
	for {
		if time.Now().After(end) {
			return nil, fmt.Errorf("wait timeout, async task: %s is running", opt.FlowID)
		}

		req := &types.ListOption{
			Filter: tools.EqualExpression("id", opt.FlowID),
			Page:   core.NewDefaultBasePage(),
		}
		flowList, err := actcli.GetDaoSet().AsyncFlow().List(kt.Kit(), req)
		if err != nil {
			logs.Errorf("list query flow failed, err: %v, flowID: %s, rid: %s", err, opt.FlowID, kt.Kit().Rid)
			return nil, err
		}

		if len(flowList.Details) == 0 {
			logs.Infof("list query flow not found, flowID: %s, rid: %s", opt.FlowID, kt.Kit().Rid)
			return nil, nil
		}

		isSkip, err := act.processResFlow(kt, opt, flowList.Details[0])
		if err != nil {
			return nil, err
		}
		// 任务已终态，无需继续处理
		if isSkip {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	return nil, nil
}

// processResFlow 检查Flow是否终态状态、解锁资源跟Flow的状态
func (act FlowWatchAction) processResFlow(kt run.ExecuteKit, opt *FlowWatchOption,
	flowInfo tableasync.AsyncFlowTable) (bool, error) {

	switch flowInfo.State {
	case enumor.FlowSuccess, enumor.FlowCancel:
		var resStatus enumor.ResFlowStatus
		if flowInfo.State == enumor.FlowSuccess {
			resStatus = enumor.SuccessResFlowStatus
		}
		if flowInfo.State == enumor.FlowCancel {
			resStatus = enumor.CancelResFlowStatus
		}
		// 解锁资源
		unlockReq := &dataproto.ResFlowLockReq{
			ResID:   opt.ResID,
			ResType: opt.ResType,
			FlowID:  opt.FlowID,
			Status:  resStatus,
		}
		return true, actcli.GetDataService().Global.LoadBalancer.ResFlowUnLock(kt.Kit(), unlockReq)
	case enumor.FlowFailed:
		// 当Flow失败时，检查资源锁定是否超时
		lockReq := &types.ListOption{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("res_id", opt.ResID),
				tools.RuleEqual("res_type", opt.ResType),
				tools.RuleEqual("owner", opt.FlowID),
			),
			Page: core.NewDefaultBasePage(),
		}
		resFlowLockList, err := actcli.GetDaoSet().LoadBalancerFlowLock().List(kt.Kit(), lockReq)
		if err != nil {
			logs.Errorf("list query flow failed, err: %v, flowID: %s, rid: %s", err, opt.FlowID, kt.Kit().Rid)
			return false, err
		}
		if len(resFlowLockList.Details) == 0 {
			return true, nil
		}

		createTime, err := time.Parse(constant.TimeStdFormat, string(resFlowLockList.Details[0].CreatedAt))
		if err != nil {
			return false, err
		}

		nowTime := time.Now()
		if nowTime.Sub(createTime).Hours() > constant.ResFlowLockExpireDays*24 {
			timeoutReq := &dataproto.ResFlowLockReq{
				ResID:   opt.ResID,
				ResType: opt.ResType,
				FlowID:  opt.FlowID,
				Status:  enumor.TimeoutResFlowStatus,
			}
			return true, actcli.GetDataService().Global.LoadBalancer.ResFlowUnLock(kt.Kit(), timeoutReq)
		}
	default:
		return false, nil
	}

	return true, nil
}

// Rollback Flow查询状态失败时的回滚Action，此处不需要回滚处理
func (act FlowWatchAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- FlowWatchAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}
