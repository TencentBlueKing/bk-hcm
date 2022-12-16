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

import "hcm/pkg/criteria/validator"

// -------------------------- Create --------------------------

// SecurityGroupCreateReq security group create request.
type SecurityGroupCreateReq[Attachment SecurityGroupAttachment] struct {
	Spec       *SecurityGroupSpecCreateReq `json:"spec" validate:"required"`
	Attachment *Attachment                 `json:"attachment" validate:"required"`
}

// Validate security group create request.
func (req *SecurityGroupCreateReq[T]) Validate() error {
	return validator.Validate.Struct(req)
}

// SecurityGroupSpecCreateReq define security group spec when create.
type SecurityGroupSpecCreateReq struct {
	Region    string  `json:"region" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Memo      *string `json:"memo" validate:"omitempty"`
	AccountID string  `json:"account_id" validate:"required"`
}

// SecurityGroupAttachment define security group attachment.
type SecurityGroupAttachment interface {
	BaseSecurityGroupAttachment | AwsSecurityGroupAttachment | AzureSecurityGroupAttachment
}

// BaseSecurityGroupAttachment define base security group attachment.
type BaseSecurityGroupAttachment struct {
	BkBizID uint64 `json:"bk_biz_id" validate:"omitempty"`
}

// AwsSecurityGroupAttachment define aws security group attachment.
type AwsSecurityGroupAttachment struct {
	BkBizID uint64 `json:"bk_biz_id" validate:"omitempty"`
	VpcID   string `json:"vpc_id" validate:"omitempty"`
}

// AzureSecurityGroupAttachment define azure security group attachment.
type AzureSecurityGroupAttachment struct {
	ResourceGroupName string `json:"resource_group_name" validate:"required"`
	BkBizID           uint64 `json:"bk_biz_id" validate:"omitempty"`
}

// -------------------------- Update --------------------------

// SecurityGroupUpdateReq tcloud security group update request.
type SecurityGroupUpdateReq struct {
	Spec *SecurityGroupSpecUpdateReq `json:"spec" validate:"required"`
}

// SecurityGroupSpecUpdateReq define security group spec when update.
type SecurityGroupSpecUpdateReq struct {
	Name string  `json:"name" validate:"omitempty"`
	Memo *string `json:"memo" validate:"omitempty"`
}

// Validate tcloud security group update request.
func (req *SecurityGroupUpdateReq) Validate() error {
	return validator.Validate.Struct(req)
}
