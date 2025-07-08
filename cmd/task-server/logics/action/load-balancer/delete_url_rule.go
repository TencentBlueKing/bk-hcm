/*
 *
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
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/slice"
)

// --------------------------[删除URLRule]-----------------------------

var _ action.Action = new(DeleteURLRuleAction)
var _ action.ParameterAction = new(DeleteURLRuleAction)

// DeleteURLRuleAction 删除负载均衡URLRule
// Deprecated 没有被使用，后续版本将移除该实现
type DeleteURLRuleAction struct{}

// DeleteURLRuleOption ...
type DeleteURLRuleOption struct {
	Vendor     enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	LbID       string        `json:"lb_id" validate:"required"`
	URLRuleIDs []string      `json:"url_rule_ids" validate:"required"`
	// ManagementDetailIDs 需要和URLRuleIDs顺序一致
	ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1"`
}

// MarshalJSON DeleteURLRuleOption.
func (opt DeleteURLRuleOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor              enumor.Vendor `json:"vendor" validate:"required"`
			LbID                string        `json:"lb_id" validate:"required"`
			URLRuleIDs          []string      `json:"url_rule_ids" validate:"required"`
			ManagementDetailIDs []string      `json:"management_detail_ids" validate:"required,min=1"`
		}{
			Vendor:              opt.Vendor,
			LbID:                opt.LbID,
			URLRuleIDs:          opt.URLRuleIDs,
			ManagementDetailIDs: opt.ManagementDetailIDs,
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return json.Marshal(req)
}

// UnmarshalJSON DeleteURLRuleOption.
func (opt *DeleteURLRuleOption) UnmarshalJSON(raw []byte) (err error) {
	opt.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())

	switch opt.Vendor {
	case enumor.TCloud:
		temp := struct {
			LbID                string   `json:"lb_id" validate:"required"`
			URLRuleIDs          []string `json:"url_rule_ids"`
			ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1"`
		}{}
		err = json.Unmarshal(raw, &temp)
		opt.LbID = temp.LbID
		opt.URLRuleIDs = temp.URLRuleIDs
		opt.ManagementDetailIDs = temp.ManagementDetailIDs
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate validate option.
func (opt DeleteURLRuleOption) Validate() error {
	if len(opt.ManagementDetailIDs) != len(opt.URLRuleIDs) {
		return errf.New(errf.InvalidParameter, "management_detail_ids and url_rule_ids must have the same length")
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act DeleteURLRuleAction) ParameterNew() (params any) {
	return new(DeleteURLRuleOption)
}

// Name return action name
func (act DeleteURLRuleAction) Name() enumor.ActionName {
	return enumor.ActionLoadBalancerDeleteUrlRule
}

// Run 删除负载均衡器的URLRule规则
func (act DeleteURLRuleAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*DeleteURLRuleOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	err := opt.Validate()
	if err != nil {
		logs.Errorf("fail to validate delete url rule option, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	// detail 状态检查
	if reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs); err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return reason, err
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	urlRuleToDetailMap := make(map[string]string, len(opt.URLRuleIDs))
	for i, urlRuleID := range opt.URLRuleIDs {
		urlRuleToDetailMap[urlRuleID] = opt.ManagementDetailIDs[i]
	}

	switch opt.Vendor {
	case enumor.TCloud:
		err := deleteTCloudUrlRule(kt.Kit(), opt, urlRuleToDetailMap)
		if err != nil {
			logs.Errorf("fail to delete tcloud url rule, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return nil, nil
}

func deleteTCloudUrlRule(kt *kit.Kit, opt *DeleteURLRuleOption, urlRuleToDetailMap map[string]string) error {
	// 删除云上url_rule（同时会删除规则表url_rule中信息，删除规则和目标组的绑定关系）
	lblUrlRuleMap := make(map[string][]string)
	lblUrlRuleSlice, listErr := listTCloudUrlRule(kt, opt.URLRuleIDs)
	if listErr != nil {
		logs.Errorf("fail to list tcloud url rule, err: %v, opt: %+v rid: %s", listErr, opt, kt.Rid)
		err := batchUpdateTaskDetailResultState(kt, opt.ManagementDetailIDs, enumor.TaskDetailFailed, nil,
			listErr)
		if err != nil {
			logs.Errorf("fail to update task detail state to failed, err: %v, opt: %+v rid: %s", err, opt, kt.Rid)
			return err
		}
		return listErr
	}
	if len(lblUrlRuleSlice) != len(opt.URLRuleIDs) {
		// 如果查询到的URLRule数量和传入的数量不一致，说明有些URLRule已经被删除或不存在
		// 需要更新任务详情状态为失败
		if err := updateTaskDetailStateByNotFoundUrlRule(kt, lblUrlRuleSlice, urlRuleToDetailMap); err != nil {
			return err
		}
	}

	for _, urlRule := range lblUrlRuleSlice {
		if _, exists := lblUrlRuleMap[urlRule.LblID]; !exists {
			lblUrlRuleMap[urlRule.LblID] = make([]string, 0)
		}
		lblUrlRuleMap[urlRule.LblID] = append(lblUrlRuleMap[urlRule.LblID], urlRule.ID)
	}

	for lblID, urlRuleIDs := range lblUrlRuleMap {
		delReq := &hcproto.TCloudRuleDeleteByIDReq{RuleIDs: urlRuleIDs}
		updateErr := actcli.GetHCService().TCloud.Clb.BatchDeleteUrlRule(kt, lblID, delReq)
		taskDetailState := enumor.TaskDetailSuccess
		if updateErr != nil {
			logs.Errorf("fail to delete tcloud url ule, err: %v, opt: %+v rid: %s",
				updateErr, opt, kt.Rid)
			taskDetailState = enumor.TaskDetailFailed
		}
		taskDetailIDs := make([]string, 0, len(urlRuleIDs))
		for _, urlRuleID := range urlRuleIDs {
			if detailID, exists := urlRuleToDetailMap[urlRuleID]; exists {
				taskDetailIDs = append(taskDetailIDs, detailID)
			}
		}
		err := batchUpdateTaskDetailResultState(kt, taskDetailIDs, taskDetailState, nil, updateErr)
		if err != nil {
			logs.Errorf("fail to update task detail state to success, err: %v, opt: %+v rid: %s",
				err, opt, kt.Rid)
			return err
		}
	}
	return nil
}

// updateTaskDetailStateByNotFoundUrlRule 未在db中查到对应的urlRule, 对应的taskDetail状态
func updateTaskDetailStateByNotFoundUrlRule(kt *kit.Kit, lblUrlRuleSlice []corelb.TCloudLbUrlRule,
	urlRuleToDetailMap map[string]string) error {

	ruleMap := make(map[string]struct{}, len(lblUrlRuleSlice))
	for _, urlRule := range lblUrlRuleSlice {
		ruleMap[urlRule.ID] = struct{}{}
	}
	for urlRuleID, detailID := range urlRuleToDetailMap {
		if _, exists := ruleMap[urlRuleID]; !exists {
			err := batchUpdateTaskDetailResultState(kt, []string{detailID},
				enumor.TaskDetailFailed,
				nil, fmt.Errorf("tcloud url rule %s not found", urlRuleID))
			if err != nil {
				logs.Errorf("fail to update task detail state to failed, err: %v, urlRuleID: %s, detailID: %s, rid: %s",
					err, urlRuleID, detailID, kt.Rid)
				return err
			}
		}
	}
	return nil
}

func listTCloudUrlRule(kt *kit.Kit, urlRuleIDs []string) ([]corelb.TCloudLbUrlRule, error) {
	lblUrlRuleSlice := make([]corelb.TCloudLbUrlRule, 0)
	for _, parts := range slice.Split(urlRuleIDs, int(core.DefaultMaxPageLimit)) {
		urlRuleReq := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", parts)),
			Page: &core.BasePage{
				Count: false,
				Start: 0,
				Limit: core.DefaultMaxPageLimit,
				Sort:  "id",
			},
		}
		urlRuleResult, err := actcli.GetDataService().TCloud.LoadBalancer.ListUrlRule(kt, urlRuleReq)
		if err != nil {
			logs.Errorf("fail to list tcloud url rule, err: %v, opt: %+v rid: %s", err, urlRuleReq, kt.Rid)
			return nil, err
		}
		lblUrlRuleSlice = append(lblUrlRuleSlice, urlRuleResult.Details...)
	}
	return lblUrlRuleSlice, nil

}

// Rollback 无需回滚
func (act DeleteURLRuleAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- DeleteURLRuleAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
