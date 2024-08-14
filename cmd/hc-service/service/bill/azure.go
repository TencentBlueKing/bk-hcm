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
	cvt "hcm/pkg/tools/converter"
)

// AzureGetBillList get azure bill list.
func (b bill) AzureGetBillList(cts *rest.Contexts) (interface{}, error) {
	req := new(hcbillservice.AzureBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := b.ad.Azure(cts.Kit, req.AccountID)
	if err != nil {
		logs.Errorf("azure request adaptor client err, req: %+v, err: %+v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}
	account, err := b.cs.DataService().Azure.Account.Get(cts.Kit.Ctx, cts.Kit.Header(), req.AccountID)
	if err != nil {
		logs.Errorf("fail to get azure account, err: %v, account_id: %s, rid: %s",
			err, req.AccountID, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.AzureBillListOption{
		SubscriptionID: cvt.PtrToVal(account.Extension).CloudSubscriptionID,
		BeginDate:      req.BeginDate,
		EndDate:        req.EndDate,
		Page:           req.Page,
	}

	list, err := cli.GetBillList(cts.Kit, opt)
	if err != nil {
		logs.Errorf("azure request adaptor list bill failed, err: %v, req: %+v, rid: %s", err, req, cts.Kit.Rid)
		return nil, err
	}

	return &hcbillservice.AzureBillListResult{
		NextLink: cvt.PtrToVal(list.NextLink),
		Details:  list.Value,
	}, nil
}

// AzureGetRootAccountBillList 获取azure 根账号下账单
func (b bill) AzureGetRootAccountBillList(cts *rest.Contexts) (any, error) {

	req := new(hcbillservice.AzureRootBillListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	cli, err := b.ad.AzureRoot(cts.Kit, req.RootAccountID)
	if err != nil {
		logs.Errorf("azure request adaptor client err, req: %+v, err: %+v, rid: %s", req, err, cts.Kit.Rid)
		return nil, err
	}

	opt := &typesBill.AzureBillListOption{
		SubscriptionID: req.SubscriptionID,
		BeginDate:      req.BeginDate,
		EndDate:        req.EndDate,
		Page:           req.Page,
	}

	list, err := cli.GetBillList(cts.Kit, opt)
	if err != nil {
		logs.Errorf("azure request adaptor list root account bill failed, err: %v, req: %+v, rid: %s",
			err, req, cts.Kit.Rid)
		return nil, err
	}

	return &hcbillservice.AzureBillListResult{
		NextLink: cvt.PtrToVal(list.NextLink),
		Details:  list.Value,
	}, nil
}
