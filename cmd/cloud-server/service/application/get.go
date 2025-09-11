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
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GetApplication ...
func (a *applicationSvc) GetApplication(cts *rest.Contexts) (interface{}, error) {
	applicationID := cts.PathParameter("application_id").String()

	application, err := a.client.DataService().Global.Application.GetApplication(
		cts.Kit.Ctx, cts.Kit.Header(), applicationID)
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if application.Applicant != cts.Kit.User {
		_, authorized, err := a.authorizer.Authorize(cts.Kit, meta.ResourceAttribute{Basic: &meta.Basic{
			Type:   meta.Application,
			Action: meta.Find,
		}})
		if err != nil {
			return nil, err
		}
		// 没有单据管理权限的用户只能查询自己的申请单
		if !authorized {
			return nil, errf.NewFromErr(errf.PermissionDenied,
				fmt.Errorf("you can not view other people's application"))
		}
	}

	// 查询审批链接
	ticket, err := a.itsmCli.GetTicketResult(cts.Kit, application.SN)
	if err != nil {
		logs.Errorf("failed to get ticket url, err: %v, ticket_id: %s, rid: %s", err, application.SN, cts.Kit.Rid)
		return nil, fmt.Errorf("call itsm get ticket url failed, err: %v", err)
	}

	if ticket == nil {
		logs.Errorf("call itsm get ticket url get empty result, ticket_id: %s, rid: %s", application.SN, cts.Kit.Rid)
		return nil, fmt.Errorf("call itsm get ticket url get empty result, ticket_id: %s", application.SN)
	}

	return &proto.ApplicationGetResp{
		ID:             application.ID,
		SN:             application.SN,
		Type:           application.Type,
		Status:         application.Status,
		Applicant:      application.Applicant,
		Content:        RemoveSenseField(application.Content),
		DeliveryDetail: application.DeliveryDetail,
		Memo:           application.Memo,
		Revision:       application.Revision,
		TicketUrl:      ticket.FrontendURL,
	}, nil
}
