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
	"hcm/pkg/thirdparty/api-gateway"
)

// WithdrawTicket 撤销单据
func (i *itsm) WithdrawTicket(kt *kit.Kit, ticketID string, operator string) (*RevokeTicketResult, error) {
	req := &RevokeTicketReq{
		SystemID: i.config.AppCode,
		TicketID: ticketID,
	}

	code, msg, res, err := apigateway.ApiGatewayCallOriginal[RevokeTicketReq, RevokeTicketResult](i.client, i.bkUserCli,
		i.config, rest.POST, kt, req, "/tickets/revoked/")

	if err != nil {
		return nil, err
	}

	// itsm成功时状态码为20000
	if code != success {
		err := fmt.Errorf("failed to call api gateway to withdraw ticket, code: %d, msg: %s", code, msg)
		logs.Errorf("%s, result: %+v, rid: %s", err, res, kt.Rid)
		return nil, err
	}

	return res, nil
}
