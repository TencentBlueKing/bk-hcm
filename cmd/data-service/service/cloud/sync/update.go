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
	dssync "hcm/pkg/api/data-service/cloud/sync"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablesync "hcm/pkg/dal/table/cloud/sync"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchUpdateAccountSD update account sync detail.
func (svc *service) BatchUpdateAccountSD(cts *rest.Contexts) (interface{}, error) {
	req := new(dssync.UpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, item := range req.Items {
			model := &tablesync.AccountSyncDetailTable{
				ResStatus:       item.ResStatus,
				ResEndTime:      item.ResEndTime,
				ResFailedReason: item.ResFailedReason,
				Reviser:         cts.Kit.User,
			}

			if err := svc.dao.AccountSyncDetail().UpdateByIDWithTx(cts.Kit, txn, item.ID, model); err != nil {
				logs.Errorf("update account sync detail by id: %s failed, err: %v, model: %+v, rid: %s", item.ID, err,
					model, cts.Kit.Rid)
				return nil, err
			}
		}
		return nil, nil
	})
	if err != nil {
		logs.Errorf("batch update account sync detail commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	return nil, nil
}
