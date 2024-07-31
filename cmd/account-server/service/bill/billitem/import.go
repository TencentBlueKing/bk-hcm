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

package billitem

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"hcm/cmd/account-server/logics/bill/puller/daily"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	billcore "hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
)

var (
	zenlayerBillItemRefType = reflect.TypeOf(billcore.ZenlayerRawBillItem{})
)

// ImportBillItems 导入账单明细
func (b *billItemSvc) ImportBillItems(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(bill.ImportBillItemReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Create}})
	if err != nil {
		return nil, err
	}

	mainAccountIDs := make([]string, 0, len(req.Items))
	for _, item := range req.Items {
		mainAccountIDs = append(mainAccountIDs, item.MainAccountID)
	}

	// 清理所有已存在的账单明细，不区分version
	itemCommonOpt := &dsbill.ItemCommonOpt{
		Vendor: vendor,
		Year:   req.BillYear,
		Month:  req.BillMonth,
	}
	billDeleteReq := &dsbill.BillItemDeleteReq{
		ItemCommonOpt: itemCommonOpt,
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", vendor),
			tools.RuleIn("main_account_id", slice.Unique(mainAccountIDs)),
			tools.RuleEqual("bill_year", req.BillYear),
			tools.RuleEqual("bill_month", req.BillMonth),
		),
	}
	if err = b.client.DataService().Global.Bill.BatchDeleteBillItem(cts.Kit,
		billDeleteReq); err != nil {
		return nil, err
	}

	billCreateReq := &dsbill.BatchBillItemCreateReq[json.RawMessage]{
		ItemCommonOpt: itemCommonOpt,
		Items:         req.Items,
	}
	_, err = b.client.DataService().Global.Bill.BatchCreateBillItem(cts.Kit, billCreateReq)
	if err != nil {
		return nil, err
	}

	if err = b.createOrUpdatePullTasks(cts.Kit, vendor, req.BillYear, req.BillMonth, req.Items); err != nil {
		return nil, err
	}
	return nil, nil
}

func (b *billItemSvc) mapPullTaskList(kt *kit.Kit, summaryMains []*dsbill.BillSummaryMainResult) (
	map[string][]*dsbill.BillDailyPullTaskResult, error) {

	result := make(map[string][]*dsbill.BillDailyPullTaskResult, len(summaryMains))
	for _, item := range summaryMains {
		dp := &daily.DailyPuller{
			RootAccountID: item.RootAccountID,
			MainAccountID: item.MainAccountID,
			ProductID:     item.ProductID,
			BkBizID:       item.BkBizID,
			Vendor:        item.Vendor,
			BillYear:      item.BillYear,
			BillMonth:     item.BillMonth,
			Version:       item.CurrentVersion,
			Client:        b.client,
		}
		pullTaskList, err := dp.GetPullTaskList(kt)
		if err != nil {
			return nil, err
		}
		result[item.MainAccountID] = pullTaskList
	}
	return result, nil
}

// create daily pull task
func (b *billItemSvc) createOrUpdatePullTasks(kt *kit.Kit, vendor enumor.Vendor, billYear int, billMonth int, items []dsbill.BillItemCreateReq[json.RawMessage]) error {

	mainAccountIDs := make([]string, 0, len(items))
	for _, item := range items {
		mainAccountIDs = append(mainAccountIDs, item.MainAccountID)
	}

	summaryMainResults, err := b.listSummaryMainByMainAccountIDs(kt, vendor, slice.Unique(mainAccountIDs), billYear, billMonth)
	if err != nil {
		return err
	}
	mapPullTasks, err := b.mapPullTaskList(kt, summaryMainResults)
	if err != nil {
		return err
	}

	accountToBillDayMap := make(map[string][]int, len(items))
	for _, summaryMain := range summaryMainResults {
		billDayList := make([]int, 0)
		for _, pullTask := range mapPullTasks[summaryMain.MainAccountID] {
			if err = b.updatePullTaskStateAndDailySummaryFlowID(kt, pullTask); err != nil {
				return err
			}
			billDayList = append(billDayList, pullTask.BillDay)
		}
		accountToBillDayMap[summaryMain.MainAccountID] = billDayList
	}

	createReqs := make(map[string]*dsbill.BillDailyPullTaskCreateReq, len(items))
	for _, item := range items {
		inSlice := slice.IsItemInSlice[int](accountToBillDayMap[item.MainAccountID], item.BillDay)
		if inSlice {
			// already has a daily pull task for this bill day, skip
			continue
		}

		// 去重, 防止items中有重复日期的账单
		key := fmt.Sprintf("%s-%s-%d-%d-%d-%d", item.RootAccountID, item.MainAccountID,
			item.BillYear, item.BillMonth, item.BillDay, item.VersionID)
		createReqs[key] = &dsbill.BillDailyPullTaskCreateReq{
			RootAccountID: item.RootAccountID,
			MainAccountID: item.MainAccountID,
			Vendor:        item.Vendor,
			ProductID:     item.ProductID,
			BkBizID:       item.BkBizID,
			BillYear:      item.BillYear,
			BillMonth:     item.BillMonth,
			BillDay:       item.BillDay,
			VersionID:     item.VersionID,
			State:         enumor.MainAccountRawBillPullStateSplit,
			Count:         0,
			Currency:      "",
			Cost:          decimal.NewFromFloat(0),
			FlowID:        "",
		}
	}

	for _, req := range createReqs {
		_, err := b.client.DataService().Global.Bill.CreateBillDailyPullTask(kt, req)
		if err != nil {
			return err
		}
	}

	return nil
}

// reset daily pull task to split state and clear daily summary flow id
func (b *billItemSvc) updatePullTaskStateAndDailySummaryFlowID(kt *kit.Kit, task *dsbill.BillDailyPullTaskResult) error {
	updateReq := &dsbill.BillDailyPullTaskUpdateReq{
		ID:                 task.ID,
		State:              enumor.MainAccountRawBillPullStateSplit,
		DailySummaryFlowID: "",
	}
	err := b.client.DataService().Global.Bill.UpdateBillDailyPullTask(kt, updateReq)
	if err != nil {
		return err
	}
	return nil
}

func (b *billItemSvc) listSummaryMainByMainAccountIDs(kt *kit.Kit, vendor enumor.Vendor, mainAccountIDs []string,
	billYear, billMonth int) ([]*dsbill.BillSummaryMainResult, error) {

	result := make([]*dsbill.BillSummaryMainResult, 0, len(mainAccountIDs))
	for _, ids := range slice.Split(mainAccountIDs, int(core.DefaultMaxPageLimit)) {
		listReq := &dsbill.BillSummaryMainListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleEqual("bill_year", billYear),
				tools.RuleEqual("bill_month", billMonth),
				tools.RuleIn("main_account_id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		list, err := b.client.DataService().Global.Bill.ListBillSummaryMain(kt, listReq)
		if err != nil {
			return nil, err
		}
		result = append(result, list.Details...)
	}
	return result, nil
}

// excelRowsIterator traverse each row in Excel file by given operation
func excelRowsIterator(kt *kit.Kit, reader io.Reader, sheetIdx, batchSize int,
	opFunc func([][]string, error) error) error {

	excel, err := excelize.OpenReader(reader)
	if err != nil {
		logs.Errorf("fialed to create excel reader, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	defer excel.Close()

	sheetName := excel.GetSheetName(sheetIdx)

	rows, err := excel.Rows(sheetName)
	if err != nil {
		logs.Errorf("fail to read rows from sheet(%s), err: %v, rid: %s", sheetName, err, kt.Rid)
		return err
	}
	defer rows.Close()
	rows.Next() // skip header row

	rowBatch := make([][]string, 0, batchSize)
	// traverse all rows
	for rows.Next() {
		columns, err := rows.Columns()
		if err != nil {
			return opFunc(nil, err)
		}
		rowBatch = append(rowBatch, columns)
		if len(rowBatch) < batchSize {
			continue
		}
		if err := opFunc(rowBatch, nil); err != nil {
			return err
		}
		rowBatch = rowBatch[:0]

	}
	return opFunc(rowBatch, nil)
}
