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

package actionsubnet

import (
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

var _ action.Action = new(DeleteAction)
var _ action.ParameterAction = new(DeleteAction)

// DeleteAction define delete cvm action.
type DeleteAction struct{}

// DeleteSubnetOption 删除子网选项
type DeleteSubnetOption struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`
	ID     string        `json:"id" validate:"required"`
}

// Validate DeleteSubnetOption.
func (opt DeleteSubnetOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ParameterNew return delete params.
func (act DeleteAction) ParameterNew() (params interface{}) {
	return new(DeleteSubnetOption)
}

// Name ...
func (act DeleteAction) Name() enumor.ActionName {
	return enumor.ActionDeleteSubnet
}

// Run ...
func (act DeleteAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*DeleteSubnetOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		err = actcli.GetHCService().TCloud.Subnet.Delete(kt.Kit(), opt.ID)
	case enumor.Aws:
		err = actcli.GetHCService().Aws.Subnet.Delete(kt.Kit(), opt.ID)
	case enumor.Gcp:
		err = actcli.GetHCService().Gcp.Subnet.Delete(kt.Kit(), opt.ID)
	case enumor.Azure:
		err = actcli.GetHCService().Azure.Subnet.Delete(kt.Kit(), opt.ID)
	case enumor.HuaWei:
		err = actcli.GetHCService().HuaWei.Subnet.Delete(kt.Kit(), opt.ID)
	default:
		return nil, fmt.Errorf("vendor: %s not support", opt.Vendor)
	}
	if err != nil {
		logs.Errorf("delete subnet failed, err: %v, opt: %+v, rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	return nil, nil
}
