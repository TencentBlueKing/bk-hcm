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

// Package clb clb异步任务资源锁的Package
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
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ClbFlowLockInterface only used for clb flow lock.
type ClbFlowLockInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tablelb.ClbFlowLockTable) error
	Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.ClbFlowLockTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, resID, resType, owner string, model *tablelb.ClbFlowLockTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesclb.ListClbFlowLockDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ ClbFlowLockInterface = new(ClbFlowLockDao)

// ClbFlowLockDao clb flow lock dao.
type ClbFlowLockDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx clb flow lock.
func (dao ClbFlowLockDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tablelb.ClbFlowLockTable) error {
	if err := model.InsertValidate(); err != nil {
		return err
	}

	tableName := table.ClbFlowLockTable
	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tablelb.ClbFlowLockColumns.ColumnExpr(), tablelb.ClbFlowLockColumns.ColonNameExpr())

	if err := dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, model); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return nil
}

// Update clb flow lock.
func (dao ClbFlowLockDao) Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.ClbFlowLockTable) error {
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
			logs.Errorf("update load balancer flow lock failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update load balancer flow lock, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx clb flow lock.
func (dao ClbFlowLockDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, resID, resType, owner string,
	model *tablelb.ClbFlowLockTable) error {

	if len(resID) == 0 {
		return errf.New(errf.InvalidParameter, "res_id is required")
	}

	if len(resType) == 0 {
		return errf.New(errf.InvalidParameter, "res_type is required")
	}

	if len(owner) == 0 {
		return errf.New(errf.InvalidParameter, "owner is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s WHERE res_id = :res_id AND res_type = :res_type AND owner = :owner`,
		model.TableName(), setExpr)

	toUpdate["res_id"] = resID
	toUpdate["res_type"] = resType
	toUpdate["owner"] = owner
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update load balancer flow lock failed, resID: %s, resType: %s, owner: %s, err: %v, rid: %v",
			resID, resType, owner, err, kt.Rid)
		return err
	}

	return nil
}

// List clb flow lock.
func (dao ClbFlowLockDao) List(kt *kit.Kit, opt *types.ListOption) (*typesclb.ListClbFlowLockDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablelb.ClbFlowLockColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ClbFlowLockTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer flow lock failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesclb.ListClbFlowLockDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablelb.ClbFlowLockColumns.FieldsNamedExpr(opt.Fields),
		table.ClbFlowLockTable, whereExpr, pageExpr)

	details := make([]tablelb.ClbFlowLockTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typesclb.ListClbFlowLockDetails{Details: details}, nil
}

// DeleteWithTx clb flow lock.
func (dao ClbFlowLockDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ClbFlowLockTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete load balancer flow lock failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
