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
	"encoding/json"
	"strings"

	actcli "hcm/cmd/task-server/logics/action/cli"
	actionflow "hcm/cmd/task-server/logics/flow"
	coretask "hcm/pkg/api/core/task"
	hcprotocvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/tools/retry"
)

func (act BatchTaskCvmResetAction) resetTCloudCvm(kt *kit.Kit, detail coretask.Detail,
	req *hcprotocvm.TCloudBatchResetCvmReq) error {

	var cloudErr error
	rangeMS := [2]uint{constant.CvmBatchTaskRetryDelayMinMS, constant.CvmBatchTaskRetryDelayMaxMS}
	policy := retry.NewRetryPolicy(0, rangeMS)
	for {
		cloudErr = actcli.GetHCService().TCloud.Cvm.ResetCvm(kt, req)
		cvmResetJson, jsonErr := json.Marshal(req)
		if jsonErr != nil {
			logs.Errorf("call hcservice api reset cvm json marshal, vendor: %s, detailID: %s, taskManageID: %s, "+
				"flowID: %s, cvmResetJson: %s, jsonErr: %+v, rid: %s", req.Vendor, detail.ID,
				detail.TaskManagementID, detail.FlowID, cvmResetJson, jsonErr, kt.Rid)
			return jsonErr
		}
		// 仅在碰到限频错误时进行重试
		if cloudErr != nil && strings.Contains(cloudErr.Error(), constant.TCloudLimitExceededErrCode) {
			if policy.RetryCount()+1 < actionflow.BatchTaskDefaultRetryTimes {
				// 	非最后一次重试，继续sleep
				logs.Errorf("call tcloud cvm reset reach rate limit, will sleep for retry, vendor: %s, "+
					"retry count: %d, err: %v, rid: %s", req.Vendor, policy.RetryCount(), cloudErr, kt.Rid)
				policy.Sleep()
				continue
			}
		}
		// 其他情况都跳过
		break
	}

	// 记录云端报错信息
	if cloudErr != nil {
		logs.Errorf("failed to call hcservice to reset cvm, vendor: %s, err: %v, detailID: %s, taskManageID: %s, "+
			"flowID: %s, rid: %s", req.Vendor, cloudErr, detail.ID, detail.TaskManagementID, detail.FlowID, kt.Rid)
		return cloudErr
	}

	return nil
}
