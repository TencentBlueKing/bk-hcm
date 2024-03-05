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

// Package hcargstpl ...
package hcargstpl

import (
	"errors"

	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"

	vpc "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/vpc/v20170312"
)

// -------------------------- Delete --------------------------

// TCloudDeleteReq define delete req.
type TCloudDeleteReq struct {
	AccountID string `json:"account_id" validate:"required"`
	ID        string `json:"id" validate:"required"`
}

// Validate request.
func (req *TCloudDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}

// -------------------------- Create --------------------------

// TCloudCreateReq tcloud create req.
type TCloudCreateReq struct {
	BkBizID        int64               `json:"bk_biz_id" validate:"omitempty"`
	AccountID      string              `json:"account_id" validate:"required"`
	Vendor         string              `json:"vendor" validate:"required"`
	Name           string              `json:"name" validate:"required"`
	Type           enumor.TemplateType `json:"type" validate:"required"`
	Templates      []*TemplateInfo     `json:"templates" validate:"omitempty"`
	GroupTemplates []string            `json:"group_templates" validate:"omitempty"`
}

// TemplateInfo define argument template's template info.
type TemplateInfo struct {
	// ip地址、协议端口等
	Address *string `json:"address,omitempty" name:"address"`
	// 备注。
	Description *string `json:"description,omitempty" name:"description"`
}

// Validate request.
func (req *TCloudCreateReq) Validate() error {
	if len(req.Templates) == 0 && len(req.GroupTemplates) == 0 {
		return errors.New("templates or group_templates is required")
	}

	if (req.Type == enumor.AddressType || req.Type == enumor.ServiceType) && len(req.Templates) == 0 {
		return errors.New("templates is required")
	}

	if (req.Type == enumor.AddressGroupType || req.Type == enumor.ServiceGroupType) && len(req.GroupTemplates) == 0 {
		return errors.New("group_templates is required")
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// TCloudUpdateReq tcloud update req.
type TCloudUpdateReq struct {
	Vendor         string          `json:"vendor" validate:"required"`
	BkBizID        int64           `json:"bk_biz_id" validate:"omitempty"`
	Name           string          `json:"name" validate:"omitempty"`
	Templates      []*TemplateInfo `json:"templates" validate:"omitempty"`
	GroupTemplates []string        `json:"group_templates" validate:"omitempty"`
}

// Validate request.
func (req *TCloudUpdateReq) Validate() error {
	if len(req.Templates) == 0 && len(req.GroupTemplates) == 0 {
		return errors.New("templates or group_templates is required")
	}

	return validator.Validate.Struct(req)
}

// CreateResult ...
type CreateResult struct {
	TemplateID *string `json:"template_id"`
}

// CreateResp ...
type CreateResp struct {
	rest.BaseResp `json:",inline"`
	Data          *CreateResult `json:"data"`
}

// -------------------------- List --------------------------

// ArgsTplListReq list req.
type ArgsTplListReq struct {
	AccountID string              `json:"account_id" validate:"required"`
	Type      enumor.TemplateType `json:"type" validate:"required"`
	Page      *core.TCloudPage    `json:"page" validate:"required"`
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

// Validate list request.
func (opt *ArgsTplListReq) Validate() error {
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
