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

// Package bill defines bill service.
package bill

import (
	"fmt"

	typesBill "hcm/pkg/adaptor/types/bill"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/core/cloud"
	hcbillservice "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// GcpGetBillList get gcp bill list.
func (b bill) GcpGetBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbillservice.GcpBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询aws账单基础表
	billInfo, err := getBillInfo[cloud.GcpBillConfigExtension](cts.Kit, req.BillAccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("gcp bill config get base info db failed, billAccID: %s, err: %+v", req.BillAccountID, err)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "bill_account_id: %s is not found", req.BillAccountID)
	}

	// 检查accountID是否存在，是否资源账号
	resAccountInfo, err := b.cs.DataService().Gcp.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), req.AccountID)
	if err != nil {
		logs.Errorf("get gcp resource account failed, accountID: %s, err: %+v", req.AccountID, err)
		return nil, err
	}
	if resAccountInfo.Type != enumor.ResourceAccount {
		return nil, fmt.Errorf("account: %s not resource account type", req.AccountID)
	}
	if resAccountInfo.Extension == nil || resAccountInfo.Extension.CloudProjectID == "" {
		return nil, fmt.Errorf("account: %s cloud_project_id is empty", req.AccountID)
	}

	cli, err := b.ad.GcpProxy(cts.Kit, req.BillAccountID)
	if err != nil {
		logs.Errorf("gcp request adaptor client err, req: %+v, err: %+v", req, err)
		return nil, err
	}

	opt := &typesBill.GcpBillListOption{
		BillAccountID: req.BillAccountID,
		AccountID:     req.AccountID,
		Month:         req.Month,
		BeginDate:     req.BeginDate,
		EndDate:       req.EndDate,
		ProjectID:     resAccountInfo.Extension.CloudProjectID,
	}
	if req.Page != nil {
		opt.Page = &typesBill.GcpBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		}
	}
	resp, count, err := cli.GetBillList(cts.Kit, opt, billInfo)
	if err != nil {
		return nil, err
	}

	return &hcbillservice.GcpBillListResult{
		Count:   count,
		Details: resp,
	}, nil
}

// GcpGetRootAccountBillList get gcp bill list.
func (b bill) GcpGetRootAccountBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbillservice.GcpRootAccountBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	// 查询aws账单基础表
	billInfo, err := getRootAccountBillConfigInfo[billcore.GcpBillConfigExtension](
		cts.Kit, req.RootAccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("gcp root account bill config get base info db failed, root account id: %s, err: %+v",
			req.RootAccountID, err)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(
			errf.RecordNotFound, "bill config for root_account_id: %s is not found", req.RootAccountID)
	}

	cli, err := b.ad.GcpRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("gcp request adaptor client err, req: %+v, err: %+v", req, err)
		return nil, err
	}

	opt := &typesBill.GcpRootAccountBillListOption{
		RootAccountID: req.RootAccountID,
		MainAccountID: req.MainAccountID,
		Month:         req.Month,
		BeginDate:     req.BeginDate,
		EndDate:       req.EndDate,
	}
	// 检查Main AccountID是否存在
	if len(req.MainAccountID) > 0 {
		mainAccountInfo, err := b.cs.DataService().Gcp.MainAccount.Get(cts.Kit, req.MainAccountID)
		if err != nil {
			logs.Errorf("get gcp main account failed, main account id: %s, err: %+v", req.MainAccountID, err)
			return nil, err
		}
		if mainAccountInfo.Extension == nil || mainAccountInfo.Extension.CloudProjectID == "" {
			return nil, fmt.Errorf("main account: %s cloud_project_id is empty", req.MainAccountID)
		}
		opt.ProjectID = mainAccountInfo.Extension.CloudProjectID
	}

	if req.Page != nil {
		opt.Page = &typesBill.GcpBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		}
	}
	resp, count, err := cli.GetRootAccountBillList(cts.Kit, opt, billInfo)
	if err != nil {
		return nil, err
	}

	return &hcbillservice.GcpBillListResult{
		Count:   count,
		Details: resp,
	}, nil
}
