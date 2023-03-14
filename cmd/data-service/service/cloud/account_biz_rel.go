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

package cloud

import (
	"fmt"

	"hcm/pkg/api/core"
	corecloud "hcm/pkg/api/core/cloud"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/logs"
	"hcm/pkg/rest"

	"github.com/jmoiron/sqlx"
)

// UpdateAccountBizRel update account biz rel.
func (a *accountSvc) UpdateAccountBizRel(cts *rest.Contexts) (interface{}, error) {
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

// ListWithAccount ...
func (a *accountSvc) ListWithAccount(cts *rest.Contexts) (interface{}, error) {
	req := new(protocloud.AccountBizRelWithAccountListReq)
	if err := cts.DecodeInto(req); err != nil {
		return nil, errf.NewFromErr(errf.DecodeRequestFailed, err)
	}

	if err := req.Validate(); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	details, err := a.dao.AccountBizRel().ListJoinAccount(cts.Kit, req.BkBizIDs)
	if err != nil {
		logs.Errorf("list account biz rels join account failed, err: %v, cvmIDs: %v, rid: %s", err,
			req.BkBizIDs, cts.Kit.Rid)
		return nil, err
	}

	accounts := make([]*protocloud.AccountBizRelWithAccount, 0, len(details.Details))
	for _, one := range details.Details {
		// 过滤账号类型
		if req.AccountType != "" && req.AccountType != one.Type {
			continue
		}

		accounts = append(accounts, &protocloud.AccountBizRelWithAccount{
			BaseAccount: corecloud.BaseAccount{
				ID:         one.ID,
				Vendor:     enumor.Vendor(one.Vendor),
				Name:       one.Name,
				Managers:   one.Managers,
				Type:       enumor.AccountType(one.Type),
				Site:       enumor.AccountSiteType(one.Site),
				SyncStatus: enumor.AccountSyncStatus(one.SyncStatus),
				Price:      one.Price,
				PriceUnit:  one.PriceUnit,
				Memo:       one.Memo,
				Revision: core.Revision{
					Creator:   one.Creator,
					Reviser:   one.Reviser,
					CreatedAt: one.CreatedAt,
					UpdatedAt: one.UpdatedAt,
				},
			},
			BkBizID:      one.BkBizID,
			RelCreator:   one.RelCreator,
			RelCreatedAt: one.RelCreatedAt,
		})
	}

	return accounts, nil
}
