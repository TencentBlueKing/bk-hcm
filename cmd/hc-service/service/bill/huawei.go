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
	hcbillservice "hcm/pkg/api/hc-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/golang/protobuf/proto"
)

// HuaWeiGetBillList get huawei bill list.
func (b bill) HuaWeiGetBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbillservice.HuaWeiBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page == nil {
		req.Page = &typesBill.HuaWeiBillPage{Offset: proto.Int32(0), Limit: proto.Int32(typesBill.HuaWeiQueryLimit)}
	}

	cli, err := b.ad.HuaWei(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("huawei request adaptor client err, req: %+v, err: %+v", req, err)
		return nil, err
	}

	opt := &typesBill.HuaWeiBillListOption{
		AccountID: req.AccountID,
		Month:     req.Month,
		Page: &typesBill.HuaWeiBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		},
	}
	resp, err := cli.GetBillList(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	return &hcbillservice.HuaWeiBillListResult{
		Count:    resp.TotalCount,
		Details:  resp.MonthlyRecords,
		Currency: resp.Currency,
	}, nil
}

// HuaWeiGetFeeRecordList get huawei fee record list
func (b bill) HuaWeiGetFeeRecordList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbillservice.HuaWeiFeeRecordListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if req.Page == nil {
		req.Page = &typesBill.HuaWeiBillPage{Offset: proto.Int32(0), Limit: proto.Int32(typesBill.HuaWeiQueryLimit)}
	}

	cli, err := b.ad.HuaWeiRoot(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("huawei request adaptor client err, req: %+v, err: %+v", req, err)
		return nil, err
	}

	opt := &typesBill.HuaWeiFeeRecordListOption{
		AccountID:     req.AccountID,
		SubAccountID:  req.SubAccountID,
		Month:         req.Month,
		BillDateBegin: req.BillDateBegin,
		BillDateEnd:   req.BillDateEnd,
		Page: &typesBill.HuaWeiBillPage{
			Offset: req.Page.Offset,
			Limit:  req.Page.Limit,
		},
	}
	resp, err := cli.GetFeeRecordList(cts.Kit, opt)
	if err != nil {
		return nil, err
	}

	return &hcbillservice.HuaWeiBillListResult{
		Count:    resp.TotalCount,
		Details:  resp.FeeRecords,
		Currency: resp.Currency,
	}, nil
}
