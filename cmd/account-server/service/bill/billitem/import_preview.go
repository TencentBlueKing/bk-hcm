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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"

	"hcm/pkg/api/account-server/bill"
	"hcm/pkg/api/core"
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
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/shopspring/decimal"
)

// ImportBillItemsPreview 导入账单明细-预览
func (b *billItemSvc) ImportBillItemsPreview(cts *rest.Contexts) (any, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if len(vendor) == 0 {
		return nil, errf.New(errf.InvalidParameter, "vendor is required")
	}

	req := new(bill.ImportBillItemPreviewReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	err := b.authorizer.AuthorizeWithPerm(cts.Kit,
		meta.ResourceAttribute{Basic: &meta.Basic{Type: meta.AccountBill, Action: meta.Create}})
	if err != nil {
		logs.Errorf("import bill item auth failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	switch vendor {
	case enumor.Zenlayer:
		return b.importZenlayerBillItemsPreview(cts.Kit, req)
	default:
		return nil, fmt.Errorf("unsupport %s vendor", vendor)
	}
}

func (b *billItemSvc) importZenlayerBillItemsPreview(kt *kit.Kit, req *bill.ImportBillItemPreviewReq) (any, error) {

	records, err := parseExcelToRecords(kt, req.ExcelFileBase64, convertStringToZenlayerRawBillItem)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, errf.New(errf.BillItemImportEmptyDataError, "empty excel file")
	}

	businessGroupIDs := make([]string, 0, len(records))
	for _, record := range records {
		businessGroupIDs = append(businessGroupIDs, *record.BusinessGroup)
	}
	cloudIDToSummaryMainMap, err := b.listSummaryMainByBusinessGroups(kt, enumor.Zenlayer,
		businessGroupIDs, req.BillYear, req.BillMonth)
	if err != nil {
		logs.Errorf("list summary main by business groups failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// convert to BillItemCreateReq
	createReqs, err := convertZenlayerRawBillItemToRawBillCreateReq(kt, req.BillYear, req.BillMonth,
		records, cloudIDToSummaryMainMap)
	if err != nil {
		logs.Errorf("convert raw bill item to createReqs failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	rate, err := b.getExchangedRate(kt, req.BillYear, req.BillMonth)
	if err != nil {
		logs.Errorf("get exchange rate, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	costMap := doCalculate(createReqs, rate)
	return bill.ImportBillItemPreviewResult{
		Items:   createReqs,
		CostMap: costMap,
	}, nil
}

type convertStringToEntityFunc[T any] func([]string) (T, error)

func convertBase64StrToReader(str bill.Base64String) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(str)))
}

func parseExcelToRecords[T any](kt *kit.Kit, base64 bill.Base64String, convertFunc convertStringToEntityFunc[T]) ([]T,
	error) {
	reader := convertBase64StrToReader(base64)
	records := make([]T, 0)
	err := excelRowsIterator(kt, reader, 0, constant.BatchOperationMaxLimit,
		func(rows [][]string) error {
			if len(rows) == 0 {
				return nil
			}
			for _, row := range rows {
				item, err := convertFunc(row)
				if err != nil {
					return err
				}
				records = append(records, item)
			}
			return nil
		})
	if err != nil {
		logs.Errorf("fail to parse excel file for bill import perview, err: %v, rid: %s", err, kt.Rid)
		return nil, errf.New(errf.BillItemImportDataError, "fail parse excel file")
	}
	return records, nil
}

func (b *billItemSvc) getExchangedRate(kt *kit.Kit, billYear, billMonth int) (*decimal.Decimal, error) {
	// 获取汇率
	listReq := &core.ListReq{
		Filter: tools.ExpressionAnd(
			tools.RuleEqual("from_currency", enumor.CurrencyUSD),
			tools.RuleEqual("to_currency", enumor.CurrencyRMB),
			tools.RuleEqual("year", billYear),
			tools.RuleEqual("month", billMonth),
		),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	result, err := b.client.DataService().Global.Bill.ListExchangeRate(kt, listReq)
	if err != nil {
		return nil, fmt.Errorf("get exchange rate from %s to %s in %d-%d failed, err %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, billYear, billMonth, err.Error())
	}
	if len(result.Details) == 0 {
		return nil, fmt.Errorf("get no exchange rate from %s to %s in %d-%d",
			enumor.CurrencyUSD, enumor.CurrencyRMB, billYear, billMonth)
	}
	if result.Details[0].ExchangeRate == nil {
		return nil, fmt.Errorf("get exchange rate is nil, from %s to %s in %d-%d",
			enumor.CurrencyUSD, enumor.CurrencyRMB, billYear, billMonth)
	}
	return result.Details[0].ExchangeRate, nil
}

func doCalculate(records []dsbill.BillItemCreateReq[json.RawMessage],
	rate *decimal.Decimal) map[enumor.CurrencyCode]*billcore.CostWithCurrency {

	retMap := make(map[enumor.CurrencyCode]*billcore.CostWithCurrency)
	for _, record := range records {
		if _, ok := retMap[record.Currency]; !ok {
			retMap[record.Currency] = &billcore.CostWithCurrency{
				Cost:     decimal.NewFromFloat(0),
				RMBCost:  decimal.NewFromFloat(0),
				Currency: record.Currency,
			}
		}
		retMap[record.Currency].Cost = retMap[record.Currency].Cost.Add(record.Cost)
		retMap[record.Currency].RMBCost = retMap[record.Currency].RMBCost.Add(record.Cost.Mul(*rate))
	}
	return retMap
}

func (b *billItemSvc) listSummaryMainByBusinessGroups(kt *kit.Kit, vendor enumor.Vendor, businessGroupIDs []string,
	billYear, billMonth int) (map[string]*dsbill.BillSummaryMain, error) {

	businessGroupIDs = slice.Unique(businessGroupIDs)
	idToCloudIDMap := make(map[string]string, len(businessGroupIDs))
	accountIDs := make([]string, 0, len(businessGroupIDs))
	for _, ids := range slice.Split(businessGroupIDs, int(filter.DefaultMaxInLimit)) {
		listReq := &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleIn("cloud_id", ids),
			),
			Page: core.NewDefaultBasePage(),
		}
		list, err := b.client.DataService().Global.MainAccount.List(kt, listReq)
		if err != nil {
			logs.Errorf("list main account failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		for _, detail := range list.Details {
			idToCloudIDMap[detail.ID] = detail.CloudID
			accountIDs = append(accountIDs, detail.ID)
		}
	}

	result := make(map[string]*dsbill.BillSummaryMain, len(accountIDs))
	summaryMains, err := b.listSummaryMainByMainAccountIDs(kt, vendor, accountIDs, billYear, billMonth)
	if err != nil {
		logs.Errorf("list summary main by main account ids failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	for _, summaryMain := range summaryMains {
		result[idToCloudIDMap[summaryMain.MainAccountID]] = summaryMain
	}

	return result, nil
}

func convertZenlayerRawBillItemToRawBillCreateReq(kt *kit.Kit, billYear, billMonth int,
	recordList []billcore.ZenlayerRawBillItem, summaryMap map[string]*dsbill.BillSummaryMain) (
	[]dsbill.BillItemCreateReq[json.RawMessage], error) {

	result := make([]dsbill.BillItemCreateReq[json.RawMessage], 0, len(recordList))
	for _, record := range recordList {
		err := validateBillYearAndMonth(converter.PtrToVal[string](record.BillingPeriod), billYear, billMonth)
		if err != nil {
			logs.Errorf("validate year and month failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}

		mainAccountCloudID := converter.PtrToVal[string](record.BusinessGroup)
		summaryMain, ok := summaryMap[mainAccountCloudID]
		if !ok {
			logs.Errorf("fail to find summary main by cloud id(%s), rid: %s", mainAccountCloudID, kt.Rid)
			return nil, fmt.Errorf("fail to find summary main by cloud id(%s)", mainAccountCloudID)
		}
		billID := converter.PtrToVal[string](record.BillID)
		if summaryMain.State != enumor.MainAccountBillSummaryStateAccounting {
			logs.Errorf("summaryMainAccount(%s) state is not accounting, can't import bill, rid: %s",
				summaryMain.ID, kt.Rid)
			return nil, fmt.Errorf("summaryMainAccount(%s) state is not accounting, can't import bill",
				summaryMain.ID)
		}

		split := strings.Split(billID, "-")
		if len(split) != 3 {
			return nil, fmt.Errorf("invalid bill id: %s, expect format: yy-mm-dd", *record.BillID)
		}
		curDay, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, err
		}
		data, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}
		result = append(result, dsbill.BillItemCreateReq[json.RawMessage]{
			RootAccountID: summaryMain.RootAccountID,
			MainAccountID: summaryMain.MainAccountID,
			Vendor:        enumor.Zenlayer,
			ProductID:     summaryMain.ProductID,
			BkBizID:       summaryMain.BkBizID,
			BillYear:      billYear,
			BillMonth:     billMonth,
			BillDay:       curDay,
			VersionID:     summaryMain.CurrentVersion,
			Currency:      enumor.CurrencyCode(converter.PtrToVal[string](record.Currency)),
			Cost:          converter.PtrToVal[decimal.Decimal](record.TotalPayable),
			Extension:     (*json.RawMessage)(&data),
		})
	}
	return result, nil
}

func validateBillYearAndMonth(curDate string, billYear, billMonth int) error {
	curYear, err := strconv.Atoi(curDate[:4])
	if err != nil {
		return err
	}
	curMonth, err := strconv.Atoi(curDate[4:])
	if err != nil {
		return err
	}
	// validate bill year and month
	if curYear != billYear || curMonth != billMonth {
		err = fmt.Errorf("invalid billID, expect: %d-%d, but got: %d-%d",
			billYear, billMonth, curYear, curMonth)
		return errf.NewFromErr(errf.BillItemImportBillDateError, err)
	}
	return nil
}

func convertStringToZenlayerRawBillItem(row []string) (billcore.ZenlayerRawBillItem, error) {
	item := billcore.ZenlayerRawBillItem{}
	for i, value := range row {
		cur := value
		field := zenlayerBillItemRefType.Field(i)

		fieldValue := reflect.ValueOf(&item).Elem().FieldByName(field.Name)
		switch fieldValue.Type().Elem().Kind() {
		case reflect.Int:
			intValue, err := strconv.Atoi(cur)
			if err != nil {
				return billcore.ZenlayerRawBillItem{}, err
			}
			fieldValue.Set(reflect.ValueOf(&intValue))
		case reflect.String:
			fieldValue.Set(reflect.ValueOf(&cur))
		case reflect.TypeOf(decimal.Decimal{}).Kind():
			// 处理 *decimal.Decimal 类型
			decValue, err := decimal.NewFromString(strings.ReplaceAll(cur, ",", ""))
			if err != nil {
				return billcore.ZenlayerRawBillItem{}, err
			}
			fieldValue.Set(reflect.ValueOf(&decValue))
		default:
			return billcore.ZenlayerRawBillItem{}, fmt.Errorf("unsupported pointer field type: %v",
				fieldValue.Type().Elem().Kind())
		}
	}
	return item, nil
}
