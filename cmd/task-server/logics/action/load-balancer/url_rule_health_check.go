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

package actionlb

import (
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
)

var _ action.Action = new(ListenerRuleUpdateHealthCheckAction)
var _ action.ParameterAction = new(ListenerRuleUpdateHealthCheckAction)

// ListenerRuleUpdateHealthCheckAction 将RS应用到监听器或者规则
type ListenerRuleUpdateHealthCheckAction struct{}

// ListenerRuleUpdateHealthCheckOption ...
type ListenerRuleUpdateHealthCheckOption struct {
	ListenerID  string                        `json:"listener_id" validate:"required"`
	CloudRuleID string                        `json:"cloud_rule_id" validate:"required"`
	Vendor      enumor.Vendor                 `json:"vendor" validate:"required"`
	HealthCheck *corelb.TCloudHealthCheckInfo `json:"health_check" validate:"required"`
}

// Validate validate option.
func (opt ListenerRuleUpdateHealthCheckOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act ListenerRuleUpdateHealthCheckAction) ParameterNew() (params any) {
	return new(ListenerRuleUpdateHealthCheckOption)
}

// Name return action name
func (act ListenerRuleUpdateHealthCheckAction) Name() enumor.ActionName {
	return enumor.ActionListenerRuleUpdateHealthCheck
}

// Run 将目标组中的RS绑定到监听器/规则中
func (act ListenerRuleUpdateHealthCheckAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*ListenerRuleUpdateHealthCheckOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if err := opt.Validate(); err != nil {
		logs.Errorf("ListenerRuleUpdateHealthCheckAction option validate failed, err: %v, params: %+v, rid: %s",
			err, opt, kt.Kit().Rid)
		return nil, err
	}

	switch opt.Vendor {
	case enumor.TCloud:
		if err := act.updateTCloudUrlRule(kt.Kit(), opt); err != nil {
			logs.Errorf("ListenerRuleUpdateHealthCheckAction update tcloud url rule failed, err: %v, params: %+v, rid: %s",
				err, opt, kt.Kit().Rid)
			return nil, err
		}
	}

	return nil, nil
}

// updateTCloudUrlRule 更新腾讯云CLB规则的健康检查配置
func (act ListenerRuleUpdateHealthCheckAction) updateTCloudUrlRule(kt *kit.Kit,
	opt *ListenerRuleUpdateHealthCheckOption) error {

	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(tools.RuleEqual("cloud_id", opt.CloudRuleID)),
		Page:   core.NewDefaultBasePage(),
	}
	resp, err := actcli.GetDataService().TCloud.LoadBalancer.ListUrlRule(kt, listReq)
	if err != nil {
		return err
	}
	if len(resp.Details) == 0 {
		logs.Errorf("ListenerRuleUpdateHealthCheckAction update url rule failed, "+
			"rule not found, params: %+v, rid: %s", opt, kt.Rid)
		return errf.New(errf.RecordNotFound, "url rule not found")
	}

	rule := resp.Details[0]
	req := &hclb.TCloudRuleUpdateReq{
		HealthCheck: opt.HealthCheck,
	}
	err = actcli.GetHCService().TCloud.Clb.UpdateUrlRule(kt, opt.ListenerID, rule.ID, req)
	if err != nil {
		logs.Errorf("ListenerRuleUpdateHealthCheckAction update url rule failed, err: %v, params: %+v, rid: %s")
		return err
	}
	return nil
}

// Rollback ...
func (act ListenerRuleUpdateHealthCheckAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- ListenerRuleUpdateHealthCheckAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
