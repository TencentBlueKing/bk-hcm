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

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	hcservice "hcm/pkg/client/hc-service"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
)

var _ action.Action = new(CvmOperationAction)
var _ action.ParameterAction = new(CvmOperationAction)

// CvmOperationAction define cvm operation action.
type CvmOperationAction struct {
	ActionName enumor.ActionName
	TCloudFunc func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error
	AwsFunc    func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error
	HuaWeiFunc func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error
	GcpFunc    func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error
	AzureFunc  func(kt *kit.Kit, cli *hcservice.Client, opt *CvmOperationOption) error
}

// CvmOperationOption operation cvm option.
type CvmOperationOption struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	AccountID string        `json:"account_id" validate:"required"`
	Region    string        `json:"region" validate:"omitempty"`
	// IDs TCloud/HuaWei/Aws 支持批量操作，Azure/Gcp 仅支持单个操作
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate operation cvm option.
func (opt CvmOperationOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	switch opt.Vendor {
	case enumor.TCloud, enumor.Aws, enumor.HuaWei:
		if len(opt.Region) == 0 {
			return fmt.Errorf("vendor: %s region is required", opt.Vendor)
		}

	case enumor.Azure, enumor.Gcp:
		if len(opt.IDs) > 1 {
			return fmt.Errorf("vendor: %s only support operation a single cvm", opt.Vendor)
		}
	default:
		return fmt.Errorf("vendor: %s not support", opt.Vendor)
	}

	return nil
}

// ParameterNew return operation cvm option.
func (act CvmOperationAction) ParameterNew() interface{} {
	return new(CvmOperationOption)
}

// Name return action name.
func (act CvmOperationAction) Name() enumor.ActionName {
	return act.ActionName
}

// Run operation cvm.
func (act CvmOperationAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*CvmOperationOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type not right")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	cli := actcli.GetHCService()

	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		err = act.TCloudFunc(kt.Kit(), cli, opt)
	case enumor.HuaWei:
		err = act.HuaWeiFunc(kt.Kit(), cli, opt)
	case enumor.Aws:
		err = act.AwsFunc(kt.Kit(), cli, opt)
	case enumor.Gcp:
		err = act.GcpFunc(kt.Kit(), cli, opt)
	case enumor.Azure:
		err = act.AzureFunc(kt.Kit(), cli, opt)
	}
	if err != nil {
		logs.Errorf("operate cvm failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	return nil, nil
}
