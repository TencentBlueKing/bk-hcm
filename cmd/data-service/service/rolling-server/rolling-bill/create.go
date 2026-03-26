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

// Package rollingbill ...
package rollingbill

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

// BatchCreateRollingBill batch create rolling bill
func (svc *service) BatchCreateRollingBill(cts *rest.Contexts) (interface{}, error) {
	req := new(rsproto.BatchCreateRollingBillReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}
	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	detailIDs, err := svc.obsDao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		details := make([]rstable.OBSBillItemRolling, 0, len(req.Bills))
		for _, createReq := range req.Bills {
			details = append(details, rstable.OBSBillItemRolling{
				BkBizID:              createReq.BkBizID,
				DeliveredCore:        createReq.DeliveredCore,
				ReturnedCore:         createReq.ReturnedCore,
				NotReturnedCore:      createReq.NotReturnedCore,
				ExemptedReturnedCore: createReq.ExemptedReturnedCore,
				Year:                 createReq.Year,
				Month:                createReq.Month,
				Day:                  createReq.Day,
				RollDate:             times.GetDataIntDate(createReq.Year, createReq.Month, createReq.Day),
				Creator:              cts.Kit.User,
				// 下面字段为obs表所需的字段
				DataDate:            createReq.DataDate,
				ProductID:           createReq.ProductID,
				BusinessSetID:       createReq.BusinessSetID,
				BusinessSetName:     createReq.BusinessSetName,
				BusinessID:          createReq.BusinessID,
				BusinessName:        createReq.BusinessName,
				BusinessModID:       createReq.BusinessModID,
				BusinessModName:     createReq.BusinessModName,
				Uin:                 createReq.Uin,
				AppID:               createReq.AppID,
				User:                createReq.User,
				CityID:              createReq.CityID,
				CampusID:            createReq.CampusID,
				IdcUnitID:           createReq.IdcUnitID,
				IdcUnitName:         createReq.IdcUnitName,
				ModuleID:            createReq.ModuleID,
				ModuleName:          createReq.ModuleName,
				ZoneID:              createReq.ZoneID,
				ZoneName:            createReq.ZoneName,
				PlatformID:          createReq.PlatformID,
				ResClassID:          createReq.ResClassID,
				ClusterID:           createReq.ClusterID,
				PlatformResID:       createReq.PlatformResID,
				BandwidthTypeID:     createReq.BandwidthTypeID,
				OperatorNameID:      createReq.OperatorNameID,
				Amount:              createReq.Amount,
				AmountInCurrentDate: createReq.AmountInCurrentDate,
				Cost:                createReq.Cost,
				ExtendDetail:        createReq.ExtendDetail,
			})
		}
		ids, err := svc.obsDao.OBSBillItemRolling().CreateWithTx(cts.Kit, txn, details)
		if err != nil {
			return nil, fmt.Errorf("create rolling bill failed, err: %v", err)
		}
		return ids, nil
	})
	if err != nil {
		return nil, err
	}

	ids, ok := detailIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("batch create rolling bill but return id type is not string, id type: %v",
			reflect.TypeOf(detailIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
