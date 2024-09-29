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
	corebill "hcm/pkg/api/core/bill"
	"hcm/pkg/rest"

	"github.com/shopspring/decimal"
)

// -------------------------- List --------------------------

// GcpBillListResult define gcp bill list result.
type GcpBillListResult struct {
	Count   int64       `json:"count"`
	Details interface{} `json:"details"`
}

// GcpBillListResp define gcp bill list resp.
type GcpBillListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *GcpBillListResult `json:"data"`
}

// GcpCreditListResult ...
type GcpCreditListResult = core.ListResultT[GcpCreditUsage]

// GcpCreditUsage ...
type GcpCreditUsage struct {
	ProjectId string               `json:"project_id"`
	Credits   []corebill.GcpCredit `json:"credits"`

	PromotionCredit        *decimal.Decimal `json:"promotion_credit"`
	BillingAccountId       string           `json:"billing_account_id"`
	Currency               string           `json:"currency"`
	CurrencyConversionRate *decimal.Decimal `json:"currency_conversion_rate"`
	Month                  string           `json:"month"`
	ProjectName            string           `json:"project_name"`
	ProjectNumber          string           `json:"project_number"`
}
