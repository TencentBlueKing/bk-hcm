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
	hcproto "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
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
	Vendor                                   enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	hcproto.TCloudBatchDeleteLoadbalancerReq `json:",inline"`
}

// MarshalJSON DeleteLoadBalancerOption.
func (opt DeleteLoadBalancerOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor                                   enumor.Vendor `json:"vendor" validate:"required"`
			hcproto.TCloudBatchDeleteLoadbalancerReq `json:",inline"`
		}{
			Vendor:                           opt.Vendor,
			TCloudBatchDeleteLoadbalancerReq: opt.TCloudBatchDeleteLoadbalancerReq,
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
		err = json.Unmarshal(raw, &opt.TCloudBatchDeleteLoadbalancerReq)
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
		err = actcli.GetHCService().TCloud.Clb.BatchDeleteLoadBalancer(kt.Kit(), &opt.TCloudBatchDeleteLoadbalancerReq)
		if err != nil {
			logs.Errorf("fail to delete tcloud load balancer, err: %v, opt: %+v rid: %s",
				err, opt.TCloudBatchDeleteLoadbalancerReq, kt.Kit().Rid)
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
