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
	"hcm/pkg/logs"
	"hcm/pkg/tools/json"

	"github.com/tidwall/gjson"
)

// SaveAddRsCloudIDKey 批量添加RS成功的监听器ID列表
const SaveAddRsCloudIDKey = "batch_add_rs_cloud_ids"

var _ action.Action = new(AddRsAction)
var _ action.ParameterAction = new(AddRsAction)

// AddRsAction define add rs action.
type AddRsAction struct{}

// AddRsOption define add rs option.
type AddRsOption struct {
	Vendor                          enumor.Vendor `json:"vendor" validate:"required"`
	hclb.TCloudBatchCreateTargetReq `json:",inline"`
}

// MarshalJSON AddRsOption.
func (opt AddRsOption) MarshalJSON() ([]byte, error) {
	var req interface{}
	switch opt.Vendor {
	case enumor.TCloud:
		req = struct {
			Vendor                          enumor.Vendor `json:"vendor" validate:"required"`
			hclb.TCloudBatchCreateTargetReq `json:",inline"`
		}{
			Vendor:                     opt.Vendor,
			TCloudBatchCreateTargetReq: opt.TCloudBatchCreateTargetReq,
		}

	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return json.Marshal(req)
}

// UnmarshalJSON AddRsOption.
func (opt *AddRsOption) UnmarshalJSON(raw []byte) (err error) {
	opt.Vendor = enumor.Vendor(gjson.GetBytes(raw, "vendor").String())

	switch opt.Vendor {
	case enumor.TCloud:
		err = json.Unmarshal(raw, &opt.TCloudBatchCreateTargetReq)
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return err
}

// Validate AddRsOption.
func (opt AddRsOption) Validate() error {
	if err := opt.Vendor.Validate(); err != nil {
		return err
	}

	var req validator.Interface
	switch opt.Vendor {
	case enumor.TCloud:
		req = &opt.TCloudBatchCreateTargetReq
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	if err := req.Validate(); err != nil {
		return err
	}

	return nil
}

// ParameterNew return request params.
func (act AddRsAction) ParameterNew() (params interface{}) {
	return new(AddRsOption)
}

// Name return action name
func (act AddRsAction) Name() enumor.ActionName {
	return enumor.ActionAddRS
}

// Run add rs.
func (act AddRsAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*AddRsOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	var result *hclb.BatchCreateResult
	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		result, err = actcli.GetHCService().TCloud.Clb.BatchAddRs(
			kt.Kit(), opt.TargetGroupID, &opt.TCloudBatchCreateTargetReq)
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}
	if err != nil {
		logs.Errorf("batch add rs failed, err: %v, result: %+v, rid: %s", err, result, kt.Kit().Rid)
		return result, err
	}

	if len(result.FailedCloudIDs) != 0 {
		return result, errf.Newf(errf.PartialFailed, "batch add rs rs partially failed, failCloudIDs: %v",
			result.FailedCloudIDs)
	}

	if err = kt.ShareData().AppendIDs(kt.Kit(), SaveAddRsCloudIDKey, result.SuccessCloudIDs...); err != nil {
		logs.Errorf("share data appendIDs failed, err: %v, rid: %s", err, kt.Kit().Rid)
		return result, err
	}

	return result, nil
}

// Rollback 批量添加RS失败时的回滚Action，此处不需要回滚处理
func (act AddRsAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof(" ----------- AddRsAction Rollback -----------, params: %s, rid: %s", params, kt.Kit().Rid)
	return nil
}
