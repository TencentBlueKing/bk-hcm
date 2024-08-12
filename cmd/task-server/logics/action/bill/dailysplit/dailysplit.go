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

package dailysplit

import (
	rawjson "encoding/json"
	"fmt"
	"path/filepath"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"
)

// DailyAccountSplitActionOption option for main account summary action
type DailyAccountSplitActionOption struct {
	RootAccountID string            `json:"root_account_id" validate:"required"`
	MainAccountID string            `json:"main_account_id" validate:"required"`
	BillYear      int               `json:"bill_year" validate:"required"`
	BillMonth     int               `json:"bill_month" validate:"required"`
	BillDay       int               `json:"bill_day" validate:"required"`
	VersionID     int               `json:"version_id" validate:"required"`
	Vendor        enumor.Vendor     `json:"vendor" validate:"required"`
	Extension     map[string]string `json:"extension"`
}

// DailyAccountSplitAction define main account summary action
type DailyAccountSplitAction struct{}

// ParameterNew return request params.
func (act DailyAccountSplitAction) ParameterNew() interface{} {
	return new(DailyAccountSplitActionOption)
}

// Name return action name
func (act DailyAccountSplitAction) Name() enumor.ActionName {
	return enumor.ActionDailyAccountSplit
}

func getFilter(opt *DailyAccountSplitActionOption, billDay int) *filter.Expression {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("version_id", opt.VersionID),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
	}
	if billDay != 0 {
		expressions = append(expressions, tools.RuleEqual("bill_day", billDay))
	}
	return tools.ExpressionAnd(expressions...)
}

func getBillItemFilter(opt *DailyAccountSplitActionOption, billDay int) *filter.Expression {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("main_account_id", opt.MainAccountID),
	}
	if billDay != 0 {
		expressions = append(expressions, tools.RuleEqual("bill_day", billDay))
	}
	return tools.ExpressionAnd(expressions...)
}

// Run pull daily bill
func (act DailyAccountSplitAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*DailyAccountSplitActionOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	pullTaskList, err := actcli.GetDataService().Global.Bill.ListBillDailyPullTask(
		kt.Kit(), &bill.BillDailyPullTaskListReq{
			Filter: getFilter(opt, opt.BillDay),
			Page: &core.BasePage{
				Start: 0,
				Limit: 31,
			},
		})
	if err != nil {
		return nil, fmt.Errorf("get pull task by opt %v failed, err %s", opt, err.Error())
	}
	if len(pullTaskList.Details) != 1 {
		return nil, fmt.Errorf("get pull task invalid length, resp %v", pullTaskList.Details)
	}
	task := pullTaskList.Details[0]

	if task.State == enumor.MainAccountRawBillPullStatePulled {
		if err := act.doDailySplit(kt, opt, task.BillDay); err != nil {
			return nil, fmt.Errorf("do splitting for %v day-%d failed, err %s", opt, task.BillDay, err.Error())
		}
		if err := act.changeTaskToSplitted(kt, task); err != nil {
			return nil, fmt.Errorf("update task %s to %s failed, err %s",
				task.ID, enumor.MainAccountRawBillPullStateSplit, err.Error())
		}
	}

	return nil, nil
}

func (act DailyAccountSplitAction) doDailySplit(
	kt run.ExecuteKit, opt *DailyAccountSplitActionOption, billDay int) error {

	// step1 清理原有当天特定version的billitem，因为有可能之前存在中途失败的脏数据了
	if err := cleanBillItem(kt.Kit(), opt, billDay); err != nil {
		return err
	}
	// step2 进行分账
	if err := splitBillItem(kt.Kit(), opt, billDay); err != nil {
		return err
	}
	// 此处不进行计算，因为有可能分账之后，产生的billitem并不在当前main account下
	return nil
}

func (act DailyAccountSplitAction) changeTaskToSplitted(
	kt run.ExecuteKit, billTask *bill.BillDailyPullTaskResult) error {

	return actcli.GetDataService().Global.Bill.UpdateBillDailyPullTask(
		kt.Kit(), &bill.BillDailyPullTaskUpdateReq{
			ID:    billTask.ID,
			State: enumor.MainAccountRawBillPullStateSplit,
		})
}

// 当前版本实现会把所有历史版本全部清理掉
// 待后续需要实现历史版本明细查看时，可分版本清理
func cleanBillItem(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int) error {
	batch := 0
	commonOpt := &bill.ItemCommonOpt{
		Vendor: opt.Vendor,
		Year:   opt.BillYear,
		Month:  opt.BillMonth,
	}
	for {
		var billListReq = &bill.BillItemListReq{
			ItemCommonOpt: commonOpt,
			ListReq:       &core.ListReq{Filter: getBillItemFilter(opt, billDay), Page: core.NewCountPage()},
		}
		result, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, billListReq)
		if err != nil {
			logs.Warnf("count bill item for %v day %d failed, err %s, rid %s", opt, billDay, err.Error(), kt.Rid)
			return fmt.Errorf("count bill item for %v day %d failed, err %s", opt, billDay, err.Error())
		}
		if result.Count > 0 {
			delReq := &bill.BillItemDeleteReq{
				ItemCommonOpt: commonOpt,
				Filter:        getBillItemFilter(opt, billDay),
			}
			if err := actcli.GetDataService().Global.Bill.BatchDeleteBillItem(kt, delReq); err != nil {
				return fmt.Errorf("delete 500 of %d bill item for %v day %d failed, err %s",
					result.Count, opt, billDay, err.Error())
			}
			logs.Infof("successfully delete batch %d bill item for %v day %d, rid %s",
				result.Count, opt, billDay, kt.Rid)
			batch = batch + 1
			continue
		}
		break
	}

	return nil
}

func splitBillItem(kt *kit.Kit, opt *DailyAccountSplitActionOption, billDay int) error {
	mainAccountInfo, err := getMainAccount(kt, opt.MainAccountID)
	if err != nil {
		return err
	}
	resp, err := actcli.GetDataService().Global.Bill.ListRawBillFileNames(kt, &bill.RawBillItemNameListReq{
		Vendor:        opt.Vendor,
		RootAccountID: opt.RootAccountID,
		MainAccountID: opt.MainAccountID,
		BillYear:      fmt.Sprintf("%d", opt.BillYear),
		BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
		Version:       fmt.Sprintf("%d", opt.VersionID),
		BillDate:      fmt.Sprintf("%02d", billDay),
	})
	if err != nil {
		return fmt.Errorf("failed to list raw bill files for %v, err %s", opt, err.Error())
	}

	splitter, err := GetSplitter(opt.Vendor)
	if err != nil {
		return fmt.Errorf("failed to get splitter for %v, err %s", opt, err.Error())
	}

	for _, filename := range resp.Filenames {
		var billItemList []bill.BillItemCreateReq[rawjson.RawMessage]
		// 后续可在该过程中，增加处理过程
		name := filepath.Base(filename)
		tmpReq := &bill.RawBillItemQueryReq{
			Vendor:        opt.Vendor,
			RootAccountID: opt.RootAccountID,
			MainAccountID: opt.MainAccountID,
			BillYear:      fmt.Sprintf("%d", opt.BillYear),
			BillMonth:     fmt.Sprintf("%02d", opt.BillMonth),
			Version:       fmt.Sprintf("%d", opt.VersionID),
			BillDate:      fmt.Sprintf("%02d", billDay),
			FileName:      name,
		}

		rawResp, err := actcli.GetDataService().Global.Bill.QueryRawBillItems(kt, tmpReq)
		if err != nil {
			return fmt.Errorf("failed to get raw bill item for %v, err %s", tmpReq, err.Error())
		}

		for _, item := range rawResp.Details {
			reqList, err := splitter.DoSplit(kt, opt, billDay, item, mainAccountInfo)
			if err != nil {
				return fmt.Errorf("batch create bill item for %s failed, err %s", filename, err.Error())
			}
			billItemList = append(billItemList, reqList...)
		}

		for _, itemsBatch := range slice.Split(billItemList, constant.BatchOperationMaxLimit) {
			createReq := &bill.BatchBillItemCreateReq[rawjson.RawMessage]{
				ItemCommonOpt: &bill.ItemCommonOpt{
					Vendor: opt.Vendor,
					Year:   opt.BillYear,
					Month:  opt.BillMonth,
				},
				Items: itemsBatch,
			}
			_, err = actcli.GetDataService().Global.Bill.BatchCreateBillItem(kt, createReq)
			if err != nil {
				return fmt.Errorf("batch create bill item for %s failed, err %s", filename, err.Error())
			}
		}
		logs.Infof("split %s successfully", filename)
	}
	return nil
}

func getMainAccount(kt *kit.Kit, mainAccountID string) (*protocore.BaseMainAccount, error) {
	var expressions []*filter.AtomRule
	expressions = append(expressions, []*filter.AtomRule{
		tools.RuleEqual("id", mainAccountID),
	}...)
	result, err := actcli.GetDataService().Global.MainAccount.List(kt, &core.ListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("get main account info by id %s failed, err %s", mainAccountID, err.Error())
	}
	if len(result.Details) != 1 {
		return nil, fmt.Errorf("get main account failed, invalid resp %v", result.Details)
	}

	return result.Details[0], nil
}
