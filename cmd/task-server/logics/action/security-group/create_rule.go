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
	hcproto "hcm/pkg/api/hc-service"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// CreateHuaweiSGRuleAction security group delete action
type CreateHuaweiSGRuleAction struct {
}

// CreateHuaweiSGRuleOption ...
type CreateHuaweiSGRuleOption struct {
	SGID    string                         `json:"sg_id" validate:"required"`
	RuleReq *hcproto.HuaWeiSGRuleCreateReq `json:"rule" validate:"omitempty"`
}

// Validate ...
func (opt *CreateHuaweiSGRuleOption) Validate() error {

	return validator.Validate.Struct(opt)
}

// ParameterNew returns parameter of
func (s CreateHuaweiSGRuleAction) ParameterNew() (params any) {
	return new(CreateHuaweiSGRuleOption)
}

// Name ActionDeleteSecurityGroup
func (s CreateHuaweiSGRuleAction) Name() enumor.ActionName {
	return enumor.ActionCreateHuaweiSGRule
}

// Run ...
func (s CreateHuaweiSGRuleAction) Run(kt run.ExecuteKit, params any) (any, error) {
	req, ok := params.(*CreateHuaweiSGRuleOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}
	return actcli.GetHCService().HuaWei.SecurityGroup.CreateSecurityGroupRule(kt.Kit(), req.SGID, req.RuleReq)

}
