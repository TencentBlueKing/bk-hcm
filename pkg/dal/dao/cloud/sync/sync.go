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

package daosync

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typessync "hcm/pkg/dal/dao/types/sync"
	"hcm/pkg/dal/table"
	tablessync "hcm/pkg/dal/table/cloud/sync"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AccountSyncDetail only used account sync detail.
type AccountSyncDetail interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablessync.AccountSyncDetailTable) ([]string, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablessync.AccountSyncDetailTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typessync.ListAccountSyncDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ AccountSyncDetail = new(AccountSyncDetailDao)

// AccountSyncDetailDao account sync detail dao.
type AccountSyncDetailDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// UpdateByIDWithTx account sync detail.
func (dao *AccountSyncDetailDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	model *tablessync.AccountSyncDetailTable) error {

	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, model.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update account sync detail failed, err: %v, id: %s, sql: %s, rid: %v", err, id,
			sql, kt.Rid)
		return err
	}

	return nil
}

// BatchCreateWithTx account sync detail with tx.
func (dao *AccountSyncDetailDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []tablessync.AccountSyncDetailTable) ([]string, error) {

	ids, err := dao.IDGen.Batch(kt, table.AccountSyncDetailTable, len(models))
	if err != nil {
		return nil, err
	}
	for index := range models {
		models[index].ID = ids[index]

		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AccountSyncDetailTable,
		tablessync.AccountSyncDetailColumns.ColumnExpr(), tablessync.AccountSyncDetailColumns.ColonNameExpr())

	err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s", table.AccountSyncDetailTable, err, sql, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.AccountSyncDetailTable, err)
	}

	return ids, nil
}

// List account sync detail.
func (dao *AccountSyncDetailDao) List(kt *kit.Kit, opt *types.ListOption) (*typessync.ListAccountSyncDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account sync detail options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablessync.AccountSyncDetailColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is dao count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountSyncDetailTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account sync detail failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &typessync.ListAccountSyncDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablessync.AccountSyncDetailColumns.FieldsNamedExpr(opt.Fields),
		table.AccountSyncDetailTable, whereExpr, pageExpr)

	details := make([]tablessync.AccountSyncDetailTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select account sync detail failed, err: %v, sql: %s, filter: %v, rid: %s", err, sql,
			opt.Filter, kt.Rid)
		return nil, err
	}

	return &typessync.ListAccountSyncDetails{Count: 0, Details: details}, nil
}

// DeleteWithTx account sync detail with tx.
func (dao *AccountSyncDetailDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountSyncDetailTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete account sync detail failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
