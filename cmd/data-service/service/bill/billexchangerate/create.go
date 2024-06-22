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

package billexchangerate

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// CreateBillExchangeRate account bill exchange rate with options
func (svc *service) CreateBillExchangeRate(cts *rest.Contexts) (any, error) {
	req := new(dsbill.BatchCreateBillExchangeRateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	idList, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		var rateList []tablebill.AccountBillExchangeRate
		for _, rate := range req.ExchangeRates {
			dbRate := tablebill.AccountBillExchangeRate{
				Year:         rate.Year,
				Month:        rate.Month,
				FromCurrency: rate.FromCurrency,
				ToCurrency:   rate.ToCurrency,
				ExchangeRate: &types.Decimal{Decimal: *rate.ExchangeRate},
				Creator:      cts.Kit.User,
				Reviser:      cts.Kit.User,
			}
			rateList = append(rateList, dbRate)
		}

		ids, err := svc.dao.AccountBillExchangeRate().CreateWithTx(cts.Kit, txn, rateList)
		if err != nil {
			logs.Errorf("fail to create bill exchange rate, err: %v, rid: %s", err, cts.Kit.Rid)
			return nil, fmt.Errorf("create account bill exchange rate list failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}
	retList, ok := idList.([]string)
	if !ok {
		return nil, fmt.Errorf("create account bill exchange rate but return ids type not []string, ids type: %v",
			reflect.TypeOf(idList).String())
	}

	return &core.BatchCreateResult{IDs: retList}, nil
}
