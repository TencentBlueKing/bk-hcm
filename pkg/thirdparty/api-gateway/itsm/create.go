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

	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	apigateway "hcm/pkg/thirdparty/api-gateway"
)

// VariableApprover 节点审批的引用变量审批人
// Note: ITSM里节点审批人可以有多种类型，上级等固定角色ITSM在提单时会自行确定，无需传参，其他海垒指定的角色则通过引用变量的方式来指定审批人
type VariableApprover struct {
	Variable  string
	Approvers []string
}

// CreateTicketParams create ticket params.
type CreateTicketParams struct {
	SystemID          string
	WorkflowKey       string
	Creator           string
	CallbackURL       string
	CallbackToken     string
	Title             string
	ContentDisplay    string
	VariableApprovers []VariableApprover
}

// CreateTicket 创建单据
func (i *itsm) CreateTicket(kt *kit.Kit, params *CreateTicketParams) (*CreateTicketResult, error) {
	req := &CreateTicketReq{
		SystemID:      params.SystemID,
		WorkflowKey:   params.WorkflowKey,
		Operator:      params.Creator,
		CallbackURL:   params.CallbackURL,
		CallbackToken: params.CallbackToken,
	}

	// 提单表单
	reqForm := make(map[string]interface{})
	reqForm["ticket__title"] = params.Title
	reqForm["application_content"] = params.ContentDisplay

	// 引用变量的审批人
	for _, v := range params.VariableApprovers {
		reqForm[v.Variable] = v.Approvers
	}
	req.FormData = reqForm

	code, msg, res, err := apigateway.ApiGatewayCallOriginal[CreateTicketReq, CreateTicketResult](i.client,
		i.bkUserCli, i.config, rest.POST, kt, req, "/ticket/create/")

	if err != nil {
		return nil, err
	}

	// itsm成功时状态码为20000
	if code != success {
		err := fmt.Errorf("failed to call api gateway to create ticket, code: %d, msg: %s", code, msg)
		logs.Errorf("%s, result: %+v, rid: %s", err, res, kt.Rid)
		return nil, err
	}

	return res, nil
}
