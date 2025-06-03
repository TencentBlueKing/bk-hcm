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

package webserver

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/thirdparty/api-gateway/itsm"
)

// ListMyApprovalTicketReq define get tickets by user req.
type ListMyApprovalTicketReq struct {
	Page core.PageWithoutSort `json:"page" validate:"required"`
}

// Validate ListMyApprovalTicketReq.
func (req *ListMyApprovalTicketReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListMyApprovalTicketResp define list my approval ticket resp.
type ListMyApprovalTicketResp struct {
	Count   int64         `json:"count"`
	Details []itsm.Ticket `json:"details"`
}

// TicketApproveReq define ticket approval req.
type TicketApproveReq struct {
	Sn          string `json:"sn" validate:"required"`
	ActivityKey string `json:"activity_key" validate:"required"`
	// StateID int                  `json:"state_id" validate:"required"`
	Action TicketApprovalAction `json:"action" validate:"required"`
	Memo   string               `json:"memo" validate:"omitempty"`
}

// Validate TicketApproveReq.
func (req *TicketApproveReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	if err := req.Action.Validate(); err != nil {
		return err
	}

	return nil
}

// TicketApprovalAction 单据审批动作。
type TicketApprovalAction string

// Validate TicketApprovalAction.
func (action TicketApprovalAction) Validate() error {
	switch action {
	case Pass:
	case Refuse:
	default:
		return fmt.Errorf("action: %s not support", action)
	}

	return nil
}

// ToItsmAction to itsm action.
func (action TicketApprovalAction) ToItsmAction() string {
	switch action {
	case Pass:
		return "approve"
	case Refuse:
		return "refuse"
	default:
		return ""
	}
}

const (
	// Pass 通过
	Pass TicketApprovalAction = "pass"
	// Refuse 拒绝
	Refuse TicketApprovalAction = "refuse"
)
