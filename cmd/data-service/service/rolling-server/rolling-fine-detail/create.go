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

// Package rollingfinedetail ...
package rollingfinedetail

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	rsproto "hcm/pkg/api/data-service/rolling-server"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	rstable "hcm/pkg/dal/table/rolling-server"
	"hcm/pkg/rest"
	"hcm/pkg/tools/times"

	"github.com/jmoiron/sqlx"
)

// BatchCreateRollingDineDetail batch create rolling fine detail
func (svc *service) BatchCreateRollingDineDetail(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.BatchCreateRollingFineDetailReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	detailIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		details := make([]rstable.RollingFineDetailTable, 0, len(req.FineDetails))
		for _, createReq := range req.FineDetails {
			details = append(details, rstable.RollingFineDetailTable{
				BkBizID:              createReq.BkBizID,
				AppliedRecordID:      createReq.AppliedRecordID,
				OrderID:              createReq.OrderID,
				SubOrderID:           createReq.SubOrderID,
				Year:                 createReq.Year,
				Month:                createReq.Month,
				Day:                  createReq.Day,
				RollDate:             times.GetDataIntDate(createReq.Year, createReq.Month, createReq.Day),
				DeliveredCore:        createReq.DeliveredCore,
				ReturnedCore:         createReq.ReturnedCore,
				ExemptedReturnedCore: createReq.ExemptedReturnedCore,
				Fine:                 createReq.Fine,
				Creator:              cts.Kit.User,
			})
		}
		ids, err := svc.dao.RollingFineDetail().CreateWithTx(cts.Kit, txn, details)
		if err != nil {
			return nil, fmt.Errorf("create rolling fine detail failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := detailIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create rolling fine detail but return id type is not string, id type: %v",
			reflect.TypeOf(detailIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
