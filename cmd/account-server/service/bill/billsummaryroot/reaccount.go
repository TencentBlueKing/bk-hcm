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
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// Reaccount reaccount root account summary
func (s *service) ReaccountRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummaryReaccountReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", req.RootAccountID),
		tools.RuleEqual("bill_year", req.BillYear),
		tools.RuleEqual("bill_month", req.BillMonth),
	}
	rootSummaryList, err := s.client.DataService().Global.Bill.ListBillSummaryRoot(
		cts.Kit, &bill.BillSummaryRootListReq{
			Filter: tools.ExpressionAnd(expressions...),
			Page: &core.BasePage{
				Start: 0,
				Limit: 1,
			},
		})
	if err != nil {
		logs.Warnf(
			"list root account bill summary failed by req %v failed, err %s, rid: %s", req, err.Error(), cts.Kit.Rid)
		return nil, err
	}
	if len(rootSummaryList.Details) == 0 {
		logs.Warnf("root account bill summary with %v no found, rid: %s", req, cts.Kit.Rid)
		return nil, fmt.Errorf("root account bill summary with %v no found, rid: %s", req, cts.Kit.Rid)
	}
	if len(rootSummaryList.Details) != 1 {
		logs.Warnf("invalid response length, resp %v, rid: %s", rootSummaryList.Details, cts.Kit.Rid)
		return nil, fmt.Errorf("invalid response length, resp %v", rootSummaryList.Details)
	}
	rootSummary := rootSummaryList.Details[0]
	updateReq := &bill.BillSummaryRootUpdateReq{
		ID:             req.RootAccountID,
		CurrentVersion: rootSummary.CurrentVersion + 1,
		State:          constant.RootAccountBillSummaryStateAccounting,
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
