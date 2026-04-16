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
	dataproto "hcm/pkg/api/data-service"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
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

	return a.buildApplicationGetResp(cts, application)
}

// GetBizApplication 业务视角下查看单据明细
func (a *applicationSvc) GetBizApplication(cts *rest.Contexts) (interface{}, error) {
	bkBizID, err := cts.PathParameter("bk_biz_id").Int64()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	applicationID := cts.PathParameter("application_id").String()
	if applicationID == "" {
		return nil, errf.Newf(errf.InvalidParameter, "application_id is required")
	}

	// 业务访问权限鉴权
	err = a.authorizer.AuthorizeWithPerm(cts.Kit, meta.ResourceAttribute{
		Basic: &meta.Basic{Type: meta.Biz, Action: meta.Access},
		BizID: bkBizID,
	})
	if err != nil {
		logs.Warnf("user %s has no access permission to biz %d, rid: %s", cts.Kit.User, bkBizID, cts.Kit.Rid)
		return nil, errf.New(errf.RecordNotFound, "application not found")
	}

	// 获取单据详情
	application, err := a.client.DataService().Global.Application.GetApplication(
		cts.Kit.Ctx, cts.Kit.Header(), applicationID)
	if err != nil {
		logs.Errorf("get application %s failed, err: %v, rid: %s", applicationID, err, cts.Kit.Rid)
		return nil, errf.New(errf.RecordNotFound, "application not found")
	}

	// 归属校验：检查 bk_biz_id 是否在 bk_biz_ids 列表中
	if !slice.IsItemInSlice(application.BkBizIDs, bkBizID) {
		logs.Warnf("application %s does not belong to biz %d, bk_biz_ids: %v, rid: %s",
			applicationID, bkBizID, application.BkBizIDs, cts.Kit.Rid)
		return nil, errf.New(errf.RecordNotFound, "application not found")
	}

	return a.buildApplicationGetResp(cts, application)
}

// buildApplicationGetResp 构建单据详情响应体
func (a *applicationSvc) buildApplicationGetResp(cts *rest.Contexts,
	application *dataproto.ApplicationResp) (*proto.ApplicationGetResp, error) {

	// 查询审批链接
	ticket, err := a.itsmCli.GetTicketResult(cts.Kit, application.SN)
	if err != nil {
		return nil, fmt.Errorf("call itsm get ticket url failed, err: %v", err)
	}

	return &proto.ApplicationGetResp{
		ID:             application.ID,
		Source:         application.Source,
		SN:             application.SN,
		Type:           application.Type,
		Operation:      application.Operation,
		Status:         application.Status,
		Applicant:      application.Applicant,
		Content:        RemoveSenseField(application.Content),
		DeliveryDetail: application.DeliveryDetail,
		Memo:           application.Memo,
		Revision:       application.Revision,
		TicketUrl:      ticket.TicketURL,
	}, nil
}
