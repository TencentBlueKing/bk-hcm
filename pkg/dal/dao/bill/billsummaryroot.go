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

// Package bill ...
package bill

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typesbill "hcm/pkg/dal/dao/types/bill"
	"hcm/pkg/dal/table"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AccountBillSummaryRoot only used for interface.
type AccountBillSummaryRoot interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, datas []*tablebill.AccountBillSummaryRoot) ([]string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*typesbill.ListAccountBillSummaryRootDetails, error)
	ListWithTx(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*typesbill.ListAccountBillSummaryRootDetails, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, billID string, updateData *tablebill.AccountBillSummaryRoot) error
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
}

// AccountBillSummaryRootDao account bill summary of root account dao
type AccountBillSummaryRootDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create account bill summary of root account with tx.
func (a AccountBillSummaryRootDao) CreateWithTx(
	kt *kit.Kit, tx *sqlx.Tx, models []*tablebill.AccountBillSummaryRoot) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := a.IDGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		models[index].ID = ids[index]

		if err = model.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		tablebill.AccountBillSummaryRootColumns.ColumnExpr(), tablebill.AccountBillSummaryRootColumns.ColonNameExpr())

	if err = a.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// ListWithTx list summary root with tx.
func (dao AccountBillSummaryRootDao) ListWithTx(kt *kit.Kit, tx *sqlx.Tx,
	opt *types.ListOption) (*typesbill.ListAccountBillSummaryRootDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account bill summary of root account options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillSummaryRootColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountBillSummaryRootTable, whereExpr)
		count, err := dao.Orm.Txn(tx).Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account bill summary of root account failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesbill.ListAccountBillSummaryRootDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablebill.AccountBillSummaryRootColumns.FieldsNamedExpr(opt.Fields),
		table.AccountBillSummaryRootTable, whereExpr, pageExpr)

	details := make([]tablebill.AccountBillSummaryRoot, 0)
	if err = dao.Orm.Txn(tx).Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &typesbill.ListAccountBillSummaryRootDetails{Details: details}, nil
}

// List get account bill summary of root account list.
func (a AccountBillSummaryRootDao) List(kt *kit.Kit, opt *types.ListOption) (
	*typesbill.ListAccountBillSummaryRootDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account bill summary of root account options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillSummaryRootColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountBillSummaryRootTable, whereExpr)
		count, err := a.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account bill summary of root account failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesbill.ListAccountBillSummaryRootDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablebill.AccountBillSummaryRootColumns.FieldsNamedExpr(opt.Fields),
		table.AccountBillSummaryRootTable, whereExpr, pageExpr)

	details := make([]tablebill.AccountBillSummaryRoot, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &typesbill.ListAccountBillSummaryRootDetails{Details: details}, nil
}

// UpdateByIDWithTx update account bill summary of root account.
func (a AccountBillSummaryRootDao) UpdateByIDWithTx(
	kt *kit.Kit, tx *sqlx.Tx, id string, updateData *tablebill.AccountBillSummaryRoot) error {

	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, table.AccountBillSummaryRootTable, setExpr)

	toUpdate["id"] = id
	_, err = a.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update account bill summary of root account failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// DeleteWithTx delete account bill summary of root account with tx.
func (a AccountBillSummaryRootDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountBillSummaryRootTable, whereExpr)

	if _, err = a.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete account bill summary of root account failed, err: %v, filter: %s, rid: %s",
			err, expr, kt.Rid)
		return err
	}

	return nil
}
