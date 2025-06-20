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

// Package actioncvm ...
package actioncvm

import (
	"fmt"

	actflow "hcm/cmd/task-server/logics/flow"
	protocvm "hcm/pkg/api/hc-service/cvm"
	hclb "hcm/pkg/api/hc-service/load-balancer"
	"hcm/pkg/api/task-server/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"
)

// --------------------------[批量操作-批量重装CVM]-----------------------------

var _ action.Action = new(BatchTaskCvmResetAction)
var _ action.ParameterAction = new(BatchTaskCvmResetAction)

// BatchTaskCvmResetAction 批量操作-批量重装CVM
type BatchTaskCvmResetAction struct{}

// ParameterNew return request params.
func (act BatchTaskCvmResetAction) ParameterNew() (params any) {
	return new(cvm.BatchTaskCvmResetOption)
}

// Name return action name
func (act BatchTaskCvmResetAction) Name() enumor.ActionName {
	return enumor.ActionResetCvm
}

// Run 批量重装CVM
func (act BatchTaskCvmResetAction) Run(kt run.ExecuteKit, params any) (result any, taskErr error) {
	opt, ok := params.(*cvm.BatchTaskCvmResetOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter,
			fmt.Sprintf("params type mismatch, BatchTaskCvmResetOption:%+v", params))
	}

	asyncKit := kt.AsyncKit()
	results := make([]*hclb.BatchCreateResult, 0, len(opt.CvmResetList))
	for i := range opt.CvmResetList {
		detailID := opt.ManagementDetailIDs[i]
		// 逐条更新结果
		ret, optErr := act.batchResetCvmSystem(asyncKit, detailID, opt.CvmResetList[i])
		// 结束后写回状态
		targetState := enumor.TaskDetailSuccess
		if optErr != nil {
			// 更新为失败
			targetState = enumor.TaskDetailFailed
		}
		err := actflow.BatchUpdateTaskDetailResultState(asyncKit, []string{detailID}, targetState, ret, optErr)
		if err != nil {
			logs.Errorf("failed to set detail to after cloud operation finished, state: %s, detailID: %s, err: %v, "+
				"optErr: %+v, ret: %+v, rid: %s", targetState, detailID, err, optErr, cvt.PtrToVal(ret), kt.Kit().Rid)
			return nil, err
		}
		if optErr != nil {
			// abort
			return nil, optErr
		}
		results = append(results, ret)
	}
	// all success
	return results, nil
}

// batchResetCvmSystem 批量重装CVM
func (act BatchTaskCvmResetAction) batchResetCvmSystem(kt *kit.Kit, detailID string,
	req *protocvm.TCloudBatchResetCvmReq) (*hclb.BatchCreateResult, error) {

	detailList, err := actflow.ListTaskDetail(kt, []string{detailID})
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
		return nil, errf.Newf(errf.InvalidParameter, "task management detail is not init, detailID: %s, "+
			"taskManageID: %s, flowID: %s, status:%s", detail.ID, detail.TaskManagementID, detail.FlowID, detail.State)
	}

	// 更新任务状态为 running
	if err = actflow.BatchUpdateTaskDetailState(kt, []string{detailID}, enumor.TaskDetailRunning); err != nil {
		return nil, fmt.Errorf("failed to update detail to running, detailID: %s, err: %v", detailID, err)
	}

	// 调用云API重装CVM
	switch req.Vendor {
	case enumor.TCloud:
		err = act.resetTCloudCvm(kt, detail, req)
		if err != nil {
			logs.Errorf("failed to reset tcloud cvm, err: %v, detailID: %s, req: %+v, rid: %s",
				err, detailID, req, kt.Rid)
			return nil, err
		}
	default:
		return nil, errf.Newf(errf.InvalidParameter, "batch hcservice cvm reset failed, invalid vendor: %s", req.Vendor)
	}

	return nil, nil
}

// Rollback 重装CVM，无法回滚
func (act BatchTaskCvmResetAction) Rollback(kt run.ExecuteKit, params any) error {
	logs.Infof(" ----------- BatchTaskCvmResetAction Rollback -----------, params: %+v, rid: %s",
		params, kt.Kit().Rid)
	return nil
}
