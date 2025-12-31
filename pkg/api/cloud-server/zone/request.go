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

package zone

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// ZoneListReq ...
type ZoneListReq struct {
	Filter *filter.Expression `json:"filter"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate ...
func (req *ZoneListReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ZoneImportReq define zone import request.
type ZoneImportReq struct {
	CloudID   string      `json:"cloud_id" validate:"required"`
	Name      string      `json:"name" validate:"required"`
	State     string      `json:"state" validate:"required"`
	Region    string      `json:"region" validate:"required"`
	NameCn    string      `json:"name_cn" validate:"required"`
	Extension interface{} `json:"extension"`
}

// Validate validate ZoneImportReq.
func (req *ZoneImportReq) Validate() error {
	return validator.Validate.Struct(req)
}
