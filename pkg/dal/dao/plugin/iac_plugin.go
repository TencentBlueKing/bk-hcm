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

// Package daoplugin defines DAO for iac_plugin.
package daoplugin

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tableplugin "hcm/pkg/dal/table/plugin"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// IaCPlugin only used for iac_plugin.
type IaCPlugin interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableplugin.IaCPluginTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
		model *tableplugin.IaCPluginTable) (int64, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableplugin.IaCPluginTable], error)
}

var _ IaCPlugin = new(Dao)

// Dao iac plugin dao.
type Dao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx batch create iac_plugin with transaction.
func (d *Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableplugin.IaCPluginTable) ([]string, error) {
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	for index := range models {
		if err := models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	ids, err := d.IDGen.Batch(kt, table.IaCPluginTable, len(models))
	if err != nil {
		return nil, err
	}
	for index := range models {
		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, table.IaCPluginTable,
		tableplugin.IaCPluginColumns.ColumnExpr(), tableplugin.IaCPluginColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.IaCPluginTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.IaCPluginTable, err)
	}

	audits := slice.Map(models, func(one tableplugin.IaCPluginTable) *tableaudit.AuditTable {
		return &tableaudit.AuditTable{
			ResID:    one.ID,
			ResName:  one.PluginName,
			ResType:  enumor.IaCPluginAuditResType,
			Action:   enumor.Create,
			BkBizID:  constant.UnassignedBiz,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: one,
			},
		}
	})
	if err = d.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// UpdateWithTx 在事务中按过滤条件更新 iac_plugin 记录，返回实际受影响的行数。
func (d *Dao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
	model *tableplugin.IaCPluginTable) (int64, error) {

	if filterExpr == nil {
		return 0, errf.New(errf.InvalidParameter, "filter expr is nil")
	}
	if model == nil {
		return 0, errf.New(errf.InvalidParameter, "iac plugin update model is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		return 0, err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf("generate update where expr for %s failed, err: %v, rid: %s", table.IaCPluginTable, err, kt.Rid)
		return 0, err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return 0, fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)
	effected, err := d.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update %s failed, err: %v, filter: %v, rid: %s", table.IaCPluginTable, err, filterExpr, kt.Rid)
		return 0, err
	}

	return effected, nil
}

// DeleteWithTx 在事务中按过滤条件删除 iac_plugin 记录。
func (d *Dao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf("generate delete where expr for %s failed, err: %v, rid: %s", table.IaCPluginTable, err, kt.Rid)
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.IaCPluginTable, whereExpr)
	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete %s failed, err: %v, filter: %v, rid: %s", table.IaCPluginTable, err, expr, kt.Rid)
		return err
	}

	return nil
}

// List list iac_plugin.
func (d *Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableplugin.IaCPluginTable], error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableplugin.IaCPluginColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.IaCPluginTable, whereExpr)
		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count %s failed, err: %v, filter: %s, rid: %s", table.IaCPluginTable, err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &types.ListResult[tableplugin.IaCPluginTable]{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableplugin.IaCPluginColumns.FieldsNamedExpr(opt.Fields),
		table.IaCPluginTable, whereExpr, pageExpr)

	details := make([]tableplugin.IaCPluginTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select %s failed, err: %v, sql: %s, filter: %v, rid: %s",
			table.IaCPluginTable, err, sql, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListResult[tableplugin.IaCPluginTable]{Count: 0, Details: details}, nil
}
