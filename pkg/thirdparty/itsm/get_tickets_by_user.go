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
	"hcm/pkg/thirdparty"
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
	ViewType    ViewType `json:"view_type" validate:"required"`
	CatalogID   int64    `json:"catalog_id" validate:"omitempty"`
	ServiceID   int64    `json:"service_id" validate:"omitempty"`
	CreateAtGte string   `json:"create_at__gte" validate:"omitempty"`
	CreateAtLte string   `json:"create_at__lte" validate:"omitempty"`
	Page        int64    `json:"page" validate:"omitempty"`
	PageSize    int64    `json:"page_size" validate:"omitempty"`
}

// Encode GetTicketsByUserReq to get request params.
func (req *GetTicketsByUserReq) Encode() map[string]string {
	v := make(map[string]string)
	if len(req.User) != 0 {
		v["user"] = req.User
	}

	if len(req.ViewType) != 0 {
		v["view_type"] = string(req.ViewType)
	}

	if req.CatalogID != 0 {
		v["catalog_id"] = strconv.FormatInt(req.CatalogID, 10)
	}

	if req.ServiceID != 0 {
		v["service_id"] = strconv.FormatInt(req.ServiceID, 10)
	}

	if len(req.CreateAtGte) != 0 {
		v["create_at__gte"] = req.CreateAtGte
	}

	if len(req.CreateAtLte) != 0 {
		v["create_at__lte"] = req.CreateAtLte
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
	Page      int64    `json:"page"`
	TotalPage int64    `json:"total_page"`
	Count     int64    `json:"count"`
	Next      string   `json:"next"`
	Previous  string   `json:"previous"`
	Items     []Ticket `json:"items"`
}

// Ticket define ticket.
type Ticket struct {
	Sn          string `json:"sn"`
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	ServiceId   int64  `json:"service_id"`
	ServiceType string `json:"service_type"`
	Meta        struct {
		Priority struct {
			Key   string `json:"key"`
			Name  string `json:"name"`
			Order int64  `json:"order"`
		} `json:"priority"`
	} `json:"meta"`
	BkBizId              int64  `json:"bk_biz_id"`
	CurrentStatus        string `json:"current_status"`
	CreateAt             string `json:"create_at"`
	Creator              string `json:"creator"`
	IsSuperviseNeeded    bool   `json:"is_supervise_needed"`
	FlowId               int64  `json:"flow_id"`
	SuperviseType        string `json:"supervise_type"`
	Supervisor           string `json:"supervisor"`
	ServiceName          string `json:"service_name"`
	CurrentStatusDisplay string `json:"current_status_display"`
	CurrentSteps         []struct {
		Id      int64  `json:"id"`
		Tag     string `json:"tag"`
		Name    string `json:"name"`
		StateID int64  `json:"state_id"`
	} `json:"current_steps"`
	PriorityName      string   `json:"priority_name"`
	CurrentProcessors string   `json:"current_processors"`
	CanComment        bool     `json:"can_comment"`
	CanOperate        bool     `json:"can_operate"`
	WaitingApprove    bool     `json:"waiting_approve"`
	Followers         []string `json:"followers"`
	CommentId         string   `json:"comment_id"`
	CanSupervise      bool     `json:"can_supervise"`
	CanWithdraw       bool     `json:"can_withdraw"`
	Sla               []string `json:"sla"`
	SlaColor          string   `json:"sla_color"`
}

// GetTicketsByUser get tickets by user.
func (i *itsm) GetTicketsByUser(kt *kit.Kit, req *GetTicketsByUserReq) (*GetTicketsByUserRespData, error) {
	resp := &struct {
		thirdparty.BaseResponse `json:",inline"`
		Data                    *GetTicketsByUserRespData `json:"data"`
	}{}

	err := i.client.Get().
		SubResourcef("/get_tickets_by_user/").
		WithParams(req.Encode()).
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Do().Into(resp)
	if err != nil {
		return nil, err
	}

	if !resp.Result || resp.Code != 0 {
		return nil, fmt.Errorf("get tickets by user failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data, nil
}
