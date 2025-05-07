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

package cloudserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/esb/cmdb"
)

// -------------------------- Create --------------------------

// SecurityGroupCreateReq security group create request.
type SecurityGroupCreateReq struct {
	Vendor    enumor.Vendor   `json:"vendor" validate:"required"`
	AccountID string          `json:"account_id" validate:"required"`
	Region    string          `json:"region" validate:"required"`
	Name      string          `json:"name" validate:"required"`
	Memo      *string         `json:"memo" validate:"omitempty"`
	Extension json.RawMessage `json:"extension" validate:"omitempty"`

	Manager     string         `json:"manager" validate:"required"`
	BakManager  string         `json:"bak_manager" validate:"required"`
	UsageBizIds []int64        `json:"usage_biz_ids" validate:"omitempty"`
	Tags        []core.TagPair `json:"tags,omitempty"`
	// internal fields
	MgmtType  enumor.MgmtType `json:"-" validate:"omitempty"`
	MgmtBizID int64           `json:"-" validate:"omitempty"`
}

// Validate security group create request.
func (req *SecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsSecurityGroupExtensionCreate ...
type AwsSecurityGroupExtensionCreate struct {
	CloudVpcID string `json:"cloud_vpc_id" validate:"omitempty"`
}

// AzureSecurityGroupExtensionCreate ...
type AzureSecurityGroupExtensionCreate struct {
	ResourceGroupName string `json:"resource_group_name" validate:"omitempty"`
}

// -------------------------- List --------------------------

// SecurityGroupListReq security group list req.
type SecurityGroupListReq struct {
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

// SecurityGroupQueryRelatedResourceCountReq security group query related resource count req.
type SecurityGroupQueryRelatedResourceCountReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate security group query related resource count request.
func (req *SecurityGroupQueryRelatedResourceCountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListSGRelBusinessResp response data of list security group related business
type ListSGRelBusinessResp struct {
	CVM          []ListSGRelBusinessItem `json:"cvm"`
	LoadBalancer []ListSGRelBusinessItem `json:"load_balancer"`
}

// ListSGRelBusinessItem item of list security group related business
type ListSGRelBusinessItem struct {
	BkBizID  int64 `json:"bk_biz_id"`
	ResCount int64 `json:"res_count"`
}

// -------------------------- Update --------------------------

// SecurityGroupUpdateReq security group update request.
type SecurityGroupUpdateReq struct {
	Name string  `json:"name"`
	Memo *string `json:"memo"`
}

// Validate security group update request.
func (req *SecurityGroupUpdateReq) Validate() error {
	if len(req.Name) == 0 && req.Memo == nil {
		return errors.New("name or memo is required")
	}

	if len(req.Name) != 0 {
		if err := validator.ValidateSecurityGroupName(req.Name); err != nil {
			return err
		}
	}

	if req.Memo != nil {
		if err := validator.ValidateSecurityGroupMemo(req.Memo); err != nil {
			return err
		}
	}

	return nil
}

// AssignSecurityGroupToBizReq define assign security group to biz req.
type AssignSecurityGroupToBizReq struct {
	BkBizID          int64    `json:"bk_biz_id" validate:"required"`
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required"`
}

// Validate assign security group to biz request.
func (req *AssignSecurityGroupToBizReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should >= 0")
	}

	if len(req.SecurityGroupIDs) == 0 {
		return errors.New("security group ids is required")
	}

	if len(req.SecurityGroupIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("security group ids should <= %d", constant.BatchOperationMaxLimit)
	}

	return nil
}

// BatchAssignBizReq define batch assign security group to biz req.
type BatchAssignBizReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100,unique"`
}

// Validate assign security group to biz request.
func (req *BatchAssignBizReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AssignBizPreviewResp define batch assign biz preview response.
type AssignBizPreviewResp struct {
	ID            string `json:"id"`
	Assignable    bool   `json:"assignable"`
	Reason        string `json:"reason"`
	AssignedBizID int64  `json:"assigned_biz_id"`
}

// SecurityGroupUpdateMgmtAttrReq security group update management attribute request.
type SecurityGroupUpdateMgmtAttrReq struct {
	MgmtType    enumor.MgmtType `json:"mgmt_type"`
	Manager     string          `json:"manager"`
	BakManager  string          `json:"bak_manager"`
	UsageBizIDs []int64         `json:"usage_biz_ids"`
	MgmtBizID   int64           `json:"mgmt_biz_id"`
}

// Validate security group update management attribute request.
func (req SecurityGroupUpdateMgmtAttrReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if req.MgmtType != "" {
		if err := req.MgmtType.Validate(); err != nil {
			return err
		}
	}

	if req.MgmtBizID == constant.UnassignedBiz {
		return errors.New("mgmt_biz_id should not be update to unassigned")
	}

	// 平台管理不可修改管理业务
	if req.MgmtType == enumor.MgmtTypePlatform {
		if req.MgmtBizID != constant.UnassignedBiz && req.MgmtBizID != 0 {
			return errors.New("platform security group can't be assigned to a management business")
		}
	}

	// 使用业务如果包含-1，则必须只能有-1
	if slices.Contains(req.UsageBizIDs, constant.AttachedAllBiz) && len(req.UsageBizIDs) > 1 {
		return errors.New("usage business has included all [-1], can not specify to include other business")
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
		return fmt.Errorf("update data cannot exceed %d items", constant.BatchOperationMaxLimit)
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
	ID         string `json:"id" validate:"required"`
	Manager    string `json:"manager" validate:"required"`
	BakManager string `json:"bak_manager" validate:"required"`
	MgmtBizID  int64  `json:"mgmt_biz_id" validate:"required"`
}

// Validate security group update management attribute item.
func (i BatchUpdateSGMgmtAttrItem) Validate() error {
	if i.MgmtBizID == constant.UnassignedBiz {
		return errors.New("mgmt_biz_id should not be update to unassigned")
	}

	return validator.Validate.Struct(i)
}

// -------------------------- Delete --------------------------

// SecurityGroupBatchDeleteReq security group update request.
type SecurityGroupBatchDeleteReq struct {
	IDs []string `json:"ids" validate:"required,max=100"`
}

// Validate security group delete request.
func (req *SecurityGroupBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Associate --------------------------

// SecurityGroupAssociateCvmReq define security group associate cvm option.
type SecurityGroupAssociateCvmReq struct {
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	CvmID           string `json:"cvm_id" validate:"required"`
}

// Validate security group associate cvm request.
func (req *SecurityGroupAssociateCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SecurityGroupAssociateSubnetReq define security group associate subnet option.
type SecurityGroupAssociateSubnetReq struct {
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	SubnetID        string `json:"subnet_id" validate:"required"`
}

// Validate security group associate subnet request.
func (req *SecurityGroupAssociateSubnetReq) Validate() error {
	return validator.Validate.Struct(req)
}

// SecurityGroupAssociateNIReq define security group associate network interface option.
type SecurityGroupAssociateNIReq struct {
	SecurityGroupID    string `json:"security_group_id" validate:"required"`
	NetworkInterfaceID string `json:"network_interface_id" validate:"required"`
}

// Validate security group associate network interface request.
func (req *SecurityGroupAssociateNIReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Get --------------------------

// SecurityGroup define security group
type SecurityGroup[Extension cloud.SecurityGroupExtension] struct {
	cloud.BaseSecurityGroup `json:",inline"`
	CvmCount                uint64     `json:"cvm_count"`
	NetworkInterfaceCount   uint64     `json:"network_interface_count"`
	SubnetCount             uint64     `json:"subnet_count"`
	Extension               *Extension `json:"extension"`
}

// BatchGetResRelatedSecurityGroupsReq ...
type BatchGetResRelatedSecurityGroupsReq struct {
	ResIDs []string `json:"res_ids" validate:"required,min=1,max=500"`
}

// Validate ...
func (req BatchGetResRelatedSecurityGroupsReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ResSGRel ...
type ResSGRel struct {
	ResID          string   `json:"res_id"`
	SecurityGroups []SGInfo `json:"security_groups"`
}

// SGInfo ...
type SGInfo struct {
	ID      string `json:"id"`
	CloudId string `json:"cloud_id"`
	Name    string `json:"name"`
}

// -------------------------- Clone --------------------------

// SecurityGroupCloneReq security group clone req.
type SecurityGroupCloneReq struct {
	Name         *string `json:"name" validate:"omitempty,min=1"`
	Manager      string  `json:"manager" validate:"required"`
	BakManager   string  `json:"bak_manager" validate:"required"`
	TargetRegion string  `json:"target_region" validate:"omitempty"`
}

// Validate ...
func (req *SecurityGroupCloneReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListSGMaintainerInfoReq define list security group usage biz maintainer request.
type ListSGMaintainerInfoReq struct {
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required,min=1,max=500"`
}

// Validate ...
func (req *ListSGMaintainerInfoReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListSGMaintainerInfoResult define list security group usage biz maintainer result.
type ListSGMaintainerInfoResult struct {
	ID            string     `json:"id"`
	Managers      []string   `json:"managers"`
	UsageBizInfos []cmdb.Biz `json:"usage_biz_infos"`
}
