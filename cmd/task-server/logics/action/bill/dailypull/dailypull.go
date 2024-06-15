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

package dailypull

import (
	"fmt"

	"hcm/cmd/task-server/logics/action/bill/dailypull/registry"
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	billproto "hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/runtime/filter"
)

var _ action.Action = new(PullDailyBillAction)
var _ action.ParameterAction = new(PullDailyBillAction)

// PullDailyBillAction define daily pull bill action
type PullDailyBillAction struct{}

// ParameterNew return request params.
func (act PullDailyBillAction) ParameterNew() interface{} {
	return new(registry.PullDailyBillOption)
}

// Name return action name
func (act PullDailyBillAction) Name() enumor.ActionName {
	return enumor.ActionPullDailyRawBill
}

// Run run pull daily bill
func (act PullDailyBillAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*registry.PullDailyBillOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	pl, ok := registry.PullerRegistry[opt.Vendor]
	if !ok {
		return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("invalid vendor %s", opt.Vendor))
	}
	result, err := pl.Pull(kt, opt)
	if err != nil {
		return nil, errf.New(errf.Aborted, err.Error())
	}

	var expressions []filter.RuleFactory
	expressions = append(expressions, []filter.RuleFactory{
		filter.AtomRule{
			Field: "root_account_id",
			Op:    filter.Equal.Factory(),
			Value: opt.RootAccountID,
		},
		filter.AtomRule{
			Field: "main_account_id",
			Op:    filter.Equal.Factory(),
			Value: opt.MainAccountID,
		},
		filter.AtomRule{
			Field: "version_id",
			Op:    filter.Equal.Factory(),
			Value: opt.VersionID,
		},
		filter.AtomRule{
			Field: "bill_year",
			Op:    filter.Equal.Factory(),
			Value: opt.BillYear,
		},
		filter.AtomRule{
			Field: "bill_month",
			Op:    filter.Equal.Factory(),
			Value: opt.BillMonth,
		},
		filter.AtomRule{
			Field: "bill_day",
			Op:    filter.Equal.Factory(),
			Value: opt.BillDay,
		},
	}...)
	filter, err := tools.And(expressions...)
	if err != nil {
		return nil, err
	}

	billCli := actcli.GetDataService().Global.Bill
	billTaskResult, err := billCli.ListBillDailyPullTask(kt.Kit(), &billproto.BillDailyPullTaskListReq{
		Filter: filter,
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		return nil, errf.New(errf.Aborted, err.Error())
	}
	if len(billTaskResult.Details) != 1 {
		return nil, errf.New(errf.Aborted, fmt.Sprintf("unexpected task length, resp %+v", billTaskResult.Details))
	}
	billTask := billTaskResult.Details[0]
	if err = billCli.UpdateBillDailyPullTask(
		kt.Kit(), &billproto.BillDailyPullTaskUpdateReq{
			ID:       billTask.ID,
			Count:    result.Count,
			Currency: result.Currency,
			Cost:     result.Cost,
			State:    constant.MainAccountRawBillPullStatePulled,
		}); err != nil {

		return nil, errf.New(errf.Aborted, err.Error())
	}
	return nil, nil
}
