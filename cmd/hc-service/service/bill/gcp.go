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
	typesBill "hcm/pkg/adaptor/types/bill"
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/api/core/cloud"
	hcbillservice "hcm/pkg/api/hc-service/bill"
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

	if req.Page == nil {
		req.Page = &typesBill.GcpBillPage{Offset: 0, Limit: core.GcpQueryLimit}
	}

	// 查询aws账单基础表
	billInfo, err := getBillInfo[cloud.GcpBillConfigExtension](cts.Kit, req.AccountID, b.cs.DataService())
	if err != nil {
		logs.Errorf("gcp bill config get base info db failed, accountID: %s, err: %+v", req.AccountID, err)
		return nil, err
	}
	if billInfo == nil {
		return nil, errf.Newf(errf.RecordNotFound, "account_id: %s is not found", req.AccountID)
	}

	cli, err := b.ad.Gcp(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("gcp request adaptor client err, req: %+v, err: %+v", req, err)
		return nil, err
	}

	opt := &typesBill.GcpBillListOption{
		AccountID: req.AccountID,
		Month:     req.Month,
		BeginDate: req.BeginDate,
		EndDate:   req.EndDate,
		Page: &typesBill.GcpBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		},
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
