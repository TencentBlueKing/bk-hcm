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
	"time"

	"hcm/cmd/account-server/logics/bill/puller/daily"
	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
	accountset "hcm/pkg/api/core/account-set"
	billcore "hcm/pkg/api/core/bill"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/iam/meta"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
	"hcm/pkg/tools/slice"
	"hcm/pkg/tools/times"

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
	mainAccountIDs = slice.Unique(mainAccountIDs)
	if len(mainAccountIDs) == 0 {
		return nil, errf.New(errf.InvalidParameter, "items.main_account_id is required")
	}

	if err = b.validateSummaryAccountState(cts.Kit, mainAccountIDs, vendor, req.BillYear, req.BillMonth); err != nil {
		return nil, err
	}

	if err = b.deleteBillItemsByMainAccountIDs(cts, vendor, req.BillYear, req.BillMonth, mainAccountIDs); err != nil {
		logs.Errorf("delete bill items failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	if err = b.createBillItems(cts, vendor, req); err != nil {
		return nil, err
	}

	if err = b.ensurePullTasks(cts.Kit, vendor, req); err != nil {
		logs.Errorf("ensure pull tasks failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}
	return nil, nil
}

func (b *billItemSvc) validateSummaryAccountState(kt *kit.Kit, mainAccountsIDs []string,
	vendor enumor.Vendor, billYear, billMonth int) error {
	summaryMains, err := b.listSummaryMainByMainAccountIDs(kt, vendor, mainAccountsIDs, billYear, billMonth)
	if err != nil {
		return err
	}
	for _, summary := range summaryMains {
		if summary.State != enumor.MainAccountBillSummaryStateAccounting {
			logs.Errorf("summaryMainAccount(%s) state is not accounting, can't import bill, rid: %s",
				summary.ID, kt.Rid)
			return fmt.Errorf("summaryMainAccount(%s) state is not accounting, can't import bill",
				summary.ID)
		}
	}
	return nil
}

func (b *billItemSvc) createBillItems(cts *rest.Contexts, vendor enumor.Vendor,
	req *bill.ImportBillItemReq) error {

	itemCommonOpt := &dsbill.ItemCommonOpt{
		Vendor: vendor,
		Year:   req.BillYear,
		Month:  req.BillMonth,
	}
	for _, items := range slice.Split(req.Items, constant.BatchOperationMaxLimit) {
		billCreateReq := &dsbill.BatchBillItemCreateReq[json.RawMessage]{
			ItemCommonOpt: itemCommonOpt,
			Items:         items,
		}
		_, err := b.client.DataService().Global.Bill.BatchCreateBillItem(cts.Kit, billCreateReq)
		if err != nil {
			logs.Errorf("create bill items failed, err: %v, rid: %s", err, cts.Kit.Rid)
			return err
		}
	}
	return nil
}

func (b *billItemSvc) deleteBillItemsByMainAccountIDs(cts *rest.Contexts, vendor enumor.Vendor, billYear, billMonth int,
	mainAccountIDs []string) error {

	if len(mainAccountIDs) == 0 {
		return nil
	}

	// 清理所有已存在的账单明细，不区分version
	itemCommonOpt := &dsbill.ItemCommonOpt{
		Vendor: vendor,
		Year:   billYear,
		Month:  billMonth,
	}
	billDeleteReq := &dsbill.BillItemDeleteReq{
		ItemCommonOpt: itemCommonOpt,
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", vendor),
			tools.RuleIn("main_account_id", slice.Unique(mainAccountIDs)),
			tools.RuleEqual("bill_year", billYear),
			tools.RuleEqual("bill_month", billMonth),
		),
	}
	err := b.client.DataService().Global.Bill.BatchDeleteBillItem(cts.Kit, billDeleteReq)
	if err != nil {
		logs.Errorf("delete bill items by main_account_ids[%v] failed, err: %v, rid: %s",
			mainAccountIDs, err, cts.Kit.Rid)
		return err
	}
	return nil
}

func (b *billItemSvc) mapAccountIDToPullTaskList(kt *kit.Kit, summaryMains []*dsbill.BillSummaryMain) (
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
func (b *billItemSvc) ensurePullTasks(kt *kit.Kit, vendor enumor.Vendor,
	req *bill.ImportBillItemReq) error {

	mainAccounts, err := b.listMainAccount(kt, vendor)
	if err != nil {
		logs.Errorf("list main account failed, err: %v, rid: %s, vendor: %s", err, kt.Rid, vendor)
		return err
	}
	mainAccountIDs := make([]string, 0, len(req.Items))
	for _, item := range mainAccounts {
		mainAccountIDs = append(mainAccountIDs, item.ID)
	}

	summaryMainResults, err := b.listSummaryMainByMainAccountIDs(kt, vendor,
		slice.Unique(mainAccountIDs), req.BillYear, req.BillMonth)
	if err != nil {
		logs.Errorf("list summary main by main account ids failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}
	mapPullTasks, err := b.mapAccountIDToPullTaskList(kt, summaryMainResults)
	if err != nil {
		logs.Errorf("map account id to pull task list failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	createReqs := make([]*dsbill.BillDailyPullTaskCreateReq, 0, len(req.Items))
	for _, summaryMain := range summaryMainResults {
		existDays := make([]int, 0)
		for _, pullTask := range mapPullTasks[summaryMain.MainAccountID] {
			if err = b.updatePullTaskStateToSplitAndResetDailySummaryFlowID(kt, pullTask.ID); err != nil {
				logs.Errorf("update pull task(%s) state failed, err: %v, rid: %s", pullTask.ID, err, kt.Rid)
				return err
			}
			existDays = append(existDays, pullTask.BillDay)
		}
		createReqs = append(createReqs, generateRemainingPullTask(existDays, summaryMain)...)
	}

	for _, createReq := range createReqs {
		if _, err = b.client.DataService().Global.Bill.CreateBillDailyPullTask(kt, createReq); err != nil {
			logs.Errorf("create bill daily pull task failed, err: %v, rid: %s", err, kt.Rid)
			return err
		}
	}
	return nil
}

func generateRemainingPullTask(existBillDays []int,
	summary *dsbill.BillSummaryMain) []*dsbill.BillDailyPullTaskCreateReq {

	days := times.GetMonthDays(summary.BillYear, time.Month(summary.BillMonth))
	result := make([]*dsbill.BillDailyPullTaskCreateReq, 0, len(days))
	for _, day := range days {
		if slice.IsItemInSlice[int](existBillDays, day) {
			continue
		}
		result = append(result, newPullTaskCreateReqFromSummaryMain(summary, day))
	}
	return result
}

func newPullTaskCreateReqFromSummaryMain(summaryMain *dsbill.BillSummaryMain,
	day int) *dsbill.BillDailyPullTaskCreateReq {

	return &dsbill.BillDailyPullTaskCreateReq{
		RootAccountID: summaryMain.RootAccountID,
		MainAccountID: summaryMain.MainAccountID,
		Vendor:        summaryMain.Vendor,
		ProductID:     summaryMain.ProductID,
		BkBizID:       summaryMain.BkBizID,
		BillYear:      summaryMain.BillYear,
		BillMonth:     summaryMain.BillMonth,
		BillDay:       day,
		VersionID:     summaryMain.CurrentVersion,
		State:         enumor.MainAccountRawBillPullStateSplit,
		Count:         0,
		Currency:      "",
		Cost:          decimal.NewFromFloat(0),
		FlowID:        "",
	}
}

func (b *billItemSvc) updatePullTaskStateToSplitAndResetDailySummaryFlowID(kt *kit.Kit, taskID string) error {
	emptyFlowID := ""
	updateReq := &dsbill.BillDailyPullTaskUpdateReq{
		ID:                 taskID,
		State:              enumor.MainAccountRawBillPullStateSplit,
		DailySummaryFlowID: &emptyFlowID,
	}
	err := b.client.DataService().Global.Bill.UpdateBillDailyPullTask(kt, updateReq)
	if err != nil {
		return err
	}
	return nil
}

func (b *billItemSvc) listSummaryMainByMainAccountIDs(kt *kit.Kit, vendor enumor.Vendor, mainAccountIDs []string,
	billYear, billMonth int) ([]*dsbill.BillSummaryMain, error) {

	result := make([]*dsbill.BillSummaryMain, 0, len(mainAccountIDs))
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
	opFunc func([][]string) error) error {

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
			return err
		}
		rowBatch = append(rowBatch, columns)
		if len(rowBatch) < batchSize {
			continue
		}
		if err := opFunc(rowBatch); err != nil {
			return err
		}
		// 清空rowBatch, 下次循环写入新数据
		rowBatch = rowBatch[:0]
	}
	return opFunc(rowBatch)
}

func (b *billItemSvc) listMainAccount(kt *kit.Kit, vendor enumor.Vendor) ([]*accountset.BaseMainAccount, error) {
	filter := tools.ExpressionAnd(
		tools.RuleEqual("vendor", vendor),
	)
	countReq := &core.ListReq{
		Filter: filter,
		Page:   core.NewCountPage(),
	}
	countResp, err := b.client.DataService().Global.MainAccount.List(kt, countReq)
	if err != nil {
		logs.Errorf("list main account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	total := countResp.Count

	result := make([]*accountset.BaseMainAccount, 0, total)
	for i := uint64(0); i < total; i += uint64(core.DefaultMaxPageLimit) {
		listReq := &core.ListReq{
			Filter: filter,
			Page: &core.BasePage{
				Start: uint32(i),
				Limit: core.DefaultMaxPageLimit,
			},
		}
		listResult, err := b.client.DataService().Global.MainAccount.List(kt, listReq)
		if err != nil {
			logs.Errorf("list main account failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		result = append(result, listResult.Details...)
	}

	return result, nil
}
