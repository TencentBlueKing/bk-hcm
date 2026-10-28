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

// Package daoplugin 提供 iac_plugin 与 iac_plugin_schema 表的数据访问对象。
package daoplugin

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// IacPluginSchema iac_plugin_schema 表 DAO 接口。
type IacPluginSchema interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableplugin.IacPluginSchemaTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
		model *tableplugin.IacPluginSchemaTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListIacPluginSchemaDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ IacPluginSchema = new(IacPluginSchemaDao)

// IacPluginSchemaDao 实现 IacPluginSchema 接口。
type IacPluginSchemaDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// BatchCreateWithTx 在事务中批量创建 iac_plugin_schema 记录。
func (d *IacPluginSchemaDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []tableplugin.IacPluginSchemaTable) ([]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "iac plugin schema models to create cannot be empty")
	}

	for index := range models {
		if err := models[index].InsertValidate(); err != nil {
			logs.Errorf("insert validate iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
	}

	ids, err := d.IDGen.Batch(kt, table.IacPluginSchemaTable, len(models))
	if err != nil {
		logs.Errorf("generate id for %s failed, err: %v, rid: %s", table.IacPluginSchemaTable, err, kt.Rid)
		return nil, err
	}
	for index := range models {
		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`,
		table.IacPluginSchemaTable,
		tableplugin.IacPluginSchemaTableColumns.ColumnExpr(),
		tableplugin.IacPluginSchemaTableColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, sql: %s, err: %v, rid: %s", table.IacPluginSchemaTable, sql, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %w", table.IacPluginSchemaTable, err)
	}

	return ids, nil
}

// UpdateWithTx 在事务中按过滤条件更新 iac_plugin_schema 记录。
func (d *IacPluginSchemaDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
	model *tableplugin.IacPluginSchemaTable) error {

	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}
	if model == nil {
		return errf.New(errf.InvalidParameter, "iac plugin schema update model is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		logs.Errorf("update validate iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf("generate update where expr for iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	opts := utils.NewFieldOptions().
		AddIgnoredFields(types.DefaultIgnoredFields...).
		AddIgnoredFields("plugin_name", "plugin_version")

	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		logs.Errorf("rearrange sql data for iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %w", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)
	if _, err = d.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue)); err != nil {
		logs.ErrorJson("update iac_plugin_schema failed, filter: %v, err: %v, rid: %s", filterExpr, err, kt.Rid)
		return err
	}

	return nil
}

// List 查询 iac_plugin_schema 记录，支持 fields 投影（如仅 schema_hash 或完整记录）。
func (d *IacPluginSchemaDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListIacPluginSchemaDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list iac plugin schema option is nil")
	}

	exprOpt := filter.NewExprOption(filter.RuleFields(tableplugin.IacPluginSchemaTableColumns.ColumnTypes()))
	if err := opt.Validate(exprOpt, core.NewDefaultPageOption()); err != nil {
		logs.Errorf("validate list iac_plugin_schema option failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf("generate where expr for iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.IacPluginSchemaTable, whereExpr)
		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count iac_plugin_schema failed, err: %v, filter: %v, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &types.ListIacPluginSchemaDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		logs.Errorf("generate page sql expr for iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		tableplugin.IacPluginSchemaTableColumns.FieldsNamedExpr(opt.Fields),
		table.IacPluginSchemaTable, whereExpr, pageExpr)

	details := make([]tableplugin.IacPluginSchemaTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select iac_plugin_schema failed, sql: %s, err: %v, filter: %v, rid: %s",
			sql, err, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListIacPluginSchemaDetails{Details: details}, nil
}

// DeleteWithTx 在事务中按过滤条件删除 iac_plugin_schema 记录。
func (d *IacPluginSchemaDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf("generate delete where expr for iac_plugin_schema failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.IacPluginSchemaTable, whereExpr)
	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete iac_plugin_schema failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
