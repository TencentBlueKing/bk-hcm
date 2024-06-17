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

package bill

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table/types"

	"github.com/shopspring/decimal"
)

// AccountBillItem 存储分账后的明细
type AccountBillItem struct {
	ID            string          `json:"id"`
	RootAccountID string          `json:"root_account_id"`
	MainAccountID string          `json:"main_account_id"`
	Vendor        enumor.Vendor   `json:"vendor"`
	ProductID     int64           `json:"product_id"`
	BkBizID       int64           `json:"bk_biz_id"`
	BillYear      int             `json:"bill_year"`
	BillMonth     int             `json:"bill_month"`
	BillDay       int             `json:"bill_day"`
	VersionID     int             `json:"version_id"`
	Currency      string          `json:"currency"`
	Cost          decimal.Decimal `json:"cost"`
	HcProductCode string          ` json:"hc_product_code"`
	HcProductName string          ` json:"hc_product_name"`
	ResAmount     decimal.Decimal ` json:"res_amount,omitempty"`
	ResAmountUnit string          `json:"res_amount_unit,omitempty"`
	Extension     types.JsonField ` json:"extension"`
	core.Revision
}
