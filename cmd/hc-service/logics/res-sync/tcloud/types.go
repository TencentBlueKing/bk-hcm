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

package tcloud

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
)

// SyncBaseParams ...
type SyncBaseParams struct {
	AccountID  string                `json:"account_id" validate:"required"`
	Region     string                `json:"region" validate:"required"`
	CloudIDs   []string              `json:"cloud_ids" validate:"required,min=1"`
	TagFilters core.MultiValueTagMap `json:"tag_filters,omitempty" validate:"max=10"`
}

// Validate ...
func (opt SyncBaseParams) Validate() error {

	if len(opt.CloudIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("cloudIDs shuold <= %d", constant.CloudResourceSyncMaxLimit)
	}

	return validator.Validate.Struct(opt)
}

// SyncResult sync result.
type SyncResult struct {
	CreatedIds []string
}

// SyncRemovedParams ...
type SyncRemovedParams struct {
	AccountID string `json:"account_id" validate:"required"`
	Region    string `json:"region" validate:"required"`
	// 为空表示所有
	CloudIDs   []string              `json:"cloud_ids,omitempty" validate:"omitempty"`
	TagFilters core.MultiValueTagMap `json:"tag_filters,omitempty"`
}

// Validate ...
func (opt SyncRemovedParams) Validate() error {

	if len(opt.CloudIDs) > constant.CloudResourceSyncMaxLimit {
		return fmt.Errorf("cloudIDs shuold <= %d", constant.CloudResourceSyncMaxLimit)
	}
	return validator.Validate.Struct(opt)
}

func getDatabaseTagKey(k string) string {
	return "tags." + k
}
