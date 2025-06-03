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
	"fmt"
	"strconv"

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway"
)

// ViewType 工单类型。
type ViewType string

const (
	// MyTODO 用户的待办单据
	MyTODO = "my_todo"
	// MyCreated 用户创建的单据
	MyCreated = "my_created"
	// MyHistory 用户的历史单据
	MyHistory = "my_history"
	// MyApproval 用户的待审批单据
	MyApproval = "my_approval"
	// MyAttention 用户关注的单据
	MyAttention = "my_attention"
	// MyDealt 用户拥有查看权限的单据
	MyDealt = "my_dealt"
)

// GetTicketsByUserReq define get tickets by user req.
type GetTicketsByUserReq struct {
	User        string   `json:"user" validate:"required"`
	ViewType    ViewType `json:"view_type" validate:"omitempty"`
	WorkflowID  string   `json:"workflow_id" validate:"omitempty"`
	CreateAtGte string   `json:"create_at__gte" validate:"omitempty"`
	CreateAtLte string   `json:"create_at__lte" validate:"omitempty"`
	Page        int64    `json:"page" validate:"omitempty"`
	PageSize    int64    `json:"page_size" validate:"omitempty"`
}

// Encode GetTicketsByUserReq to get request params.
func (req *GetTicketsByUserReq) Encode() map[string]string {
	v := make(map[string]string)
	if len(req.User) != 0 {
		v["current_processors__in"] = req.User
	}

	if len(req.ViewType) != 0 {
		v["view_type"] = string(req.ViewType)
	}

	if len(req.WorkflowID) != 0 {
		v["workflow_key__in"] = req.WorkflowID
	}

	if len(req.CreateAtGte) != 0 && len(req.CreateAtLte) != 0 {
		v["created_at__range"] = fmt.Sprintf("%s,%s", req.CreateAtGte, req.CreateAtLte)
	}

	if req.Page != 0 {
		v["page"] = strconv.FormatInt(req.Page, 10)
	}

	if req.PageSize != 0 {
		v["page_size"] = strconv.FormatInt(req.PageSize, 10)
	}

	return v
}

// GetTicketsByUserRespData define get ticket by user resp data.
type GetTicketsByUserRespData struct {
	Page     int64    `json:"page"`
	PageSize int64    `json:"page_size"`
	Count    int64    `json:"count"`
	Results  []Ticket `json:"results"`
}

// Ticket define ticket.
type Ticket struct {
	Sn            string `json:"sn"`
	ID            string `json:"id"`
	Title         string `json:"title"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
	EndAt         string `json:"end_at"`
	WorkflowID    string `json:"workflow_id"`
	StatusDisplay string `json:"status_display"`
	CurrentSteps  []struct {
		TicketID    string `json:"ticket_id"`
		Name        string `json:"name"`
		ActivityKey string `json:"activity_key"`
	} `json:"current_steps"`
	CurrentProcessors []struct {
		TicketID      string `json:"ticket_id"`
		Processor     string `json:"processor"`
		ProcessorType string `json:"processor_type"`
	} `json:"current_processors"`
	FrontendURL   string `json:"frontend_url"`
	ApproveResult bool   `json:"approve_result"`
}

// GetTicketsByUser get tickets by user.
func (i *itsm) GetTicketsByUser(kt *kit.Kit, req *GetTicketsByUserReq) (*GetTicketsByUserRespData, error) {

	code, msg, res, err := apigateway.ApiGatewayCallOriginalWithoutReq[GetTicketsByUserRespData](i.client, i.bkUserCli,
		i.config, rest.GET, kt, req.Encode(), "/ticket/list/")

	if err != nil {
		return nil, err
	}

	// itsm成功时状态码为20000
	if code != success {
		err := fmt.Errorf("failed to call api gateway to get ticket by user, code: %d, msg: %s", code, msg)
		logs.Errorf("%s, result: %+v, rid: %s", err, res, kt.Rid)
		return nil, err
	}

	return res, nil
}
