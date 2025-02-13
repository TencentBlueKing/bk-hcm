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

	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
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

// -------------------------- Delete --------------------------

// SecurityGroupBatchDeleteReq security group update request.
type SecurityGroupBatchDeleteReq struct {
	IDs []string `json:"ids" validate:"required"`
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
