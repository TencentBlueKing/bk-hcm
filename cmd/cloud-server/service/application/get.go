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
	"fmt"

	proto "hcm/pkg/api/cloud-server/application"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/rest"
)

// Get ...
func (a *applicationSvc) Get(cts *rest.Contexts) (interface{}, error) {
	applicationID := cts.PathParameter("application_id").String()

	application, err := a.client.DataService().Global.Application.Get(cts.Kit.Ctx, cts.Kit.Header(), applicationID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 只能查询自己的申请单
	if application.Applicant != cts.Kit.User {
		return nil, errf.NewFromErr(errf.PermissionDenied, fmt.Errorf("you can not view other people's application"))
	}

	// 查询审批链接
	ticket, err := a.itsmCli.GetTicketResult(cts.Kit, application.SN)
	if err != nil {
		return nil, fmt.Errorf("call itsm get ticket url failed, err: %v", err)
	}

	return &proto.ApplicationGetResp{
		ID:        application.ID,
		SN:        application.SN,
		Type:      application.Type,
		Status:    application.Status,
		Applicant: application.Applicant,
		// 暂时不需要，需要时再将JSON解析成struct或map
		Content:        "",
		DeliveryDetail: application.DeliveryDetail,
		Memo:           application.Memo,
		Revision:       application.Revision,
		TicketUrl:      ticket.TicketURL,
	}, nil
}
