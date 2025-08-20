/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package account

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
)

// ResCondSyncReq sync condition
type ResCondSyncReq struct {
	Regions  []string `json:"regions,required" validate:"min=1,max=5"`
	CloudIDs []string `json:"cloud_ids,omitempty" validate:"max=20"`

	TagFilters core.MultiValueTagMap `json:"tag_filters,omitempty" validate:"max=5"`
}

// Validate ...
func (r *ResCondSyncReq) Validate() error {
	if len(r.CloudIDs) > 0 {
		if len(r.Regions) > 1 {
			return fmt.Errorf("regions must be one when cloud_ids is specified, got: %v", r.Regions)
		}
	}
	return validator.Validate.Struct(r)
}

// AzureResCondSyncReq azure resource condition sync request
type AzureResCondSyncReq struct {
	ResourceGroupNames []string `json:"resource_group_names,required" validate:"min=1,max=5"`
	CloudIDs           []string `json:"cloud_ids,omitempty" validate:"max=20"`
}

// Validate ...
func (r *AzureResCondSyncReq) Validate() error {
	if len(r.CloudIDs) > 0 {
		if len(r.ResourceGroupNames) > 1 {
			return fmt.Errorf("resource_group_names must be one when cloud_ids is specified, got: %v", r.ResourceGroupNames)
		}
	}
	return validator.Validate.Struct(r)
}
