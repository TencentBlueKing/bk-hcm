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
	Create(kt *kit.Kit, tx *sqlx.Tx, model *tableasync.AsyncFlowTable) (string, error)
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableasync.AsyncFlowTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tableasync.AsyncFlowTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tableasync.AsyncFlowTable) error
	UpdateStateByCAS(kt *kit.Kit, tx *sqlx.Tx, info *typesasync.UpdateFlowInfo) error
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

// Create async flow.
func (dao *AsyncFlowDao) Create(kt *kit.Kit, tx *sqlx.Tx, model *tableasync.AsyncFlowTable) (string, error) {

	id, err := dao.IDGen.One(kt, table.AsyncFlowTable)
	if err != nil {
		return "", err
	}
	model.ID = id

	if err = model.InsertValidate(); err != nil {
		return "", err
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AsyncFlowTable,
		tableasync.AsyncFlowColumns.ColumnExpr(), tableasync.AsyncFlowColumns.ColonNameExpr())
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, model)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s", table.AsyncFlowTable, err, sql, kt.Rid)
		return "", fmt.Errorf("insert %s failed, err: %v", table.AsyncFlowTable, err)
	}

	return id, nil
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

	opts := utils.NewFieldOptions().AddBlankedFields("worker").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).
			Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
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

	opts := utils.NewFieldOptions().AddBlankedFields("worker").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, model.TableName(), setExpr)

	toUpdate["id"] = id
	effected, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
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

// UpdateStateByCAS update async flow state by CAS.
func (dao *AsyncFlowDao) UpdateStateByCAS(kt *kit.Kit, tx *sqlx.Tx, info *typesasync.UpdateFlowInfo) error {

	if err := info.Validate(); err != nil {
		return err
	}

	setSql := "set state = :target"
	if info.Worker != nil {
		setSql += ", worker = :worker"
	}

	if info.Reason != nil {
		setSql += ", reason = :reason"
	}

	sql := fmt.Sprintf(`update %s %s where id = :id and state = :source`, table.AsyncFlowTable, setSql)

	whereValue := map[string]interface{}{
		"id":     info.ID,
		"source": info.Source,
		"target": info.Target,
		"worker": info.Worker,
		"reason": info.Reason,
	}
	effected, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).
		Update(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.Errorf("update async flow failed, err: %v, id: %s, sql: %s, rid: %v", err, info.ID, sql, kt.Rid)
		return err
	}

	if effected == 0 {
		return errf.Newf(errf.RecordNotUpdate, "flow[%s] update state: `%s`->`%s`, worker: %+v failed",
			info.ID, info.Source, info.Target, info.Worker)
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

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
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

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).
			Count(kt.Ctx, sql, whereValue)
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
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
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

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
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
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
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
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete async flow failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
