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

package cloudserver

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// ListWithCvmReq ...
type ListWithCvmReq struct {
	Fields []string           `json:"fields" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
	// NotEqualDiskID 关联关系表中 disk_id 不等于该值的查询条件。
	NotEqualDiskID string `json:"not_equal_disk_id" validate:"omitempty"`
}

// Validate ...
func (req *ListWithCvmReq) Validate() error {
	return validator.Validate.Struct(req)
}
