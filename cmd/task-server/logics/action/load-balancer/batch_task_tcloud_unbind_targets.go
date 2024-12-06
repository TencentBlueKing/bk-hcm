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
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

// --------------------------[批量操作-解绑RS]-----------------------------

var _ action.Action = new(BatchTaskUnBindTargetAction)
var _ action.ParameterAction = new(BatchTaskUnBindTargetAction)

// BatchTaskUnBindTargetAction 批量操作-解绑RS
type BatchTaskUnBindTargetAction struct{}

// BatchTaskUnBindTargetOption ...
type BatchTaskUnBindTargetOption struct {
	Vendor         enumor.Vendor `json:"vendor" validate:"required"`
	LoadBalancerID string        `json:"lb_id" validate:"required"`
	// ManagementDetailIDs 对应的详情行id列表，需要和批量绑定的Targets参数长度对应
	ManagementDetailIDs []string                       `json:"management_detail_ids" validate:"required,max=20"`
	LblList             []*hclb.TCloudBatchUnbindRsReq `json:"lbl_list" validate:"required,max=20"`
}

// Validate validate option.
func (opt BatchTaskUnBindTargetOption) Validate() error {

	switch opt.Vendor {
	case enumor.TCloud:
	default:
		return fmt.Errorf("unsupport vendor for batch unbind rs: %s", opt.Vendor)
	}

	if opt.LblList == nil {
		return errf.New(errf.InvalidParameter, "lbl_list is required")
	}
	if len(opt.ManagementDetailIDs) != len(opt.LblList) {
		return errf.Newf(errf.InvalidParameter, "management_detail_ids and lbl list num not match, %d != %d",
			len(opt.ManagementDetailIDs), len(opt.LblList))
	}
	return validator.Validate.Struct(opt)
}

// ParameterNew return request params.
func (act BatchTaskUnBindTargetAction) ParameterNew() (params any) {
	return new(BatchTaskUnBindTargetOption)
}

// Name return action name
func (act BatchTaskUnBindTargetAction) Name() enumor.ActionName {
	return enumor.ActionBatchTaskTCloudUnBindTarget
}

// Run 批量解绑RS
func (act BatchTaskUnBindTargetAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*BatchTaskUnBindTargetOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	results := make([]*hclb.BatchCreateResult, 0, len(opt.LblList))
	for i := range opt.LblList {
		detailID := opt.ManagementDetailIDs[i]
		// 逐条更新结果，结束后写回状态
		ret, optErr := act.batchListenerUnbindRs(kt.Kit(), opt.LoadBalancerID, detailID, opt.LblList[i])
		targetState := enumor.TaskDetailSuccess
		if optErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := batchUpdateTaskDetailResultState(kt.Kit(), []string{detailID}, targetState, ret, optErr)
		if err != nil {
			logs.Errorf("failed to set detail to [%s] after cloud operation finished, err: %v, rid: %s",
				targetState, err, kt.Kit().Rid)
			return nil, err
		}
		if optErr != nil {
			// abort
			return nil, err
		}
		results = append(results, ret)
	}
	// all success
	return results, nil
}

// batchListenerUnbindRs 批量解绑监听器的RS
func (act BatchTaskUnBindTargetAction) batchListenerUnbindRs(kt *kit.Kit, lbID, detailID string,
	req *hclb.TCloudBatchUnbindRsReq) (*hclb.BatchCreateResult, error) {

	detailList, err := listTaskDetail(kt, []string{detailID})
	if err != nil {
		logs.Errorf("failed to query task detail, err: %v, detailID: %s, rid: %s", err, detailID, kt.Rid)
		return nil, err
	}

	detail := detailList[0]
	if detail.State == enumor.TaskDetailCancel {
		// 任务被取消，跳过该任务, 直接成功即可
		return nil, nil
	}
	if detail.State != enumor.TaskDetailInit {
		return nil, errf.Newf(errf.InvalidParameter, "task management detail(%s) status(%s) is not init",
			detail.ID, detail.State)
	}

	// 更新任务状态为 running
	if err = batchUpdateTaskDetailState(kt, []string{detailID}, enumor.TaskDetailRunning); err != nil {
		return nil, fmt.Errorf("failed to update detail to running, detailID: %s, err: %v", detailID, err)
	}

	// 调用云API解绑rs，支持幂等
	lblResp, err := actcli.GetHCService().TCloud.Clb.BatchRemoveListenerTarget(kt, lbID, req)
	if err != nil {
		logs.Errorf("failed to call hc to listener unbind rs, err: %v, lbID: %s, detailID: %s, rid: %s",
			err, lbID, detailID, kt.Rid)
		return nil, err
	}
	return lblResp, nil
}

// Rollback 解绑rs支持重入，无需回滚
func (act BatchTaskUnBindTargetAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskUnBindTargetAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
