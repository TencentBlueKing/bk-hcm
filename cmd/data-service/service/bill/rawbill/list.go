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

// Package rawbill ...
package rawbill

import (
	"fmt"

	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/rest"
)

// ListRawBill list cloud raw bill
func (s *service) ListRawBill(cts *rest.Contexts) (interface{}, error) {
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

	path := fmt.Sprintf("rawbills/%s/%s/%s/%s/%s/%s/%s/",
		vendor, rootAccountID, accoundID, billYear, billMonth, version, billDate)

	logs.Infof("get path %s\n", path)

	filenames, err := s.ostore.ListItems(cts.Request.Request.Context(), path)
	if err != nil {
		return nil, errf.NewFromErr(errf.Aborted, err)
	}

	return &dsbill.RawBillItemNameListResult{
		Filenames: filenames,
	}, nil
}
