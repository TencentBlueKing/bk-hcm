/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package daoapplystat package dao apply statistics config.
package daoapplystat

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableapplystat "hcm/pkg/dal/table/cvm-apply-order-statistics-config"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Interface only used for cvm apply order statistics config.
type Interface interface {
	List(kt *kit.Kit, opt *types.ListOption) (
		*types.ListResult[tableapplystat.CvmApplyOrderStatisticsConfigTable], error)
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
		models []*tableapplystat.CvmApplyOrderStatisticsConfigTable) ([]string, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, f *filter.Expression) error
}

var _ Interface = new(Dao)

// Dao cvm apply order statistics config dao.
type Dao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create cvm apply order statistics config with tx.
func (d Dao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tableapplystat.CvmApplyOrderStatisticsConfigTable) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := d.IDGen.Batch(kt, table.CvmApplyOrderStatisticsConfigTable, len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		if err = model.InsertValidate(); err != nil {
			return nil, err
		}

		model.ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, table.CvmApplyOrderStatisticsConfigTable,
		tableapplystat.CvmApplyOrderStatisticsConfigTableColumns.ColumnExpr(),
		tableapplystat.CvmApplyOrderStatisticsConfigTableColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, sql: %s, err: %v, rid: %s", table.CvmApplyOrderStatisticsConfigTable, sql, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.CvmApplyOrderStatisticsConfigTable, err)
	}

	return ids, nil
}

// UpdateWithTx ... 此方法当前未实现。如需使用，请重新实现。
func (d Dao) UpdateWithTx() error {
	return fmt.Errorf("UpdateWithTx is not implemented yet")
}

// List ...
func (d Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableapplystat.CvmApplyOrderStatisticsConfigTable], error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list cvm apply order statistics config options is nil")
	}

	columnTypes := tableapplystat.CvmApplyOrderStatisticsConfigTableColumns.ColumnTypes()
	if err := opt.ValidateExcludeFilter(
		filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.CvmApplyOrderStatisticsConfigTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson(
				"count cvm apply order statistics config failed, err: %v, filter: %v, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListResult[tableapplystat.CvmApplyOrderStatisticsConfigTable]{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		tableapplystat.CvmApplyOrderStatisticsConfigTableColumns.FieldsNamedExpr(opt.Fields),
		table.CvmApplyOrderStatisticsConfigTable, whereExpr, pageExpr)

	details := make([]tableapplystat.CvmApplyOrderStatisticsConfigTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}
	return &types.ListResult[tableapplystat.CvmApplyOrderStatisticsConfigTable]{
		Count: 0, Details: details}, nil
}

// DeleteWithTx delete cvm apply order statistics config with tx.
func (d Dao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.CvmApplyOrderStatisticsConfigTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson(
			"delete cvm apply order statistics config failed, err: %v, filter: %v, rid: %s",
			err, expr, kt.Rid)
		return err
	}

	return nil
}
