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

package itsm

import (
	"hcm/pkg/criteria/validator"
)

const (
	success = 20000
)

// OperatorType 审批操作人类型
type OperatorType string

const (
	// OperatorUser 用户操作审批
	OperatorUser OperatorType = "user"
)

// ----------------------------- create ticket -----------------------------

// CreateTicketReq create ticket request
type CreateTicketReq struct {
	SystemID      string                 `json:"system_id"`
	WorkflowKey   string                 `json:"workflow_key" validate:"required"`
	Operator      string                 `json:"operator" validate:"required"`
	FormData      map[string]interface{} `json:"form_data" validate:"required"`
	CallbackURL   string                 `json:"callback_url"`
	CallbackToken string                 `json:"callback_token"`
}

// Validate CreateTicketReq validate
func (c *CreateTicketReq) Validate() error {
	return validator.Validate.Struct(c)
}

// CreateTicketResult create ticket result
type CreateTicketResult struct {
	ID string `json:"id"`
	// SN Notice：v4版本的SN和v3不是一个含义，虽然我们本地依然沿用SN的叫法，但是在v4版本中，SN并不使用，而是用ticket_id进行查询
	SN          string `json:"sn"`
	FrontendURL string `json:"frontend_url"`
}

// ----------------------------- revoke ticket -----------------------------

// RevokeTicketReq revoke ticket request
type RevokeTicketReq struct {
	SystemID string `json:"system_id" validate:"required"`
	TicketID string `json:"ticket_id" validate:"required"`
}

// Validate RevokeTicketReq validate
func (r *RevokeTicketReq) Validate() error {
	return validator.Validate.Struct(r)
}

// RevokeTicketResult revoke ticket result
type RevokeTicketResult struct {
	Result bool `json:"result"`
}

// ----------------------------- approve ticket -----------------------------

// ApproveTasksReq get approve tasks request
type ApproveTasksReq struct {
	TicketID    string `json:"ticket_id" validate:"required"`
	ActivityKey string `json:"activity_key" validate:"required"`
}

// Validate GetApproveTasksReq validate
func (r *ApproveTasksReq) Validate() error {
	return validator.Validate.Struct(r)
}

// ApproveTasksResult get approve tasks result
type ApproveTasksResult struct {
	Items []struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		ActivityKey   string `json:"activity_key"`
		Desc          string `json:"desc"`
		Type          string `json:"type"`
		Status        string `json:"status"`
		StatusDisplay string `json:"status_display"`
	} `json:"items"`
}

// HandleApproveReq define handle approve req.
type HandleApproveReq struct {
	SystemID     string       `json:"system_id" validate:"required"`
	TicketID     string       `json:"ticket_id" validate:"required"`
	TaskID       string       `json:"task_id" validate:"required"`
	OperatorType OperatorType `json:"operator_type" validate:"required"`
	Operator     string       `json:"operator" validate:"required"`
	Action       string       `json:"action" validate:"required"`
	Remark       string       `json:"remark"`
}

// Validate HandleApproveReq validate.
func (r *HandleApproveReq) Validate() error {
	return validator.Validate.Struct(r)
}

// HandleApproveResult define handle approve result.
type HandleApproveResult struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}
