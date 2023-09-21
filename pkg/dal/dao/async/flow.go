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

package daoasync

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typesasync "hcm/pkg/dal/dao/types/async"
	"hcm/pkg/dal/table"
	tableasync "hcm/pkg/dal/table/async"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AsyncFlow only used async flow.
type AsyncFlow interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableasync.AsyncFlowTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tableasync.AsyncFlowTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tableasync.AsyncFlowTable) error
	UpdateByIDCAS(kt *kit.Kit, id string, state enumor.FlowState) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesasync.ListAsyncFlows, error)
	ListWithTx(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*typesasync.ListAsyncFlows, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ AsyncFlow = new(AsyncFlowDao)

// AsyncFlowDao async flow dao.
type AsyncFlowDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// Update async flow.
func (dao *AsyncFlowDao) Update(kt *kit.Kit, expr *filter.Expression, model *tableasync.AsyncFlowTable) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update async flow failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			return nil, errf.New(errf.RecordNotUpdate, "record not update")
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx async flow.
func (dao *AsyncFlowDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	model *tableasync.AsyncFlowTable) error {

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
	effected, err := dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update async flow failed, err: %v, id: %s, sql: %s, rid: %v", err, id,
			sql, kt.Rid)
		return err
	}

	if effected == 0 {
		return errf.New(errf.RecordNotUpdate, "record not update")
	}

	return nil
}

// UpdateByIDCAS update async flow CAS.
func (dao *AsyncFlowDao) UpdateByIDCAS(kt *kit.Kit, id string, state enumor.FlowState) error {

	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	sql := fmt.Sprintf(`update %s set state = '%s' where id = :id and state = '%s'`,
		table.AsyncFlowTable, state, enumor.FlowPending)

	whereValue := map[string]interface{}{
		"id": id,
	}
	effected, err := dao.Orm.Do().Update(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.Errorf("update async flow failed, err: %v, id: %s, sql: %s, rid: %v", err, id,
			sql, kt.Rid)
		return err
	}

	if effected == 0 {
		return errf.New(errf.RecordNotUpdate, "record not update")
	}

	return nil
}

// BatchCreateWithTx async flow with tx.
func (dao *AsyncFlowDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []tableasync.AsyncFlowTable) ([]string, error) {

	ids, err := dao.IDGen.Batch(kt, table.AsyncFlowTable, len(models))
	if err != nil {
		return nil, err
	}
	for index := range models {
		models[index].ID = ids[index]

		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AsyncFlowTable,
		tableasync.AsyncFlowColumns.ColumnExpr(), tableasync.AsyncFlowColumns.ColonNameExpr())

	err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s", table.AsyncFlowTable, err, sql, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.AsyncFlowTable, err)
	}

	return ids, nil
}

// ListWithTx async flow with tx.
func (dao *AsyncFlowDao) ListWithTx(kt *kit.Kit, tx *sqlx.Tx,
	opt *types.ListOption) (*typesasync.ListAsyncFlows, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list async flow options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableasync.AsyncFlowColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is dao count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AsyncFlowTable, whereExpr)

		count, err := dao.Orm.Txn(tx).Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count async flow failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesasync.ListAsyncFlows{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableasync.AsyncFlowColumns.FieldsNamedExpr(opt.Fields),
		table.AsyncFlowTable, whereExpr, pageExpr)

	details := make([]tableasync.AsyncFlowTable, 0)
	if err = dao.Orm.Txn(tx).Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select async flow failed, err: %v, sql: %s, filter: %v, rid: %s", err, sql,
			opt.Filter, kt.Rid)
		return nil, err
	}

	return &typesasync.ListAsyncFlows{Count: 0, Details: details}, nil
}

// List async flow.
func (dao *AsyncFlowDao) List(kt *kit.Kit, opt *types.ListOption) (*typesasync.ListAsyncFlows, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list async flow options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableasync.AsyncFlowColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is dao count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AsyncFlowTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count async flow failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesasync.ListAsyncFlows{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableasync.AsyncFlowColumns.FieldsNamedExpr(opt.Fields),
		table.AsyncFlowTable, whereExpr, pageExpr)

	details := make([]tableasync.AsyncFlowTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select async flow failed, err: %v, sql: %s, filter: %v, rid: %s", err, sql,
			opt.Filter, kt.Rid)
		return nil, err
	}

	return &typesasync.ListAsyncFlows{Count: 0, Details: details}, nil
}

// DeleteWithTx async flow with tx.
func (dao *AsyncFlowDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AsyncFlowTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete async flow failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
