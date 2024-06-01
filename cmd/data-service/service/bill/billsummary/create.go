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

package billsummary

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dsbill "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchCreateBillSummary create account bill summary with options
func (svc *service) BatchCreateBillSummary(cts *rest.Contexts) (interface{}, error) {
	req := new(dsbill.BillSummaryCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	id, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		summary := tablebill.AccountBillSummary{
			FirstAccountID:  string(req.FirstAccountID),
			SecondAccountID: string(req.SecondAccountID),
			Vendor:          req.Vendor,
			ProductID:       req.ProductID,
			BkBizID:         req.BkBizID,
			BillYear:        req.BillYear,
			BillMonth:       req.BillMonth,
			CurrentVersion:  req.CurrentVersion,
		}
		ids, err := svc.dao.AccountBillSummary().CreateWithTx(cts.Kit, txn, []tablebill.AccountBillSummary{
			summary,
		})
		if err != nil {
			return nil, fmt.Errorf("create account bill summary failed, err: %v", err)
		}
		if len(ids) != 1 {
			return nil, fmt.Errorf("create account bill summary expect 1 puller ID: %v", ids)
		}
		return ids[0], nil
	})
	if err != nil {
		return nil, err
	}
	idStr, ok := id.(string)
	if !ok {
		return nil, fmt.Errorf("create account bill summary but return id type not string, id type: %v",
			reflect.TypeOf(id).String())
	}

	return &core.CreateResult{ID: idStr}, nil
}
