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

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// --------------------------[TCloud创建7层规则]-----------------------------

var _ action.Action = new(BatchTaskTCloudCreateL7RuleAction)
var _ action.ParameterAction = new(BatchTaskTCloudCreateL7RuleAction)

// BatchTaskTCloudCreateL7RuleAction TCloud创建7层规则
type BatchTaskTCloudCreateL7RuleAction struct{}

// BatchTaskTCloudCreateL7RuleOption ...
type BatchTaskTCloudCreateL7RuleOption struct {
	LoadBalancerID                 string   `json:"lb_id" validate:"required"`
	ListenerID                     string   `json:"listener_id" validate:"required"`
	ManagementDetailIDs            []string `json:"management_detail_ids" validate:"required,min=1,max=20"`
	*hclb.TCloudRuleBatchCreateReq `json:"inline" validate:"required,dive,required"`
}

// Validate validate option.
func (opt BatchTaskTCloudCreateL7RuleOption) Validate() error {
	if len(opt.ManagementDetailIDs) != len(opt.TCloudRuleBatchCreateReq.Rules) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and rules length not match: %d! = %d",
			len(opt.ManagementDetailIDs), len(opt.TCloudRuleBatchCreateReq.Rules))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskTCloudCreateL7RuleAction) ParameterNew() (params any) {
	return new(BatchTaskTCloudCreateL7RuleOption)
}

// Name return action name
func (act BatchTaskTCloudCreateL7RuleAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskTCloudCreateL7Rule
}

// Run 创建监听器
func (act BatchTaskTCloudCreateL7RuleAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskTCloudCreateL7RuleOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	// 查询 负载均衡、监听器是否正确
	lb, _, err := getListenerWithLb(kt.Kit(), opt.ListenerID)
	if err != nil {
		logs.Errorf("fail to get listener with lb, err: %v, listner: %s, rid: %s", err, opt.ListenerID, kt.Kit().Rid)
		return nil, err
	}
	if lb.ID != opt.LoadBalancerID {
		return nil, errf.Newf(errf.InvalidParameter, "loadbalancer id mismatch, want: %s, got: %s",
			opt.LoadBalancerID, lb.ID)
	}

	detailList, err := listTaskDetail(kt.Kit(), opt.ManagementDetailIDs)
	if err != nil {
		return fmt.Sprintf("task detail query failed"), err
	}
	for _, detail := range detailList {
		if detail.State == enumor.TaskDetailCancel {
			// 任务被取消，跳过该批次
			return fmt.Sprintf("task detail %s canceled", detail.ID), nil
		}
		if detail.State != enumor.TaskDetailInit {
			return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
				detail.ID, detail.State)
		}
	}

	nonExistsReq, err := act.skipExistsRule(kt.Kit(), opt)
	if err != nil {
		return nil, err
	}
	if len(nonExistsReq.Rules) == 0 {
		// 已存在跳过
		logs.Infof("all rule exists, skip, rid: %s", kt.Kit().Rid)
		return "all rule exists", nil
	}

	// 进入创建
	// 更新任务状态为 running
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning, nil); err != nil {
		return fmt.Sprintf("fail to update detail to running"), err
	}

	defer func() {
		// 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if taskErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, targetState, taskErr)
		if err != nil {
			logs.Errorf("fail to set detail to %s after cloud operation finished, err: %v, rid: %s",
				targetState, err, kt.Kit().Rid)
		}
	}()

	lblResp, err := actcli.GetHCService().TCloud.Clb.BatchCreateUrlRule(kt.Kit(), lb.ID, nonExistsReq)
	if err != nil {
		logs.Errorf("fail to call hc to create tcloud l7 rules, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, err
	}
	// all success
	return lblResp, nil
}

// 查询规则是否存在，返回不存在的规则入参。如果存在且参数一样跳过，如果存在但不符合入参则报错。
func (act BatchTaskTCloudCreateL7RuleAction) skipExistsRule(kt *kit.Kit, opt *BatchTaskTCloudCreateL7RuleOption) (
	nonExistReq *hclb.TCloudRuleBatchCreateReq, err error) {

	nonExistReq = &hclb.TCloudRuleBatchCreateReq{
		Rules: make([]hclb.TCloudRuleCreate, 0, len(opt.Rules)),
	}
	for i := range opt.Rules {
		reqRule := opt.Rules[i]

		// 查询是否已经存在对应规则
		listRuleReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("lb_id", opt.LoadBalancerID),
				tools.RuleEqual("lbl_id", opt.ListenerID),
				tools.RuleEqual("domain", reqRule.Domains[0]),
				tools.RuleEqual("url", reqRule.Url),
			),
			Page: core.NewDefaultBasePage(),
		}
		ruleResp, err := actcli.GetDataService().TCloud.LoadBalancer.ListUrlRule(kt, listRuleReq)
		if err != nil {
			return nil, fmt.Errorf("fail to query listener, err: %v", err)
		}

		if len(ruleResp.Details) == 0 {
			// 不存在直接创建
			nonExistReq.Rules = append(nonExistReq.Rules, reqRule)
			continue
		}
		dbRule := ruleResp.Details[0]
		if err := act.checkRuleMatch(reqRule, dbRule); err != nil {
			return nil, fmt.Errorf("check rule exists failed, err: %v, idx: %d", err, i)
		}
		// 存在但是不冲突也加入创建
		nonExistReq.Rules = append(nonExistReq.Rules, reqRule)
	}

	return nonExistReq, nil
}

func (act BatchTaskTCloudCreateL7RuleAction) checkRuleMatch(req hclb.TCloudRuleCreate, db corelb.TCloudLbUrlRule) (
	err error) {

	if req.SessionExpireTime != nil && db.SessionExpire != cvt.PtrToVal(req.SessionExpireTime) {
		return fmt.Errorf("url rule(%s) session expire time mismatch, want: %d, got: %d",
			req.Url, cvt.PtrToVal(req.SessionExpireTime), db.SessionExpire)
	}
	if req.Scheduler != nil && db.Scheduler != cvt.PtrToVal(req.Scheduler) {
		return fmt.Errorf("url rule(%s) scheduler mismatch, want: %s, got: %s",
			req.Url, cvt.PtrToVal(req.Scheduler), db.Scheduler)
	}
	if len(req.Domains) > 0 && req.Domains[0] != db.Domain {
		return fmt.Errorf("url rule(%s) domain mismatch, want: %+v, got: %+v",
			req.Url, req.Domains, db.Domain)
	}
	if req.HealthCheck != nil && isHealthCheckChange(req.HealthCheck, db.HealthCheck, true) {
		return fmt.Errorf("url rule(%s) health check mismatch, want: %+v, got: %+v",
			req.Url, cvt.PtrToVal(req.HealthCheck), db.HealthCheck)
	}
	if req.Certificates != nil && isListenerCertChange(req.Certificates, db.Certificate) {
		return fmt.Errorf("url rule(%s) certificates mismatch, want: %+v, got: %+v",
			req.Url, cvt.PtrToVal(req.Certificates), db.Certificate)
	}

	return nil
}

// Rollback 支持重入，无需回滚
func (act BatchTaskTCloudCreateL7RuleAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskTCloudCreateL7RuleAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
