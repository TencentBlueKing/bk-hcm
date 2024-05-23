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

// Package resflow 资源与异步任务关系的Package
package resflow

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typeslb "hcm/pkg/dal/dao/types/load-balancer"
	"hcm/pkg/dal/table"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ResourceFlowRelInterface only used for resource flow rel.
type ResourceFlowRelInterface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.ResourceFlowRelTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.ResourceFlowRelTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablelb.ResourceFlowRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListResourceFlowRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ ResourceFlowRelInterface = new(ResourceFlowRelDao)

// ResourceFlowRelDao resource flow rel dao.
type ResourceFlowRelDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx resource flow rel.
func (dao ResourceFlowRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.ResourceFlowRelTable) (
	[]string, error) {

	tableName := table.ResourceFlowRelTable
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
		tablelb.ResourceFlowRelColumns.ColumnExpr(), tablelb.ResourceFlowRelColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return ids, nil
}

// Update resource flow rel.
func (dao ResourceFlowRelDao) Update(kt *kit.Kit, expr *filter.Expression,
	model *tablelb.ResourceFlowRelTable) error {

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
			logs.Errorf("update load balancer flow rel failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update load balancer flow rel, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx resource flow rel.
func (dao ResourceFlowRelDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	model *tablelb.ResourceFlowRelTable) error {

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
		logs.Errorf("update load balancer flow rel failed, id: %s, err: %v, rid: %v", id, err, kt.Rid)
		return err
	}

	return nil
}

// List resource flow rel.
func (dao ResourceFlowRelDao) List(kt *kit.Kit, opt *types.ListOption) (
	*typeslb.ListResourceFlowRelDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablelb.ResourceFlowRelColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ResourceFlowRelTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer flow rel failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typeslb.ListResourceFlowRelDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablelb.ResourceFlowRelColumns.FieldsNamedExpr(opt.Fields),
		table.ResourceFlowRelTable, whereExpr, pageExpr)

	details := make([]tablelb.ResourceFlowRelTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typeslb.ListResourceFlowRelDetails{Details: details}, nil
}

// DeleteWithTx resource flow rel.
func (dao ResourceFlowRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ResourceFlowRelTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete load balancer flow rel failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
