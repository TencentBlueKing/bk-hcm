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
	"fmt"

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
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"

	"github.com/tidwall/gjson"
)

// --------------------------[删除负载均衡]-----------------------------

var _ action.Action = new(DeleteLoadBalancerAction)
var _ action.ParameterAction = new(DeleteLoadBalancerAction)

// DeleteLoadBalancerAction 删除负载均衡
type DeleteLoadBalancerAction struct{}

// DeleteLoadBalancerOption ...
type DeleteLoadBalancerOption struct {
	Vendor                             enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	hcproto.BatchDeleteLoadBalancerReq `json:",inline"`
}

// MarshalJSON DeleteLoadBalancerOption.
func (opt DeleteLoadBalancerOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor                             enumor.Vendor `json:"vendor" validate:"required"`
			hcproto.BatchDeleteLoadBalancerReq `json:",inline"`
		}{
			Vendor:                     opt.Vendor,
			BatchDeleteLoadBalancerReq: opt.BatchDeleteLoadBalancerReq,
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return json.Marshal(req)
}

// UnmarshalJSON DeleteLoadBalancerOption.
func (opt *DeleteLoadBalancerOption) UnmarshalJSON(raw []byte) (err error) {
	opt.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())

	switch opt.Vendor {
	case enumor.TCloud:
		err = json.Unmarshal(raw, &opt.BatchDeleteLoadBalancerReq)
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate validate option.
func (opt DeleteLoadBalancerOption) Validate() error {

	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act DeleteLoadBalancerAction) ParameterNew() (params any) {
	return new(DeleteLoadBalancerOption)
}

// Name return action name
func (act DeleteLoadBalancerAction) Name() enumor.ActionName {
	return enumor.ActionDeleteLoadBalancer
}

// Run 将目标组中的RS绑定到监听器/规则中
func (act DeleteLoadBalancerAction) Run(kt run.ExecuteKit, params any) (any, error) {

	opt, ok := params.(*DeleteLoadBalancerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		err = actcli.GetHCService().TCloud.Clb.BatchDeleteLoadBalancer(kt.Kit(), &opt.BatchDeleteLoadBalancerReq)
		if err != nil {
			logs.Errorf("[%s] fail to delete tcloud load balancer, err: %v, opt: %+v rid: %s",
				opt.Vendor, err, opt.BatchDeleteLoadBalancerReq, kt.Kit().Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return nil, nil
}

// Rollback 无需回滚
func (act DeleteLoadBalancerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- DeleteLoadBalancerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}

// --------------------------[删除URLRule]-----------------------------

var _ action.Action = new(DeleteURLRuleAction)
var _ action.ParameterAction = new(DeleteURLRuleAction)

// DeleteURLRuleAction 删除负载均衡URLRule
type DeleteURLRuleAction struct{}

// DeleteURLRuleOption ...
type DeleteURLRuleOption struct {
	Vendor     enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	LbID       string        `json:"lb_id" validate:"required"`
	URLRuleIDs []string      `json:"url_rule_ids" validate:"required"`
}

// MarshalJSON DeleteURLRuleOption.
func (opt DeleteURLRuleOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor     enumor.Vendor `json:"vendor" validate:"required"`
			LbID       string        `json:"lb_id" validate:"required"`
			URLRuleIDs []string      `json:"url_rule_ids" validate:"required"`
		}{
			Vendor:     opt.Vendor,
			LbID:       opt.LbID,
			URLRuleIDs: opt.URLRuleIDs,
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
			LbID       string   `json:"lb_id" validate:"required"`
			URLRuleIDs []string `json:"url_rule_ids"`
		}{}
		err = json.Unmarshal(raw, &temp)
		opt.LbID = temp.LbID
		opt.URLRuleIDs = temp.URLRuleIDs
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate validate option.
func (opt DeleteURLRuleOption) Validate() error {

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
	switch opt.Vendor {
	case enumor.TCloud:
		// 删除云上url_rule（同时会删除规则表url_rule中信息，删除规则和目标组的绑定关系）
		lblUrlRuleMap := make(map[string][]string)
		lblUrlRuleSlice := make([]corelb.TCloudLbUrlRule, 0)
		urlRuleReq := &core.ListReq{
			Filter: tools.ExpressionAnd(tools.RuleIn("id", opt.URLRuleIDs)),
			Page:   core.NewDefaultBasePage(),
		}
		for {
			urlRuleResult, err := actcli.GetDataService().TCloud.LoadBalancer.ListUrlRule(kt.Kit(), urlRuleReq)
			if err != nil {
				logs.Errorf("fail to list tcloud url ule, err: %v, opt: %+v rid: %s",
					err, opt, kt.Kit().Rid)
				return nil, err
			}
			lblUrlRuleSlice = append(lblUrlRuleSlice, urlRuleResult.Details...)

			if uint(len(urlRuleResult.Details)) < core.DefaultMaxPageLimit {
				break
			}

			urlRuleReq.Page.Start += uint32(core.DefaultMaxPageLimit)
		}
		if lblUrlRuleSlice == nil || len(lblUrlRuleSlice) != len(opt.URLRuleIDs) {
			logs.Errorf("fail to list tcloud url ule, url rule not found, opt: %+v rid: %s", opt, kt.Kit().Rid)
			return nil, fmt.Errorf("fail to list tcloud url ule, url rule not found")
		}

		for _, urlRule := range lblUrlRuleSlice {
			if _, exists := lblUrlRuleMap[urlRule.LblID]; !exists {
				lblUrlRuleMap[urlRule.LblID] = make([]string, 0)
			}
			lblUrlRuleMap[urlRule.LblID] = append(lblUrlRuleMap[urlRule.LblID], urlRule.ID)
		}

		for lblID, urlRuleIDs := range lblUrlRuleMap {
			delReq := &hcproto.TCloudRuleDeleteByIDReq{RuleIDs: urlRuleIDs}
			err := actcli.GetHCService().TCloud.Clb.BatchDeleteUrlRule(kt.Kit(), lblID, delReq)
			if err != nil {
				logs.Errorf("fail to delete tcloud url ule, err: %v, opt: %+v rid: %s",
					err, opt, kt.Kit().Rid)
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return nil, nil
}

// Rollback 无需回滚
func (act DeleteURLRuleAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- DeleteURLRuleAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}

// --------------------------[删除Listener]-----------------------------

var _ action.Action = new(DeleteListenerAction)
var _ action.ParameterAction = new(DeleteListenerAction)

// DeleteListenerAction 删除负载均衡监听器
type DeleteListenerAction struct{}

// DeleteListenerOption ...
type DeleteListenerOption struct {
	Vendor      enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	LbID        string        `json:"lb_id" validate:"required"`
	ListenerIDs []string      `json:"url_rule_ids" validate:"required"`
}

// MarshalJSON DeleteListenerOption.
func (opt DeleteListenerOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor      enumor.Vendor `json:"vendor" validate:"required"`
			LbID        string        `json:"lb_id" validate:"required"`
			ListenerIDs []string      `json:"url_rule_ids" validate:"required"`
		}{
			Vendor:      opt.Vendor,
			LbID:        opt.LbID,
			ListenerIDs: opt.ListenerIDs,
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return json.Marshal(req)
}

// UnmarshalJSON DeleteListenerOption.
func (opt *DeleteListenerOption) UnmarshalJSON(raw []byte) (err error) {
	opt.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())

	switch opt.Vendor {
	case enumor.TCloud:
		temp := struct {
			LbID        string   `json:"lb_id" validate:"required"`
			ListenerIDs []string `json:"url_rule_ids"`
		}{}
		err = json.Unmarshal(raw, &temp)
		opt.LbID = temp.LbID
		opt.ListenerIDs = temp.ListenerIDs
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate validate option.
func (opt DeleteListenerOption) Validate() error {

	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act DeleteListenerAction) ParameterNew() (params any) {
	return new(DeleteListenerOption)
}

// Name return action name
func (act DeleteListenerAction) Name() enumor.ActionName {
	return enumor.ActionLoadBalancerDeleteListener
}

// Run 删除负载均衡器的Listener监听器
func (act DeleteListenerAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*DeleteListenerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	switch opt.Vendor {
	case enumor.TCloud:
		// 删除云上listener（会同步删除本地listener表和url_rule表中数据，删除规则和目标组的绑定关系）
		err := actcli.GetHCService().TCloud.Clb.DeleteListener(kt.Kit(), &core.BatchDeleteReq{IDs: opt.ListenerIDs})
		if err != nil {
			logs.Errorf("fail to delete tcloud listener, err: %v, opt: %+v rid: %s",
				err, opt, kt.Kit().Rid)
			return nil, err
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return nil, nil
}

// Rollback 无需回滚
func (act DeleteListenerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- DeleteListenerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
