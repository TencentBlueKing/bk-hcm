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

package core

import (
	"fmt"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
)

// BaseDeleteOption defines basic options to delete a cloud resource.
type BaseDeleteOption struct {
	ResourceID string `json:"resource_id"`
}

// Validate BaseDeleteOption.
func (b BaseDeleteOption) Validate() error {
	if len(b.ResourceID) == 0 {
		return errf.New(errf.InvalidParameter, "resource_id is required")
	}

	return nil
}

// BaseRegionalDeleteOption defines basic options to delete a regional cloud resource.
type BaseRegionalDeleteOption struct {
	BaseDeleteOption `json:",inline"`
	Region           string `json:"region"`
}

// Validate BaseRegionalDeleteOption.
func (b BaseRegionalDeleteOption) Validate() error {
	if err := b.BaseDeleteOption.Validate(); err != nil {
		return err
	}

	if len(b.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	return nil
}

// AzureDeleteOption defines basic options to delete an azure cloud resource.
type AzureDeleteOption struct {
	BaseDeleteOption  `json:",inline"`
	ResourceGroupName string `json:"resource_group_name"`
}

// Validate AzureDeleteOption.
func (a AzureDeleteOption) Validate() error {
	if err := a.BaseDeleteOption.Validate(); err != nil {
		return err
	}

	if len(a.ResourceGroupName) == 0 {
		return errf.New(errf.InvalidParameter, "resource group name is required")
	}

	return nil
}

// TCloudListOption defines basic tencent cloud list options.
type TCloudListOption struct {
	Region   string      `json:"region" validate:"required"`
	CloudIDs []string    `json:"cloud_ids" validate:"omitempty"`
	Page     *TCloudPage `json:"page" validate:"required"`
}

// Validate tcloud list option.
func (opt TCloudListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if len(opt.CloudIDs) > TCloudQueryLimit {
		return errf.New(errf.InvalidParameter, "tcloud resource ids length should <= 100")
	}

	if err := opt.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// AwsListOption defines basic aws list options.
type AwsListOption struct {
	Region   string   `json:"region" validate:"required"`
	CloudIDs []string `json:"cloud_ids" validate:"omitempty"`
	Page     *AwsPage `json:"page" validate:"omitempty"`
}

// Validate aws list option.
func (a AwsListOption) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.CloudIDs) != 0 {
		if a.Page != nil {
			return errf.New(errf.InvalidParameter, "only one of resource ids and page can be set")
		}

		if len(a.CloudIDs) > AwsQueryLimit {
			return errf.New(errf.InvalidParameter, "aws resource ids length should <= 1000")
		}

		return nil
	}

	if a.Page != nil {
		if err := a.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// GcpListOption defines basic gcp list options.
type GcpListOption struct {
	Page      *GcpPage `json:"page" validate:"required"`
	Zone      string   `json:"zone" validate:"omitempty"`
	CloudIDs  []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks []string `json:"self_links" validate:"omitempty"`
}

// Validate gcp list option.
func (a GcpListOption) Validate() error {

	if err := validator.Validate.Struct(a); err != nil {
		return err
	}

	if len(a.CloudIDs) > GcpQueryLimit {
		return errf.Newf(errf.InvalidParameter, "gcp resource ids length should <= %d", GcpQueryLimit)
	}

	if len(a.SelfLinks) > GcpQueryLimit {
		return errf.Newf(errf.InvalidParameter, "gcp resource self link length should <= %d", GcpQueryLimit)
	}

	if err := a.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// AzureListOption defines basic azure list options.
// TODO confirm resource group product form
type AzureListOption struct {
	// TODO 这是一个基础的option，不要随意添加个性化参数到这里，之后把  NetworkInterfaceName 去掉。
	ResourceGroupName    string   `json:"resource_group_name"`
	NetworkInterfaceName string   `json:"network_interface_name"`
	CloudIDs             []string `json:"cloud_ids" validate:"omitempty"`
	SelfLinks            []string `json:"self_links" validate:"omitempty"`
}

// Validate aws page.
func (a AzureListOption) Validate() error {
	if len(a.ResourceGroupName) == 0 {
		return errf.New(errf.InvalidParameter, "resource group name must be set")
	}

	return nil
}

// HuaWeiListOption defines basic huawei list options.
type HuaWeiListOption struct {
	Region   string      `json:"region" validate:"required"`
	CloudIDs []string    `json:"cloud_i_ds" validate:"omitempty"`
	Page     *HuaWeiPage `json:"page" validate:"required"`
}

// Validate huawei list option.
func (a HuaWeiListOption) Validate() error {
	if err := validator.Validate.Struct(a); err != nil {
		return nil
	}

	if len(a.CloudIDs) > HuaWeiQueryLimit {
		return errf.New(errf.InvalidParameter, "huawei resource ids length should <= 2000")
	}

	if err := a.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// AzureListByIDOption azure list by id option.
type AzureListByIDOption struct {
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	CloudIDs          []string `json:"cloud_ids" validate:"required"`
}

// Validate azure list by id option.
func (opt AzureListByIDOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return nil
	}

	if len(opt.CloudIDs) == 0 {
		return fmt.Errorf("cloud_ids shuold > 1")
	}

	return nil
}
