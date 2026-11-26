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

package region

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// RegionListReq ...
type RegionListReq struct {
	Filter *filter.Expression    `json:"filter" validate:"required"`
	Page   *core.PageWithoutSort `json:"page" validate:"required"`
}

// Validate ...
func (req *RegionListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// RegionBatchUpdateSyncEnableReq define region batch update sync_enable request.
type RegionBatchUpdateSyncEnableReq struct {
	IDs        []string `json:"ids" validate:"required,min=1"`
	SyncEnable *bool    `json:"sync_enable" validate:"required"`
}

// Validate validate RegionBatchUpdateSyncEnableReq.
func (req *RegionBatchUpdateSyncEnableReq) Validate() error {
	if len(req.IDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}
