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

	reader := getReader(req.ExcelFileBase64)
	records := make([]billcore.ZenlayerRawBillItem, 0)
	err := excelRowsIterator(kt, reader, 0, constant.BatchOperationMaxLimit,
		func(rows [][]string, err error) error {
			if len(rows) == 0 {
				return nil
			}
			for _, row := range rows {
				item, err := convertStringToZenlayerRawBillItem(row)
				if err != nil {
					return err
				}
				records = append(records, item)
			}
			return nil
		})
	if err != nil {
		logs.Errorf("fail parse excel file, err: %v, rid: %s", err, kt.Rid)

		return nil, errf.New(errf.BillItemImportDataError, "fail parse excel file")
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
		return nil, err
	}

	// convert to BillItemCreateReq
	createReqs, err := convertZenlayerToRawBillCreateReq(kt, req.BillYear, req.BillMonth,
		records, cloudIDToSummaryMainMap)
	if err != nil {
		return nil, err
	}

	rate, err := b.getExchangedRate(kt, req.BillYear, req.BillMonth)
	if err != nil {
		return nil, err
	}
	costMap, err := doCalculate(createReqs, rate)
	if err != nil {
		return nil, err
	}

	return bill.ImportBillItemPreviewResult{
		Items:   createReqs,
		CostMap: costMap,
	}, nil
}

func (b *billItemSvc) getExchangedRate(kt *kit.Kit, billYear, billMonth int) (*decimal.Decimal, error) {
	// 获取汇率
	result, err := b.client.DataService().Global.Bill.ListExchangeRate(kt, &core.ListReq{
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
	})
	if err != nil {
		return nil, fmt.Errorf("get exchange rate from %s to %s in %d-%d failed, err %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, billYear, billMonth, err.Error())
	}
	if len(result.Details) == 0 {
		return nil, fmt.Errorf("get no exchange rate from %s to %s in %d-%d, rid %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, billYear, billMonth, kt.Rid)
	}
	if result.Details[0].ExchangeRate == nil {
		return nil, fmt.Errorf("get exchange rate is nil, from %s to %s in %d-%d, rid %s",
			enumor.CurrencyUSD, enumor.CurrencyRMB, billYear, billMonth, kt.Rid)
	}
	return result.Details[0].ExchangeRate, nil
}

func doCalculate(records []dsbill.BillItemCreateReq[json.RawMessage], rate *decimal.Decimal) (
	map[enumor.CurrencyCode]*billcore.CostWithCurrency, error) {

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
	return retMap, nil
}

func (b *billItemSvc) listSummaryMainByBusinessGroups(kt *kit.Kit, vendor enumor.Vendor, businessGroupIDs []string,
	billYear, billMonth int) (map[string]*dsbill.BillSummaryMainResult, error) {

	businessGroupIDs = slice.Unique(businessGroupIDs)
	idToCloudIDMap := make(map[string]string, len(businessGroupIDs))
	accountIDs := make([]string, 0, len(businessGroupIDs))
	for _, ids := range slice.Split(businessGroupIDs, int(core.DefaultMaxPageLimit)) {
		list, err := b.client.DataService().Global.MainAccount.List(kt, &core.ListReq{
			Filter: tools.ExpressionAnd(
				tools.RuleEqual("vendor", vendor),
				tools.RuleIn("cloud_id", ids),
			),
			Page: core.NewDefaultBasePage(),
		})
		if err != nil {
			return nil, err
		}
		for _, detail := range list.Details {
			idToCloudIDMap[detail.ID] = detail.CloudID
			accountIDs = append(accountIDs, detail.ID)
		}
	}

	result := make(map[string]*dsbill.BillSummaryMainResult, len(accountIDs))
	summaryMains, err := b.listSummaryMainByMainAccountIDs(kt, vendor, accountIDs, billYear, billMonth)
	if err != nil {
		return nil, err
	}
	for _, summaryMain := range summaryMains {
		result[idToCloudIDMap[summaryMain.MainAccountID]] = summaryMain
	}

	return result, nil
}

func convertZenlayerToRawBillCreateReq(kt *kit.Kit, billYear, billMonth int,
	recordList []billcore.ZenlayerRawBillItem, summaryMap map[string]*dsbill.BillSummaryMainResult) (
	[]dsbill.BillItemCreateReq[json.RawMessage], error) {

	result := make([]dsbill.BillItemCreateReq[json.RawMessage], 0, len(recordList))
	for _, record := range recordList {
		mainAccountCloudID := *record.BusinessGroup

		tmp := dsbill.BillItemCreateReq[json.RawMessage]{}

		tmp.Vendor = enumor.Zenlayer
		summaryMain, ok := summaryMap[mainAccountCloudID]
		if !ok {
			logs.Errorf("fail to find summary main by cloud id(%s), rid: %s", mainAccountCloudID, kt.Rid)
			return nil, fmt.Errorf("fail to find summary main by cloud id(%s)", mainAccountCloudID)
		}
		tmp.RootAccountID = summaryMain.RootAccountID
		tmp.MainAccountID = summaryMain.MainAccountID
		tmp.ProductID = summaryMain.ProductID
		tmp.BkBizID = summaryMain.BkBizID
		tmp.VersionID = summaryMain.CurrentVersion

		split := strings.Split(*record.BillID, "-")
		if len(split) != 3 {
			return nil, fmt.Errorf("invalid bill id: %s, expect format: yy-mm-dd", *record.BillID)
		}
		curDay, err := strconv.Atoi(split[2])
		if err != nil {
			return nil, err
		}
		curYear, err := strconv.Atoi((*record.BillingPeriod)[:4])
		if err != nil {
			return nil, err
		}
		curMonth, err := strconv.Atoi((*record.BillingPeriod)[4:])
		if err != nil {
			return nil, err
		}

		// validate bill year and month
		if curYear != billYear || curMonth != billMonth {

			return nil, errf.NewFromErr(errf.BillItemImportBillDateError,
				fmt.Errorf("invalid billID, expect: %d-%d, but got: %d-%d",
					billYear, billMonth, curYear, curMonth))
		}

		tmp.BillYear = billYear
		tmp.BillMonth = billMonth
		tmp.BillDay = curDay

		if record.TotalPayable != nil {
			tmp.Cost = *record.TotalPayable
		}
		if record.Currency != nil {
			tmp.Currency = enumor.CurrencyCode(*record.Currency)
		}

		data, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}
		tmp.Extension = (*json.RawMessage)(&data)

		result = append(result, tmp)
	}
	return result, nil
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
			return billcore.ZenlayerRawBillItem{}, fmt.Errorf("unsupported pointer field type: %v", fieldValue.Type().Elem().Kind())
		}
	}
	return item, nil
}

func getReader(str string) io.Reader {
	return base64.NewDecoder(base64.StdEncoding, bytes.NewReader([]byte(str)))
}
