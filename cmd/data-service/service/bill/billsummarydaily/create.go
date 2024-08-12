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

package billsummarydaily

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// CreateBillSummaryDaily create bill summary daily with options
func (svc *service) CreateBillSummaryDaily(cts *rest.Contexts) (interface{}, error) {
	req := new(dsbill.BillSummaryDailyCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	id, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		summary := &tablebill.AccountBillSummaryDaily{
			RootAccountID:      req.RootAccountID,
			MainAccountID:      req.MainAccountID,
			RootAccountCloudID: req.RootAccountCloudID,
			MainAccountCloudID: req.MainAccountCloudID,
			Vendor:             req.Vendor,
			ProductID:          req.ProductID,
			BkBizID:            req.BkBizID,
			BillYear:           req.BillYear,
			BillMonth:          req.BillMonth,
			BillDay:            req.BillDay,
			VersionID:          req.VersionID,
			Currency:           req.Currency,
			Cost:               &types.Decimal{Decimal: req.Cost},
			Count:              req.Count,
			Creator:            cts.Kit.User,
			Reviser:            cts.Kit.User,
		}
		ids, err := svc.dao.AccountBillSummaryDaily().CreateWithTx(
			cts.Kit, txn, []*tablebill.AccountBillSummaryDaily{
				summary,
			})
		if err != nil {
			return nil, fmt.Errorf("create account bill summary daily failed, err: %v", err)
		}
		if len(ids) != 1 {
			return nil, fmt.Errorf("create account bill summary daily expect 1 puller ID: %v", ids)
		}
		return ids[0], nil
	})
	if err != nil {
		return nil, err
	}
	idStr, ok := id.(string)
	if !ok {
		return nil, fmt.Errorf("create account bill summary daily but return id type not string, id type: %v",
			reflect.TypeOf(id).String())
	}

	return &core.CreateResult{ID: idStr}, nil
}
