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
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// --------------------------[同步TCloud负载均衡]-----------------------------

var _ action.Action = new(SyncTCloudLoadBalancerAction)
var _ action.ParameterAction = new(SyncTCloudLoadBalancerAction)

// SyncTCloudLoadBalancerAction 同步TCloud负载均衡
type SyncTCloudLoadBalancerAction struct{}

// SyncTCloudLoadBalancerOption ...
type SyncTCloudLoadBalancerOption struct {
	*sync.TCloudSyncReq `json:",inline" validate:"required"`
}

// Validate validate option.
func (opt SyncTCloudLoadBalancerOption) Validate() error {

	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act SyncTCloudLoadBalancerAction) ParameterNew() (params any) {
	return new(SyncTCloudLoadBalancerOption)
}

// Name return action name
func (act SyncTCloudLoadBalancerAction) Name() enumor.ActionName {
	return enumor.ActionSyncTCloudLoadBalancer
}

// Run 将目标组中的RS绑定到监听器/规则中
func (act SyncTCloudLoadBalancerAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*SyncTCloudLoadBalancerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	err := actcli.GetHCService().TCloud.Clb.SyncLoadBalancer(kt.Kit(), opt.TCloudSyncReq)
	if err != nil {
		logs.Errorf("fail to sync load balancer, err: %v, req: %v rid: %s", err, opt.TCloudSyncReq, kt.Kit().Rid)
		return nil, err
	}

	return nil, nil
}

// Rollback 支持重入，无需回滚
func (act SyncTCloudLoadBalancerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- SyncTCloudLoadBalancerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
