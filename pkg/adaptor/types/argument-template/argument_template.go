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

package argstpl

import (
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/tools/converter"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// -------------------------- List Vpc Task Result --------------------------

// TCloudVpcTaskResultOption defines options to list tcloud vpc task result instances.
type TCloudVpcTaskResultOption struct {
	TaskID string `json:"task_id" validate:"required"`
}

// Validate tcloud vpc task result list option.
func (opt TCloudVpcTaskResultOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- List --------------------------

// TCloudListOption defines options to list tcloud argument template instances.
type TCloudListOption struct {
	Page *core.TCloudPage `json:"page" validate:"omitempty"`
	// 过滤条件。
	// ------ IP地址模版 ------
	// - address-template-name - IP地址模板名称。
	// - address-template-id - IP地址模板实例ID，例如：ipm-mdunqeb6。
	// - address-ip - IP地址。
	// ------ IP地址模版组 ------
	// - address-template-group-name - String - （过滤条件）IP地址模板集合名称。
	// - address-template-group-id - String - （过滤条件）IP地址模板实集合例ID，例如：ipmg-mdunqeb6。
	// ------ 协议端口模版 ------
	// - service-template-name - 协议端口模板名称。
	// - service-template-id - 协议端口模板实例ID，例如：ppm-e6dy460g。
	// - service-port- 协议端口。
	// ------ 协议端口模版集合 ------
	// - service-template-group-name - String - （过滤条件）协议端口模板集合名称。
	// - service-template-group-id - String - （过滤条件）协议端口模板集合实例ID，例如：ppmg-e6dy460g。
	Filters []*vpc.Filter `json:"Filters,omitempty" name:"Filters"`
}

// Validate tcloud argument template list option.
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

// -------------------------- Delete --------------------------

// TCloudDeleteOption defines options to operation tcloud argument template instances.
type TCloudDeleteOption struct {
	CloudID string `json:"cloud_id" validate:"required"`
}

// Validate tcloud argument template operation option.
func (opt TCloudDeleteOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create Address --------------------------

// TCloudCreateAddressOption defines options to create tcloud argument template address instances.
type TCloudCreateAddressOption struct {
	TemplateName   string             `json:"template_name" validate:"required"`
	AddressesExtra []*vpc.AddressInfo `json:"addresses_extra" validate:"required"`
}

// Validate tcloud argument template operation option.
func (opt TCloudCreateAddressOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Address --------------------------

// TCloudUpdateAddressOption defines options to update tcloud argument template address instances.
type TCloudUpdateAddressOption struct {
	TemplateID     string             `json:"template_id" validate:"required"`
	TemplateName   string             `json:"template_name" validate:"omitempty"`
	AddressesExtra []*vpc.AddressInfo `json:"addresses_extra" validate:"omitempty"`
}

// Validate tcloud argument template operation option.
func (opt TCloudUpdateAddressOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create Address Group --------------------------

// TCloudCreateAddressGroupOption defines options to create tcloud argument template address group instances.
type TCloudCreateAddressGroupOption struct {
	TemplateGroupName string   `json:"template_group_name" validate:"required"`
	TemplateIDs       []string `json:"template_ids" validate:"required"`
}

// Validate tcloud argument template operation option.
func (opt TCloudCreateAddressGroupOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Address Group --------------------------

// TCloudUpdateAddressGroupOption defines options to update tcloud argument template address group instances.
type TCloudUpdateAddressGroupOption struct {
	TemplateGroupID   string   `json:"template_group_id" validate:"required"`
	TemplateGroupName string   `json:"template_group_name" validate:"omitempty"`
	TemplateIDs       []string `json:"template_ids" validate:"omitempty"`
}

// Validate tcloud argument template operation option.
func (opt TCloudUpdateAddressGroupOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create Service Port --------------------------

// TCloudCreateServiceOption defines options to create tcloud argument template service instances.
type TCloudCreateServiceOption struct {
	TemplateName  string              `json:"template_name" validate:"required"`
	ServicesExtra []*vpc.ServicesInfo `json:"services_extra" validate:"required"`
}

// Validate tcloud argument template operation option.
func (opt TCloudCreateServiceOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Service Port --------------------------

// TCloudUpdateServiceOption defines options to update tcloud argument template service instances.
type TCloudUpdateServiceOption struct {
	TemplateID    string              `json:"template_id" validate:"required"`
	TemplateName  string              `json:"template_name" validate:"omitempty"`
	ServicesExtra []*vpc.ServicesInfo `json:"services_extra" validate:"omitempty"`
}

// Validate tcloud argument template operation option.
func (opt TCloudUpdateServiceOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Create Service Port Group --------------------------

// TCloudCreateServiceGroupOption defines options to create tcloud argument template service group instances.
type TCloudCreateServiceGroupOption struct {
	TemplateGroupName string   `json:"template_group_name" validate:"required"`
	TemplateIDs       []string `json:"template_ids" validate:"required"`
}

// Validate tcloud argument template operation option.
func (opt TCloudCreateServiceGroupOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// -------------------------- Update Service Port Group --------------------------

// TCloudUpdateServiceGroupOption defines options to update tcloud argument template service group instances.
type TCloudUpdateServiceGroupOption struct {
	TemplateGroupID   string   `json:"template_group_id" validate:"required"`
	TemplateGroupName string   `json:"template_group_name" validate:"omitempty"`
	TemplateIDs       []string `json:"template_ids" validate:"omitempty"`
}

// Validate tcloud argument template operation option.
func (opt TCloudUpdateServiceGroupOption) Validate() error {
	return validator.Validate.Struct(opt)
}

// TCloudArgsTplAddress for argument template address Instance
type TCloudArgsTplAddress struct {
	*vpc.AddressTemplate
}

// GetCloudID ...
func (at TCloudArgsTplAddress) GetCloudID() string {
	return converter.PtrToVal(at.AddressTemplateId)
}

// TCloudArgsTplAddressGroup for argument template address group Instance
type TCloudArgsTplAddressGroup struct {
	*vpc.AddressTemplateGroup
}

// GetCloudID ...
func (at TCloudArgsTplAddressGroup) GetCloudID() string {
	return converter.PtrToVal(at.AddressTemplateGroupId)
}

// TCloudArgsTplService for argument template service Instance
type TCloudArgsTplService struct {
	*vpc.ServiceTemplate
}

// GetCloudID ...
func (at TCloudArgsTplService) GetCloudID() string {
	return converter.PtrToVal(at.ServiceTemplateId)
}

// TCloudArgsTplServiceGroup for argument template service group Instance
type TCloudArgsTplServiceGroup struct {
	*vpc.ServiceTemplateGroup
}

// GetCloudID ...
func (at TCloudArgsTplServiceGroup) GetCloudID() string {
	return converter.PtrToVal(at.ServiceTemplateGroupId)
}
