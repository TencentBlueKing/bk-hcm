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
	"fmt"

	apicore "hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- Create --------------------------

// TCloudSecurityGroupCreateReq tcloud security group create request.
type TCloudSecurityGroupCreateReq struct {
	Region    string  `json:"region" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Memo      *string `json:"memo" validate:"omitempty"`
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id" validate:"required"`

	Tags []apicore.TagPair `json:"tags,omitempty"`

	MgmtType    enumor.MgmtType `json:"mgmt_type" validate:"required"`
	MgmtBizID   int64           `json:"mgmt_biz_id" validate:"required"`
	Manager     string          `json:"manager" validate:"required"`
	BakManager  string          `json:"bak_manager" validate:"required"`
	UsageBizIds []int64         `json:"usage_biz_ids" validate:"omitempty"`
}

// Validate tcloud security group create request.
func (req *TCloudSecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// HuaWeiSecurityGroupCreateReq tcloud security group create request.
type HuaWeiSecurityGroupCreateReq struct {
	Region    string  `json:"region" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Memo      *string `json:"memo" validate:"omitempty"`
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id" validate:"required"`

	MgmtType    enumor.MgmtType `json:"mgmt_type" validate:"required"`
	MgmtBizID   int64           `json:"mgmt_biz_id" validate:"required"`
	Manager     string          `json:"manager" validate:"required"`
	BakManager  string          `json:"bak_manager" validate:"required"`
	UsageBizIds []int64         `json:"usage_biz_ids" validate:"omitempty"`
}

// Validate tcloud security group create request.
func (req *HuaWeiSecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsSecurityGroupCreateReq tcloud security group create request.
type AwsSecurityGroupCreateReq struct {
	Region     string  `json:"region" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	Memo       *string `json:"memo" validate:"omitempty"`
	AccountID  string  `json:"account_id" validate:"required"`
	BkBizID    int64   `json:"bk_biz_id" validate:"required"`
	CloudVpcID string  `json:"cloud_vpc_id" validate:"required"`

	MgmtType    enumor.MgmtType   `json:"mgmt_type" validate:"required"`
	MgmtBizID   int64             `json:"mgmt_biz_id" validate:"required"`
	Manager     string            `json:"manager" validate:"required"`
	BakManager  string            `json:"bak_manager" validate:"required"`
	UsageBizIds []int64           `json:"usage_biz_ids" validate:"omitempty"`
	Tags        []apicore.TagPair `json:"tags,omitempty"`
}

// Validate tcloud security group create request.
func (req *AwsSecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureSecurityGroupCreateReq tcloud security group create request.
type AzureSecurityGroupCreateReq struct {
	Region            string  `json:"region" validate:"required"`
	Name              string  `json:"name" validate:"required"`
	Memo              *string `json:"memo" validate:"omitempty"`
	AccountID         string  `json:"account_id" validate:"required"`
	BkBizID           int64   `json:"bk_biz_id" validate:"required"`
	ResourceGroupName string  `json:"resource_group_name" validate:"required"`

	MgmtType    enumor.MgmtType   `json:"mgmt_type" validate:"required"`
	MgmtBizID   int64             `json:"mgmt_biz_id" validate:"required"`
	Manager     string            `json:"manager" validate:"required"`
	BakManager  string            `json:"bak_manager" validate:"required"`
	UsageBizIds []int64           `json:"usage_biz_ids" validate:"omitempty"`
	Tags        []apicore.TagPair `json:"tags,omitempty"`
}

// Validate tcloud security group create request.
func (req *AzureSecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// SecurityGroupUpdateReq security group update request.
type SecurityGroupUpdateReq struct {
	Name string  `json:"name" validate:"omitempty"`
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate tcloud security group update request.
func (req *SecurityGroupUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AzureSecurityGroupUpdateReq azure security group update request.
type AzureSecurityGroupUpdateReq struct {
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate azure security group update request.
func (req *AzureSecurityGroupUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Sync --------------------------

// SecurityGroupSyncReq security group sync request.
type SecurityGroupSyncReq struct {
	AccountID         string   `json:"account_id" validate:"required"`
	Region            string   `json:"region" validate:"omitempty"`
	ResourceGroupName string   `json:"resource_group_name" validate:"omitempty"`
	CloudIDs          []string `json:"cloud_ids" validate:"omitempty"`
}

// Validate security group sync request.
func (req *SecurityGroupSyncReq) Validate() error {
	if len(req.CloudIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("operate sync count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Associate --------------------------

// SecurityGroupAssociateCvmReq define security group bind cvm option.
type SecurityGroupAssociateCvmReq struct {
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	CvmID           string `json:"cvm_id" validate:"required"`
}

// Validate security group cvm bind option.
func (opt SecurityGroupAssociateCvmReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureSecurityGroupAssociateSubnetReq define security group bind subnet option.
type AzureSecurityGroupAssociateSubnetReq struct {
	SecurityGroupID string `json:"security_group_id" validate:"required"`
	SubnetID        string `json:"subnet_id" validate:"required"`
}

// Validate security group subnet bind option.
func (opt AzureSecurityGroupAssociateSubnetReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// AzureSecurityGroupAssociateNIReq define security group bind network interface option.
type AzureSecurityGroupAssociateNIReq struct {
	SecurityGroupID    string `json:"security_group_id" validate:"required"`
	NetworkInterfaceID string `json:"network_interface_id" validate:"required"`
}

// Validate security group network interface bind option.
func (opt AzureSecurityGroupAssociateNIReq) Validate() error {
	return validator.Validate.Struct(opt)
}

// ListSecurityGroupStatisticReq define tcloud list security group statistic request.
type ListSecurityGroupStatisticReq struct {
	SecurityGroupIDs []string `json:"security_group_ids" validate:"required"`
	Region           string   `json:"region" validate:"required"`
	AccountID        string   `json:"account_id" validate:"required"`
}

// Validate tcloud list security group statistic request.
func (req *ListSecurityGroupStatisticReq) Validate() error {
	if len(req.SecurityGroupIDs) == 0 {
		return fmt.Errorf("security group ids should not be empty")
	}
	if len(req.SecurityGroupIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("security group ids count should <= %d", constant.CloudResourceSyncMaxLimit)
	}
	return validator.Validate.Struct(req)
}

// ListSecurityGroupStatisticResp ...
type ListSecurityGroupStatisticResp struct {
	Details []*SecurityGroupStatisticItem `json:"details"`
}

// SecurityGroupStatisticItem ...
type SecurityGroupStatisticItem struct {
	ID        string                           `json:"id"`
	Resources []SecurityGroupStatisticResource `json:"resources"`
}

// SecurityGroupStatisticResource ...
type SecurityGroupStatisticResource struct {
	ResName string `json:"res_name"`
	Count   int64  `json:"count"`
}

// -------------------------- Clone --------------------------

// TCloudSecurityGroupCloneReq tcloud security group clone request.
type TCloudSecurityGroupCloneReq struct {
	SecurityGroupID string            `json:"security_group_id" validate:"required"`
	Manager         string            `json:"manager" validate:"required"`
	BakManager      string            `json:"bak_manager" validate:"required"`
	ManagementBizID int64             `json:"mgmt_biz_id" validate:"required"`
	Tags            []apicore.TagPair `json:"tags" validate:"omitempty"`
	TargetRegion    string            `json:"target_region" validate:"omitempty"`
	GroupName       string            `json:"group_name" validate:"required"`
}

// Validate tcloud security group clone request.
func (req *TCloudSecurityGroupCloneReq) Validate() error {
	return validator.Validate.Struct(req)
}
