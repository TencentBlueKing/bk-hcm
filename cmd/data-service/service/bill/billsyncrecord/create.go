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

package billsyncrecord

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
	cvt "hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// CreateBillSyncRecord account bill sync record with options
func (svc *service) CreateBillSyncRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(dsbill.BatchBillSyncRecordCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	idList, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		var itemList []tablebill.AccountBillSyncRecord
		for _, item := range req.Items {
			item := tablebill.AccountBillSyncRecord{
				Vendor:    item.Vendor,
				BillYear:  item.BillYear,
				BillMonth: item.BillMonth,
				State:     item.State,
				Currency:  item.Currency,
				Count:     cvt.ValToPtr(item.Count),
				Cost:      &types.Decimal{Decimal: item.Cost},
				RMBCost:   &types.Decimal{Decimal: item.RMBCost},
				Detail:    cvt.ValToPtr(types.JsonField("[]")),
				Operator:  cts.Kit.User,
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			}
			itemList = append(itemList, item)
		}

		ids, err := svc.dao.AccountBillSyncRecord().CreateWithTx(
			cts.Kit, txn, itemList)
		if err != nil {
			return nil, fmt.Errorf("create account bill sync record failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}
	retList, ok := idList.([]string)
	if !ok {
		return nil, fmt.Errorf("create account bill sync record but return ids type not []string, ids type: %v",
			reflect.TypeOf(idList).String())
	}

	return &core.BatchCreateResult{IDs: retList}, nil
}
