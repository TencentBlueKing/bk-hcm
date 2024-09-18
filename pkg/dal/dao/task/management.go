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

package task

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tasktype "hcm/pkg/dal/dao/types/task"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/task"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Management defines task Management dao operations.
type Management interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, managements []task.ManagementTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, management *task.ManagementTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*tasktype.ListTaskManagements, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Management = new(ManagementDao)

// ManagementDao task Management dao.
type ManagementDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewManagementDao create a task management dao.
func NewManagementDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) Management {
	return &ManagementDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// CreateWithTx create task managements with transaction.
func (d *ManagementDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, managements []task.ManagementTable) ([]string, error) {
	if len(managements) == 0 {
		return nil, errf.New(errf.InvalidParameter, "managements to create cannot be empty")
	}

	ids, err := d.idGen.Batch(kt, table.TaskManagementTable, len(managements))
	if err != nil {
		return nil, err
	}

	for idx := range managements {
		managements[idx].ID = ids[idx]
		managements[idx].Creator = kt.User
		if err = managements[idx].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.TaskManagementTable,
		task.ManagementColumns.ColumnExpr(), task.ManagementColumns.ColonNameExpr())

	err = d.orm.Txn(tx).BulkInsert(kt.Ctx, sql, managements)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.TaskManagementTable, err)
	}

	return ids, nil
}

// UpdateWithTx update task management with transaction.
func (d *ManagementDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	management *task.ManagementTable) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	management.Reviser = kt.User
	if err := management.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(management, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, management.TableName(), setExpr, whereExpr)

	effected, err := d.orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update task management failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update task management, but data not found, filter: %v, rid: %v", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List task managements.
func (d *ManagementDao) List(kt *kit.Kit, opt *types.ListOption,
	whereOpts ...*filter.SQLWhereOption) (*tasktype.ListTaskManagements, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list task management options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(task.ManagementColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereOpt := tools.DefaultSqlWhereOption
	if len(whereOpts) != 0 && whereOpts[0] != nil {
		err := whereOpts[0].Validate()
		if err != nil {
			return nil, err
		}
		whereOpt = whereOpts[0]
	}

	if opt.Filter == nil {
		opt.Filter = tools.AllExpression()
	}
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.TaskManagementTable, whereExpr)

		count, err := d.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count task managements failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &tasktype.ListTaskManagements{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, task.ManagementColumns.FieldsNamedExpr(opt.Fields),
		table.TaskManagementTable, whereExpr, pageExpr)

	managements := make([]task.ManagementTable, 0)
	if err = d.orm.Do().Select(kt.Ctx, &managements, sql, whereValue); err != nil {
		return nil, err
	}

	return &tasktype.ListTaskManagements{Details: managements}, nil
}

// DeleteWithTx delete task management with transaction.
func (d *ManagementDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.TaskManagementTable, whereExpr)
	if _, err = d.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete task management failed, err: %v, filter: %v, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
