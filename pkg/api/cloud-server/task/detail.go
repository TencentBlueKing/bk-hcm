/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Detail System) available.
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

package task

import (
	"hcm/pkg/criteria/validator"
)

// DetailStateCountReq task detail state count request.
type DetailStateCountReq struct {
	IDs []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate DetailStateSummary.
func (req DetailStateCountReq) Validate() error {
	return validator.Validate.Struct(req)
}

// DetailStateCountResult defines task detail state count result.
type DetailStateCountResult struct {
	Details []DetailStateSummary `json:"details"`
}

// DetailStateSummary defines task detail state summary.
type DetailStateSummary struct {
	ID      string `json:"id"`
	Success int    `json:"success"`
	Failed  int    `json:"failed"`
	Init    int    `json:"init"`
	Running int    `json:"running"`
	Cancel  int    `json:"cancel"`
	Total   int    `json:"total"`
}
