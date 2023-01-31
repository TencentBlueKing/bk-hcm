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

package hcservice

import (
	"errors"

	"hcm/pkg/criteria/validator"
)

// HuaWeiSGRuleCreateReq define huawei security group create request.
type HuaWeiSGRuleCreateReq struct {
	AccountID   string              `json:"account_id" validate:"required"`
	EgressRule  *HuaWeiSGRuleCreate `json:"egress_rule" validate:"omitempty"`
	IngressRule *HuaWeiSGRuleCreate `json:"ingress_rule" validate:"omitempty"`
}

// Validate huawei security group rule create request.
func (req *HuaWeiSGRuleCreateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.EgressRule == nil && req.IngressRule == nil {
		return errors.New("egress rule or ingress rule is required")
	}

	if req.EgressRule != nil && req.IngressRule != nil {
		return errors.New("egress rule or ingress rule only one is allowed")
	}

	return nil
}

// HuaWeiSGRuleCreate define huawei sg rule spec when create.
type HuaWeiSGRuleCreate struct {
	Memo               *string `json:"memo"`
	Ethertype          *string `json:"ethertype"`
	Protocol           *string `json:"protocol"`
	RemoteIPPrefix     *string `json:"remote_ip_prefix"`
	CloudRemoteGroupID *string `json:"cloud_remote_group_id"`
	Port               *string `json:"port"`
	Action             *string `json:"action"`
	Priority           int64   `json:"priority"`
}
