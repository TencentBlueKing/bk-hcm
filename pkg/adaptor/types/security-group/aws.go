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

package securitygroup

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	"github.com/aws/aws-sdk-go/service/ec2"
)

// -------------------------- Create --------------------------

// AwsCreateOption define aws security group create option.
type AwsCreateOption struct {
	Region      string  `json:"region" validate:"required"`
	CloudVpcID  string  `json:"cloud_vpc_id" validate:"omitempty"`
	Name        string  `json:"name" validate:"required,lte=255"`
	Description *string `json:"description" validate:"omitempty,lte=255"`
}

// Validate aws security group create option.
func (opt AwsCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// AwsListOption define aws security group list option.
type AwsListOption struct {
	Region   string        `json:"region" validate:"required"`
	CloudIDs []string      `json:"cloud_ids" validate:"omitempty"`
	Page     *core.AwsPage `json:"page" validate:"omitempty"`
}

// Validate security group list option.
func (opt AwsListOption) Validate() error {
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

// -------------------------- Delete --------------------------

// AwsDeleteOption aws security group delete option.
type AwsDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate security group delete option.
func (opt AwsDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Associate --------------------------

// AwsAssociateCvmOption define security group bind cvm option.
type AwsAssociateCvmOption struct {
	Region               string `json:"region" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
	CloudCvmID           string `json:"cloud_cvm_id" validate:"required"`
}

// Validate security group cvm bind option.
func (opt AwsAssociateCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// AwsSG for ec2 SecurityGroup
type AwsSG struct {
	*ec2.SecurityGroup
}

// GetCloudID ...
func (sg AwsSG) GetCloudID() string {
	return converter.PtrToVal(sg.GroupId)
}
