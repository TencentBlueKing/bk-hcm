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

package accountbizrel

import (
	"fmt"

	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// UpdateAccountBizRel update account biz rel.
func (a *service) UpdateAccountBizRel(cts *rest.Contexts) (interface{}, error) {
	accountID := cts.PathParameter("account_id").String()

	req := new(protocloud.AccountBizRelUpdateReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, err
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	_, err := a.dao.Txn().AutoTxn(cts.Kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ftr := tools.EqualExpression("account_id", accountID)
		if err := a.dao.AccountBizRel().DeleteWithTx(cts.Kit, txn, ftr); err != nil {
			return nil, fmt.Errorf("delete account_biz_rels failed, err: %v", err)
		}

		rels := make([]*tablecloud.AccountBizRelTable, len(req.BkBizIDs))
		for index, bizID := range req.BkBizIDs {
			rels[index] = &tablecloud.AccountBizRelTable{
				BkBizID:   bizID,
				AccountID: accountID,
				Creator:   cts.Kit.User,
			}
		}
		if err := a.dao.AccountBizRel().BatchCreateWithTx(cts.Kit, txn, rels); err != nil {
			return nil, fmt.Errorf("batch create account_biz_rels failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		return nil, err
	}

	return nil, err
}
