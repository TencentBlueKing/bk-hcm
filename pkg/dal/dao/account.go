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

package dao

import (
	"fmt"
	"reflect"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Account supplies all the account related operations.
type Account interface {
	// Create one account instance
	Create(kt *kit.Kit, account *table.Account) (uint64, error)
	// Update one account's info
	Update(kt *kit.Kit, filter *filter.Expression, account *table.Account) error
	// List accounts with options.
	List(kt *kit.Kit, opts *types.ListAccountsOption) (*types.ListAccountDetails, error)
	// Delete one account instance.
	Delete(kt *kit.Kit, filter *filter.Expression) error
	// BatchCreate account instances.
	// TODO: 这里仅展示批量创建如何使用，账号应该没有批量创建接口，之后删掉
	BatchCreate(kt *kit.Kit, accounts []table.Account) ([]uint64, error)
}

var _ Account = new(accountDao)

type accountDao struct {
	orm      orm.Interface
	auditDao audit.AuditDao
}

// Create one account instance.
func (ad *accountDao) Create(kt *kit.Kit, account *table.Account) (uint64, error) {
	if account == nil {
		return 0, errf.New(errf.InvalidParameter, "account is nil")
	}

	if err := account.ValidateCreate(); err != nil {
		return 0, errf.New(errf.InvalidParameter, err.Error())
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, account.TableName(), table.AccountColumns.ColumnExpr(),
		table.AccountColumns.ColonNameExpr())

	result, err := ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		id, err := ad.orm.Txn(txn).Insert(kt.Ctx, sql, account)
		if err != nil {
			return 0, fmt.Errorf("insert account failed, err: %v", err)
		}

		// audit this to be create account details.
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

// Update an account instance.
func (ad *accountDao) Update(kit *kit.Kit, expr *filter.Expression, updateField *table.Account) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	if updateField == nil {
		return errf.New(errf.InvalidParameter, "update field is nil")
	}

	if err := updateField.ValidateUpdate(); err != nil {
		return errf.New(errf.InvalidParameter, err.Error())
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
	}
	whereExpr, err := expr.SQLWhereExpr(sqlOpt)
	if err != nil {
		return err
	}

	opts := table.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields("id")
	setExpr, toUpdate, err := table.RearrangeSQLDataWithOption(updateField, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set expr failed, err: %v", err)
	}

	ab := ad.auditDao.Decorator(kit, enumor.Account).PrepareUpdate(whereExpr, toUpdate)
	sql := fmt.Sprintf(`UPDATE %s %s %s`, updateField.TableName(), setExpr, whereExpr)

	_, err = ad.orm.AutoTxn(kit, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := ad.orm.Txn(txn).Update(kit.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("update account: %d failed, err: %v, rid: %v", updateField.ID, err, kit.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account, but record not found, filter: %v, rid: %v", expr, kit.Rid)
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

// List account's detail info with the filter's expression.
func (ad *accountDao) List(kt *kit.Kit, opts *types.ListAccountsOption) (*types.ListAccountDetails, error) {
	if opts == nil {
		return nil, errf.New(errf.InvalidParameter, "list account options is nil")
	}

	if err := opts.Validate(types.DefaultPageOption); err != nil {
		return nil, err
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
	}
	whereExpr, err := opts.Filter.SQLWhereExpr(sqlOpt)
	if err != nil {
		return nil, err
	}

	var sql string
	if opts.Page.Count {
		// this is a count request, then do count operation only.
		sql = fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountTable, whereExpr)
		count, err := ad.orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			return nil, err
		}

		return &types.ListAccountDetails{Count: count, Details: make([]*table.Account, 0)}, nil
	}

	pageExpr, err := opts.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql = fmt.Sprintf(`SELECT %s FROM %s %s %s`, table.AccountColumns.NamedExpr(),
		table.AccountTable, whereExpr, pageExpr)

	list := make([]*table.Account, 0)
	err = ad.orm.Do().Select(kt.Ctx, &list, sql)
	if err != nil {
		return nil, err
	}

	return &types.ListAccountDetails{Count: 0, Details: list}, nil
}

// Delete an account instance.
func (ad *accountDao) Delete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter is nil")
	}

	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
	}
	whereExpr, err := expr.SQLWhereExpr(sqlOpt)
	if err != nil {
		return err
	}

	ab := ad.auditDao.Decorator(kt, enumor.Account).PrepareDelete(whereExpr)
	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountTable, whereExpr)
	_, err = ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, option *orm.TxnOption) (interface{}, error) {
		// delete the account at first.
		err := ad.orm.Txn(txn).Delete(kt.Ctx, sql)
		if err != nil {
			return nil, err
		}

		// audit delete account details.
		if err := ab.Do(txn); err != nil {
			return nil, fmt.Errorf("audit delete account failed, err: %v", err)
		}

		return nil, nil
	})
	if err != nil {
		logs.ErrorJson("delete account failed, filter: %v, err: %v, rid: %v", expr, err, kt.Rid)
		return fmt.Errorf("delete account, but run txn failed, err: %v", err)
	}

	return nil
}

// BatchCreate account instances.
// TODO: 这里仅展示批量创建如何使用，账号应该没有批量创建接口，之后删掉
func (ad *accountDao) BatchCreate(kt *kit.Kit, accounts []table.Account) ([]uint64, error) {

	if accounts == nil {
		return nil, errf.New(errf.InvalidParameter, "accounts is nil")
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AccountTable, table.AccountColumns.ColumnExpr(),
		table.AccountColumns.ColonNameExpr())

	result, err := ad.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		ids, err := ad.orm.Txn(txn).BulkInsert(kt.Ctx, sql, accounts)
		if err != nil {
			return 0, fmt.Errorf("insert accounts failed, err: %v", err)
		}

		// audit this to be create accounts details.
		for index := range accounts {
			accounts[index].ID = ids[index]
		}

		if err := ad.auditDao.Decorator(kt, enumor.Account).AuditBatchCreate(txn, accounts); err != nil {
			return 0, fmt.Errorf("audit create accounts failed, err: %v", err)
		}
		return ids, nil
	})

	if err != nil {
		logs.Errorf("create accounts, but do auto txn failed, err: %v, rid: %s", err, kt.Rid)
		return nil, fmt.Errorf("create accounts, but auto run txn failed, err: %v", err)
	}

	ids, ok := result.([]uint64)
	if !ok {
		logs.Errorf("insert accounts return id type not is uint64, id type: %v, rid: %s",
			reflect.TypeOf(result).String(), kt.Rid)
	}

	return ids, nil
}
