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

package typesasync

import (
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	tableasync "hcm/pkg/dal/table/async"
)

// ListAsyncFlows list async flows.
type ListAsyncFlows struct {
	Count   uint64                      `json:"count,omitempty"`
	Details []tableasync.AsyncFlowTable `json:"details,omitempty"`
}

// UpdateFlowInfo define update flow info.
type UpdateFlowInfo struct {
	ID     string             `json:"id" validate:"required"`
	Source enumor.FlowState   `json:"source" validate:"required"`
	Target enumor.FlowState   `json:"target" validate:"required"`
	Reason *tableasync.Reason `json:"reason" validate:"omitempty"`
	Worker *string            `json:"worker" validate:"omitempty"`
}

// Validate UpdateFlowInfo.
func (info *UpdateFlowInfo) Validate() error {
	return validator.Validate.Struct(info)
}
