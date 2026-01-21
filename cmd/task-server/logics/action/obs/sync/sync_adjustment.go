/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package sync

import (
	"fmt"
	"time"

	actcli "hcm/cmd/task-server/logics/action/cli"
	"hcm/pkg/api/core"
	protocore "hcm/pkg/api/core/account-set"
	"hcm/pkg/api/core/bill"
	dataproto "hcm/pkg/api/data-service/account-set"
	"hcm/pkg/async/action"
	"hcm/pkg/async/action/run"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	typesdao "hcm/pkg/dal/dao/types"
	tableobs "hcm/pkg/dal/table/obs"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
	"github.com/shopspring/decimal"
)

const adjustmentSetIndex = "adjustment"
const adjustmentProductName = "BillAdjustment"
const adjustmentProductCode = "BillAdjustment"

// AdjustmentOption option for sync adjustment action
type AdjustmentOption struct {
	Vendor    enumor.Vendor `json:"vendor" validate:"required"`
	BillYear  int           `json:"bill_year" validate:"required"`
	BillMonth int           `json:"bill_month" validate:"required"`
}

// adjustmentOperator ...
type adjustmentOperator struct {
	clean  func(kt *kit.Kit, opt *AdjustmentOption) error
	insert func(*kit.Kit, []*bill.AdjustmentItem) error
}

var _ action.Action = new(SyncAdjustmentAction)
var _ action.ParameterAction = new(SyncAdjustmentAction)

// SyncAdjustmentAction define sync action
type SyncAdjustmentAction struct {
	exchangeRateMap map[enumor.CurrencyCode]*decimal.Decimal
	rootAccountMap  map[string]*dataproto.RootAccountGetBaseResult
}

// ParameterNew return request params.
func (act SyncAdjustmentAction) ParameterNew() interface{} {
	return new(AdjustmentOption)
}

// Name return action name
func (act SyncAdjustmentAction) Name() enumor.ActionName {
	return enumor.ActionObsAdjustmentSync
}

// Run ...
func (act SyncAdjustmentAction) Run(kt run.ExecuteKit, params interface{}) (interface{}, error) {
	opt, ok := params.(*AdjustmentOption)
	if !ok {
		return nil, errf.New(errf.InvalidParameter, "params type mismatch")
	}

	var operator adjustmentOperator
	switch opt.Vendor {
	case enumor.HuaWei:
		operator.clean = act.cleanHuawei
		operator.insert = act.convertHuawei
	case enumor.Aws:
		operator.clean = act.cleanAws
		operator.insert = act.convertAws
	case enumor.Gcp:
		operator.clean = act.cleanGcp
		operator.insert = act.convertGcp
	case enumor.Zenlayer:
		operator.clean = act.cleanZenlayer
		operator.insert = act.convertZenlayer
	default:
		return nil, fmt.Errorf("unsupported obs adjustment vendor %s", opt.Vendor)
	}
	err := operator.clean(kt.Kit(), opt)
	if err != nil {
		logs.Errorf("fail to clean adjustment, err: %v, opt: %+v, rid: %s", err, opt, kt.Kit().Rid)
		return nil, err
	}

	// 1. 查询调账明细
	page := &core.BasePage{Start: 0, Limit: core.DefaultMaxPageLimit, Sort: "id"}
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("vendor", opt.Vendor),
			tools.RuleEqual("bill_year", opt.BillYear),
			tools.RuleEqual("bill_month", opt.BillMonth),
			tools.RuleEqual("state", enumor.BillAdjustmentStateConfirmed),
		),
		Page: page,
	}
	for {
		adjustmentResp, err := actcli.GetDataService().Global.Bill.ListBillAdjustmentItem(kt.Kit(), listReq)
		if err != nil {
			logs.Errorf("fail to list bill adjustment sync obs, err: %v, req: %v, rid: %s", err, listReq, kt.Kit().Rid)
			return "fail to list bill adjustment sync obs", err
		}
		if len(adjustmentResp.Details) == 0 {
			return nil, nil
		}

		if err := operator.insert(kt.Kit(), adjustmentResp.Details); err != nil {
			logs.Errorf("fail to convert adjustment to obs bill, err: %v, vendor: %s, start: %d, rid: %s",
				err, opt.Vendor, page.Start, kt.Kit().Rid)
			return nil, err
		}
		logs.Infof("[%s] create obs adjustment bill for successfully, time: %d-%d, offset: %d, count: %d, rid: %s",
			opt.Vendor, opt.BillYear, opt.BillMonth, page.Start, len(adjustmentResp.Details), kt.Kit().Rid)

		if uint(len(adjustmentResp.Details)) < core.DefaultMaxPageLimit {
			break
		}
		page.Start += uint32(core.DefaultMaxPageLimit)
	}

	logs.Infof("sync obs adjustment for %s %d-%d done, rid: %s", opt.Vendor, opt.BillYear, opt.BillMonth, kt.Kit().Rid)
	return nil, nil
}

// convertHuawei ...
func (act SyncAdjustmentAction) convertHuawei(kt *kit.Kit, adjItems []*bill.AdjustmentItem) error {

	obsItems := make([]*tableobs.OBSBillItemHuawei, len(adjItems))
	mainAccountMap := make(map[string]*dataproto.MainAccountGetResult[protocore.HuaWeiMainAccountExtension])

	for i := range adjItems {
		adj := adjItems[i]
		exchangeRate, err := act.getCNYExchangeRate(kt, adj.Currency, adj.BillYear, adj.BillMonth)
		if err != nil {
			// 仅标记调用
			logs.Errorf("fail to get exchange rate for huawei adj, rid: %s", kt.Rid)
			return err
		}

		if _, ok := mainAccountMap[adj.MainAccountID]; !ok {
			info, err := actcli.GetDataService().HuaWei.MainAccount.Get(kt, adj.MainAccountID)
			if err != nil {
				logs.Errorf("fail to get huawei main account %s, err: %v, rid: %s", adj.MainAccountID, err, kt.Rid)
				return err
			}
			mainAccountMap[adj.MainAccountID] = info
		}
		mainAccount := mainAccountMap[adj.MainAccountID]
		yearM := adj.BillYear*100 + adj.BillMonth
		floatRate, _ := exchangeRate.Float64()

		// -- convert --
		// OBS 要求数据，决定汇率
		var accountType = "HW国际区"
		if mainAccount.Site == enumor.MainAccountChinaSite {
			accountType = "国内账单"
		}
		fetchTime := time.Now()
		obsItem := &tableobs.OBSBillItemHuawei{
			SetIndex:      adjustmentSetIndex,
			Vendor:        string(adj.Vendor),
			MainAccountID: adj.MainAccountID,
			BillYear:      int64(adj.BillYear),
			BillMonth:     int64(adj.BillMonth),

			CloudServiceType: adjustmentProductCode,
			AccountName:      mainAccount.Extension.CloudMainAccountName,
			AccountType:      accountType,
			ProductId:        int32(mainAccount.OpProductID),
			YearMonth:        int32(yearM),
			FetchTime:        fetchTime.Format(constant.DateTimeLayout),
			TotalCount:       1,
			Rate:             floatRate,
			RealCost:         &types.Decimal{Decimal: adj.Cost.Mul(cvt.PtrToVal(exchangeRate))},

			ProductName:  adj.Memo,
			ResourceName: adj.Memo,
		}
		obsItems[i] = obsItem
	}
	// insert to obs data base
	_, err := actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		if _, err := actcli.GetObsDaoSet().OBSBillItemHuawei().CreateWithTx(kt, txn, obsItems); err != nil {
			logs.Errorf("create huawei obs adjustment item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("create obs bill txn failed, err %s", err.Error())
	}
	return nil
}

func (act SyncAdjustmentAction) convertAws(kt *kit.Kit, adjItems []*bill.AdjustmentItem) error {
	obsItems := make([]*tableobs.OBSBillItemAws, len(adjItems))
	mainAccountMap := make(map[string]*dataproto.MainAccountGetResult[protocore.AwsMainAccountExtension])
	for i := range adjItems {
		adj := adjItems[i]
		exchangeRate, err := act.getCNYExchangeRate(kt, adj.Currency, adj.BillYear, adj.BillMonth)
		if err != nil {
			// 仅标记调用
			logs.Errorf("fail to get exchange rate for aws adj, rid: %s", kt.Rid)
			return err
		}

		if _, ok := mainAccountMap[adj.MainAccountID]; !ok {
			info, err := actcli.GetDataService().Aws.MainAccount.Get(kt, adj.MainAccountID)
			if err != nil {
				logs.Errorf("fail to get aws main account %s, err: %v, rid: %s", adj.MainAccountID, err, kt.Rid)
				return err
			}
			mainAccountMap[adj.MainAccountID] = info
		}
		rootInfo, err := act.getRootAccount(kt, adj.RootAccountID)
		if err != nil {
			logs.Errorf("fail to get root account id for aws adjustment, err: %v, rid: %s", err, kt.Rid)
			return err
		}

		mainAccount := mainAccountMap[adj.MainAccountID]
		yearM := adj.BillYear*100 + adj.BillMonth
		floatRate, _ := exchangeRate.Float64()

		// -- convert --
		// OBS 要求数据格式 1 国内 2 国际
		var regionCode = int32(2)
		if mainAccount.Site == enumor.MainAccountChinaSite {
			regionCode = 1
		}

		adjCost, err := adj.GetCost()
		if err != nil {
			return fmt.Errorf("get adjustment item cost failed, err: %+v", err)
		}

		obsItem := &tableobs.OBSBillItemAws{
			SetIndex:      adjustmentSetIndex,
			Vendor:        string(adj.Vendor),
			MainAccountID: adj.MainAccountID,
			BillYear:      int64(adj.BillYear),
			BillMonth:     int64(adj.BillMonth),
			YearMonth:     int32(yearM),
			Rate:          floatRate,
			// OBS要求，aws账单独立配置，Cost保存原始数据即可
			Cost:              &types.Decimal{Decimal: adjCost},
			ProductID:         int32(mainAccount.OpProductID),
			LinkedAccountName: mainAccount.Extension.CloudMainAccountName,
			Region:            regionCode,
			// OBS要求，存入标识
			Memo: "ieg上报",
			// OBS 要求，OBS外币金额写入line_item_unblended_cost字段中
			LineItemUnblendedCost:       adjCost.String(),
			LineItemNetUnblendedCost:    adjCost.String(),
			LineItemProductCode:         adjustmentProductCode,
			ProductProductName:          adjustmentProductName,
			LineItemUsageAccountID:      mainAccount.CloudID,
			LineItemCurrencyCode:        string(adj.Currency),
			BillPayerAccountID:          rootInfo.CloudID,
			LineItemLineItemDescription: adj.Memo,
		}
		obsItems[i] = obsItem
	}
	// insert to obs data base
	_, err := actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		if _, err := actcli.GetObsDaoSet().OBSBillItemAws().CreateWithTx(kt, txn, obsItems); err != nil {
			logs.Errorf("create aws obs adjustment item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("create obs bill txn failed, err %s", err.Error())
	}
	return nil
}

// convertHuawei ...
func (act SyncAdjustmentAction) convertGcp(kt *kit.Kit, adjItems []*bill.AdjustmentItem) error {
	obsItems := make([]*tableobs.OBSBillItemGcp, len(adjItems))
	mainAccountMap := make(map[string]*dataproto.MainAccountGetResult[protocore.GcpMainAccountExtension])
	for i := range adjItems {
		adj := adjItems[i]
		exchangeRate, err := act.getCNYExchangeRate(kt, adj.Currency, adj.BillYear, adj.BillMonth)
		if err != nil {
			// 仅标记调用
			logs.Errorf("fail to get exchange rate for gcp adj, rid: %s", kt.Rid)
			return err
		}

		if _, ok := mainAccountMap[adj.MainAccountID]; !ok {
			info, err := actcli.GetDataService().Gcp.MainAccount.Get(kt, adj.MainAccountID)
			if err != nil {
				logs.Errorf("fail to get gcp main account %s, err: %v, rid: %s", adj.MainAccountID, err, kt.Rid)
				return err
			}
			mainAccountMap[adj.MainAccountID] = info
		}
		mainAccount := mainAccountMap[adj.MainAccountID]
		yearM := adj.BillYear*100 + adj.BillMonth
		floatRate, _ := exchangeRate.Float64()
		fetchTime := time.Now()

		// -- convert --

		obsItem := &tableobs.OBSBillItemGcp{
			SetIndex:      adjustmentSetIndex,
			Vendor:        string(adj.Vendor),
			MainAccountID: adj.MainAccountID,
			BillYear:      int64(adj.BillYear),
			BillMonth:     int64(adj.BillMonth),
			YearMonth:     int32(yearM),
			Rate:          floatRate,

			Cost:                   adj.Cost.InexactFloat64(),
			ProductId:              int32(mainAccount.OpProductID),
			Currency:               string(adj.Currency),
			CurrencyConversionRate: floatRate,
			UsageAmount:            1,
			UsageUnit:              "",
			FetchTime:              fetchTime.Format(constant.DateTimeLayout),
			RealCost:               adj.Cost.Mul(cvt.PtrToVal(exchangeRate)).InexactFloat64(),

			ServiceId:          adjustmentProductCode,
			ServiceDescription: adjustmentProductName,
			ProjectId:          mainAccount.CloudID,

			SkuDescription: adj.Memo,
		}
		obsItems[i] = obsItem
	}
	// insert to obs data base
	_, err := actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		if _, err := actcli.GetObsDaoSet().OBSBillItemGcp().CreateWithTx(kt, txn, obsItems); err != nil {
			logs.Errorf("create gcp obs adjustment item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("create obs bill txn failed, err %s", err.Error())
	}
	return nil
}

func (act SyncAdjustmentAction) convertZenlayer(kt *kit.Kit, adjItems []*bill.AdjustmentItem) error {
	obsItems := make([]*tableobs.OBSBillItemZenlayer, len(adjItems))
	mainAccountMap := make(map[string]*dataproto.MainAccountGetResult[protocore.ZenlayerMainAccountExtension])
	for i := range adjItems {
		adj := adjItems[i]
		exchangeRate, err := act.getCNYExchangeRate(kt, adj.Currency, adj.BillYear, adj.BillMonth)
		if err != nil {
			// 仅标记调用
			logs.Errorf("fail to get exchange rate for zenlayer adj, rid: %s", kt.Rid)
			return err
		}

		if _, ok := mainAccountMap[adj.MainAccountID]; !ok {
			info, err := actcli.GetDataService().Zenlayer.MainAccount.Get(kt, adj.MainAccountID)
			if err != nil {
				logs.Errorf("fail to get zenlayer main account %s, err: %v, rid: %s", adj.MainAccountID, err, kt.Rid)
				return err
			}
			mainAccountMap[adj.MainAccountID] = info
		}
		mainAccount := mainAccountMap[adj.MainAccountID]
		yearM := adj.BillYear*100 + adj.BillMonth
		floatRate, _ := exchangeRate.Float64()

		// -- convert --

		obsItem := &tableobs.OBSBillItemZenlayer{
			SetIndex:      adjustmentSetIndex,
			Vendor:        string(adj.Vendor),
			MainAccountID: adj.MainAccountID,
			BillYear:      int64(adj.BillYear),
			BillMonth:     int64(adj.BillMonth),
			YearMonth:     int32(yearM),
			Rate:          floatRate,
			ProductID:     int32(mainAccount.OpProductID),

			Cost:       &types.Decimal{Decimal: adj.Cost},
			Currency:   string(adj.Currency),
			Type:       adjustmentProductCode,
			PayContent: adjustmentProductName,
			RealCost:   &types.Decimal{Decimal: adj.Cost.Mul(cvt.PtrToVal(exchangeRate))},
		}
		obsItems[i] = obsItem
	}
	// insert to obs data base
	_, err := actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (any, error) {
		if _, err := actcli.GetObsDaoSet().OBSBillItemZenlayer().CreateWithTx(kt, txn, obsItems); err != nil {
			logs.Errorf("create zenlayer obs adjustment item failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return fmt.Errorf("create obs bill txn failed, err %s", err.Error())
	}
	return nil
}

func (act SyncAdjustmentAction) cleanHuawei(kt *kit.Kit, adjOpt *AdjustmentOption) error {

	deleteFilter := tools.ExpressionAnd(tools.RuleEqual("set_index", adjustmentSetIndex))
	countOpt := &typesdao.ListOption{
		Filter: deleteFilter,
		Page:   core.NewCountPage(),
	}
	for {
		countResp, err := actcli.GetObsDaoSet().OBSBillItemHuawei().List(kt, countOpt)
		if err != nil {
			logs.Errorf("%s fail to count obs adjustment, err: %s, rid: %s",
				adjOpt.Vendor, err, kt.Rid)
			return err
		}
		if cvt.PtrToVal(countResp.Count) == 0 {
			break
		}
		_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {

			if err := actcli.GetObsDaoSet().OBSBillItemHuawei().DeleteWithTx(
				kt, txn, deleteFilter, uint64(core.DefaultMaxPageLimit)); err != nil {
				logs.Errorf("%s delete obs bill adjustment by filter %v failed, err: %s, setIndex:%s, rid: %s",
					adjOpt.Vendor, deleteFilter, err.Error(), kt.Rid)
				return nil, err
			}

			return nil, nil
		})
		if err != nil {
			return fmt.Errorf("delete obs bill txn failed: %w", err)
		}
		logs.Infof("%s delete previous obs data for adjustment successfully, count: %d, rid: %s",
			adjOpt.Vendor, countResp.Count, kt.Rid)
	}
	return nil
}
func (act SyncAdjustmentAction) cleanAws(kt *kit.Kit, adjOpt *AdjustmentOption) error {

	deleteFilter := tools.ExpressionAnd(tools.RuleEqual("set_index", adjustmentSetIndex))
	countOpt := &typesdao.ListOption{
		Filter: deleteFilter,
		Page:   core.NewCountPage(),
	}
	for {
		countResp, err := actcli.GetObsDaoSet().OBSBillItemAws().List(kt, countOpt)
		if err != nil {
			logs.Errorf("%s fail to count obs adjustment, err: %s, rid: %s",
				adjOpt.Vendor, err, kt.Rid)
			return err
		}
		if countResp.Count == 0 {
			break
		}
		_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {

			if err := actcli.GetObsDaoSet().OBSBillItemAws().DeleteWithTx(
				kt, txn, deleteFilter, uint64(core.DefaultMaxPageLimit)); err != nil {
				logs.Errorf("%s delete obs bill adjustment by filter %v failed, err: %s, setIndex:%s, rid: %s",
					adjOpt.Vendor, deleteFilter, err.Error(), kt.Rid)
				return nil, err
			}

			return nil, nil
		})
		if err != nil {
			return fmt.Errorf("delete obs bill txn failed: %w", err)
		}
		logs.Infof("%s delete previous obs data for adjustment successfully, count: %d, rid: %s",
			adjOpt.Vendor, countResp.Count, kt.Rid)
	}
	return nil
}
func (act SyncAdjustmentAction) cleanGcp(kt *kit.Kit, adjOpt *AdjustmentOption) error {

	deleteFilter := tools.ExpressionAnd(tools.RuleEqual("set_index", adjustmentSetIndex))
	countOpt := &typesdao.ListOption{
		Filter: deleteFilter,
		Page:   core.NewCountPage(),
	}
	for {
		countResp, err := actcli.GetObsDaoSet().OBSBillItemGcp().List(kt, countOpt)
		if err != nil {
			logs.Errorf("%s fail to count obs adjustment, err: %s, rid: %s",
				adjOpt.Vendor, err, kt.Rid)
			return err
		}
		if countResp.Count == 0 {
			break
		}
		_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {

			if err := actcli.GetObsDaoSet().OBSBillItemGcp().DeleteWithTx(
				kt, txn, deleteFilter, uint64(core.DefaultMaxPageLimit)); err != nil {
				logs.Errorf("%s delete obs bill adjustment by filter %v failed, err: %s, setIndex:%s, rid: %s",
					adjOpt.Vendor, deleteFilter, err.Error(), kt.Rid)
				return nil, err
			}

			return nil, nil
		})
		if err != nil {
			return fmt.Errorf("delete obs bill txn failed: %w", err)
		}
		logs.Infof("%s delete previous obs data for adjustment successfully, count: %d, rid: %s",
			adjOpt.Vendor, countResp.Count, kt.Rid)
	}
	return nil
}

func (act SyncAdjustmentAction) cleanZenlayer(kt *kit.Kit, adjOpt *AdjustmentOption) error {

	deleteFilter := tools.ExpressionAnd(tools.RuleEqual("set_index", adjustmentSetIndex))
	countOpt := &typesdao.ListOption{
		Filter: deleteFilter,
		Page:   core.NewCountPage(),
	}
	for {
		countResp, err := actcli.GetObsDaoSet().OBSBillItemZenlayer().List(kt, countOpt)
		if err != nil {
			logs.Errorf("%s fail to count obs adjustment, err: %s, rid: %s",
				adjOpt.Vendor, err, kt.Rid)
			return err
		}
		if countResp.Count == 0 {
			break
		}
		_, err = actcli.GetObsDaoSet().Txn().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {

			if err := actcli.GetObsDaoSet().OBSBillItemZenlayer().DeleteWithTx(
				kt, txn, deleteFilter, uint64(core.DefaultMaxPageLimit)); err != nil {
				logs.Errorf("%s delete obs bill adjustment by filter %v failed, err: %s, setIndex:%s, rid: %s",
					adjOpt.Vendor, deleteFilter, err.Error(), kt.Rid)
				return nil, err
			}

			return nil, nil
		})
		if err != nil {
			return fmt.Errorf("delete obs bill txn failed: %w", err)
		}
		logs.Infof("%s delete previous obs data for adjustment successfully, count: %d, rid: %s",
			adjOpt.Vendor, countResp.Count, kt.Rid)
	}
	return nil
}

func (act SyncAdjustmentAction) getRootAccount(kt *kit.Kit, rootAccountID string) (
	*dataproto.RootAccountGetBaseResult, error) {

	if act.rootAccountMap == nil {
		act.rootAccountMap = make(map[string]*dataproto.RootAccountGetBaseResult)
	}
	if root, ok := act.rootAccountMap[rootAccountID]; ok {
		return root, nil
	}

	rootInfo, err := actcli.GetDataService().Global.RootAccount.GetBasicInfo(kt, rootAccountID)
	if err != nil {
		logs.Errorf("fail to get root account info of adjustment, root account: %s, err:%v, rid: %s",
			rootAccountID, err, kt.Rid)
		return nil, err
	}
	act.rootAccountMap[rootAccountID] = rootInfo
	return rootInfo, nil
}

func (act SyncAdjustmentAction) getCNYExchangeRate(kt *kit.Kit, fromCurrency enumor.CurrencyCode,
	billYear, billMonth int) (*decimal.Decimal, error) {

	if act.exchangeRateMap == nil {
		act.exchangeRateMap = make(map[enumor.CurrencyCode]*decimal.Decimal)
	}
	if rate, ok := act.exchangeRateMap[fromCurrency]; ok {
		return rate, nil
	}

	if fromCurrency == enumor.CurrencyCNY {
		act.exchangeRateMap[fromCurrency] = cvt.ValToPtr(decimal.NewFromInt(1))
		return act.exchangeRateMap[fromCurrency], nil
	}
	rate, err := getExchangeRate(kt, fromCurrency, enumor.CurrencyCNY, billYear, billMonth)
	if err != nil {
		logs.Errorf("fail to get CNY exchange rate, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	act.exchangeRateMap[fromCurrency] = rate
	return rate, err
}
