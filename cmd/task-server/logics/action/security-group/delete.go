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

package actionsg

import (
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/logs"
)

// DeleteSgAction security group delete action
type DeleteSgAction struct {
}

// DeleteSGOption ...
type DeleteSGOption struct {
	Vendor enumor.Vendor `json:"vendor" validate:"required"`
	ID     string        `json:"id" validate:"required"`
}

// Validate ...
func (opt *DeleteSGOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// ParameterNew returns parameter of
func (s DeleteSgAction) ParameterNew() (params interface{}) {
	return new(DeleteSGOption)
}

// Name ActionDeleteSecurityGroup
func (s DeleteSgAction) Name() enumor.ActionName {
	return enumor.ActionDeleteSecurityGroup
}

// Run ...
func (s DeleteSgAction) Run(kt run.ExecuteKit, params any) (any, error) {
	opt, ok := params.(*DeleteSGOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if err := opt.Validate(); err != nil {
		return nil, err
	}

	cli := actcli.GetHCService()
	var err error
	switch opt.Vendor {
	case enumor.TCloud:
		err = cli.TCloud.SecurityGroup.DeleteSecurityGroup(kt.Kit(), opt.ID)
	case enumor.Aws:
		err = cli.Aws.SecurityGroup.DeleteSecurityGroup(kt.Kit(), opt.ID)
	case enumor.HuaWei:
		err = cli.HuaWei.SecurityGroup.DeleteSecurityGroup(kt.Kit(), opt.ID)
	case enumor.Azure:
		err = cli.Azure.SecurityGroup.DeleteSecurityGroup(kt.Kit(), opt.ID)
	default:
		return nil, errf.Newf(errf.Unknown, "id: %s vendor: %s not support", opt.ID, opt.Vendor)
	}
	if err != nil {
		logs.Errorf("delete security group failed, err: %v, vendor: %s, opt: %v, rid: %s",
			err, opt.Vendor, opt, kt.Kit().Rid)
		return nil, err
	}

	return nil, nil
}
