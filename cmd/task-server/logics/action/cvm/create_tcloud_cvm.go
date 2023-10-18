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
	"errors"

	actcli "hcm/cmd/task-server/logics/action/cli"
	hccvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
)

var _ action.Action = new(CreateTCloudCvmAction)
var _ action.ParameterAction = new(CreateTCloudCvmAction)

// CreateTCloudCvmAction define create tcloud cvm action.
type CreateTCloudCvmAction struct{}

// ParameterNew return request params.
func (act CreateTCloudCvmAction) ParameterNew() (params interface{}) {
	return new(hccvm.TCloudBatchCreateReq)
}

// Name return action name
func (act CreateTCloudCvmAction) Name() enumor.ActionName {
	return enumor.ActionCreateTCloudCvm
}

// Run create tcloud cvm.
func (act CreateTCloudCvmAction) Run(kt run.ExecuteKit, params interface{}) error {
	req, ok := params.(*hccvm.TCloudBatchCreateReq)
	if !ok {
		return errf.New(errf.InvalidParameter, "params type not right")
	}

	result, err := actcli.GetHCService().TCloud.Cvm.BatchCreateCvm(kt.Kit(), req)
	if err != nil {
		logs.Errorf("batch create tcloud cvm failed, err: %v, result: %v, rid: %s", err, result, kt.Kit().Rid)
		return err
	}

	if len(result.SuccessCloudIDs) != 0 {
		if err = kt.ShareData().AppendSuccessCloudIDs(kt.Kit(), result.SuccessCloudIDs...); err != nil {
			logs.Errorf("append success cloud ids failed, err: %v, result: %+v, rid: %s", err, result, kt.Kit().Rid)
			return err
		}
	}

	if len(result.FailedCloudIDs) != 0 {
		if err = kt.ShareData().AppendFailedCloudIDs(kt.Kit(), result.SuccessCloudIDs...); err != nil {
			logs.Errorf("append failed cloud ids failed, err: %v, result: %+v, rid: %s", err, result, kt.Kit().Rid)
			return err
		}
	}

	if len(result.UnknownCloudIDs) != 0 {
		if err = kt.ShareData().AppendUnknownCloudIDs(kt.Kit(), result.SuccessCloudIDs...); err != nil {
			logs.Errorf("append unknown cloud ids failed, err: %v, result: %+v, rid: %s", err, result, kt.Kit().Rid)
			return err
		}
	}

	if len(result.FailedMessage) != 0 {
		return errors.New(result.FailedMessage)
	}

	return nil
}
