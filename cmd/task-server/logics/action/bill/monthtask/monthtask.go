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

// Package monthtask ...
package monthtask

import (
	"encoding/json"
	"fmt"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
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

	"github.com/shopspring/decimal"
)

// MonthTaskActionOption option for month task action
type MonthTaskActionOption struct {
	Type          enumor.MonthTaskType `json:"type" validate:"required"`
	RootAccountID string               `json:"root_account_id" validate:"required"`
	BillYear      int                  `json:"bill_year" validate:"required"`
	BillMonth     int                  `json:"bill_month" validate:"required"`
	Vendor        enumor.Vendor        `json:"vendor" validate:"required"`
}

// MonthTaskAction month task action
type MonthTaskAction struct{}

// ParameterNew generate parameter for new month task
func (act MonthTaskAction) ParameterNew() interface{} {
	return new(MonthTaskActionOption)
}

// Name return month task action name
func (act MonthTaskAction) Name() enumor.ActionName {
	return enumor.ActionMonthTaskAction
}

// Run month task
func (act MonthTaskAction) Run(kt run.ExecuteKit, param interface{}) (interface{}, error) {
	opt, ok := param.(*MonthTaskActionOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "param type mismatch")
	}
	runner, err := GetRunner(opt.Vendor)
	if err != nil {
		return nil, err
	}
	switch opt.Type {
	case enumor.MonthTaskTypePull:
		if err := act.runPull(kt.Kit(), runner, opt); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.MonthTaskTypeSplit:
		if err := act.runSplit(kt.Kit(), runner, opt); err != nil {
			return nil, err
		}
		return nil, nil
	case enumor.MonthTaskTypeSummary:
		if err := act.runSummary(kt.Kit(), opt); err != nil {
			return nil, err
		}
		return nil, nil
	default:
		return nil, errf.New(errf.InvalidParameter, fmt.Sprintf(
			"invalid month task type %s", opt.Type))
	}
}

func (act MonthTaskAction) runPull(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption) error {
	for {
		task, err := getMonthPullTask(kt, opt)
		if err != nil {
			return err
		}
		rawBillItemList, isFinished, err := runner.Pull(
			kt, opt.RootAccountID, opt.BillYear, opt.BillMonth, task.PullIndex)
		if err != nil {
			return err
		}
		lenRawBillItemList := len(rawBillItemList)
		filename := fmt.Sprintf("%d-%d.csv", task.PullIndex, lenRawBillItemList)
		storeReq := &bill.RawBillCreateReq{
			RawBillPathParam: bill.RawBillPathParam{
				Vendor:        opt.Vendor,
				RootAccountID: task.RootAccountID,
				MainAccountID: enumor.MonthRawBillPathName,
				BillYear:      fmt.Sprintf("%d", task.BillYear),
				BillMonth:     fmt.Sprintf("%02d", task.BillMonth),
				BillDate:      enumor.MonthRawBillSpecialDatePathName,
				Version:       fmt.Sprintf("%d", task.VersionID),
				FileName:      filename,
			},

			Items: rawBillItemList,
		}
		databillCli := actcli.GetDataService().Global.Bill
		_, err = databillCli.CreateRawBill(kt, storeReq)
		if err != nil {
			logs.Warnf("failed to create month raw bill, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return fmt.Errorf("failed to create month raw bill, opt: %+v, err: %s", opt, err.Error())
		}
		logs.Infof("month task %+v pulled %d records, continue", opt, lenRawBillItemList)
		if isFinished {
			if err := databillCli.UpdateBillMonthPullTask(kt, &bill.BillMonthTaskUpdateReq{
				ID:        task.ID,
				Count:     task.Count + uint64(lenRawBillItemList),
				PullIndex: task.PullIndex + uint64(lenRawBillItemList),
				State:     enumor.RootAccountMonthBillTaskStatePulled,
			}); err != nil {
				logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
				return err
			}
			return nil
		}
		if err := databillCli.UpdateBillMonthPullTask(kt, &bill.BillMonthTaskUpdateReq{
			ID:        task.ID,
			Count:     task.Count + uint64(lenRawBillItemList),
			PullIndex: task.PullIndex + uint64(lenRawBillItemList),
		}); err != nil {
			logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return err
		}
	}
}

func (act MonthTaskAction) runSplit(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption) error {
	// step1 清理原有月度任务的billitem，因为有可能之前存在中途失败的脏数据了
	if err := act.cleanBillItem(kt, opt); err != nil {
		return err
	}
	// step2 进行分账
	var splitMainAccountMap = make(map[string]struct{})
	curlIndex := uint64(0)
	for {
		task, err := getMonthPullTask(kt, opt)
		if err != nil {
			logs.Errorf("fail to get month task for splitting, err: %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		cnt, isFinished, err := act.split(kt, runner, opt, task, splitMainAccountMap, curlIndex)
		if err != nil {
			return err
		}

		mtUpdate := &bill.BillMonthTaskUpdateReq{
			ID:         task.ID,
			SplitIndex: curlIndex + uint64(cnt),
		}
		if isFinished {
			var itemList = make([]billcore.MonthTaskSummaryDetailItem, 0, len(splitMainAccountMap))
			for mainAccountID := range splitMainAccountMap {
				itemList = append(itemList, billcore.MonthTaskSummaryDetailItem{MainAccountID: mainAccountID})
			}
			itemListJSON, err := json.Marshal(itemList)
			if err != nil {
				logs.Errorf("fail to marshal month task summary detail, err: %v, rid: %s", err, kt.Rid)
				return err
			}
			mtUpdate.SummaryDetail = string(itemListJSON)
			mtUpdate.State = enumor.RootAccountMonthBillTaskStateSplit
			if err := actcli.GetDataService().Global.Bill.UpdateBillMonthPullTask(kt, mtUpdate); err != nil {
				logs.Warnf("failed to update month pull task to finished, opt: %+v, err: %s, rid: %s",
					opt, err.Error(), kt.Rid)
				return err
			}
			return nil
		}
		if err := actcli.GetDataService().Global.Bill.UpdateBillMonthPullTask(kt, mtUpdate); err != nil {
			logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return err
		}
	}
}

func (act MonthTaskAction) split(
	kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption,
	monthTask *bill.BillMonthTaskResult, accountMap map[string]struct{}, offset uint64) (
	cnt int, finished bool, err error) {

	limit := runner.GetBatchSize(kt)
	if offset >= monthTask.PullIndex {
		return 0, true, nil
	}
	isFinished := false
	if offset+limit > monthTask.PullIndex {
		limit = monthTask.PullIndex - offset
		isFinished = true
	}

	name := fmt.Sprintf("%d-%d.csv", offset, limit)
	tmpReq := &bill.RawBillItemQueryReq{
		Vendor:        monthTask.Vendor,
		RootAccountID: monthTask.RootAccountID,
		MainAccountID: enumor.MonthRawBillPathName,
		BillYear:      fmt.Sprintf("%d", monthTask.BillYear),
		BillMonth:     fmt.Sprintf("%02d", monthTask.BillMonth),
		Version:       fmt.Sprintf("%d", monthTask.VersionID),
		BillDate:      enumor.MonthRawBillSpecialDatePathName,
		FileName:      name,
	}
	resp, err := actcli.GetDataService().Global.Bill.QueryRawBillItems(kt, tmpReq)
	if err != nil {
		logs.Warnf("failed to get raw bill item for %v, err %s, rid: %s", tmpReq, err.Error(), kt.Rid)
		return 0, false, fmt.Errorf("failed to get raw bill item for %v, err %s", tmpReq, err.Error())
	}
	tmpBillItemList, err := runner.Split(kt, monthTask.RootAccountID, monthTask.BillYear, monthTask.BillMonth,
		resp.Details)
	if err != nil {
		logs.Warnf("failed to split bill item, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
		return 0, false, err
	}
	commonOpt := &bill.ItemCommonOpt{Vendor: opt.Vendor, Year: opt.BillYear, Month: opt.BillMonth}
	for i, itemsBatch := range slice.Split(tmpBillItemList, constant.BatchOperationMaxLimit) {
		for idx := range itemsBatch {
			// force bill day as zero to represent month bill
			itemsBatch[idx].BillDay = 0
			accountMap[itemsBatch[idx].MainAccountID] = struct{}{}
		}
		createReq := &bill.BatchRawBillItemCreateReq{ItemCommonOpt: commonOpt, Items: itemsBatch}
		_, err = actcli.GetDataService().Global.Bill.BatchCreateBillItem(kt, createReq)
		if err != nil {
			logs.Warnf("failed to batch create bill item of batch idx %d, err: %s, opt: %+v, rid: %s",
				i, err.Error(), opt, kt.Rid)
			return 0, false, fmt.Errorf("failed to batch create bill item, err: %s, opt: %+v", err.Error(), opt)
		}
	}

	logs.Infof("split bill item for opt %+v done, offset: %d, limit: %d", opt, offset, limit)
	return len(resp.Details), isFinished, nil
}

func getCleanBillItemFilter(opt *MonthTaskActionOption) *filter.Expression {
	expressions := []*filter.AtomRule{
		// do not set main_account_id
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		// special day 0 for month bill
		tools.RuleEqual("bill_day", 0),
	}
	return tools.ExpressionAnd(expressions...)
}

func (act MonthTaskAction) cleanBillItem(kt *kit.Kit, opt *MonthTaskActionOption) error {
	batch := 0
	commonOpt := &bill.ItemCommonOpt{Vendor: opt.Vendor, Year: opt.BillYear, Month: opt.BillMonth}

	for {
		listReq := &bill.BillItemListReq{
			ItemCommonOpt: commonOpt,
			ListReq:       &core.ListReq{Filter: getCleanBillItemFilter(opt), Page: core.NewCountPage()},
		}
		result, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, listReq)
		if err != nil {
			logs.Warnf("count bill item for %+v failed, err %s, rid %s", opt, err.Error(), kt.Rid)
			return fmt.Errorf("count bill item for %+vfailed, err %s", opt, err.Error())
		}
		delReq := &bill.BillItemDeleteReq{ItemCommonOpt: commonOpt, Filter: getCleanBillItemFilter(opt)}
		if result.Count > 0 {
			if err := actcli.GetDataService().Global.Bill.BatchDeleteBillItem(kt, delReq); err != nil {
				return fmt.Errorf("delete 500 of %d bill item for %+v failed, err %s",
					result.Count, opt, err.Error())
			}
			count := min(result.Count, constant.BatchOperationMaxLimit)
			logs.Infof("successfully delete batch %d bill item for %+v, rid %s", count, opt, kt.Rid)
			batch = batch + 1
			continue
		}
		break
	}
	return nil
}

func (act MonthTaskAction) runMainAccountSummary(kt *kit.Kit, opt *MonthTaskActionOption,
	task *bill.BillMonthTaskResult, itemList []billcore.MonthTaskSummaryDetailItem) error {

	commonOpt := &bill.ItemCommonOpt{Vendor: opt.Vendor, Year: opt.BillYear, Month: opt.BillMonth}
	for i, item := range itemList {
		if item.IsFinished {
			continue
		}
		flt := tools.ExpressionAnd(
			tools.RuleEqual("root_account_id", opt.RootAccountID),
			tools.RuleEqual("main_account_id", item.MainAccountID),
			tools.RuleEqual("bill_year", opt.BillYear),
			tools.RuleEqual("bill_month", opt.BillMonth),
			tools.RuleEqual("vendor", task.Vendor),
			// special day 0 for month bill
			tools.RuleEqual("bill_day", 0),
		)
		listReq := &bill.BillItemListReq{
			ItemCommonOpt: commonOpt,
			ListReq:       &core.ListReq{Filter: flt, Page: core.NewCountPage()},
		}
		result, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, listReq)
		if err != nil {
			logs.Warnf("count bill item for %+v %+v failed, err: %s, rid: %s", opt, item, err.Error(), kt.Rid)
			return fmt.Errorf("count bill item for %+v %+v failed, err: %s", opt, item, err.Error())
		}
		currency := enumor.CurrencyUSD
		cost := decimal.NewFromFloat(0)
		count := result.Count
		limit := uint64(500)

		for start := uint64(0); start < result.Count; start = start + limit {
			listReq := &bill.BillItemListReq{
				ItemCommonOpt: commonOpt,
				ListReq: &core.ListReq{Filter: flt,
					Page: &core.BasePage{Start: uint32(start), Limit: uint(limit)}},
			}
			result, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, listReq)
			if err != nil {
				logs.Warnf("get %d-%d bill item for %+v %+v failed, err: %s, rid: %s",
					start, limit, opt, item, err.Error(), kt.Rid)
				return err
			}
			for _, item := range result.Details {
				if len(item.Currency) != 0 && len(currency) == 0 {
					currency = item.Currency
				}
				cost = cost.Add(item.Cost)
			}
		}
		itemList[i].IsFinished = true
		itemList[i].Currency = currency
		itemList[i].Cost = cost
		itemList[i].Count = count
		marshalDetail, err := json.Marshal(itemList)
		if err != nil {
			logs.Warnf("marshal detail failed, err: %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		if err := actcli.GetDataService().Global.Bill.UpdateBillMonthPullTask(kt, &bill.BillMonthTaskUpdateReq{
			ID:            task.ID,
			SummaryDetail: string(marshalDetail),
		}); err != nil {
			logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return err
		}
	}
	return nil
}

func (act MonthTaskAction) runSummary(kt *kit.Kit, opt *MonthTaskActionOption) error {
	task, err := getMonthPullTask(kt, opt)
	if err != nil {
		return err
	}
	var itemList []billcore.MonthTaskSummaryDetailItem
	if task.SummaryDetail != "" {
		if err := json.Unmarshal([]byte(task.SummaryDetail), &itemList); err != nil {
			logs.Warnf("decode %s to []billcore.MonthTaskSummaryDetailItem failed, err: %s, rid: %s",
				task.SummaryDetail, err.Error(), kt.Rid)
			return err
		}
	}

	if err := act.runMainAccountSummary(kt, opt, task, itemList); err != nil {
		return err
	}

	if err := actcli.GetDataService().Global.Bill.UpdateBillMonthPullTask(kt, &bill.BillMonthTaskUpdateReq{
		ID:    task.ID,
		State: enumor.RootAccountMonthBillTaskStateAccounted,
	}); err != nil {
		logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
		return err
	}
	return nil
}

func getMonthPullTask(kt *kit.Kit, opt *MonthTaskActionOption) (*bill.BillMonthTaskResult, error) {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillMonthPullTask(kt, &bill.BillMonthTaskListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		logs.Warnf("get month pull task failed, opt: %+v, err: %s, rid: %s", err.Error(), kt.Rid)
		return nil, fmt.Errorf("get month pull task failed, opt: %+v, err: %s", opt, err.Error())
	}
	if len(result.Details) != 1 {
		logs.Warnf("get invalid length month pull task, resp: %v, rid: %s", result, kt.Rid)
		return nil, fmt.Errorf("get invalid length month pull task, resp: %v", result)
	}
	return result.Details[0], nil
}
