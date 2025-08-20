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
	apicore "hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// -------------------------- Create --------------------------

// TCloudCreateOption define security group create option.
type TCloudCreateOption struct {
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"required,lte=60"`
	Description *string `json:"description" validate:"omitempty,lte=100"`

	Tags []apicore.TagPair `json:"tags" validate:"omitempty"`
}

// Validate security group create option.
func (opt TCloudCreateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update --------------------------

// TCloudUpdateOption define tcloud security group update option.
type TCloudUpdateOption struct {
	CloudID     string  `json:"cloud_id" validate:"required"`
	Region      string  `json:"region" validate:"required"`
	Name        string  `json:"name" validate:"omitempty,lte=60"`
	Description *string `json:"description" validate:"omitempty,lte=100"`
}

// Validate security group update option.
func (opt TCloudUpdateOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// TCloudListOption define tcloud security group list option.
type TCloudListOption struct {
	Region   string           `json:"region" validate:"required"`
	CloudIDs []string         `json:"cloud_ids" validate:"omitempty"`
	Page     *core.TCloudPage `json:"page" validate:"omitempty"`

	TagFilters apicore.MultiValueTagMap `json:"tag_filters"`
}

// Validate tcloud security group list option.
func (opt TCloudListOption) Validate() error {
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

// -------------------------- Associate --------------------------

// TCloudAssociateCvmOption define security group bind cvm option.
type TCloudAssociateCvmOption struct {
	Region               string `json:"region" validate:"required"`
	CloudSecurityGroupID string `json:"cloud_security_group_id" validate:"required"`
	CloudCvmID           string `json:"cloud_cvm_id" validate:"required"`
}

// Validate security group cvm bind option.
func (opt TCloudAssociateCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudBatchAssociateCvmOption define security group bind cvm option.
type TCloudBatchAssociateCvmOption struct {
	Region               string   `json:"region" validate:"required"`
	CloudSecurityGroupID string   `json:"cloud_security_group_id" validate:"required"`
	CloudCvmIDs          []string `json:"cloud_cvm_ids" validate:"required"`
}

// Validate security group cvm bind option.
func (opt TCloudBatchAssociateCvmOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Delete --------------------------

// TCloudDeleteOption tcloud security group delete option.
type TCloudDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
	Region  string `json:"region" validate:"required"`
}

// Validate security group delete option.
func (opt TCloudDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudSG for vpc SecurityGroup
type TCloudSG struct {
	*vpc.SecurityGroup
}

// GetCloudID ...
func (sg TCloudSG) GetCloudID() string {
	return converter.PtrToVal(sg.SecurityGroupId)
}

// TCloudSecurityGroupAssociationStatistic for vpc SecurityGroupAssociationStatistics
type TCloudSecurityGroupAssociationStatistic struct {
	*vpc.SecurityGroupAssociationStatistics
}

// TCloudSecurityGroupCloneOption ...
type TCloudSecurityGroupCloneOption struct {
	GroupName       string            `json:"group_name" validate:"required,lte=60"`
	Region          string            `json:"region" validate:"required"`
	RemoteRegion    string            `json:"remote_region" validate:"omitempty"`
	SecurityGroupID string            `json:"security_group_id" validate:"required"`
	Tags            []apicore.TagPair `json:"tags" validate:"omitempty"`
}

// Validate security group clone option.
func (opt TCloudSecurityGroupCloneOption) Validate() error {
	return validator.Validate.Struct(opt)
}
