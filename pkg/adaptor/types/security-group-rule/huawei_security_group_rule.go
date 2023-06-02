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

package securitygrouprule

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
)

// -------------------------- Create --------------------------

// HuaWeiCreateOption huawei security group rule create option.
type HuaWeiCreateOption struct {
	Region               string        `json:"region" validate:"required"`
	CloudSecurityGroupID string        `json:"cloud_security_group_id" validate:"required"`
	Rule                 *HuaWeiCreate `json:"rule" validate:"required"`
}

// Validate huawei security group rule create option.
func (opt HuaWeiCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// HuaWeiCreate huawei security group rule.
type HuaWeiCreate struct {
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

// HuaWeiDeleteOption huawei security group delete option.
type HuaWeiDeleteOption struct {
	Region      string `json:"region" validate:"required"`
	CloudRuleID string `json:"cloud_rule_id" validate:"required"`
}

// Validate huawei security group rule delete option.
func (opt HuaWeiDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// HuaWeiListOption define huawei security group rule list option.
type HuaWeiListOption struct {
	Region               string           `json:"region" validate:"required"`
	CloudSecurityGroupID string           `json:"cloud_security_group_id" validate:"required"`
	Page                 *core.HuaWeiPage `json:"page" validate:"omitempty"`
}

// Validate huawei security group rule list option.
func (opt HuaWeiListOption) Validate() error {
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

// HuaWeiSGRule for model SecurityGroupRule
type HuaWeiSGRule struct {
	model.SecurityGroupRule
}

// GetCloudID ...
func (sgrule HuaWeiSGRule) GetCloudID() string {
	return sgrule.Id
}
