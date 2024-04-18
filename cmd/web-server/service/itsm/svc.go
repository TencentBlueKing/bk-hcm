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
	"net/http"

	"hcm/cmd/web-server/service/capability"
	webserver "hcm/pkg/api/web-server"
	"hcm/pkg/client"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	itsm2 "hcm/pkg/thirdparty/api-gateway/itsm"
)

// InitService initial the service
func InitService(c *capability.Capability) {
	svr := &service{
		client:  c.ApiClient,
		itsmCli: c.ItsmCli,
	}

	h := rest.NewHandler()
	h.Add("ListMyApprovalTicket", http.MethodPost, "/tickets/types/my_approval/list", svr.ListMyApprovalTicket)
	h.Add("TicketApprove", http.MethodPost, "/tickets/approve", svr.TicketApprove)

	h.Load(c.WebService)
}

type service struct {
	client  *client.ClientSet
	itsmCli itsm2.Client
}

// ListMyApprovalTicket 查询待我审批的单据。
func (svc *service) ListMyApprovalTicket(cts *rest.Contexts) (interface{}, error) {
	req := new(webserver.ListMyApprovalTicketReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	serviceIDs, err := svc.client.CloudServer().ApprovalProcess.GetApprovalProcessServiceID(cts.Kit)
	if err != nil {
		logs.Errorf("call cloud-server to get approval process service id failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	result := &webserver.ListMyApprovalTicketResp{
		// 总数量
		Count:   0,
		Details: nil,
	}
	for _, serviceID := range serviceIDs {
		getReq := &itsm2.GetTicketsByUserReq{
			ServiceID: serviceID,
			User:      cts.Kit.User,
			ViewType:  itsm2.MyApproval,
			Page:      (int64(req.Page.Start) / int64(req.Page.Limit)) + 1,
			PageSize:  int64(req.Page.Limit),
		}
		resp, err := svc.itsmCli.GetTicketsByUser(cts.Kit, getReq)
		if err != nil {
			logs.Errorf("request itsm get tickets by user failed, err: %v, req: %v, rid: %s", err, getReq, cts.Kit.Rid)
			return nil, err
		}
		result.Details = append(result.Details, resp.Items...)
		result.Count += resp.Count
	}

	return result, nil
}

// TicketApprove 审批单据。
func (svc *service) TicketApprove(cts *rest.Contexts) (interface{}, error) {
	req := new(webserver.TicketApproveReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	getReq := &itsm2.ApproveReq{
		Sn:       req.Sn,
		StateID:  req.StateID,
		Approver: cts.Kit.User,
		Action:   req.Action.ToItsmAction(),
		Remark:   req.Memo,
	}
	if err := svc.itsmCli.Approve(cts.Kit, getReq); err != nil {
		logs.Errorf("request itsm ticket approve failed, err: %v, req: %v, rid: %s", err, getReq, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
