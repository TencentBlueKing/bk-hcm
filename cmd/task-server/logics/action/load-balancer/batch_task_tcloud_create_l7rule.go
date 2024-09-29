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
	"errors"
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
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and rules length not match: %d != %d",
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
	// detail 状态检查
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
	// 规则检查
	ruleCheckResult, err := act.checkExistsRule(kt.Kit(), opt)
	if err != nil {
		logs.Errorf("fail to check exists rule, err: %v, rid: %s", err, kt.Kit().Rid)
		return nil, err
	}
	// 1. 有错误，写入错误信息
	for i := range ruleCheckResult.Mismatch {
		mismatch := ruleCheckResult.Mismatch[i]
		ids := []string{mismatch.DetailID}
		err := batchUpdateTaskDetailResultState(kt.Kit(), ids, enumor.TaskDetailFailed, nil, mismatch.Error)
		if err != nil {
			logs.Errorf("fail to set detail to %s, err: %v, detail: %s, rid: %s",
				enumor.TaskDetailFailed, err, mismatch.DetailID, kt.Kit().Rid)
			// 继续尝试处理其他情况
		}
		continue
	}
	// 2. 已存在直接成功，写入已存在的id
	for i := range ruleCheckResult.Exists {
		exists := ruleCheckResult.Exists[i]
		ids := []string{exists.DetailID}
		createResult := &core.CloudCreateResult{
			ID:      exists.Rule.ID,
			CloudID: exists.Rule.CloudID,
		}
		reason := errors.New("rule exists, skip")
		err := batchUpdateTaskDetailResultState(kt.Kit(), ids, enumor.TaskDetailSuccess, createResult, reason)
		if err != nil {
			logs.Errorf("fail to set detail to success, err: %v, detail: %s, rid: %s",
				err, exists.DetailID, kt.Kit().Rid)
			// 继续尝试处理其他情况
		}
		continue
	}
	// 3. 创建不存在的规则
	return act.createNonExists(kt.Kit(), ruleCheckResult.NonExists, opt, lb)
}

func (act BatchTaskTCloudCreateL7RuleAction) createNonExists(kt *kit.Kit, nonExists []RuleCheckInfo,
	opt *BatchTaskTCloudCreateL7RuleOption, lb *corelb.BaseLoadBalancer) (any, error) {

	if len(nonExists) == 0 {
		return "no rule should be created", nil
	}

	nonExistIds := make([]string, len(nonExists))
	ruleCreateReq := new(hclb.TCloudRuleBatchCreateReq)
	for i := range nonExists {
		nonExistIds[i] = nonExists[i].DetailID
		ruleCreateReq.Rules = append(ruleCreateReq.Rules, opt.Rules[nonExists[i].Index])
	}

	// 更新任务状态为 running
	if err := batchUpdateTaskDetailState(kt, nonExistIds, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update detail to running, err: %v, detail ids: %s, rid: %s",
			err, nonExistIds, kt.Rid)
		return fmt.Sprintf("fail to update detail state to running"), err
	}

	lblResp, createErr := actcli.GetHCService().TCloud.Clb.BatchCreateUrlRule(kt, lb.ID, ruleCreateReq)
	// 更新为失败
	if createErr != nil {
		logs.Errorf("fail to call hc to create tcloud l7 rules, err: %v, req: %+v, rid: %s",
			createErr, ruleCreateReq, kt.Rid)
		err := batchUpdateTaskDetailResultState(kt, nonExistIds, enumor.TaskDetailFailed, lblResp, createErr)
		if err != nil {
			logs.Errorf("fail to set detail to failed after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
		}
		return lblResp, err
	}
	// 更新为成功
	for i := range nonExists {
		detailID := []string{nonExists[i].DetailID}
		var ret = &core.CloudCreateResult{CloudID: lblResp.SuccessCloudIDs[i]}
		err := batchUpdateTaskDetailResultState(kt, detailID, enumor.TaskDetailSuccess, ret, nil)
		if err != nil {
			logs.Errorf("fail to set detail to success after cloud operation, err: %v, rid: %s",
				err, kt.Rid)
			// 继续尝试更新其他结果
		}
	}
	return lblResp, nil
}

// RuleCheckInfo ...
type RuleCheckInfo struct {
	Index    int
	DetailID string
	// rule queried from db
	Rule *corelb.TCloudLbUrlRule
	// error will be set if it doesn't match
	Error error
}

// GetDetailID ...
func (r RuleCheckInfo) GetDetailID() string {
	return r.DetailID
}

// RuleCheckSummary ...
type RuleCheckSummary struct {
	NonExists []RuleCheckInfo
	Exists    []RuleCheckInfo
	Mismatch  []RuleCheckInfo
}

// 查询规则是否存在，返回不存在的规则入参。如果存在且参数一样跳过，如果存在但不符合入参则报错。
func (act BatchTaskTCloudCreateL7RuleAction) checkExistsRule(kt *kit.Kit, opt *BatchTaskTCloudCreateL7RuleOption) (
	checkResult *RuleCheckSummary, err error) {

	checkResult = new(RuleCheckSummary)
	for i := range opt.Rules {
		result := RuleCheckInfo{Index: i, DetailID: opt.ManagementDetailIDs[i]}
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
			logs.Errorf("query url rule failed, err: %v, req: %+v, rid: %s", err, listRuleReq, kt.Rid)
			return nil, fmt.Errorf("fail to query url rule, err: %v", err)
		}

		if len(ruleResp.Details) == 0 {
			// 不存在直接创建
			checkResult.NonExists = append(checkResult.NonExists, result)
			continue
		}
		result.Rule = &ruleResp.Details[0]
		if err := act.checkRuleMatch(reqRule, result.Rule); err != nil {
			result.Error = fmt.Errorf("rule exist but %v", err)
			checkResult.Mismatch = append(checkResult.Mismatch, result)
			continue
		}
		checkResult.Exists = append(checkResult.Exists, result)
	}
	return checkResult, nil
}

func (act BatchTaskTCloudCreateL7RuleAction) checkRuleMatch(req hclb.TCloudRuleCreate, db *corelb.TCloudLbUrlRule) (
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
