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

// AccountBillSummaryDaily only used for interface.
type AccountBillSummaryDaily interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []tablebill.AccountBillSummaryDaily) ([]string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*typesbill.ListAccountBillSummaryDailyDetails, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, billID string, updateData *tablebill.AccountBillSummaryDaily) error
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
}

// AccountBillSummaryDailyDao account bill summary daliy dao
type AccountBillSummaryDailyDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create account bill summary daliy with tx.
func (a AccountBillSummaryDailyDao) CreateWithTx(
	kt *kit.Kit, tx *sqlx.Tx, models []tablebill.AccountBillSummaryDaily) (
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
		tablebill.AccountBillSummaryDailyColumns.ColumnExpr(), tablebill.AccountBillSummaryDailyColumns.ColonNameExpr())

	if err = a.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// List get account bill summary daliy list.
func (a AccountBillSummaryDailyDao) List(kt *kit.Kit, opt *types.ListOption) (
	*typesbill.ListAccountBillSummaryDailyDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account bill summary daily options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillSummaryColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountBillSummaryDailyTable, whereExpr)
		count, err := a.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account bill summary daily failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesbill.ListAccountBillSummaryDailyDetails{Count: &count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablebill.AccountBillSummaryDailyColumns.FieldsNamedExpr(opt.Fields),
		table.AccountBillSummaryDailyTable, whereExpr, pageExpr)

	details := make([]tablebill.AccountBillSummaryDaily, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &typesbill.ListAccountBillSummaryDailyDetails{Details: details}, nil
}

// Update update account bill summary daliy.
func (a AccountBillSummaryDailyDao) UpdateByIDWithTx(
	kt *kit.Kit, tx *sqlx.Tx, id string, updateData *tablebill.AccountBillSummaryDaily) error {

	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, table.AccountBillSummaryDailyTable, setExpr)

	toUpdate["id"] = id
	_, err = a.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update account bill summary daily failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// DeleteWithTx delete account bill summary daliy with tx.
func (a AccountBillSummaryDailyDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountBillSummaryDailyTable, whereExpr)

	if _, err = a.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete account bill summary daily failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
