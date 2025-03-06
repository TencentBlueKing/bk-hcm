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
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/iam/meta"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ConfirmRootAccountSummary reaccount root account summary
func (s *service) ConfirmRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummaryConfirmReq)
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
		rootSummary.State != enumor.RootAccountBillSummaryStateConfirmed {
		logs.Warnf("bill of root account %s in %d-%02d is in state %s, cannot do confirm",
			rootSummary.RootAccountID, req.BillYear, req.BillMonth, rootSummary.State)
		return nil, fmt.Errorf("bill of root account %s %d-%02d is in state %s, cannot do confirm",
			rootSummary.RootAccountID, req.BillYear, req.BillMonth, rootSummary.State)
	}

	updateReq := &bill.BillSummaryRootUpdateReq{
		ID:    rootSummary.ID,
		State: enumor.RootAccountBillSummaryStateConfirmed,
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
