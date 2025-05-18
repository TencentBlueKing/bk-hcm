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

package billsummarymain

import (
	"fmt"

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
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/cmdb"
	"hcm/pkg/tools/maps"
	"hcm/pkg/tools/slice"
)

// ListMainAccountSummary list main account summary with options
func (s *service) ListMainAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.MainAccountSummaryListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	authParam := meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Find}}
	if err := s.authorizer.AuthorizeWithPerm(cts.Kit, authParam); err != nil {
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
	listReq := &dsbillapi.BillSummaryMainListReq{Filter: expression, Page: req.Page}
	summary, err := s.client.DataService().Global.Bill.ListBillSummaryMain(cts.Kit, listReq)
	if err != nil {
		return nil, err
	}
	if len(summary.Details) == 0 {
		return summary, nil
	}

	ret := &asbillapi.MainAccountSummaryListResult{
		Count:   0,
		Details: make([]*asbillapi.MainAccountSummaryResult, 0, len(summary.Details)),
	}

	mainAccountIDs, rootAccountIDs := s.getAccountIds(summary)
	mainMap, err := s.listMainAccount(cts.Kit, mainAccountIDs)
	if err != nil {
		logs.Errorf("list main account for summary main failed, err: %v, main ids: %v, rid: %s",
			err, mainAccountIDs, cts.Kit.Rid)
		return nil, err
	}
	rootMap, err := s.listRootAccount(cts.Kit, rootAccountIDs)
	if err != nil {
		logs.Errorf("list root account for summary main failed, err: %v, root ids: %v, rid: %s",
			err, rootAccountIDs, cts.Kit.Rid)
		return nil, err
	}

	for _, detail := range summary.Details {
		mainAccount, ok := mainMap[detail.MainAccountID]
		if !ok {
			return nil, fmt.Errorf("main account %s(%s) of summary main %s not found",
				detail.MainAccountCloudID, detail.MainAccountID, detail.ID)
		}
		rootAccount, ok := rootMap[detail.RootAccountID]
		if !ok {
			return nil, fmt.Errorf("root account: %s(%s) of summary main %s not found",
				detail.RootAccountCloudID, detail.RootAccountID, detail.ID)
		}
		tmp := &asbillapi.MainAccountSummaryResult{
			BillSummaryMain: detail,
			MainAccountName: mainAccount.Name,
			RootAccountName: rootAccount.Name,
		}
		ret.Details = append(ret.Details, tmp)
	}

	return ret, nil
}

func (s *service) getAccountIds(summary *dsbillapi.BillSummaryMainListResult) (
	mainAccountIDs []string, rootAccountIDs []string) {

	mainAccountIDMap := make(map[string]struct{})
	rootAccountIDMap := make(map[string]struct{})
	for _, detail := range summary.Details {
		mainAccountIDMap[detail.MainAccountID] = struct{}{}
		rootAccountIDMap[detail.RootAccountID] = struct{}{}
	}

	mainAccountIDs = maps.Keys(mainAccountIDMap)
	rootAccountIDs = maps.Keys(rootAccountIDMap)
	return mainAccountIDs, rootAccountIDs
}

func (s *service) listMainAccount(kt *kit.Kit, accountIDs []string) (map[string]*accountset.BaseMainAccount, error) {

	accountMap := make(map[string]*accountset.BaseMainAccount, len(accountIDs))
	if len(accountIDs) == 0 {
		return accountMap, nil
	}

	listOpt := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", accountIDs),
		),
		Page:   core.NewDefaultBasePage(),
		Fields: []string{"id", "cloud_id", "name"},
	}
	accountResult, err := s.client.DataService().Global.MainAccount.List(kt, listOpt)
	if err != nil {
		logs.Errorf("fail to list main account, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	for _, detail := range accountResult.Details {
		accountMap[detail.ID] = detail
	}
	return accountMap, nil
}

func (s *service) listRootAccount(kt *kit.Kit, accountIDs []string) (map[string]*accountset.BaseRootAccount, error) {

	rootNameMap := make(map[string]*accountset.BaseRootAccount)

	if len(accountIDs) == 0 {
		return rootNameMap, nil
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
	for i := range accountResp.Details {
		account := accountResp.Details[i]
		rootNameMap[account.ID] = account
	}
	return rootNameMap, nil
}

func (s *service) listBiz(kt *kit.Kit, ids []int64) (map[int64]string, error) {
	ids = slice.Unique(ids)
	if len(ids) == 0 {
		return nil, nil
	}

	data := make(map[int64]string)
	for _, split := range slice.Split(ids, int(filter.DefaultMaxInLimit)) {
		rules := []cmdb.Rule{
			&cmdb.AtomRule{
				Field:    "bk_biz_id",
				Operator: cmdb.OperatorIn,
				Value:    split,
			},
		}
		expression := &cmdb.QueryFilter{Rule: &cmdb.CombinedRule{Condition: "AND", Rules: rules}}

		params := &cmdb.SearchBizParams{
			BizPropertyFilter: expression,
			Fields:            []string{"bk_biz_id", "bk_biz_name"},
		}
		resp, err := s.cmdbCli.SearchBusiness(kt, params)
		if err != nil {
			logs.Errorf("call cmdb search business api failed, err: %v, rid: %s", err, kt.Rid)
			return nil, fmt.Errorf("call cmdb search business api failed, err: %v", err)
		}

		for _, biz := range resp.Info {
			data[biz.BizID] = biz.BizName
		}
	}

	return data, nil
}
