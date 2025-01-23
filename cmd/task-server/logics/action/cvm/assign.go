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

package actioncvm

import (
	"fmt"

	"hcm/cmd/cloud-server/logics/cvm"
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	corecvm "hcm/pkg/api/core/cloud/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/tools"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/logs"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"
)

var _ action.Action = new(AssignCvmAction)
var _ action.ParameterAction = new(AssignCvmAction)

// AssignCvmAction 分配主机Action
type AssignCvmAction struct{}

// AssignCvmOption assign cvm option.
type AssignCvmOption struct {
	BizID     int64  `json:"bk_biz_id" validate:"required"`
	BkCloudID *int64 `json:"bk_cloud_id" validate:"required"`
}

// Validate AssignCvmOption.
func (opt AssignCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ParameterNew ...
func (act AssignCvmAction) ParameterNew() (params interface{}) {
	return new(AssignCvmOption)
}

// Name return action name.
func (act AssignCvmAction) Name() enumor.ActionName {
	return enumor.ActionAssignCvm
}

// Run assign cvm action.
func (act AssignCvmAction) Run(kt run.ExecuteKit, params interface{}) (result interface{}, err error) {
	opt, ok := params.(*AssignCvmOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	idsStr, exist := kt.ShareData().Get(SaveCreateCvmCloudIDKey)
	if !exist {
		return nil, fmt.Errorf("cvm_cloud_ids is required for assign")
	}

	cli := actcli.GetDataService()

	cloudIDs := tableasync.ParseIDsStr(idsStr)
	split := slice.Split(cloudIDs, constant.BatchOperationMaxLimit)
	cvmIDs := make([]string, 0, len(cloudIDs))
	for _, partIDs := range split {
		listReq := &core.ListReq{
			Filter: tools.ContainersExpression("cloud_id", partIDs),
			Page:   core.NewDefaultBasePage(),
		}
		listResult, err := cli.Global.Cvm.ListCvm(kt.Kit(), listReq)
		if err != nil {
			logs.Errorf("list cvm failed, err: %v, cloudIDs: %v, rid: %s", err, partIDs, kt.Kit().Rid)
			return nil, err
		}

		ids := slice.Map(listResult.Details, func(cvm corecvm.BaseCvm) string {
			return cvm.ID
		})

		assignedCvmInfo := make([]cvm.AssignedCvmInfo, 0, len(ids))
		for _, id := range ids {
			assignedCvmInfo = append(assignedCvmInfo, cvm.AssignedCvmInfo{
				CvmID:     id,
				BkBizID:   opt.BizID,
				BkCloudID: converter.PtrToVal(opt.BkCloudID),
			})
		}
		if err = cvm.Assign(kt.Kit(), cli, assignedCvmInfo); err != nil {
			logs.Errorf("assign cvm failed, err: %v, ids: %+v, rid: %s", err, ids, kt.Kit().Rid)
			return nil, err
		}
		cvmIDs = append(cvmIDs, ids...)
	}

	return &AssignCvmResult{IDs: cvmIDs}, nil
}

// AssignCvmResult assign cvm result.
type AssignCvmResult struct {
	IDs []string `json:"ids"`
}
