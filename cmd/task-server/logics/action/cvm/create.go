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
	"fmt"

	"github.com/tidwall/gjson"

	actcli "hcm/cmd/task-server/logics/action/cli"
	hccvm "hcm/pkg/api/hc-service/cvm"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"
)

// SaveCreateCvmCloudIDKey 创建成功的主机ID列表
const SaveCreateCvmCloudIDKey = "create_cvm_cloud_ids"

var _ action.Action = new(CreateCvmAction)
var _ action.ParameterAction = new(CreateCvmAction)

// CreateCvmAction define create cvm action.
type CreateCvmAction struct{}

// CreateOption define create cvm option.
type CreateOption struct {
	Vendor                     enumor.Vendor `json:"vendor" validate:"required"`
	hccvm.TCloudBatchCreateReq `json:",inline"`
	hccvm.AwsBatchCreateReq    `json:",inline"`
	hccvm.HuaWeiBatchCreateReq `json:",inline"`
	hccvm.GcpBatchCreateReq    `json:",inline"`
	hccvm.AzureCreateReq       `json:",inline"`
}

// MarshalJSON CreateOption.
func (opt CreateOption) MarshalJSON() ([]byte, error) {

	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor                     enumor.Vendor `json:"vendor" validate:"required"`
			hccvm.TCloudBatchCreateReq `json:",inline"`
		}{
			Vendor:               opt.Vendor,
			TCloudBatchCreateReq: opt.TCloudBatchCreateReq,
		}
	case enumor.Aws:
		req = struct {
			Vendor                  enumor.Vendor `json:"vendor" validate:"required"`
			hccvm.AwsBatchCreateReq `json:",inline"`
		}{
			Vendor:            opt.Vendor,
			AwsBatchCreateReq: opt.AwsBatchCreateReq,
		}
	case enumor.HuaWei:
		req = struct {
			Vendor                     enumor.Vendor `json:"vendor" validate:"required"`
			hccvm.HuaWeiBatchCreateReq `json:",inline"`
		}{
			Vendor:               opt.Vendor,
			HuaWeiBatchCreateReq: opt.HuaWeiBatchCreateReq,
		}
	case enumor.Gcp:
		req = struct {
			Vendor                  enumor.Vendor `json:"vendor" validate:"required"`
			hccvm.GcpBatchCreateReq `json:",inline"`
		}{
			Vendor:            opt.Vendor,
			GcpBatchCreateReq: opt.GcpBatchCreateReq,
		}
	case enumor.Azure:
		req = struct {
			Vendor               enumor.Vendor `json:"vendor" validate:"required"`
			hccvm.AzureCreateReq `json:",inline"`
		}{
			Vendor:         opt.Vendor,
			AzureCreateReq: opt.AzureCreateReq,
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return json.Marshal(req)
}

// UnmarshalJSON CreateOption.
func (opt *CreateOption) UnmarshalJSON(raw []byte) (err error) {
	opt.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())

	switch opt.Vendor {
	case enumor.TCloud:
		err = json.Unmarshal(raw, &opt.TCloudBatchCreateReq)
	case enumor.Aws:
		err = json.Unmarshal(raw, &opt.AwsBatchCreateReq)
	case enumor.HuaWei:
		err = json.Unmarshal(raw, &opt.HuaWeiBatchCreateReq)
	case enumor.Gcp:
		err = json.Unmarshal(raw, &opt.GcpBatchCreateReq)
	case enumor.Azure:
		err = json.Unmarshal(raw, &opt.AzureCreateReq)
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate CreateOption.
func (opt CreateOption) Validate() error {
	if err := opt.Vendor.Validate(); err != nil {
		return err
	}

	var req validator.Interface
	switch opt.Vendor {
	case enumor.TCloud:
		req = &opt.TCloudBatchCreateReq
	case enumor.Aws:
		req = &opt.AwsBatchCreateReq
	case enumor.HuaWei:
		req = &opt.HuaWeiBatchCreateReq
	case enumor.Gcp:
		req = &opt.GcpBatchCreateReq
	case enumor.Azure:
		req = &opt.AzureCreateReq
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	if err := req.Validate(); err != nil {
		return err
	}

	return nil
}

// ParameterNew return request params.
func (act CreateCvmAction) ParameterNew() (params interface{}) {
	return new(CreateOption)
}

// Name return action name
func (act CreateCvmAction) Name() enumor.ActionName {
	return enumor.ActionCreateCvm
}

// Run create cvm.
func (act CreateCvmAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*CreateOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	var result *hccvm.BatchCreateResult
	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		result, err = actcli.GetHCService().TCloud.Cvm.BatchCreateCvm(kt.Kit(), &opt.TCloudBatchCreateReq)
	case enumor.Aws:
		result, err = actcli.GetHCService().Aws.Cvm.BatchCreateCvm(kt.Kit(), &opt.AwsBatchCreateReq)
	case enumor.HuaWei:
		result, err = actcli.GetHCService().HuaWei.Cvm.BatchCreateCvm(kt.Kit(), &opt.HuaWeiBatchCreateReq)
	case enumor.Gcp:
		result, err = actcli.GetHCService().Gcp.Cvm.BatchCreateCvm(kt.Kit(), &opt.GcpBatchCreateReq)
	case enumor.Azure:
		var azureResult *hccvm.AzureCreateResp
		azureResult, err = actcli.GetHCService().Azure.Cvm.CreateCvm(kt.Kit(), &opt.AzureCreateReq)
		if azureResult != nil {
			result = &hccvm.BatchCreateResult{
				SuccessCloudIDs: []string{azureResult.CloudID},
			}
		}
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}
	if err != nil {
		logs.Errorf("batch create cvm failed, err: %v, result: %+v, rid: %s", err, result, kt.Kit().Rid)
		return result, err
	}

	if len(result.FailedMessage) != 0 {
		return result, errors.New(result.FailedMessage)
	}

	if err = kt.ShareData().AppendIDs(kt.Kit(), SaveCreateCvmCloudIDKey, result.SuccessCloudIDs...); err != nil {
		logs.Errorf("share data appendIDs failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}
