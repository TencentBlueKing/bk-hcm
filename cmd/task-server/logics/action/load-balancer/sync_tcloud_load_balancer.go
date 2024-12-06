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

// --------------------------[同步TCloud负载均衡]-----------------------------

var _ action.Action = new(SyncTCloudLoadBalancerAction)
var _ action.ParameterAction = new(SyncTCloudLoadBalancerAction)

// SyncTCloudLoadBalancerAction 同步TCloud负载均衡
type SyncTCloudLoadBalancerAction struct{}

// SyncTCloudLoadBalancerOption ...
type SyncTCloudLoadBalancerOption struct {
	Vendor              enumor.Vendor `json:"vendor" validate:"required"`
	*sync.TCloudSyncReq `json:",inline" validate:"required"`
}

// Validate validate option.
func (opt SyncTCloudLoadBalancerOption) Validate() error {
	switch opt.Vendor {
	case enumor.TCloud:
	default:
		return fmt.Errorf("unsupport vendor for sync load balancer: %s", opt.Vendor)
	}
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
func (act SyncTCloudLoadBalancerAction) Run(et run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*SyncTCloudLoadBalancerOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	// 这里如果不重设rid会导致rid长度超长
	// TODO 改造rid模式，增加独立的spanid字段标记层次关系
	kt := et.KitWithNewRid()
	logs.Infof("reset rid to %s for sync load balancer, old rid: %s", kt.Rid, et.Kit().Rid)
	switch opt.Vendor {
	case enumor.TCloud:
		taskErr = actcli.GetHCService().TCloud.Clb.SyncLoadBalancer(kt, opt.TCloudSyncReq)
	default:
		return nil, fmt.Errorf("unsupport vendor for sync load balancer: %s", opt.Vendor)
	}
	if taskErr != nil {
		logs.Errorf("[%s] fail to sync load balancer, err: %v, req: %v rid: %s",
			opt.Vendor, taskErr, opt.TCloudSyncReq, kt.Rid)
		return nil, taskErr
	}

	return nil, nil
}

// Rollback 支持重入，无需回滚
func (act SyncTCloudLoadBalancerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- SyncTCloudLoadBalancerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
