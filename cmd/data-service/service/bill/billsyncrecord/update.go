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

	dataservice "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// UpdateBillSyncRecord account with options
func (svc *service) UpdateBillSyncRecord(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BillSyncRecordUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	billSyncRecord := &tablebill.AccountBillSyncRecord{
		ID:       req.ID,
		Operator: req.Operator,
		Currency: req.Currency,
		Count:    req.Count,
		State:    req.State,
		Detail:   req.Detail,
		Reviser:  cts.Kit.User,
	}
	if req.Cost != nil {
		billSyncRecord.Cost = &types.Decimal{Decimal: *req.Cost}
	}
	if req.RMBCost != nil {
		billSyncRecord.RMBCost = &types.Decimal{Decimal: *req.RMBCost}
	}
	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.AccountBillSyncRecord().UpdateByIDWithTx(
			cts.Kit, txn, billSyncRecord.ID, billSyncRecord); err != nil {
			return nil, fmt.Errorf("update bill sync record failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
