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

package billpuller

import (
	"fmt"

	dataservice "hcm/pkg/api/data-service/bill"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// UpdateBillPuller update bill puller with options
func (svc *service) UpdateBillPuller(cts *rest.Contexts) (interface{}, error) {
	req := new(dataservice.BillPullerUpdateReq)

	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	billPuller := &tablebill.AccountBillPuller{
		ID:                    req.ID,
		FirstAccountID:        req.FirstAccountID,
		SecondAccountID:       req.SecondAccountID,
		Vendor:                req.Vendor,
		ProductID:             req.ProductID,
		BkBizID:               req.BkBizID,
		PullMode:              req.PullMode,
		SyncPeriod:            req.SyncPeriod,
		BillDelay:             req.BillDelay,
		FinalBillCalendarDate: req.FinalBillCalendarDate,
	}
	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err := svc.dao.AccountBillPuller().UpdateByIDWithTx(cts.Kit, txn, billPuller.ID, billPuller); err != nil {
			return nil, fmt.Errorf("update bill puller failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
