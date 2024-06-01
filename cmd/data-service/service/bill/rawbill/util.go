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
	"encoding/csv"
	"fmt"

	dsbill "hcm/pkg/api/data-service/bill"
)

func generateFilePath(req *dsbill.RawBillCreateReq) string {
	return fmt.Sprintf("rawbills/%s/%s/%s/%s/%s/%s/%s/%s",
		req.Vendor, req.FirstAccountID, req.AccountID, req.BillYear,
		req.BillMonth, req.Version, req.BillDate, req.FileName)
}

// 生成csv内容
func generateCSV(items []dsbill.RawBillItem) (*bytes.Buffer, error) {
	// 创建CSV文件的缓冲区
	var buffer bytes.Buffer
	csvWriter := csv.NewWriter(&buffer)
	for _, item := range items {
		record := []string{
			item.Region,
			item.HcProductCode,
			item.HcProductName,
			item.BillCurrency,
			item.BillCost.String(),
			item.ResAmount.String(),
			item.ResAmountUnit,
			string(item.Extension),
		}
		err := csvWriter.Write(record)
		if err != nil {
			return nil, fmt.Errorf("write record to csv failed, err %s", err.Error())
		}
	}
	// 刷新缓冲区，确保所有记录都写入
	csvWriter.Flush()
	return &buffer, nil
}
