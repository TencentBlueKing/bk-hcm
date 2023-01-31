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

package types

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// HuaWeiSGRuleCreateOption huawei security group rule create option.
type HuaWeiSGRuleCreateOption struct {
	Region               string              `json:"region" validate:"required"`
	CloudSecurityGroupID string              `json:"cloud_security_group_id" validate:"required"`
	Rule                 *HuaWeiSGRuleCreate `json:"rule" validate:"required"`
}

// Validate huawei security group rule create option.
func (opt HuaWeiSGRuleCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// HuaWeiSGRuleCreate huawei security group rule.
type HuaWeiSGRuleCreate struct {
	Description        *string                      `json:"description"`
	Ethertype          *string                      `json:"ethertype"`
	Protocol           *string                      `json:"protocol"`
	RemoteIPPrefix     *string                      `json:"remote_ip_prefix"`
	CloudRemoteGroupID *string                      `json:"cloud_remote_group_id"`
	Port               *string                      `json:"port"`
	Action             *string                      `json:"action"`
	Priority           *string                      `json:"priority"`
	Type               enumor.SecurityGroupRuleType `json:"type"`
}

// -------------------------- Delete --------------------------

// HuaWeiSGRuleDeleteOption huawei security group delete option.
type HuaWeiSGRuleDeleteOption struct {
	Region      string `json:"region" validate:"required"`
	CloudRuleID string `json:"cloud_rule_id" validate:"required"`
}

// Validate huawei security group rule delete option.
func (opt HuaWeiSGRuleDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// HuaWeiSGRuleListOption define huawei security group rule list option.
type HuaWeiSGRuleListOption struct {
	Region               string           `json:"region" validate:"required"`
	CloudSecurityGroupID string           `json:"cloud_security_group_id" validate:"required"`
	Page                 *core.HuaweiPage `json:"page" validate:"omitempty"`
}

// Validate huawei security group rule list option.
func (opt HuaWeiSGRuleListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}
