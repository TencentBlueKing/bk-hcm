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
	"reflect"

	"github.com/jmoiron/sqlx"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// Account supplies all the cloud account related operations.
type Account interface {
	// Create one account instance
	Create(kt *kit.Kit, account *tablecloud.AccountModel) (uint64, error)
	// Update one account's info
	Update(kt *kit.Kit, expr *filter.Expression, account *tablecloud.AccountModel) error
	// List accounts with options.
	List(kt *kit.Kit, opt *types.ListOption) ([]*tablecloud.AccountModel, error)
}

var _ Account = new(AccountDao)

type AccountDao struct {
	orm      orm.Interface
	auditDao audit.AuditDao
}

func NewAccountDao(orm orm.Interface, auditDao audit.AuditDao) *AccountDao {
	return &AccountDao{orm, auditDao}
}

// Create one account instance.
func (ad *AccountDao) Create(kt *kit.Kit, account *tablecloud.AccountModel) (uint64, error) {
	if account == nil {
		return 0, errf.New(errf.InvalidParameter, "cloud account is nil")
	}

	sql := account.GenerateInsertSQL()

	result, err := ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		id, err := ad.orm.Txn(txn).Insert(kt.Ctx, sql, account)
		if err != nil {
			return 0, fmt.Errorf("insert account failed, err: %v", err)
		}

		account.ID = id
		if err := ad.auditDao.Decorator(kt, enumor.Account).AuditCreate(txn, account); err != nil {
			return 0, fmt.Errorf("audit create account failed, err: %v", err)
		}
		return id, nil
	})
	if err != nil {
		logs.Errorf("create account, but do auto txn failed, err: %v, rid: %s", err, kt.Rid)
		return 0, fmt.Errorf("create account, but auto run txn failed, err: %v", err)
	}

	id, ok := result.(uint64)
	if !ok {
		logs.Errorf("insert account return id type not is uint64, id type: %v, rid: %s",
			reflect.TypeOf(result).String(), kt.Rid)
	}

	return id, nil
}

func (ad *AccountDao) Update(kt *kit.Kit, expr *filter.Expression, account *tablecloud.AccountModel) error {
	sql, err := account.GenerateUpdateSQL(expr)
	if err != nil {
		return err
	}

	whereExpr, _ := table.GenerateWhereExpr(expr)
	toUpdate := account.GenerateUpdateFieldKV()

	ab := ad.auditDao.Decorator(kt, enumor.Account).PrepareUpdate(whereExpr, toUpdate)

	_, err = ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := ad.orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update account: %d failed, err: %v, rid: %v", account.ID, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account, but record not found, filter: %v, rid: %v", expr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		if err := ab.Do(txn); err != nil {
			return nil, fmt.Errorf("do account update audit failed, err: %v", err)
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (ad *AccountDao) List(kt *kit.Kit, opt *types.ListOption) ([]*tablecloud.AccountModel, error) {
	account := new(tablecloud.AccountModel)
	listSQL, err := account.GenerateListSQL(opt)
	if err != nil {
		return nil, err
	}

	accounts := make([]*tablecloud.AccountModel, 0)
	err = ad.orm.Do().Select(kt.Ctx, &accounts, listSQL)
	if err != nil {
		return nil, err
	}

	return accounts, nil
}
