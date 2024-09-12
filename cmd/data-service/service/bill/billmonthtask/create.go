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

package billmonthtask

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
	"github.com/shopspring/decimal"
)

// CreateBillMonthTask create bill puller with options
func (svc *service) CreateBillMonthTask(cts *rest.Contexts) (interface{}, error) {
	req := new(dsbill.BillMonthTaskCreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}
	taskID, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		puller := &tablebill.AccountBillMonthTask{
			RootAccountID:      req.RootAccountID,
			RootAccountCloudID: req.RootAccountCloudID,
			Type:               req.Type,
			Vendor:             req.Vendor,
			BillYear:           req.BillYear,
			BillMonth:          req.BillMonth,
			VersionID:          req.VersionID,
			State:              req.State,
			SummaryDetail:      types.JsonField("[]"),
			Creator:            cts.Kit.User,
			Reviser:            cts.Kit.User,
			Cost:               &types.Decimal{Decimal: decimal.Zero},
		}
		taskIDs, err := svc.dao.AccountBillMonthPullTask().BatchCreateWithTx(
			cts.Kit, txn, []*tablebill.AccountBillMonthTask{
				puller,
			})
		if err != nil {
			return nil, fmt.Errorf("create account bill month task failed, err: %v", err)
		}
		if len(taskIDs) != 1 {
			return nil, fmt.Errorf("create account bill month task expect 1 puller ID: %v", taskIDs)
		}
		return taskIDs[0], nil
	})
	if err != nil {
		return nil, err
	}
	id, ok := taskID.(string)
	if !ok {
		return nil, fmt.Errorf("create account bill month task but return id type not string, id type: %v",
			reflect.TypeOf(taskID).String())
	}

	return &core.CreateResult{ID: id}, nil
}
