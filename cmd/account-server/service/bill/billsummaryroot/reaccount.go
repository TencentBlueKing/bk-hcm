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
	"fmt"

	asbillapi "hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/client"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// ReaccountRootAccountSummary reaccount root account summary
func (s *service) ReaccountRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummaryReaccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	err := s.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Update}})
	if err != nil {
		return nil, err
	}

	rootSummary, err := getRootSummary(s.client, cts.Kit, req.RootAccountID, req.BillYear, req.BillMonth)
	if err != nil {
		return nil, err
	}

	if rootSummary.State != enumor.RootAccountBillSummaryStateAccounted &&
		rootSummary.State != enumor.RootAccountBillSummaryStateConfirmed &&
		rootSummary.State != enumor.RootAccountBillSummaryStateSynced {
		logs.Warnf("bill of root account %s in %d-%02d is in state %s, cannot do reaccount",
			rootSummary.RootAccountID, req.BillYear, req.BillMonth, rootSummary.State)
		return nil, fmt.Errorf("bill of root account %s %d-%02d is in state %s, cannot do reaccount",
			rootSummary.RootAccountID, req.BillYear, req.BillMonth, rootSummary.State)
	}

	updateReq := &bill.BillSummaryRootUpdateReq{
		ID:             rootSummary.ID,
		CurrentVersion: rootSummary.CurrentVersion + 1,
		State:          enumor.RootAccountBillSummaryStateAccounting,
	}
	if err := s.client.DataService().Global.Bill.UpdateBillSummaryRoot(cts.Kit, updateReq); err != nil {
		logs.Warnf("failed to update root account bill summary %s to version %d state %s, err %s, rid %s",
			updateReq.ID, updateReq.CurrentVersion, updateReq.State, err.Error(), cts.Kit.Rid)
		return nil, err
	}
	logs.Infof("successfully update root account bill summary %s to version %d state %s, rid %s",
		updateReq.ID, updateReq.CurrentVersion, updateReq.State, cts.Kit.Rid)
	return nil, nil
}

func getRootSummary(
	client *client.ClientSet, kt *kit.Kit,
	rootAccountID string, billYear, billMonth int) (
	*billcore.SummaryRoot, error) {

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", rootAccountID),
		tools.RuleEqual("bill_year", billYear),
		tools.RuleEqual("bill_month", billMonth),
	}
	rootSummaryList, err := client.DataService().Global.Bill.ListBillSummaryRoot(
		kt, &bill.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		logs.Warnf(
			"list root account bill summary failed by req %s/%d/%d failed, err %s, rid: %s",
			rootAccountID, billYear, billMonth, err.Error(), kt.Rid)
		return nil, err
	}
	if len(rootSummaryList.Details) == 0 {
		logs.Warnf("root account bill summary with %s/%d/%d no found, rid: %s",
			rootAccountID, billYear, billMonth, kt.Rid)
		return nil, fmt.Errorf("root account bill summary with %s/%d/%d no found, rid: %s",
			rootAccountID, billYear, billMonth, kt.Rid)
	}
	if len(rootSummaryList.Details) != 1 {
		logs.Warnf("invalid response length, resp %v, rid: %s", rootSummaryList.Details, kt.Rid)
		return nil, fmt.Errorf("invalid response length, resp %v", rootSummaryList.Details)
	}
	return rootSummaryList.Details[0], nil
}
