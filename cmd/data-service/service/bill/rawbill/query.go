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

package rawbill

import (
	"bytes"
	"fmt"

	"encoding/csv"

	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

	"github.com/shopspring/decimal"
)

// QueryRawBillDetail query cloud raw bill details
func (s *service) QueryRawBillDetail(cts *rest.Contexts) (interface{}, error) {
	vendor := enumor.Vendor(cts.PathParameter("vendor").String())
	if err := vendor.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	rootAccountID := cts.PathParameter("root_account_id").String()
	accoundID := cts.PathParameter("account_id").String()
	billYear := cts.PathParameter("bill_year").String()
	billMonth := cts.PathParameter("bill_month").String()
	version := cts.PathParameter("version").String()
	billDate := cts.PathParameter("bill_date").String()
	name := cts.PathParameter("bill_name").String()

	path := fmt.Sprintf("rawbills/%s/%s/%s/%s/%s/%s/%s/%s",
		vendor, rootAccountID, accoundID, billYear, billMonth, version, billDate, name)

	// 创建CSV文件的缓冲区
	var buffer bytes.Buffer
	if err := s.ostore.Download(cts.Kit, path, &buffer); err != nil {
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	csvReader := csv.NewReader(&buffer)
	csvLines, err := csvReader.ReadAll()
	if err != nil {
		return nil, errf.NewFromErr(errf.Aborted, err)
	}
	var itemList []*dsbill.RawBillItem
	for _, csvlineArr := range csvLines {
		if len(csvlineArr) != 8 {
			return nil, errf.NewFromErr(errf.Aborted, fmt.Errorf("bill csv line invalid length, %v", csvlineArr))
		}
		cost, err := decimal.NewFromString(csvlineArr[4])
		if err != nil {
			return nil, errf.NewFromErr(errf.Aborted, fmt.Errorf("bill csv line invalid cost, %v", csvlineArr))
		}
		resAmount, err := decimal.NewFromString(csvlineArr[5])
		if err != nil {
			return nil, errf.NewFromErr(errf.Aborted, fmt.Errorf("bill csv line invalid resAmount, %v", csvlineArr))
		}
		item := &dsbill.RawBillItem{
			Region:        csvlineArr[0],
			HcProductCode: csvlineArr[1],
			HcProductName: csvlineArr[2],
			BillCurrency:  enumor.CurrencyCode(csvlineArr[3]),
			BillCost:      cost,
			ResAmount:     resAmount,
			ResAmountUnit: csvlineArr[6],
			Extension:     types.JsonField(csvlineArr[7]),
		}
		itemList = append(itemList, item)
	}

	count := uint64(len(itemList))
	return &dsbill.RawBillItemQueryResult{
		Count:   &count,
		Details: itemList,
	}, nil
}
