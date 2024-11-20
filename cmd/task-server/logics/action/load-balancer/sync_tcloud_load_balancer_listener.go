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
	"hcm/pkg/api/hc-service/sync"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// --------------------------[同步TCloud负载均衡监听器]-----------------------------

var _ action.Action = new(SyncTCloudLoadBalancerListenerAction)
var _ action.ParameterAction = new(SyncTCloudLoadBalancerListenerAction)

// SyncTCloudLoadBalancerListenerAction 同步TCloud负载均衡监听器
type SyncTCloudLoadBalancerListenerAction struct{}

// SyncTCloudLoadBalancerListenerOption ...
type SyncTCloudLoadBalancerListenerOption struct {
	Vendor                      enumor.Vendor `json:"vendor" validate:"required"`
	*sync.TCloudListenerSyncReq `json:",inline" validate:"required"`
}

// Validate validate option.
func (opt SyncTCloudLoadBalancerListenerOption) Validate() error {
	switch opt.Vendor {
	case enumor.TCloud:
	default:
		return fmt.Errorf("unsupport vendor for sync load balancer listener: %s", opt.Vendor)
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act SyncTCloudLoadBalancerListenerAction) ParameterNew() (params any) {
	return new(SyncTCloudLoadBalancerListenerOption)
}

// Name return action name
func (act SyncTCloudLoadBalancerListenerAction) Name() enumor.ActionName {
	return enumor.SyncTCloudLoadBalancerListener
}

// Run 将目标组中的RS绑定到监听器/规则中
func (act SyncTCloudLoadBalancerListenerAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*SyncTCloudLoadBalancerListenerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	switch opt.Vendor {
	case enumor.TCloud:
		taskErr = actcli.GetHCService().TCloud.Clb.SyncLoadBalancerListener(kt.Kit(), opt.TCloudListenerSyncReq)
	default:
		return nil, fmt.Errorf("unsupport vendor for sync load balancer listener: %s", opt.Vendor)
	}
	if taskErr != nil {
		logs.Errorf("[%s] fail to sync load balancer listener, err: %v, req: %v rid: %s",
			opt.Vendor, taskErr, opt.TCloudListenerSyncReq, kt.Kit().Rid)
		return nil, taskErr
	}

	return nil, nil
}

// Rollback 支持重入，无需回滚
func (act SyncTCloudLoadBalancerListenerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- SyncTCloudLoadBalancerListenerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
