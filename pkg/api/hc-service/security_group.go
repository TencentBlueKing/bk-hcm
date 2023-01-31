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
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/vpc/v3/model"
	tcloud "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// -------------------------- Create --------------------------

// TCloudSecurityGroupCreateReq tcloud security group create request.
type TCloudSecurityGroupCreateReq struct {
	Region    string  `json:"region" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Memo      *string `json:"memo" validate:"omitempty"`
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id" validate:"required"`
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
}

// Validate tcloud security group create request.
func (req *HuaWeiSecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AwsSecurityGroupCreateReq tcloud security group create request.
type AwsSecurityGroupCreateReq struct {
	Region    string  `json:"region" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Memo      *string `json:"memo" validate:"omitempty"`
	AccountID string  `json:"account_id" validate:"required"`
	BkBizID   int64   `json:"bk_biz_id" validate:"required"`
	VpcID     string  `json:"vpc_id" validate:"omitempty"`
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
}

// Validate tcloud security group create request.
func (req *AzureSecurityGroupCreateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// SecurityGroupUpdateReq tcloud security group update request.
type SecurityGroupUpdateReq struct {
	Name string  `json:"name" validate:"omitempty"`
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate tcloud security group update request.
func (req *SecurityGroupUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Sync --------------------------

type SecurityGroupSyncReq struct {
	AccountID         string `json:"account_id" validate:"required"`
	Region            string `json:"region" validate:"omitempty"`
	ResourceGroupName string `json:"resource_group_name" validate:"omitempty"`
}

// Validate security group sync request.
func (req *SecurityGroupSyncReq) Validate() error {
	return validator.Validate.Struct(req)
}

type SecurityGroupSyncDS struct {
	IsUpdated       bool
	HcSecurityGroup cloud.BaseSecurityGroup
}

type SecurityGroupSyncHuaWeiDiff struct {
	SecurityGroup model.SecurityGroup
}

type SecurityGroupSyncTCloudDiff struct {
	SecurityGroup *tcloud.SecurityGroup
}

type SecurityGroupSyncAwsDiff struct {
	SecurityGroup *ec2.SecurityGroup
}

type SecurityGroupSyncAzureDiff struct {
	SecurityGroup *armnetwork.SecurityGroup
}
