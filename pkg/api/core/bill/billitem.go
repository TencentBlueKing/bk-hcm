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

package bill

import (
	rawjson "encoding/json"
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/bssintl/v2/model"
	"github.com/shopspring/decimal"
)

// BaseBillItem 存储分账后的明细
type BaseBillItem struct {
	ID             string          `json:"id,omitempty"`
	RootAccountID  string          `json:"root_account_id"`
	MainAccountID  string          `json:"main_account_id"`
	Vendor         enumor.Vendor   `json:"vendor" validate:"required"`
	ProductID      int64           `json:"product_id" validate:"omitempty"`
	BkBizID        int64           `json:"bk_biz_id" validate:"omitempty"`
	BillYear       int             `json:"bill_year" validate:"required"`
	BillMonth      int             `json:"bill_month" validate:"required"`
	BillDay        int             `json:"bill_day" validate:"required"`
	VersionID      int             `json:"version_id" validate:"required"`
	Currency       string          `json:"currency" validate:"required"`
	Cost           decimal.Decimal `json:"cost" validate:"required"`
	HcProductCode  string          `json:"hc_product_code,omitempty"`
	HcProductName  string          `json:"hc_product_name,omitempty"`
	ResAmount      decimal.Decimal `json:"res_amount,omitempty"`
	ResAmountUnit  string          `json:"res_amount_unit,omitempty"`
	*core.Revision `json:",inline"`
}

// BillItem ...
type BillItem[E BillItemExtension] struct {
	*BaseBillItem `json:",inline"`
	Extension     *E `json:"extension,omitempty"`
}

// BillItemRaw ...
type BillItemRaw struct {
	*BaseBillItem `json:",inline"`
	Extension     rawjson.RawMessage `json:"extension,omitempty"`
}

// TCloudBillItem ...
type TCloudBillItem = BillItem[TCloudBillItemExtension]

// HuaweiBillItem ...
type HuaweiBillItem = BillItem[HuaweiBillItemExtension]

// AzureBillItem ...
type AzureBillItem = BillItem[AzureBillItemExtension]

// AwsBillItem ...
type AwsBillItem = BillItem[AwsBillItemExtension]

// GcpBillItem ...
type GcpBillItem = BillItem[GcpBillItemExtension]

// KaopuBillItem ...
type KaopuBillItem = BillItem[KaopuBillItemExtension]

// ZenlayerBillItem ...
type ZenlayerBillItem = BillItem[ZenlayerBillItemExtension]

// BillItemExtension 账单详情
type BillItemExtension interface {
	TCloudBillItemExtension |
		HuaweiBillItemExtension |
		AwsBillItemExtension |
		AzureBillItemExtension |
		GcpBillItemExtension |
		KaopuBillItemExtension |
		ZenlayerBillItemExtension |
		rawjson.RawMessage
}

// TCloudBillItemExtension ...
type TCloudBillItemExtension struct {
}

// AwsBillItemExtension ...
type AwsBillItemExtension struct {
}

// HuaweiBillItemExtension ...
type HuaweiBillItemExtension struct {
	*model.ResFeeRecordV2 `json:",inline"`
}

// GcpBillItemExtension ...
type GcpBillItemExtension struct {
}

// AzureBillItemExtension ...
type AzureBillItemExtension struct {
}

// KaopuBillItemExtension ...
type KaopuBillItemExtension struct {
}

// ZenlayerBillItemExtension ...
type ZenlayerBillItemExtension struct {
}
