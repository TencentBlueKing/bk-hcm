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
	asbillapi "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/rest"
)

// ListMainAccountSummary list root account summary with options
func (s *service) ListMainAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.MainAccountSummaryListReq)
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
			return nil, err
		}
	}

	summary, err := s.client.DataService().Global.Bill.ListBillSummaryMain(cts.Kit, &dsbillapi.BillSummaryMainListReq{
		Filter: expression,
		Page:   req.Page,
	})
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

	accountIDs := make([]string, 0, len(summary.Details))
	for _, detail := range summary.Details {
		accountIDs = append(accountIDs, detail.MainAccountID)
	}

	// fetch account
	listOpt := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleIn("id", accountIDs),
		),
		Page: core.NewDefaultBasePage(),
	}
	accountResult, err := s.client.DataService().Global.MainAccount.List(cts.Kit, listOpt)
	if err != nil {
		return nil, err
	}

	accountMap := make(map[string]*accountset.BaseMainAccount, len(accountIDs))
	for _, detail := range accountResult.Details {
		accountMap[detail.ID] = detail
	}

	for _, detail := range summary.Details {
		account := accountMap[detail.MainAccountID]

		tmp := &asbillapi.MainAccountSummaryResult{
			BillSummaryMainResult: *detail,
			MainAccountCloudID:    account.CloudID,
			MainAccountCloudName:  account.Name,
		}
		ret.Details = append(ret.Details, tmp)
	}

	return ret, nil
}
