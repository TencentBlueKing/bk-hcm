/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package actionlb

import (
	"fmt"
	"strings"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"
	cvt "hcm/pkg/tools/converter"
	"hcm/pkg/tools/retry"
)

// --------------------------[创建TCloud监听器]-----------------------------

var _ action.Action = new(BatchTaskTCloudCreateListenerAction)
var _ action.ParameterAction = new(BatchTaskTCloudCreateListenerAction)

// BatchTaskTCloudCreateListenerAction 创建TCloud监听器
type BatchTaskTCloudCreateListenerAction struct{}

// BatchTaskTCloudCreateListenerOption ...
type BatchTaskTCloudCreateListenerOption struct {
	ManagementDetailIDs []string                        `json:"management_detail_ids" validate:"required,min=1,max=20"`
	Listeners           []*hclb.TCloudListenerCreateReq `json:"Listeners,required,min=1,max=20,dive,required"`
	// 是否在某个监听器创建失败的时候停止后续操作执行，默认不停止
	AbortOnFailed bool `json:"abort_on_failed"`
}

// Validate validate option.
func (opt BatchTaskTCloudCreateListenerOption) Validate() error {
	if len(opt.ManagementDetailIDs) != len(opt.Listeners) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and listeners length not match: %d! = %d",
			len(opt.ManagementDetailIDs), len(opt.Listeners))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskTCloudCreateListenerAction) ParameterNew() (params any) {
	return new(BatchTaskTCloudCreateListenerOption)
}

// Name return action name
func (act BatchTaskTCloudCreateListenerAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskTCloudCreateListener
}

// Run 创建监听器
func (act BatchTaskTCloudCreateListenerAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskTCloudCreateListenerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	asyncKit := kt.AsyncKit()

	results := make([]*core.CloudCreateResult, 0, len(opt.Listeners))
	for i := range opt.Listeners {
		detailID := opt.ManagementDetailIDs[i]
		// 逐条更新结果
		ret, createErr := act.createSingleListener(asyncKit, detailID, opt.Listeners[i]) // 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if createErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := batchUpdateTaskDetailResultState(asyncKit, []string{detailID}, targetState, ret, createErr)
		if err != nil {
			logs.Errorf("fail to set detail to %s after cloud operation finished, err: %v, rid: %s",
				targetState, err, asyncKit.Rid)
			return nil, err
		}
		if targetState == enumor.TaskDetailFailed && opt.AbortOnFailed {
			// abort
			return nil, err
		}
		results = append(results, ret)
	}
	// all success
	return results, nil
}

func (act BatchTaskTCloudCreateListenerAction) createSingleListener(kt *kit.Kit, detailId string,
	req *hclb.TCloudListenerCreateReq) (*core.CloudCreateResult, error) {

	detailList, err := listTaskDetail(kt, []string{detailId})
	if err != nil {
		logs.Errorf("fail to query task detail, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	detail := detailList[0]
	if detail.State == enumor.TaskDetailCancel {
		// 任务被取消，跳过该任务, 直接成功即可
		return nil, nil
	}
	if detail.State != enumor.TaskDetailInit {
		return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
			detail.ID, detail.State)
	}
	lbl, err := act.checkListenerExists(kt, req)
	if err != nil {
		return nil, err
	}
	if lbl != nil {
		// 已存在且参数一致，认为创建成功
		return &core.CloudCreateResult{ID: lbl.ID, CloudID: lbl.CloudID}, nil
	}

	// 更新任务状态为 running
	if err := batchUpdateTaskDetailState(kt, []string{detailId}, enumor.TaskDetailRunning); err != nil {
		return nil, fmt.Errorf("fail to update detail to running, err: %v", err)
	}

	var lblResp *hclb.ListenerCreateResult
	rangeMS := [2]uint{BatchTaskDefaultRetryDelayMinMS, BatchTaskDefaultRetryDelayMaxMS}
	policy := retry.NewRetryPolicy(0, rangeMS)
	for policy.RetryCount() < BatchTaskDefaultRetryTimes {

		lblResp, err = actcli.GetHCService().TCloud.Clb.CreateListener(kt, req)
		// 仅在碰到限频错误时进行重试
		if err != nil && strings.Contains(err.Error(), constant.TCloudLimitExceededErrCode) {
			if policy.RetryCount()+1 < BatchTaskDefaultRetryTimes {
				// 	非最后一次重试，继续sleep
				logs.Errorf("call tcloud reach rate limit, will sleep for retry, retry count: %d, err: %v, rid: %s",
					policy.RetryCount(), err, kt.Rid)
				policy.Sleep()
				continue
			}
		}
		// 其他情况都跳过
		break
	}

	if err != nil {
		logs.Errorf("fail to call hc to create listener, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return &core.CloudCreateResult{ID: lblResp.ID, CloudID: lblResp.CloudID}, nil
}

// 检查监听器是否存在，不存在，不返回错误。存在会返回数据库监听器实例，如果存在但是参数一直则不返回错误
func (act BatchTaskTCloudCreateListenerAction) checkListenerExists(kt *kit.Kit, req *hclb.TCloudListenerCreateReq) (
	lbl *corelb.TCloudListener, err error) {

	// 查询是否已经存在对应监听器
	lbReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", enumor.TCloud),
			tools.RuleEqual("lb_id", req.LbID),
			tools.RuleEqual("protocol", req.Protocol),
			tools.RuleEqual("port", req.Port),
		),
		Page: core.NewDefaultBasePage(),
	}
	lblResp, err := actcli.GetDataService().TCloud.LoadBalancer.ListListener(kt, lbReq)
	if err != nil {
		return nil, fmt.Errorf("fail to query listener, err: %v", err)
	}
	if len(lblResp.Details) == 0 {
		// 不存在，不返回错误
		return nil, nil
	}
	// 存在则判断是否和入参一致
	lbl = cvt.ValToPtr(lblResp.Details[0])

	if req.Name != lbl.Name {
		return lbl, fmt.Errorf("listener(%s) already exist, name mismatch, want: %s, db: %s",
			lbl.CloudID, req.Name, lbl.Name)
	}
	if req.BkBizID != lbl.BkBizID {
		return lbl, fmt.Errorf("listener(%s) already exist, biz id mismatch, want: %d, db: %d",
			lbl.CloudID, req.BkBizID, lbl.BkBizID)
	}
	if req.SniSwitch != lbl.SniSwitch {
		return lbl, fmt.Errorf("listener(%s) already exist, sni switch mismatch, want: %d, db: %d",
			lbl.CloudID, req.SniSwitch, lbl.SniSwitch)
	}
	if req.EndPort != nil {
		if lbl.Extension == nil {
			return lbl, fmt.Errorf("listener(%s) already exist, session expire mismatch, want: %d, db no ext",
				lbl.CloudID, req.SessionExpire)
		}
		if !assert.IsPtrInt64Equal(lbl.Extension.EndPort, req.EndPort) {
			return lbl, fmt.Errorf("listener(%s) already exist, session expire mismatch, want: %+v, db: %+v",
				lbl.CloudID, req.SessionExpire, lbl.Extension.EndPort)
		}
	}
	if req.Certificate != nil {
		if lbl.Extension == nil {
			return lbl, fmt.Errorf("listener(%s) already exist, cert mismatch, want: %+v, db no ext",
				lbl.CloudID, req.Certificate)
		}
		if isListenerCertChange(req.Certificate, lbl.Extension.Certificate) {
			return lbl, fmt.Errorf("listener(%s) already exist, cert mismatch, want: %+v, got: %+v", lbl.CloudID,
				req.Certificate, lbl.Extension.Certificate)
		}

	}

	if req.Protocol.IsLayer7Protocol() {
		return lbl, nil
	}
	// 对于四层需要继续查询规则

	if err := act.checkL4RuleExists(kt, lbl, req); err != nil {
		return lbl, fmt.Errorf("listener(%s) already exist, %v", lbl.CloudID, err)
	}
	return lbl, nil

}

// Rollback 支持重入，无需回滚
func (act BatchTaskTCloudCreateListenerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskTCloudCreateListenerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}

func (act BatchTaskTCloudCreateListenerAction) checkL4RuleExists(kt *kit.Kit, lbl *corelb.TCloudListener,
	req *hclb.TCloudListenerCreateReq) error {

	ruleReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("lbl_id", lbl.ID),
		),
		Page: core.NewDefaultBasePage(),
	}
	ruleResp, err := actcli.GetDataService().TCloud.LoadBalancer.ListUrlRule(kt, ruleReq)
	if err != nil {
		return fmt.Errorf("fail to query listener rule, err: %v", err)
	}
	if len(ruleResp.Details) == 0 {
		return fmt.Errorf("tcloud url rule not found for l4 listener, id: %s(%s)", lbl.CloudID, lbl.ID)
	}
	// 存在则判断是否和入参一致
	rule := ruleResp.Details[0]
	if len(req.Scheduler) > 0 && rule.Scheduler != req.Scheduler {
		return fmt.Errorf("scheduler mismatch, want: %s, db: %s", req.Scheduler, rule.Scheduler)
	}
	if req.SessionExpire > 0 && rule.SessionExpire != req.SessionExpire {
		return fmt.Errorf("session expire mismatch, want: %d, db: %d", req.SessionExpire, rule.SessionExpire)
	}
	if len(req.Scheduler) > 0 && rule.Scheduler != req.Scheduler {
		return fmt.Errorf("scheduler mismatch, want: %s, db: %s", req.Scheduler, rule.Scheduler)
	}
	if req.SessionType != nil && rule.SessionType != cvt.PtrToVal(req.SessionType) {
		return fmt.Errorf("session type mismatch, want: %s, db: %s", cvt.PtrToVal(req.SessionType), rule.SessionType)
	}
	return nil
}
