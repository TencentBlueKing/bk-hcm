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

package billsummaryroot

import (
	asbillapi "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	cvt "hcm/pkg/tools/converter"
)

// ListRootAccountSummary list root account summary with options
func (s *service) ListRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummaryListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}})
	if err != nil {
		return nil, err
	}

	var expression = tools.ExpressionAnd(
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	)
	if req.Filter != nil {
		var err error
		expression, err = tools.And(req.Filter, expression)
		if err != nil {
			logs.Errorf("build filter expression failed, error: %v, rid: %s", err, cts.Kit.Rid)
			return nil, err
		}
	}

	listReq := &dsbillapi.BillSummaryRootListReq{
		Filter: expression,
		Page:   req.Page,
	}
	summaryResp, err := s.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, listReq)
	if err != nil {
		logs.Errorf("fail to list summary root, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	if len(summaryResp.Details) == 0 {
		return asbillapi.BillSummaryRootListResult{Count: cvt.PtrToVal(summaryResp.Count)}, nil
	}

	rootIDMap := make(map[string]struct{})
	summaryList := summaryResp.Details
	for i := range summaryList {
		rootIDMap[summaryList[i].RootAccountID] = struct{}{}
	}

	rootAccountIds := cvt.MapKeyToSlice(rootIDMap)
	rootMap, err := s.listRootAccount(cts.Kit, rootAccountIds)
	if err != nil {
		logs.Errorf("fail to get root account name for summary root, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	details := make([]asbillapi.BillSummaryRootResult, len(summaryResp.Details))
	for idx := range summaryResp.Details {
		summary := summaryResp.Details[idx]
		details[idx] = asbillapi.BillSummaryRootResult{
			SummaryRoot:     summary,
			RootAccountName: cvt.PtrToVal(rootMap[summary.RootAccountID]).Name,
		}
	}

	return asbillapi.BillSummaryRootListResult{Count: cvt.PtrToVal(summaryResp.Count), Details: details}, nil
}

func (s *service) listRootAccount(kt *kit.Kit, accountIDs []string) (map[string]*accountset.BaseRootAccount, error) {

	if len(accountIDs) == 0 {
		return map[string]*accountset.BaseRootAccount{}, nil
	}

	rootAccountReq := &core.ListReq{
		Filter: tools.ContainersExpression("id", accountIDs),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "cloud_id", "name"},
	}
	accountResp, err := s.client.DataService().Global.RootAccount.List(kt, rootAccountReq)
	if err != nil {
		logs.Errorf("fail to list root account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	rootNameMap := make(map[string]*accountset.BaseRootAccount)
	for i := range accountResp.Details {
		account := accountResp.Details[i]
		rootNameMap[account.ID] = account
	}
	return rootNameMap, nil
}
