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

package cloud

import (
	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// ListResUsageBizRelResult ...
type ListResUsageBizRelResult = core.ListResultT[corecloud.ResUsageBizRel]

// ResUsageBizRelUpdateReq 覆盖更新
type ResUsageBizRelUpdateReq struct {
	UsageBizIDs []int64       `json:"usage_biz_ids" validate:"required,max=3000"`
	ResCloudID  string        `json:"res_cloud_id" validate:"omitempty,max=255"`
	ResVendor   enumor.Vendor `json:"res_vendor" validate:"omitempty,max=64"`
}

// Validate ...
func (r *ResUsageBizRelUpdateReq) Validate() error {
	return validator.Validate.Struct(r)
}
