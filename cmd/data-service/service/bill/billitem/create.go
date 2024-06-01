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

// CreateBillItem create bill item with options
func (svc *service) CreateBillItem(cts *rest.Contexts) (interface{}, error) {
	req := make(dsbill.BatchBillItemCreateReq, 0)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	for _, item := range req {
		if err := item.Validate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	idList, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		var summaryList []tablebill.AccountBillItem
		for _, item := range req {
			summary := tablebill.AccountBillItem{
				FirstAccountID:  item.FirstAccountID,
				SecondAccountID: item.SecondAccountID,
				Vendor:          item.Vendor,
				ProductID:       item.ProductID,
				BkBizID:         item.BkBizID,
				BillYear:        item.BillYear,
				BillMonth:       item.BillMonth,
				BillDay:         item.BillDay,
				VersionID:       item.VersionID,
				Currency:        item.Currency,
				Cost:            item.Cost,
				RMBCost:         item.RMBCost,
			}
			summaryList = append(summaryList, summary)
		}

		ids, err := svc.dao.AccountBillItem().CreateWithTx(
			cts.Kit, txn, summaryList)
		if err != nil {
			return nil, fmt.Errorf("create account bill item list failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}
	retList, ok := idList.([]string)
	if !ok {
		return nil, fmt.Errorf("create account bill item but return ids type not []string, ids type: %v",
			reflect.TypeOf(idList).String())
	}

	return &core.BatchCreateResult{IDs: retList}, nil
}
