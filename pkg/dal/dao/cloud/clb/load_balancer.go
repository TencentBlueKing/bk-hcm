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

// Package clb 负载均衡的Package
package clb

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typesclb "hcm/pkg/dal/dao/types/clb"
	"hcm/pkg/dal/table"
	tableclb "hcm/pkg/dal/table/cloud/clb"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ClbInterface only used for clb.
type ClbInterface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tableclb.LoadBalancerTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tableclb.LoadBalancerTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tableclb.LoadBalancerTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesclb.ListClbDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ ClbInterface = new(ClbDao)

// ClbDao clb dao.
type ClbDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx clb.
func (dao ClbDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tableclb.LoadBalancerTable) ([]string, error) {
	tableName := table.LoadBalancerTable
	ids, err := dao.IDGen.Batch(kt, tableName, len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		if err = model.InsertValidate(); err != nil {
			return nil, err
		}
		model.ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tableclb.LoadBalancerColumns.ColumnExpr(), tableclb.LoadBalancerColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return ids, nil
}

// Update clb.
func (dao ClbDao) Update(kt *kit.Kit, expr *filter.Expression, model *tableclb.LoadBalancerTable) error {
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

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effect, err := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.Errorf("update load balancer failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update load balancer, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx clb.
func (dao ClbDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tableclb.LoadBalancerTable) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, model.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update load balancer failed, id: %s, err: %v, rid: %v", id, err, kt.Rid)
		return err
	}

	return nil
}

// List clb.
func (dao ClbDao) List(kt *kit.Kit, opt *types.ListOption) (*typesclb.ListClbDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableclb.LoadBalancerColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.LoadBalancerTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesclb.ListClbDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableclb.LoadBalancerColumns.FieldsNamedExpr(opt.Fields),
		table.LoadBalancerTable, whereExpr, pageExpr)

	details := make([]tableclb.LoadBalancerTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typesclb.ListClbDetails{Details: details}, nil
}

// DeleteWithTx clb.
func (dao ClbDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.LoadBalancerTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete load balancer failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// ListClbByIDs clb
func ListClbByIDs(kt *kit.Kit, orm orm.Interface, ids []string) (map[string]tableclb.LoadBalancerTable, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE id IN (:ids)`, tableclb.LoadBalancerColumns.FieldsNamedExpr(nil),
		table.LoadBalancerTable)

	clbs := make([]tableclb.LoadBalancerTable, 0)
	if err := orm.Do().Select(kt.Ctx, &clbs, sql, map[string]interface{}{"ids": ids}); err != nil {
		return nil, err
	}

	idMap := make(map[string]tableclb.LoadBalancerTable, len(ids))
	for _, one := range clbs {
		idMap[one.ID] = one
	}

	return idMap, nil
}
