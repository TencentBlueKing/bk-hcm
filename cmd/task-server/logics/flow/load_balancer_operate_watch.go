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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typesasync "hcm/pkg/dal/dao/types/async"
	tableasync "hcm/pkg/dal/table/async"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"

	"github.com/jmoiron/sqlx"
)

var _ action.Action = new(FlowSlaveOperateWatchAction)
var _ action.ParameterAction = new(FlowSlaveOperateWatchAction)

// FlowSlaveOperateWatchAction define load balancer operate watch.
type FlowSlaveOperateWatchAction struct{}

// FlowSlaveOperateWatchOption define slave operate watch option.
type FlowSlaveOperateWatchOption struct {
	FlowID string `json:"flow_id" validate:"required"`
	// 资源ID，比如负载均衡ID
	ResID string `json:"res_id" validate:"required"`
	// 资源类型
	ResType enumor.CloudResourceType `json:"res_type" validate:"required"`
	// 子资源ID数组，比如目标组ID
	SubResIDs []string `json:"sub_res_ids" validate:"omitempty"`
	// 子资源类型
	SubResType enumor.CloudResourceType `json:"sub_res_type" validate:"omitempty"`
	// 任务类型
	TaskType enumor.TaskType `json:"task_type" validate:"required"`
}

// Validate FlowSlaveOperateWatchOption.
func (opt FlowSlaveOperateWatchOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act FlowSlaveOperateWatchAction) ParameterNew() (params interface{}) {
	return new(FlowSlaveOperateWatchOption)
}

// Name return action name
func (act FlowSlaveOperateWatchAction) Name() enumor.ActionName {
	return enumor.ActionFlowSlaveOperateWatch
}

// Run flow watch.
func (act FlowSlaveOperateWatchAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*FlowSlaveOperateWatchOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	logs.Infof("flow slave operate watch start, mainFlowID: %s, resID: %s, resType: %s, subResType: %s, "+
		"taskType: %s, rid: %s", opt.FlowID, opt.ResID, opt.ResType, opt.SubResType, opt.TaskType, kt.Kit().Rid)

	start := time.Now()
	end := start.Add(OperateWatchTimeout)
	for {
		if time.Now().After(end) {
			logs.Errorf("flow slave operate watch wait timeout, mainFlowID: %s, resID: %s, resType: %s, "+
				"taskType: %s, waitedSec: %.3f, rid: %s", opt.FlowID, opt.ResID, opt.ResType, opt.TaskType,
				time.Since(start).Seconds(), kt.Kit().Rid)
			return nil, fmt.Errorf("wait timeout, async task flow: %s is running", opt.FlowID)
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
func (act FlowSlaveOperateWatchAction) processResFlow(kt run.ExecuteKit, opt *FlowSlaveOperateWatchOption,
	flowInfo tableasync.AsyncFlowTable) (bool, error) {

	switch flowInfo.State {
	case enumor.FlowSuccess, enumor.FlowCancel, enumor.FlowFailed:
		// 主flow已终态，terminalObservedDelaySec 是主flow终态到watch观察到该终态的延迟
		logs.Infof("watch observed main flow terminal state, mainFlowID: %s, resID: %s, resType: %s, state: %s, "+
			"terminalObservedDelaySec: %.3f, rid: %s", opt.FlowID, opt.ResID, opt.ResType, flowInfo.State,
			elapsedSecSince(string(flowInfo.UpdatedAt)), kt.Kit().Rid)

		// 当Flow失败时，检查资源锁定是否超时
		resFlowLockList, err := act.queryResFlowLock(kt, opt)
		if err != nil {
			return false, err
		}
		if len(resFlowLockList) == 0 {
			// 主flow已终态，但该flow名下没有锁记录，说明锁已被提前释放或从未创建成功，属异常
			logs.Warnf("main flow reached terminal state but no res flow lock owned by it, mainFlowID: %s, "+
				"resID: %s, resType: %s, state: %s, rid: %s", opt.FlowID, opt.ResID, opt.ResType, flowInfo.State,
				kt.Kit().Rid)
			return true, nil
		}

		var resStatus enumor.ResFlowStatus
		if flowInfo.State == enumor.FlowSuccess {
			resStatus = enumor.SuccessResFlowStatus
		}
		if flowInfo.State == enumor.FlowCancel || flowInfo.State == enumor.FlowFailed {
			resStatus = enumor.CancelResFlowStatus
		}

		if err := act.updateTGListenerRuleRelBindStatus(kt.Kit(), opt, flowInfo.State); err != nil {
			return false, err
		}

		// 解锁资源
		if err = act.processUnlockResFlow(kt, opt, resStatus); err != nil {
			logs.Errorf("unlock res flow failed, err: %v, mainFlowID: %s, resID: %s, resType: %s, resStatus: %s, "+
				"rid: %s", err, opt.FlowID, opt.ResID, opt.ResType, resStatus, kt.Kit().Rid)
			return true, err
		}

		// terminalToUnlockSec 是主flow终态到锁真正释放的总延迟，撞锁问题的时间窗即落在这段
		logs.Infof("unlock res flow success, mainFlowID: %s, resID: %s, resType: %s, resStatus: %s, "+
			"terminalToUnlockSec: %.3f, rid: %s", opt.FlowID, opt.ResID, opt.ResType, resStatus,
			elapsedSecSince(string(flowInfo.UpdatedAt)), kt.Kit().Rid)

		return true, nil
	case enumor.FlowInit:
		// 需要检查资源是否已锁定
		resFlowLockList, err := act.queryResFlowLock(kt, opt)
		if err != nil {
			return false, err
		}
		if len(resFlowLockList) == 0 {
			return true, nil
		}

		// 如已锁定资源，则需要更新Flow状态为Pending
		err = act.updateFlowStateByCAS(kt.Kit(), opt.FlowID, enumor.FlowInit, enumor.FlowPending)
		if err != nil {
			logs.Errorf("call taskserver to update flow state failed, err: %v, mainFlowID: %s, resID: %s, "+
				"resType: %s, rid: %s", err, opt.FlowID, opt.ResID, opt.ResType, kt.Kit().Rid)
			return false, err
		}

		logs.Infof("main flow state updated to pending after lock confirmed, mainFlowID: %s, resID: %s, "+
			"resType: %s, rid: %s", opt.FlowID, opt.ResID, opt.ResType, kt.Kit().Rid)

		return false, nil
	default:
		return false, nil
	}
}
func (act FlowSlaveOperateWatchAction) processUnlockResFlow(kt run.ExecuteKit, opt *FlowSlaveOperateWatchOption,
	status enumor.ResFlowStatus) error {

	unlockReq := &dataproto.ResFlowLockReq{
		ResID:   opt.ResID,
		ResType: opt.ResType,
		FlowID:  opt.FlowID,
		Status:  status,
	}
	return actcli.GetDataService().Global.LoadBalancer.ResFlowUnLock(kt.Kit(), unlockReq)
}

// elapsedSecSince 返回从 timeStr 表示的时刻到当前的秒数，timeStr 为 DB 时间字段（RFC3339）。
// 解析失败时返回 -1，表示该耗时不可用，避免因日志取值影响主流程。
func elapsedSecSince(timeStr string) float64 {
	t, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		return -1
	}

	return time.Since(t).Seconds()
}

func (act FlowSlaveOperateWatchAction) queryResFlowLock(kt run.ExecuteKit, opt *FlowSlaveOperateWatchOption) (
	[]tablelb.ResourceFlowLockTable, error) {

	// 当Flow失败时，检查资源锁定是否超时
	lockReq := &types.ListOption{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("res_id", opt.ResID),
			tools.RuleEqual("res_type", opt.ResType),
			tools.RuleEqual("owner", opt.FlowID),
		),
		Page: core.NewDefaultBasePage(),
	}
	resFlowLockList, err := actcli.GetDaoSet().ResourceFlowLock().List(kt.Kit(), lockReq)
	if err != nil {
		logs.Errorf("list query flow lock failed, err: %v, flowID: %s, rid: %s", err, opt.FlowID, kt.Kit().Rid)
		return nil, err
	}
	return resFlowLockList.Details, nil
}

func (act FlowSlaveOperateWatchAction) updateFlowStateByCAS(kt *kit.Kit, flowID string,
	source, target enumor.FlowState) error {

	_, err := actcli.GetDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		info := &typesasync.UpdateFlowInfo{
			ID:     flowID,
			Source: source,
			Target: target,
		}
		if err := actcli.GetDaoSet().AsyncFlow().UpdateStateByCAS(kt, txn, info); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("call taskserver to update flow watch pending state failed, err: %v, flowID: %s, "+
			"source: %s, target: %s, rid: %s", err, flowID, source, target, kt.Rid)
		return err
	}

	logs.Infof("flow state updated by cas success, flowID: %s, state: %s -> %s, rid: %s",
		flowID, source, target, kt.Rid)

	return nil
}

// updateTGListenerRuleRelBindStatus 更新目标组与监听器的绑定状态
func (act FlowSlaveOperateWatchAction) updateTGListenerRuleRelBindStatus(kt *kit.Kit,
	opt *FlowSlaveOperateWatchOption, flowState enumor.FlowState) error {

	if opt == nil || opt.TaskType != enumor.ApplyTargetGroupType || opt.SubResType != enumor.TargetGroupCloudResType {
		return nil
	}

	var bindStatus enumor.BindingStatus
	switch flowState {
	case enumor.FlowSuccess:
		bindStatus = enumor.SuccessBindingStatus
	case enumor.FlowCancel, enumor.FlowFailed:
		bindStatus = enumor.FailedBindingStatus
	default:
		return nil
	}

	for _, targetGroupID := range opt.SubResIDs {
		if err := actcli.GetDataService().Global.LoadBalancer.BatchUpdateListenerRuleRelStatusByTGID(kt, targetGroupID,
			&dataproto.TGListenerRelStatusUpdateReq{BindingStatus: bindStatus}); err != nil {
			return err
		}
	}
	return nil
}

// Rollback Flow查询状态失败时的回滚Action，此处不需要回滚处理
func (act FlowSlaveOperateWatchAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- FlowSlaveOperateWatchAction Rollback -----------, params: %s, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
