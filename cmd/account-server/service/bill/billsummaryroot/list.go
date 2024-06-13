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
	dsbillapi "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
)

// List list root account summary with options
func (s *service) ListRootAccountSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(asbillapi.RootAccountSummaryListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	var expressions []filter.RuleFactory
	expressions = append(expressions, []filter.RuleFactory{
		filter.AtomRule{
			Field: "bill_year",
			Op:    filter.Equal.Factory(),
			Value: req.BillYear,
		},
		filter.AtomRule{
			Field: "bill_month",
			Op:    filter.Equal.Factory(),
			Value: req.BillMonth,
		},
	}...)
	if req.Filter != nil {
		expressions = append(expressions, req.Filter)
	}
	bizFilter, err := tools.And(
		expressions...)
	if err != nil {
		return nil, err
	}

	return s.client.DataService().Global.Bill.ListBillSummaryRoot(cts.Kit, &dsbillapi.BillSummaryRootListReq{
		Filter: bizFilter,
		Page:   req.Page,
	})
}
