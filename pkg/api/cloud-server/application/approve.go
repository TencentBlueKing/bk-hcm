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

package application

import (
	"hcm/pkg/criteria/validator"
)

// ItsmApproveResult ...
type ItsmApproveResult struct {
	CallbackToken string                  `json:"callback_token" validate:"required"`
	Ticket        ItsmApproveTicketDetail `json:"ticket" validate:"required"`
}

// Validate ...
func (i *ItsmApproveResult) Validate() error {
	return validator.Validate.Struct(i)
}

// ItsmApproveTicketDetail ...
type ItsmApproveTicketDetail struct {
	ID            string  `json:"id" validate:"required"`
	SN            string  `json:"sn"`
	Title         string  `json:"title" validate:"required"`
	ApproveResult *bool   `json:"approve_result" validate:"required"`
	EndAt         *string `json:"end_at"`
	Status        string  `json:"status" validate:"required"`
	WorkflowID    string  `json:"workflow_id" validate:"required"`
}

// Validate ...
func (i *ItsmApproveTicketDetail) Validate() error {
	return validator.Validate.Struct(i)
}
