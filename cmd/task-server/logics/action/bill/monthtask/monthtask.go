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
	"fmt"
	"path/filepath"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	"hcm/pkg/api/data-service/bill"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
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
	Step          enumor.MonthTaskStep `json:"stage" validate:"required"`
	RootAccountID string               `json:"root_account_id" validate:"required"`
	BillYear      int                  `json:"bill_year" validate:"required"`
	BillMonth     int                  `json:"bill_month" validate:"required"`
	Vendor        enumor.Vendor        `json:"vendor" validate:"required"`
	Extension     map[string]string    `json:"extension"`
}

// Validate ...
func (o MonthTaskActionOption) Validate() error {
	return validator.Validate.Struct(o)
}
func (o MonthTaskActionOption) String() string {
	return fmt.Sprintf("[%s]%s %s %d-%2d", o.Vendor, o.Type, o.RootAccountID, o.BillYear, o.BillMonth)
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
	runner, err := GetRunner(opt.Vendor, opt.Type)
	if err != nil {
		return nil, err
	}
	switch opt.Step {
	case enumor.MonthTaskStepPull:
		if err := act.runPull(kt.Kit(), runner, opt); err != nil {
			logs.Errorf("fail to pull month task, opt: %s, err: %v, rid: %s", opt.String(), err, kt.Kit().Rid)
			return nil, err
		}
		return nil, nil
	case enumor.MonthTaskStepSplit:
		if err := act.runSplit(kt.Kit(), runner, opt); err != nil {
			logs.Errorf("fail to split month task, opt: %s, err: %v, rid: %s", opt.String(), err, kt.Kit().Rid)
			return nil, err
		}
		return nil, nil
	case enumor.MonthTaskStepSummary:
		if err := act.runSummary(kt.Kit(), runner, opt); err != nil {
			logs.Errorf("fail to summary month task, opt: %s, err: %v, rid: %s", opt.String(), err, kt.Kit().Rid)
			return nil, err
		}
		return nil, nil
	default:
		return nil, errf.New(errf.InvalidParameter, fmt.Sprintf(
			"invalid month task type %s", opt.Step))
	}
}

func (act MonthTaskAction) runPull(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption) error {

	// 清除原始账单
	if err := act.cleanRawBills(kt, opt); err != nil {
		logs.Errorf("fail to clean raw bills for pull month bill, err:%v, rid: %s", err, kt.Rid)
		return err
	}

	for {
		task, err := getMonthTask(kt, opt)
		if err != nil {
			return err
		}

		rawBillItemList, isFinished, err := runner.Pull(kt, opt, task.PullIndex)
		if err != nil {
			return err
		}
		lenRawBillItemList := len(rawBillItemList)
		if lenRawBillItemList == 0 {
			logs.Infof("month task %s pulled 0 records, skip, rid: %s", task.String(), kt.Rid)
			return nil
		}
		filename := getMonthTaskRawBillFilename(task, task.PullIndex, uint64(lenRawBillItemList))
		storeReq := &bill.RawBillCreateReq{
			RawBillPathParam: bill.RawBillPathParam{
				Vendor:        opt.Vendor,
				RootAccountID: task.RootAccountID,
				MainAccountID: enumor.MonthRawBillPathName,
				BillYear:      fmt.Sprintf("%d", task.BillYear),
				BillMonth:     fmt.Sprintf("%02d", task.BillMonth),
				// 将类型作为特殊日期
				BillDate: string(task.Type),
				Version:  fmt.Sprintf("%d", task.VersionID),
				FileName: filename,
			},

			Items: rawBillItemList,
		}
		databillCli := actcli.GetDataService().Global.Bill
		_, err = databillCli.CreateRawBill(kt, storeReq)
		if err != nil {
			logs.Errorf("failed to create month raw bill, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return fmt.Errorf("failed to create month raw bill, opt: %+v, err: %s", opt, err.Error())
		}
		logs.Infof("month task %+v pulled %d records, continue", opt, lenRawBillItemList)
		if isFinished {
			updateToPulledReq := &bill.BillMonthTaskUpdateReq{
				ID:        task.ID,
				Count:     task.Count + uint64(lenRawBillItemList),
				PullIndex: task.PullIndex + uint64(lenRawBillItemList),
				State:     enumor.RootAccountMonthBillTaskStatePulled,
			}
			if err := databillCli.UpdateBillMonthTask(kt, updateToPulledReq); err != nil {
				logs.Errorf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
				return err
			}
			return nil
		}
		updateIdxReq := &bill.BillMonthTaskUpdateReq{
			ID:        task.ID,
			Count:     task.Count + uint64(lenRawBillItemList),
			PullIndex: task.PullIndex + uint64(lenRawBillItemList),
		}
		if err := databillCli.UpdateBillMonthTask(kt, updateIdxReq); err != nil {
			logs.Errorf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return err
		}
	}
}

func (act MonthTaskAction) runSplit(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption) error {
	// step1 清理原有月度任务的bill item，因为有可能之前存在中途失败的脏数据了
	if err := act.cleanBillItem(kt, runner, opt); err != nil {
		return err
	}
	// step2 进行分账
	var splitMainAccountMap = make(map[string]struct{})
	curlIndex := uint64(0)
	for {
		task, err := getMonthTask(kt, opt)
		if err != nil {
			logs.Errorf("fail to get month task for splitting, err: %s, rid: %s", err.Error(), kt.Rid)
			return err
		}
		cnt, isFinished, err := act.split(kt, runner, opt, task, splitMainAccountMap, curlIndex)
		if err != nil {
			logs.Errorf("fail to split for bill month task, err: %v, opt: %+v, rid: %s", err, opt, kt.Rid)
			return err
		}
		curlIndex += uint64(cnt)

		mtUpdate := &bill.BillMonthTaskUpdateReq{
			ID:         task.ID,
			SplitIndex: curlIndex,
		}
		if isFinished {
			var itemList = make([]billcore.MonthTaskSummaryDetailItem, 0, len(splitMainAccountMap))
			for mainAccountID := range splitMainAccountMap {
				itemList = append(itemList, billcore.MonthTaskSummaryDetailItem{MainAccountID: mainAccountID})
			}
			mtUpdate.SummaryDetail = itemList
			mtUpdate.State = enumor.RootAccountMonthBillTaskStateSplit
			if err := actcli.GetDataService().Global.Bill.UpdateBillMonthTask(kt, mtUpdate); err != nil {
				logs.Warnf("failed to update month pull task to finished, opt: %+v, err: %s, rid: %s",
					opt, err.Error(), kt.Rid)
				return err
			}
			return nil
		}
		if err := actcli.GetDataService().Global.Bill.UpdateBillMonthTask(kt, mtUpdate); err != nil {
			logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
			return err
		}
	}
}

func (act MonthTaskAction) split(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption,
	monthTask *billcore.MonthTask, accountMap map[string]struct{}, offset uint64) (
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

	name := getMonthTaskRawBillFilename(monthTask, offset, limit)
	tmpReq := &bill.RawBillItemQueryReq{
		Vendor:        monthTask.Vendor,
		RootAccountID: monthTask.RootAccountID,
		MainAccountID: enumor.MonthRawBillPathName,
		BillYear:      fmt.Sprintf("%d", monthTask.BillYear),
		BillMonth:     fmt.Sprintf("%02d", monthTask.BillMonth),
		Version:       fmt.Sprintf("%d", monthTask.VersionID),
		BillDate:      string(opt.Type),
		FileName:      name,
	}
	resp, err := actcli.GetDataService().Global.Bill.QueryRawBillItems(kt, tmpReq)
	if err != nil {
		logs.Errorf("failed to get raw bill item for %v, err: %v, rid: %s", tmpReq, err, kt.Rid)
		return 0, false, fmt.Errorf("failed to get raw bill item for %v, err %s", tmpReq, err.Error())
	}
	tmpBillItemList, err := runner.Split(kt, opt, resp.Details)
	if err != nil {
		logs.Errorf("failed to split bill item, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return 0, false, err
	}
	commonOpt := &bill.ItemCommonOpt{Vendor: opt.Vendor, Year: opt.BillYear, Month: opt.BillMonth}
	for i, itemsBatch := range slice.Split(tmpBillItemList, constant.BatchOperationMaxLimit) {
		for idx := range itemsBatch {
			// force bill day as zero to represent month bill
			itemsBatch[idx].BillDay = enumor.MonthTaskSpecialBillDay
			accountMap[itemsBatch[idx].MainAccountID] = struct{}{}
		}
		createReq := &bill.BatchRawBillItemCreateReq{ItemCommonOpt: commonOpt, Items: itemsBatch}
		_, err = actcli.GetDataService().Global.Bill.BatchCreateBillItem(kt, createReq)
		if err != nil {
			logs.Warnf("failed to batch create bill item of batch idx %d, err: %v, opt: %+v, rid: %s",
				i, err, opt, kt.Rid)
			return 0, false, fmt.Errorf("failed to batch create bill item, err: %v, opt: %+v", err, opt)
		}
	}

	logs.Infof("split bill item for opt %+v done, cnt: %d, offset: %d, limit: %d, rid: %s",
		opt, len(resp.Details), offset, limit, kt.Rid)
	return len(resp.Details), isFinished, nil
}

func getMonthTaskRawBillFilename(monthTask *billcore.MonthTask, offset uint64, limit uint64) string {
	name := fmt.Sprintf("%s-%d-%d.csv", monthTask.Type, offset, limit)
	return name
}

func getCleanBillItemFilter(opt *MonthTaskActionOption, productCodes []string) *filter.Expression {
	expressions := []*filter.AtomRule{
		// do not set main_account_id
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("vendor", opt.Vendor),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		// special day 0 for month bill
		tools.RuleEqual("bill_day", enumor.MonthTaskSpecialBillDay),
		tools.RuleIn("hc_product_code", productCodes),
	}
	return tools.ExpressionAnd(expressions...)
}

func (act MonthTaskAction) cleanBillItem(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption) error {

	batch := 0
	commonOpt := &bill.ItemCommonOpt{Vendor: opt.Vendor, Year: opt.BillYear, Month: opt.BillMonth}
	delFilter := getCleanBillItemFilter(opt, runner.GetHcProductCodes())
	for {
		listReq := &bill.BillItemListReq{
			ItemCommonOpt: commonOpt,
			ListReq:       &core.ListReq{Filter: delFilter, Page: core.NewCountPage()},
		}
		result, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, listReq)
		if err != nil {
			logs.Warnf("count bill item for %+v failed, err %s, rid %s", opt, err.Error(), kt.Rid)
			return fmt.Errorf("count bill item for %+v failed, err %s", opt, err.Error())
		}
		delReq := &bill.BillItemDeleteReq{ItemCommonOpt: commonOpt, Filter: delFilter}
		if result.Count > 0 {
			if err := actcli.GetDataService().Global.Bill.BatchDeleteBillItem(kt, delReq); err != nil {
				return fmt.Errorf("delete 100 of %d bill item for %+v failed, err %s",
					result.Count, opt, err.Error())
			}
			count := min(result.Count, constant.BatchOperationMaxLimit)
			logs.Infof("successfully delete (%d/%d) bill item, batch_idx: %d, opt: %+v, rid: %s",
				count, result.Count, batch, opt, kt.Rid)
			batch = batch + 1
			continue
		}
		break
	}
	return nil
}

func (act MonthTaskAction) runMainAccountSummary(kt *kit.Kit, codes []string, task *billcore.MonthTask,
	itemList []billcore.MonthTaskSummaryDetailItem) error {

	commonOpt := &bill.ItemCommonOpt{Vendor: task.Vendor, Year: task.BillYear, Month: task.BillMonth}
	for i, item := range itemList {
		if item.IsFinished {
			continue
		}
		flt := tools.ExpressionAnd(
			tools.RuleEqual("root_account_id", task.RootAccountID),
			tools.RuleEqual("main_account_id", item.MainAccountID),
			tools.RuleEqual("bill_year", task.BillYear),
			tools.RuleEqual("bill_month", task.BillMonth),
			tools.RuleEqual("vendor", task.Vendor),
			// special day 0 for month bill
			tools.RuleEqual("bill_day", enumor.MonthTaskSpecialBillDay),
			tools.RuleIn("hc_product_code", codes),
		)
		listReq := &bill.BillItemListReq{
			ItemCommonOpt: commonOpt,
			ListReq:       &core.ListReq{Filter: flt, Page: core.NewCountPage()},
		}
		result, err := actcli.GetDataService().Global.Bill.ListBillItem(kt, listReq)
		if err != nil {
			logs.Warnf("count bill item for %s %+v failed, err: %v, rid: %s", task.String(), item, err, kt.Rid)
			return fmt.Errorf("count bill item for %+v %+v failed, err: %v", task.String(), item, err)
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
				logs.Warnf("get %d-%d bill item for %s %+v failed, err: %v, rid: %s",
					start, limit, task.String(), item, err, kt.Rid)
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

		req := &bill.BillMonthTaskUpdateReq{
			ID:            task.ID,
			SummaryDetail: itemList,
		}
		if err := actcli.GetDataService().Global.Bill.UpdateBillMonthTask(kt, req); err != nil {
			logs.Errorf("failed to update month pull task: %s, err: %v, rid: %s", task.String(), err, kt.Rid)
			return err
		}
	}
	return nil
}

func (act MonthTaskAction) runSummary(kt *kit.Kit, runner MonthTaskRunner, opt *MonthTaskActionOption) error {
	task, err := getMonthTask(kt, opt)
	if err != nil {
		return err
	}

	if err := act.runMainAccountSummary(kt, runner.GetHcProductCodes(), task, task.SummaryDetail); err != nil {
		return err
	}

	req := &bill.BillMonthTaskUpdateReq{
		ID:    task.ID,
		State: enumor.RootAccountMonthBillTaskStateAccounted,
	}
	if err := actcli.GetDataService().Global.Bill.UpdateBillMonthTask(kt, req); err != nil {
		logs.Warnf("failed to update month pull task, opt: %+v, err: %s, rid: %s", opt, err.Error(), kt.Rid)
		return err
	}
	return nil
}

func getMonthTask(kt *kit.Kit, opt *MonthTaskActionOption) (*billcore.MonthTask, error) {
	expressions := []*filter.AtomRule{
		tools.RuleEqual("root_account_id", opt.RootAccountID),
		tools.RuleEqual("bill_year", opt.BillYear),
		tools.RuleEqual("bill_month", opt.BillMonth),
		tools.RuleEqual("type", opt.Type),
	}
	result, err := actcli.GetDataService().Global.Bill.ListBillMonthTask(kt, &bill.BillMonthTaskListReq{
		Filter: tools.ExpressionAnd(expressions...),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	})
	if err != nil {
		logs.Warnf("get month task failed, opt: %+v, err: %v, rid: %s", opt, err, kt.Rid)
		return nil, fmt.Errorf("get month task failed, opt: %+v, err: %s", opt, err.Error())
	}
	if len(result.Details) != 1 {
		logs.Errorf("get invalid length month task, resp: %+v, rid: %s", result, kt.Rid)
		return nil, fmt.Errorf("get invalid length month task, resp: %v", result)
	}
	return result.Details[0], nil
}

func (act MonthTaskAction) cleanRawBills(kt *kit.Kit, opt *MonthTaskActionOption) error {

	task, err := getMonthTask(kt, opt)
	if err != nil {
		logs.Errorf("fail to list month task, err: %s, rid: %s", err, kt.Rid)
		return err
	}

	listResult, err := actcli.GetDataService().Global.Bill.ListRawBillFileNames(
		kt, &bill.RawBillItemNameListReq{
			Vendor:        task.Vendor,
			RootAccountID: task.RootAccountID,
			MainAccountID: enumor.MonthRawBillPathName,
			BillYear:      fmt.Sprintf("%d", task.BillYear),
			BillMonth:     fmt.Sprintf("%02d", task.BillMonth),
			BillDate:      string(task.Type),
			Version:       fmt.Sprintf("%d", task.VersionID),
		})
	if err != nil {
		logs.Warnf("list raw bill filenames failed, err: %s, vendor: %s, rid: %s", err, task.Vendor, kt.Rid)
		return err
	}
	for _, filename := range listResult.Filenames {
		name := filepath.Base(filename)
		req := &bill.RawBillDeleteReq{
			RawBillPathParam: bill.RawBillPathParam{
				Vendor:        task.Vendor,
				RootAccountID: task.RootAccountID,
				MainAccountID: enumor.MonthRawBillPathName,
				BillYear:      fmt.Sprintf("%d", task.BillYear),
				BillMonth:     fmt.Sprintf("%02d", task.BillMonth),
				BillDate:      string(task.Type),
				Version:       fmt.Sprintf("%d", task.VersionID),
				FileName:      name,
			},
		}
		if err := actcli.GetDataService().Global.Bill.DeleteRawBill(kt, req); err != nil {
			logs.Warnf("delete raw bill %s failed, err: %s, vendor: %s, rid: %s",
				filename, err.Error(), task.Vendor, kt.Rid)
			return err
		}
	}
	return nil
}
