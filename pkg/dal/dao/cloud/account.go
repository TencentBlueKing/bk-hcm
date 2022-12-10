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

	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Account only used for account.
type Account interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, account *corecloud.Account) (uint64, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *corecloud.Account) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListAccountDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Account = new(AccountDao)

// AccountDao account dao.
type AccountDao struct {
	Orm orm.Interface
}

// CreateWithTx account with tx.
func (a AccountDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *corecloud.Account) (uint64, error) {
	account, err := cloud.ConvAccountTable(model)
	if err != nil {
		return 0, err
	}

	if err := account.InsertValidate(); err != nil {
		return 0, err
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, account.TableName(), cloud.AccountColumns.ColumnExpr(),
		cloud.AccountColumns.ColonNameExpr())

	id, err := a.Orm.Txn(tx).Insert(kt.Ctx, sql, account)
	if err != nil {
		return 0, fmt.Errorf("insert %s failed, err: %v", account.TableName(), err)
	}

	return id, nil
}

// Update accounts.
func (a AccountDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *corecloud.Account) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	account, err := cloud.ConvAccountTable(model)
	if err != nil {
		return err
	}

	if err = account.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, err := filterExpr.SQLWhereExpr(types.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(account, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, account.TableName(), setExpr, whereExpr)

	_, err = a.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := a.Orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.ErrorJson("update account failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List accounts.
func (a AccountDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListAccountDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.AccountColumns.ColumnTypes())),
		types.DefaultPageOption); err != nil {
		return nil, err
	}

	whereExpr, err := opt.Filter.SQLWhereExpr(types.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountTable, whereExpr)

		count, err := a.Orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			logs.ErrorJson("count accounts failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListAccountDetails{Count: count}, nil
	}

	pageExpr, err := opt.Page.SQLExpr(types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.AccountColumns.NamedExpr(),
		table.AccountTable, whereExpr, pageExpr)

	list := make([]*cloud.AccountTable, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &list, sql); err != nil {
		return nil, err
	}

	details, err := cloud.ConvAccountList(list)
	if err != nil {
		logs.Errorf("conv account list failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.ListAccountDetails{Count: 0, Details: details}, nil
}

// DeleteWithTx account with tx.
func (a AccountDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, err := filterExpr.SQLWhereExpr(types.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountTable, whereExpr)
	if err = a.Orm.Txn(tx).Delete(kt.Ctx, sql); err != nil {
		logs.ErrorJson("delete account failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
