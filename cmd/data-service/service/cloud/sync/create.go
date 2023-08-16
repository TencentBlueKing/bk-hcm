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

package sync

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dssync "hcm/pkg/api/data-service/cloud/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablesync "hcm/pkg/dal/table/cloud/sync"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchCreateAccountSD create account sync detail.
func (svc *service) BatchCreateAccountSD(cts *rest.Contexts) (interface{}, error) {
	req := new(dssync.CreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	asdIds, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablesync.AccountSyncDetailTable, 0, len(req.Items))
		for _, item := range req.Items {
			models = append(models, tablesync.AccountSyncDetailTable{
				Vendor:          item.Vendor,
				AccountID:       item.AccountID,
				ResName:         item.ResName,
				ResStatus:       item.ResStatus,
				ResEndTime:      item.ResEndTime,
				ResFailedReason: item.ResFailedReason,
				Creator:         cts.Kit.User,
				Reviser:         cts.Kit.User,
			})
		}
		ids, err := svc.dao.AccountSyncDetail().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create account sync detail failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create account sync detail commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := asdIds.([]string)
	if !ok {
		return nil, fmt.Errorf("create account sync detail but return id type not string, id type: %v",
			reflect.TypeOf(asdIds).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
