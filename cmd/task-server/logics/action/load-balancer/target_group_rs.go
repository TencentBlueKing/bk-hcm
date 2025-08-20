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
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"

	"github.com/tidwall/gjson"
)

// SaveRsCloudIDKey 批量操作RS成功的监听器ID列表
const SaveRsCloudIDKey = "batch_rs_cloud_ids"

// --------------------------[批量添加RS到目标组]-----------------------------

var _ action.Action = new(AddTargetToGroupAction)
var _ action.ParameterAction = new(AddTargetToGroupAction)

// AddTargetToGroupAction define add rs action.
type AddTargetToGroupAction struct{}

// OperateRsOption define operate rs option.
type OperateRsOption struct {
	Vendor                           enumor.Vendor `json:"vendor" validate:"required"`
	ManagementDetailIDs              []string      `json:"management_detail_ids" validate:"required,min=1"`
	hclb.TCloudBatchOperateTargetReq `json:",inline"`
}

// MarshalJSON marshal json.
func (opt OperateRsOption) MarshalJSON() ([]byte, error) {
	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor                           enumor.Vendor `json:"vendor" validate:"required"`
			hclb.TCloudBatchOperateTargetReq `json:",inline"`
			ManagementDetailIDs              []string `json:"management_detail_ids" validate:"required,min=1"`
		}{
			Vendor:                      opt.Vendor,
			TCloudBatchOperateTargetReq: opt.TCloudBatchOperateTargetReq,
			ManagementDetailIDs:         opt.ManagementDetailIDs,
		}

	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return json.Marshal(req)
}

// UnmarshalJSON unmarshal json.
func (opt *OperateRsOption) UnmarshalJSON(raw []byte) (err error) {
	opt.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())
	// raw byte to []string array
	err = json.Unmarshal([]byte(gjson.GetBytes(raw, "management_detail_ids").String()), &opt.ManagementDetailIDs)

	switch opt.Vendor {
	case enumor.TCloud:
		err = json.Unmarshal(raw, &opt.TCloudBatchOperateTargetReq)
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate validate option.
func (opt OperateRsOption) Validate() error {
	if err := opt.Vendor.Validate(); err != nil {
		return err
	}

	var req validator.Interface
	switch opt.Vendor {
	case enumor.TCloud:
		req = &opt.TCloudBatchOperateTargetReq
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	if err := req.Validate(); err != nil {
		return err
	}

	return nil
}

// ParameterNew return request params.
func (act AddTargetToGroupAction) ParameterNew() (params interface{}) {
	return new(OperateRsOption)
}

// Name return action name
func (act AddTargetToGroupAction) Name() enumor.ActionName {
	return enumor.ActionTargetGroupAddRS
}

// Run add target.
func (act AddTargetToGroupAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*OperateRsOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs); err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return reason, err
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	var result *hclb.BatchCreateResult
	var err error
	taskDetailState := enumor.TaskDetailSuccess
	defer func() {
		// 更新任务状态
		if err := batchUpdateTaskDetailResultState(kt.Kit(), opt.ManagementDetailIDs, taskDetailState,
			result, err); err != nil {
			logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		}
	}()
	switch opt.Vendor {
	case enumor.TCloud:
		result, err = actcli.GetHCService().TCloud.Clb.BatchAddRs(
			kt.Kit(), opt.TargetGroupID, &opt.TCloudBatchOperateTargetReq)
	default:
		taskDetailState = enumor.TaskDetailFailed
		err = fmt.Errorf("vendor: %s not support", opt.Vendor)
		return nil, err
	}
	if err != nil {
		taskDetailState = enumor.TaskDetailFailed
		logs.Errorf("batch add rs failed, err: %v, result: %+v, rid: %s", err, result, kt.Kit().Rid)
		return result, err
	}

	if len(result.FailedCloudIDs) != 0 {
		taskDetailState = enumor.TaskDetailFailed
		return result, errf.Newf(errf.PartialFailed, "batch add rs rs partially failed, failCloudIDs: %v",
			result.FailedCloudIDs)
	}

	if err = kt.ShareData().AppendIDs(kt.Kit(), SaveRsCloudIDKey, result.SuccessCloudIDs...); err != nil {
		taskDetailState = enumor.TaskDetailFailed
		logs.Errorf("share data appendIDs failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}

// Rollback 批量添加RS失败时的回滚Action，此处不需要回滚处理
func (act AddTargetToGroupAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- AddTargetToGroupAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}

// --------------------------[批量移除RS]-----------------------------

var _ action.Action = new(RemoveTargetAction)
var _ action.ParameterAction = new(RemoveTargetAction)

// RemoveTargetAction define remove rs action.
type RemoveTargetAction struct{}

// ParameterNew return request params.
func (act RemoveTargetAction) ParameterNew() (params interface{}) {
	return new(OperateRsOption)
}

// Name return action name
func (act RemoveTargetAction) Name() enumor.ActionName {
	return enumor.ActionTargetGroupRemoveRS
}

// Run remove rs.
func (act RemoveTargetAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*OperateRsOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs); err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return reason, err
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}
	var result *hclb.BatchCreateResult
	var err error
	taskDetailState := enumor.TaskDetailSuccess
	defer func() {
		// 更新任务状态
		if err := batchUpdateTaskDetailResultState(kt.Kit(), opt.ManagementDetailIDs, taskDetailState,
			result, err); err != nil {
			logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		}
	}()
	switch opt.Vendor {
	case enumor.TCloud:
		result, err = actcli.GetHCService().TCloud.Clb.BatchRemoveTarget(
			kt.Kit(), opt.TargetGroupID, &opt.TCloudBatchOperateTargetReq)
	default:
		taskDetailState = enumor.TaskDetailFailed
		err = fmt.Errorf("vendor: %s not support for remove target", opt.Vendor)
		return nil, err
	}
	if err != nil {
		taskDetailState = enumor.TaskDetailFailed
		logs.Errorf("batch remove rs failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}

// Rollback 批量移除RS失败时的回滚Action，此处不需要回滚处理
func (act RemoveTargetAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- RemoveTargetAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}

// --------------------------[批量修改RS端口]-----------------------------

var _ action.Action = new(ModifyTargetPortAction)
var _ action.ParameterAction = new(ModifyTargetPortAction)

// ModifyTargetPortAction define modify rs port action.
type ModifyTargetPortAction struct{}

// ParameterNew return request params.
func (act ModifyTargetPortAction) ParameterNew() (params interface{}) {
	return new(OperateRsOption)
}

// Name return action name
func (act ModifyTargetPortAction) Name() enumor.ActionName {
	return enumor.ActionTargetGroupModifyPort
}

// Run modify target port.
func (act ModifyTargetPortAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*OperateRsOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}
	if reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs); err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return reason, err
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	var result *hclb.BatchCreateResult
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
		err = actcli.GetHCService().TCloud.Clb.BatchModifyTargetPort(
			kt.Kit(), opt.TargetGroupID, &opt.TCloudBatchOperateTargetReq)
	default:
		taskDetailState = enumor.TaskDetailFailed
		err = fmt.Errorf("vendor: %s not support for modify target port", opt.Vendor)
		return nil, err
	}
	if err != nil {
		taskDetailState = enumor.TaskDetailFailed
		logs.Errorf("batch modify target port failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}

// Rollback 批量修改RS端口失败时的回滚Action，此处不需要回滚处理
func (act ModifyTargetPortAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- ModifyTargetPortAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}

// --------------------------[批量修改RS权重]-----------------------------

var _ action.Action = new(ModifyTargetWeightAction)
var _ action.ParameterAction = new(ModifyTargetWeightAction)

// ModifyTargetWeightAction define modify target weight action.
type ModifyTargetWeightAction struct{}

// ParameterNew return request params.
func (act ModifyTargetWeightAction) ParameterNew() (params interface{}) {
	return new(OperateRsOption)
}

// Name return action name
func (act ModifyTargetWeightAction) Name() enumor.ActionName {
	return enumor.ActionTargetGroupModifyWeight
}

// Run modify target port.
func (act ModifyTargetWeightAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*OperateRsOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if reason, err := validateDetailListStatus(kt.Kit(), opt.ManagementDetailIDs); err != nil {
		logs.Errorf("validate detail list status failed, err: %v, reason: %s, rid: %s", err, reason, kt.Kit().Rid)
		return reason, err
	}
	if err := batchUpdateTaskDetailState(kt.Kit(), opt.ManagementDetailIDs, enumor.TaskDetailRunning); err != nil {
		logs.Errorf("fail to update task detail state, err: %v, opt: %+v rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	var result *hclb.BatchCreateResult
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
		err = actcli.GetHCService().TCloud.Clb.BatchModifyTargetWeight(
			kt.Kit(), opt.TargetGroupID, &opt.TCloudBatchOperateTargetReq)
	default:
		taskDetailState = enumor.TaskDetailFailed
		err = fmt.Errorf("vendor: %s not support for modify target weight", opt.Vendor)
		return nil, err
	}
	if err != nil {
		logs.Errorf("batch modify target weight failed, err: %v, rid: %s", err, kt.Kit().Rid)
		taskDetailState = enumor.TaskDetailFailed
		return result, err
	}

	return result, nil
}

// Rollback 批量修改RS权重失败时的回滚Action，此处不需要回滚处理
func (act ModifyTargetWeightAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- ModifyTargetWeightAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}

func validateDetailListStatus(kt *kit.Kit, detailIDs []string) (string, error) {
	// detail 状态检查
	detailList, err := listTaskDetail(kt, detailIDs)
	if err != nil {
		return fmt.Sprintf("task detail query failed"), err
	}
	for _, detail := range detailList {
		if detail.State == enumor.TaskDetailCancel {
			// 任务被取消，跳过该批次
			return fmt.Sprintf("task detail %s canceled", detail.ID), nil
		}
		if detail.State != enumor.TaskDetailInit {
			return "", errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
				detail.ID, detail.State)
		}
	}
	return "", nil
}
