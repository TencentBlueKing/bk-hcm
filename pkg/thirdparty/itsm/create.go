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
	"strings"

	"hcm/pkg/kit"
	"hcm/pkg/thirdparty"
)

// VariableApprover 节点审批的引用变量审批人
// Note: ITSM里节点审批人可以有多种类型，上级等固定角色ITSM在提单时会自行确定，无需传参，其他海垒指定的角色则通过引用变量的方式来指定审批人
type VariableApprover struct {
	Variable  string
	Approvers []string
}

// CreateTicketParams create ticket params.
type CreateTicketParams struct {
	ServiceID         int64
	Creator           string
	CallbackURL       string
	Title             string
	ContentDisplay    string
	VariableApprovers []VariableApprover
}

type createTicketResult struct {
	SN string `json:"sn"`
}

type createTicketResp struct {
	thirdparty.BaseResponse `json:",inline"`
	Data                    *createTicketResult `json:"data"`
}

// CreateTicket 创建单据
func (i *itsm) CreateTicket(kt *kit.Kit, params *CreateTicketParams) (string, error) {
	// 提单表单
	fields := []map[string]interface{}{
		{"key": "title", "value": params.Title},
		{"key": "application_content", "value": params.ContentDisplay},
	}
	// 引用变量的审批人
	for _, v := range params.VariableApprovers {
		fields = append(fields, map[string]interface{}{
			"key":   v.Variable,
			"value": strings.Join(v.Approvers, ","),
		})
	}

	req := map[string]interface{}{
		"service_id": params.ServiceID,
		"creator":    params.Creator,
		"meta":       map[string]string{"callback_url": params.CallbackURL},
		"fields":     fields,
	}

	resp := new(createTicketResp)

	err := i.client.Post().
		SubResourcef("/create_ticket/").
		WithContext(kt.Ctx).
		WithHeaders(i.header(kt)).
		Body(req).
		Do().Into(resp)
	if err != nil {
		return "", err
	}
	if !resp.Result || resp.Code != 0 {
		return "", fmt.Errorf("create ticket failed, code: %d, msg: %s", resp.Code, resp.Message)
	}

	return resp.Data.SN, nil
}
