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

package subaccount

import (
	"fmt"
	"reflect"

	"hcm/pkg/api/core"
	dssubaccount "hcm/pkg/api/data-service/cloud/sub-account"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	tablesubaccount "hcm/pkg/dal/table/cloud/sub-account"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// BatchCreateSubAccount create sub account.
func (svc *service) BatchCreateSubAccount(cts *rest.Contexts) (interface{}, error) {

	req := new(dssubaccount.CreateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	accountIDs, err := svc.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		models := make([]tablesubaccount.Table, 0, len(req.Items))
		for _, item := range req.Items {
			models = append(models, tablesubaccount.Table{
				CloudID:   item.CloudID,
				Name:      item.Name,
				Vendor:    item.Vendor,
				Site:      item.Site,
				AccountID: item.AccountID,
				Extension: tabletype.JsonField(item.Extension),
				Managers:  item.Managers,
				BkBizIDs:  item.BkBizIDs,
				Memo:      item.Memo,
				Creator:   cts.Kit.User,
				Reviser:   cts.Kit.User,
			})
		}
		ids, err := svc.dao.SubAccount().BatchCreateWithTx(cts.Kit, txn, models)
		if err != nil {
			return nil, fmt.Errorf("batch create sub account failed, err: %v", err)
		}

		return ids, nil
	})
	if err != nil {
		logs.Errorf("batch create sub account commit txn failed, err: %v, rid: %s", err, cts.Kit.Rid)
		return nil, err
	}

	ids, ok := accountIDs.([]string)
	if !ok {
		return nil, fmt.Errorf("create sub account but return id type not string, id type: %v",
			reflect.TypeOf(accountIDs).String())
	}

	return &core.BatchCreateResult{IDs: ids}, nil
}
