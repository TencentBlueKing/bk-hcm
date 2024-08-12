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

// Package dailypull ...
package dailypull

import (
	"fmt"
	"path/filepath"

	"hcm/cmd/task-server/logics/action/bill/dailypull/registry"
	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	databill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	cvt "hcm/pkg/tools/converter"
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

// Rollback clean old raw bills
func (act PullDailyBillAction) Rollback(kt run.ExecuteKit, params interface{}) error {
	logs.Infof("rollback daily pull bill action, rid: %s", kt.Kit().Rid)
	return nil
}

// Run run pull daily bill
func (act PullDailyBillAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*registry.PullDailyBillOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	pl, ok := registry.PullerRegistry[opt.Vendor]
	if !ok {
		return nil, errf.New(errf.InvalidParameter, fmt.Sprintf("invalid vendor for pull raw bill %s", opt.Vendor))
	}

	// clean old raw bill item
	err := act.cleanRawBills(kt, opt)
	if err != nil {
		logs.Errorf("fail to clean raw bills before pull, err: %v, vendor:%s, rid: %s", err, opt.Vendor, kt.Kit().Rid)
		return nil, err
	}

	result, err := pl.Pull(kt, opt)
	if err != nil {
		return nil, errf.New(errf.Aborted, err.Error())
	}

	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
		tools.RuleEqual("version_id", opt.VersionID),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		tools.RuleEqual("bill_day", opt.BillDay),
	}
	filter := tools.ExpressionAnd(expressions...)

	billCli := actcli.GetDataService().Global.Bill
	billTaskResult, err := billCli.ListBillDailyPullTask(kt.Kit(), &databill.BillDailyPullTaskListReq{
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
		kt.Kit(), &databill.BillDailyPullTaskUpdateReq{
			ID:       billTask.ID,
			Count:    result.Count,
			Currency: result.Currency,
			Cost:     cvt.ValToPtr(result.Cost),
			State:    enumor.MainAccountRawBillPullStatePulled,
		}); err != nil {
		return nil, errf.New(errf.Aborted, err.Error())
	}
	logs.Infof("update daily pull task %s to count %d, currency %s, cost %s, state %s, rid: %s",
		billTask.ID, result.Count, result.Currency, result.Cost,
		enumor.MainAccountRawBillPullStatePulled, kt.Kit().Rid)
	return nil, nil
}

func (act *PullDailyBillAction) cleanRawBills(kt run.ExecuteKit, opt *registry.PullDailyBillOption) error {
	listResult, err := actcli.GetDataService().Global.Bill.ListRawBillFileNames(
		kt.Kit(), &databill.RawBillItemNameListReq{
			Vendor:        opt.Vendor,
			RootAccountID: opt.RootAccountID,
			MainAccountID: opt.MainAccountID,
			BillYear:      fmt.Sprintf("%d", opt.BillYear),
			BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
			BillDate:      fmt.Sprintf("%02d", opt.BillDay),
			Version:       fmt.Sprintf("%d", opt.VersionID),
		})
	if err != nil {
		logs.Warnf("list raw bill filenames failed, err: %s, vendor:%s, rid: %s", err.Error(), opt.Vendor, kt.Kit().Rid)
		return err
	}
	for _, filename := range listResult.Filenames {
		name := filepath.Base(filename)
		if err := actcli.GetDataService().Global.Bill.DeleteRawBill(kt.Kit(), &databill.RawBillDeleteReq{
			RawBillPathParam: databill.RawBillPathParam{
				Vendor:        opt.Vendor,
				RootAccountID: opt.RootAccountID,
				MainAccountID: opt.MainAccountID,
				BillYear:      fmt.Sprintf("%d", opt.BillYear),
				BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
				BillDate:      fmt.Sprintf("%02d", opt.BillDay),
				Version:       fmt.Sprintf("%d", opt.VersionID),
				FileName:      name,
			},
		}); err != nil {
			logs.Warnf("delete raw bill %s failed, err: %s, vendor: %s, rid: %s",
				filename, err.Error(), opt.Vendor, kt.Kit().Rid)
			return err
		}
	}
	return nil
}
