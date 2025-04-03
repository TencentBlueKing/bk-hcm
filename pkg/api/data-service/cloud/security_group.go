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

package cloud

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// -------------------------- Create --------------------------

// SecurityGroupBatchCreateReq security group create request.
type SecurityGroupBatchCreateReq[Extension cloud.SecurityGroupExtension] struct {
	SecurityGroups []SecurityGroupBatchCreate[Extension] `json:"security_groups" validate:"required"`
}

// SecurityGroupBatchCreate define security group batch create.
type SecurityGroupBatchCreate[Extension cloud.SecurityGroupExtension] struct {
	CloudID          string          `json:"cloud_id" validate:"required"`
	Region           string          `json:"region" validate:"required"`
	Name             string          `json:"name" validate:"required"`
	Memo             *string         `json:"memo" validate:"omitempty"`
	AccountID        string          `json:"account_id" validate:"required"`
	BkBizID          int64           `json:"bk_biz_id" validate:"required"`
	MgmtType         enumor.MgmtType `json:"mgmt_type" validate:"lte=64"`
	MgmtBizID        int64           `json:"mgmt_biz_id" `
	Manager          string          `json:"manager" validate:"lte=64"`
	BakManager       string          `json:"bak_manager" validate:"lte=64"`
	Extension        *Extension      `json:"extension" validate:"required"`
	UsageBizIds      []int64         `json:"usage_biz_ids" validate:"omitempty"`
	CloudCreatedTime string          `json:"cloud_created_time" validate:"required"`
	CloudUpdateTime  string          `json:"cloud_update_time" validate:"required"`
	Tags             core.TagMap     `json:"tags" validate:"required"`
}

// Validate security group create request.
func (req *SecurityGroupBatchCreateReq[T]) Validate() error {
	if len(req.SecurityGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// SecurityGroupBatchUpdateReq security group batch update request.
type SecurityGroupBatchUpdateReq[Extension cloud.SecurityGroupExtension] struct {
	SecurityGroups []SecurityGroupBatchUpdate[Extension] `json:"security_groups" validate:"required"`
}

// SecurityGroupBatchUpdate define security group batch update.
type SecurityGroupBatchUpdate[Extension cloud.SecurityGroupExtension] struct {
	ID               string          `json:"id" validate:"required"`
	Name             string          `json:"name" validate:"omitempty"`
	BkBizID          int64           `json:"bk_biz_id" validate:"omitempty"`
	MgmtType         enumor.MgmtType `json:"mgmt_type" validate:"lte=64"`
	MgmtBizID        int64           `json:"mgmt_biz_id" validate:"omitempty"`
	Manager          string          `json:"manager" validate:"lte=64"`
	BakManager       string          `json:"bak_manager" validate:"lte=64"`
	Memo             *string         `json:"memo" validate:"omitempty"`
	Extension        *Extension      `json:"extension" validate:"omitempty"`
	CloudCreatedTime string          `json:"cloud_created_time" validate:"omitempty"`
	CloudUpdateTime  string          `json:"cloud_update_time" validate:"omitempty"`
	Tags             core.TagMap     `json:"tags" validate:"omitempty"`
}

// Validate security group update request.
func (req *SecurityGroupBatchUpdateReq[T]) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.SecurityGroups) == 0 {
		return errors.New("security group is required")
	}

	if len(req.SecurityGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// SecurityGroupCommonInfoBatchUpdateReq define security group common info batch update req.
type SecurityGroupCommonInfoBatchUpdateReq struct {
	IDs     []string `json:"ids" validate:"required"`
	BkBizID int64    `json:"bk_biz_id" validate:"required"`
}

// Validate security group common info batch update req.
func (req *SecurityGroupCommonInfoBatchUpdateReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.IDs) == 0 {
		return errors.New("ids required")
	}

	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// BatchUpdateSecurityGroupMgmtAttrReq security group update management attribute request.
type BatchUpdateSecurityGroupMgmtAttrReq struct {
	SecurityGroups []BatchUpdateSGMgmtAttrItem `json:"security_groups" validate:"required,min=1"`
}

// Validate security group batch update management attribute request.
func (req BatchUpdateSecurityGroupMgmtAttrReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if len(req.SecurityGroups) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group count should <= %d", constant.BatchOperationMaxLimit)
	}

	for _, item := range req.SecurityGroups {
		if err := item.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// BatchUpdateSGMgmtAttrItem security group update management attribute item.
type BatchUpdateSGMgmtAttrItem struct {
	ID         string          `json:"id" validate:"required"`
	MgmtType   enumor.MgmtType `json:"mgmt_type"`
	MgmtBizID  int64           `json:"mgmt_biz_id"`
	Manager    string          `json:"manager"`
	BakManager string          `json:"bak_manager"`
	Vendor     enumor.Vendor   `json:"vendor"`
	CloudID    string          `json:"cloud_id"`
}

// Validate security group update management attribute item.
func (i BatchUpdateSGMgmtAttrItem) Validate() error {
	return validator.Validate.Struct(i)
}

// -------------------------- List --------------------------

// SecurityGroupListReq security group list req.
type SecurityGroupListReq struct {
	Field  []string           `json:"field" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate security group list request.
func (req *SecurityGroupListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SecurityGroupListResult define security group list result.
type SecurityGroupListResult struct {
	Count   uint64                    `json:"count,omitempty"`
	Details []cloud.BaseSecurityGroup `json:"details,omitempty"`
}

// SecurityGroupListResp define security group list resp.
type SecurityGroupListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *SecurityGroupListResult `json:"data"`
}

// SecurityGroupExtListResult define security group with extension list result.
type SecurityGroupExtListResult[T cloud.SecurityGroupExtension] struct {
	Count   uint64                   `json:"count,omitempty"`
	Details []cloud.SecurityGroup[T] `json:"details,omitempty"`
}

// SecurityGroupExtListResp define list resp.
type SecurityGroupExtListResp[T cloud.SecurityGroupExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *SecurityGroupExtListResult[T] `json:"data"`
}

// -------------------------- Delete --------------------------

// SecurityGroupBatchDeleteReq security group delete request.
type SecurityGroupBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate security group delete request.
func (req *SecurityGroupBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Get --------------------------

// SecurityGroupGetResp define security group get resp.
type SecurityGroupGetResp[T cloud.SecurityGroupExtension] struct {
	rest.BaseResp `json:",inline"`
	Data          *cloud.SecurityGroup[T] `json:"data"`
}

// CountSecurityGroupRuleReq ...
type CountSecurityGroupRuleReq struct {
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,min=1"`
}

// Validate list security group rule count req.
func (req *CountSecurityGroupRuleReq) Validate() error {
	return validator.Validate.Struct(req)
}

// CountSecurityGroupRuleResp define list security group rule count resp.
type CountSecurityGroupRuleResp = map[string]int64
