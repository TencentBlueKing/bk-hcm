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
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/runtime/filter"
)

// ListSecurityGroupDetails list security group details.
type ListSecurityGroupDetails struct {
	Count   uint64                     `json:"count,omitempty"`
	Details []cloud.SecurityGroupTable `json:"details,omitempty"`
}

// ListTCloudSGRuleDetails list tcloud security group rule details.
type ListTCloudSGRuleDetails struct {
	Count   uint64                               `json:"count,omitempty"`
	Details []cloud.TCloudSecurityGroupRuleTable `json:"details,omitempty"`
}

// ListAwsSGRuleDetails list aws security group rule details.
type ListAwsSGRuleDetails struct {
	Count   uint64                            `json:"count,omitempty"`
	Details []cloud.AwsSecurityGroupRuleTable `json:"details,omitempty"`
}

// ListHuaWeiSGRuleDetails list huawei security group rule details.
type ListHuaWeiSGRuleDetails struct {
	Count   uint64                               `json:"count,omitempty"`
	Details []cloud.HuaWeiSecurityGroupRuleTable `json:"details,omitempty"`
}

// ListAzureSGRuleDetails list azure security group rule details.
type ListAzureSGRuleDetails struct {
	Count   uint64                              `json:"count,omitempty"`
	Details []cloud.AzureSecurityGroupRuleTable `json:"details,omitempty"`
}

// SGRuleListOption defines options to list security group rule.
type SGRuleListOption struct {
	SecurityGroupID string
	Fields          []string
	Filter          *filter.Expression
	Page            *core.BasePage
}

// Validate list option.
func (opt *SGRuleListOption) Validate(eo *filter.ExprOption, po *core.PageOption) error {
	if len(opt.SecurityGroupID) == 0 {
		return errf.New(errf.InvalidParameter, "security group is required")
	}

	if opt.Filter == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	if opt.Page == nil {
		return errf.New(errf.InvalidParameter, "page is required")
	}

	if eo == nil {
		return errf.New(errf.InvalidParameter, "filter expr option is required")
	}

	if po == nil {
		return errf.New(errf.InvalidParameter, "page option is required")
	}

	if err := opt.Filter.Validate(eo); err != nil {
		return err
	}

	if err := opt.Page.Validate(po); err != nil {
		return err
	}

	return nil
}
