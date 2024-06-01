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

// AccountBillDailyPullTask interface for operating account bill puller
type AccountBillDailyPullTask interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, pullers []*tablebill.AccountBillDailyPullTask) ([]string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*typesbill.ListAccountBillDailyPullTaskDetails, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, eipID string, updateData *tablebill.AccountBillDailyPullTask) error
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
}

var _ AccountBillDailyPullTask = new(AccountBillDailyPullTaskDao)

// AccountBillDailyPullTaskDao dao for AccountBillDailyPullTask
type AccountBillDailyPullTaskDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// BatchCreateWithTx batch create bill daily pull task
func (abpDao AccountBillDailyPullTaskDao) BatchCreateWithTx(
	kt *kit.Kit, tx *sqlx.Tx, abPullers []*tablebill.AccountBillDailyPullTask) ([]string, error) {
	if len(abPullers) == 0 {
		return nil, errf.New(errf.InvalidParameter, "account bill daily pull task model data is required")
	}
	for _, i := range abPullers {
		if err := i.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`,
		table.AccountBillDailyPullTaskTable, tablebill.AccountBillDailyPullTaskColumns.ColumnExpr(),
		tablebill.AccountBillDailyPullTaskColumns.ColonNameExpr(),
	)

	ids, err := abpDao.IDGen.Batch(kt, table.AccountBillDailyPullTaskTable, len(abPullers))
	if err != nil {
		return nil, err
	}

	for idx, d := range abPullers {
		d.ID = ids[idx]
	}

	err = abpDao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, abPullers)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.AccountBillDailyPullTaskTable, err)
	}
	return ids, nil
}

// List list account bill daily pull task
func (abpDao AccountBillDailyPullTaskDao) List(kt *kit.Kit, opt *types.ListOption) (
	*typesbill.ListAccountBillDailyPullTaskDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account bill daily pull task options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillDailyPullTaskColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountBillDailyPullTaskTable, whereExpr)
		count, err := abpDao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account bill daily pull task failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesbill.ListAccountBillDailyPullTaskDetails{Count: &count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablebill.AccountBillDailyPullTaskColumns.FieldsNamedExpr(opt.Fields),
		table.AccountBillDailyPullTaskTable, whereExpr, pageExpr)

	details := make([]tablebill.AccountBillDailyPullTask, 0)
	if err = abpDao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &typesbill.ListAccountBillDailyPullTaskDetails{Details: details}, nil
}

// UpdateByIDWithTx update account bill daily pull task
func (abpDao AccountBillDailyPullTaskDao) UpdateByIDWithTx(
	kt *kit.Kit, tx *sqlx.Tx, pullerID string, updateData *tablebill.AccountBillDailyPullTask) error {

	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, table.AccountBillDailyPullTaskTable, setExpr)

	toUpdate["id"] = pullerID
	_, err = abpDao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update account bill daily pull task failed, err: %v, id: %s, rid: %v", err, pullerID, kt.Rid)
		return err
	}

	return nil
}

// DeleteWithTx delete account bill daily pull task
func (abpDao AccountBillDailyPullTaskDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountBillDailyPullTaskTable, whereExpr)

	if _, err = abpDao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete account bill daily pull task failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
