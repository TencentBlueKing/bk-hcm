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

	"hcm/pkg/criteria/errf"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/thirdparty/api-gateway"
)

// TicketResult ticket result.
type TicketResult struct {
	ID string `json:"id"`
	// SN Notice：v4版本的SN和v3不是一个含义，虽然我们本地依然沿用SN的叫法，但是在v4版本中，SN并不使用，而是用ticket_id进行查询
	SN            string  `json:"sn"`
	FrontendURL   string  `json:"frontend_url"`
	Title         string  `json:"title"`
	ApproveResult *bool   `json:"approve_result"`
	EndAt         *string `json:"end_at"`
	Status        string  `json:"status"`
}

// batchQueryTicketResult 查询单据结果
func (i *itsm) batchQueryTicketResult(kt *kit.Kit, ticketID string) (*TicketResult, error) {
	params := map[string]string{
		"id": ticketID,
	}
	code, msg, res, err := apigateway.ApiGatewayCallOriginalWithoutReq[TicketResult](i.client, i.bkUserCli,
		i.config, rest.GET, kt, params, "/ticket/detail/")

	if err != nil {
		return nil, err
	}

	// itsm成功时状态码为20000
	if code != success {
		err := fmt.Errorf("failed to call api gateway to query ticket result, code: %d, msg: %s", code, msg)
		logs.Errorf("%s, result: %+v, rid: %s", err, res, kt.Rid)
		return nil, err
	}

	return res, nil
}

// GetTicketResult 查询单据结果
func (i *itsm) GetTicketResult(kt *kit.Kit, sn string) (result *TicketResult, err error) {
	result, err = i.batchQueryTicketResult(kt, sn)
	if err != nil {
		return nil, err
	}

	if len(result.ID) == 0 {
		return result, errf.New(errf.RecordNotFound, "itsm returns empty result")
	}
	return result, nil
}
