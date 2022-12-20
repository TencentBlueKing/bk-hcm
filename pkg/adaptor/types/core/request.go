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
	"hcm/pkg/criteria/errf"
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

// TCloudListOption defines basic tencent cloud list options.
type TCloudListOption struct {
	Region      string      `json:"region"`
	ResourceIDs []string    `json:"resource_ids,omitempty"`
	Page        *TCloudPage `json:"page,omitempty"`
}

// Validate tcloud list option.
func (t TCloudListOption) Validate() error {
	if len(t.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(t.ResourceIDs) != 0 {
		if t.Page != nil {
			return errf.New(errf.InvalidParameter, "only one of resource ids and page can be set")
		}

		if len(t.ResourceIDs) > TCloudQueryLimit {
			return errf.New(errf.InvalidParameter, "tcloud resource ids length should <= 100")
		}

		return nil
	}

	if t.Page == nil {
		return errf.New(errf.InvalidParameter, "one of tcloud resource ids and page is required")
	}

	if err := t.Page.Validate(); err != nil {
		return err
	}

	return nil
}

// AwsListOption defines basic aws list options.
type AwsListOption struct {
	Region      string   `json:"region"`
	ResourceIDs []string `json:"resource_ids,omitempty"`
	Page        *AwsPage `json:"page,omitempty"`
}

// Validate aws list option.
func (a AwsListOption) Validate() error {
	if len(a.Region) == 0 {
		return errf.New(errf.InvalidParameter, "region is required")
	}

	if len(a.ResourceIDs) != 0 {
		if a.Page != nil {
			return errf.New(errf.InvalidParameter, "only one of resource ids and page can be set")
		}

		if len(a.ResourceIDs) > AwsQueryLimit {
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
	ResourceIDs []string `json:"resource_ids,omitempty"`
	Page        *GcpPage `json:"page,omitempty"`
}

// Validate gcp list option.
func (a GcpListOption) Validate() error {
	if len(a.ResourceIDs) != 0 {
		if a.Page != nil {
			return errf.New(errf.InvalidParameter, "only one of resource ids and page can be set")
		}

		if len(a.ResourceIDs) > GcpQueryLimit {
			return errf.New(errf.InvalidParameter, "gcp resource ids length should <= 500")
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

// AzureListOption defines basic azure list options.
// TODO confirm resource group product form
type AzureListOption struct {
	ResourceGroupName string `json:"resource_group_name"`
}

// Validate aws page.
func (a AzureListOption) Validate() error {
	if len(a.ResourceGroupName) == 0 {
		return errf.New(errf.InvalidParameter, "resource group name must be set")
	}

	return nil
}

// HuaweiListOption defines basic huawei list options.
type HuaweiListOption struct {
	ResourceIDs []string    `json:"resource_ids,omitempty"`
	Page        *HuaweiPage `json:"page,omitempty"`
}

// Validate huawei list option.
func (a HuaweiListOption) Validate() error {
	if len(a.ResourceIDs) != 0 {
		if a.Page != nil {
			return errf.New(errf.InvalidParameter, "only one of resource ids and page can be set")
		}

		if len(a.ResourceIDs) > HuaweiQueryLimit {
			return errf.New(errf.InvalidParameter, "huawei resource ids length should <= 2000")
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
