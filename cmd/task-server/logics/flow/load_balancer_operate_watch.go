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

var _ action.Action = new(LoadBalancerOperateWatchAction)
var _ action.ParameterAction = new(LoadBalancerOperateWatchAction)

// LoadBalancerOperateWatchAction define load balancer operate watch.
type LoadBalancerOperateWatchAction struct{}

// LoadBalancerOperateWatchOption define load balancer operate watch option.
type LoadBalancerOperateWatchOption struct {
	FlowID string `json:"flow_id" validate:"required"`
	// 资源ID，比如负载均衡ID
	ResID string `json:"res_id" validate:"required"`
	// 资源类型
	ResType enumor.CloudResourceType `json:"res_type" validate:"required"`
	// 子资源ID数组，比如目标组ID
	SubResIDs []string `json:"sub_res_ids" validate:"required"`
	// 子资源类型
	SubResType enumor.CloudResourceType `json:"sub_res_type" validate:"required"`
	// 任务类型
	TaskType enumor.TaskType `json:"task_type" validate:"required"`
}

// Validate LoadBalancerOperateWatchOption.
func (opt LoadBalancerOperateWatchOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act LoadBalancerOperateWatchAction) ParameterNew() (params interface{}) {
	return new(LoadBalancerOperateWatchOption)
}

// Name return action name
func (act LoadBalancerOperateWatchAction) Name() enumor.ActionName {
	return enumor.ActionLoadBalancerOperateWatch
}

// Run flow watch.
func (act LoadBalancerOperateWatchAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*LoadBalancerOperateWatchOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	end := time.Now().Add(OperateWatchTimeout)
	for {
		if time.Now().After(end) {
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
func (act LoadBalancerOperateWatchAction) processResFlow(kt run.ExecuteKit, opt *LoadBalancerOperateWatchOption,
	flowInfo tableasync.AsyncFlowTable) (bool, error) {

	switch flowInfo.State {
	case enumor.FlowSuccess, enumor.FlowCancel, enumor.FlowFailed:
		// 当Flow失败时，检查资源锁定是否超时
		resFlowLockList, err := act.queryResFlowLock(kt, opt)
		if err != nil {
			return false, err
		}
		if len(resFlowLockList) == 0 {
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
		err = act.processUnlockResFlow(kt, opt, resStatus)
		return true, err
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
			logs.Errorf("call taskserver to update flow state failed, err: %v, flowID: %s", err, opt.FlowID)
			return false, err
		}
		return false, nil
	default:
		return false, nil
	}
}
func (act LoadBalancerOperateWatchAction) processUnlockResFlow(kt run.ExecuteKit, opt *LoadBalancerOperateWatchOption,
	status enumor.ResFlowStatus) error {

	unlockReq := &dataproto.ResFlowLockReq{
		ResID:   opt.ResID,
		ResType: opt.ResType,
		FlowID:  opt.FlowID,
		Status:  status,
	}
	return actcli.GetDataService().Global.LoadBalancer.ResFlowUnLock(kt.Kit(), unlockReq)
}

func (act LoadBalancerOperateWatchAction) queryResFlowLock(kt run.ExecuteKit, opt *LoadBalancerOperateWatchOption) (
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

func (act LoadBalancerOperateWatchAction) updateFlowStateByCAS(kt *kit.Kit, flowID string,
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
	return nil
}

// updateTGListenerRuleRelBindStatus 更新目标组与监听器的绑定状态
func (act LoadBalancerOperateWatchAction) updateTGListenerRuleRelBindStatus(kt *kit.Kit,
	opt *LoadBalancerOperateWatchOption, flowState enumor.FlowState) error {

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
func (act LoadBalancerOperateWatchAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- LoadBalancerOperateWatchAction Rollback -----------, params: %s, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
