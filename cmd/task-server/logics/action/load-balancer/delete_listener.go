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
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// --------------------------[删除Listener]-----------------------------

var _ action.Action = new(DeleteListenerAction)
var _ action.ParameterAction = new(DeleteListenerAction)

// DeleteListenerAction 删除负载均衡监听器
// Deprecated 没有被使用，后续版本将移除该实现
type DeleteListenerAction struct{}

// DeleteListenerOption ...
type DeleteListenerOption struct {
	Vendor              enumor.Vendor `json:"vendor,omitempty" validate:"required"`
	LbID                string        `json:"lb_id" validate:"required"`
	ListenerIDs         []string      `json:"url_rule_ids" validate:"required,max=20,min=1"`
	ManagementDetailIDs []string      `json:"management_detail_ids" validate:"required,min=1"`
}

// MarshalJSON DeleteListenerOption.
func (opt DeleteListenerOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor              enumor.Vendor `json:"vendor" validate:"required"`
			LbID                string        `json:"lb_id" validate:"required"`
			ListenerIDs         []string      `json:"url_rule_ids" validate:"required"`
			ManagementDetailIDs []string      `json:"management_detail_ids" validate:"required,min=1"`
		}{
			Vendor:              opt.Vendor,
			LbID:                opt.LbID,
			ListenerIDs:         opt.ListenerIDs,
			ManagementDetailIDs: opt.ManagementDetailIDs,
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
			LbID                string   `json:"lb_id" validate:"required"`
			ListenerIDs         []string `json:"url_rule_ids"`
			ManagementDetailIDs []string `json:"management_detail_ids" validate:"required,min=1"`
		}{}
		err = json.Unmarshal(raw, &temp)
		opt.LbID = temp.LbID
		opt.ListenerIDs = temp.ListenerIDs
		opt.ManagementDetailIDs = temp.ManagementDetailIDs
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate validate option.
func (opt DeleteListenerOption) Validate() error {
	if len(opt.ListenerIDs) != len(opt.ManagementDetailIDs) {
		return errf.New(errf.InvalidParameter, "management_detail_ids and listener_ids must have the same length")
	}
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
	if err := opt.Validate(); err != nil {
		logs.Errorf("fail to validate delete listener option, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	if reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs); err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return reason, err
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	var err error
	taskDetailState := enumor.TaskDetailSuccess
	defer func() {
		// 更新任务状态
		if err := batchUpdateTaskDetailResultState(kt.Kit(), opt.ManagementDetailIDs, taskDetailState,
			nil, err); err != nil {
			logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		}
	}()

	switch opt.Vendor {
	case enumor.TCloud:
		// 删除云上listener（会同步删除本地listener表和url_rule表中数据，删除规则和目标组的绑定关系）
		err = actcli.GetHCService().TCloud.Clb.DeleteListener(kt.Kit(),
			&core.BatchDeleteReq{IDs: opt.ListenerIDs})
		if err != nil {
			taskDetailState = enumor.TaskDetailFailed
			logs.Errorf("fail to delete tcloud listener, err: %v, opt: %+v rid: %s",
				err, opt, kt.Kit().Rid)
			return nil, err
		}
	default:
		taskDetailState = enumor.TaskDetailFailed
		err = fmt.Errorf("vendor: %s not support for delete listener action", opt.Vendor)
		return nil, err
	}

	return nil, nil
}

// Rollback 无需回滚
func (act DeleteListenerAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- DeleteListenerAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
