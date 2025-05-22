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

package daoselection

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableselection "hcm/pkg/dal/table/cloud-selection"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// IdcInterface only used for idc.
type IdcInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tableselection.IdcTable) (string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableselection.IdcTable], error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tableselection.IdcTable) error
}

var _ IdcInterface = new(IdcDao)

// IdcDao idc dao.
type IdcDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// UpdateByIDWithTx with tx.
func (dao IdcDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tableselection.IdcTable) error {
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
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update idc failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// CreateWithTx with tx.
func (dao IdcDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tableselection.IdcTable) (string, error) {
	if err := model.InsertValidate(); err != nil {
		return "", err
	}

	id, err := dao.IDGen.One(kt, table.CloudSelectionIdcTable)
	if err != nil {
		return "", err
	}
	model.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, model.TableName(), tableselection.IdcTableColumns.ColumnExpr(),
		tableselection.IdcTableColumns.ColonNameExpr())

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Insert(kt.Ctx, sql, model)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, model: %+v, rid: %s", err, sql, model, kt.Rid)
		return "", fmt.Errorf("insert %s failed, err: %v", model.TableName(), err)
	}

	return id, nil
}

// List idc.
func (dao IdcDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableselection.IdcTable], error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableselection.IdcTableColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is dao count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.CloudSelectionIdcTable, whereExpr)

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count idc failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListResult[tableselection.IdcTable]{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableselection.IdcTableColumns.FieldsNamedExpr(opt.Fields),
		table.CloudSelectionIdcTable, whereExpr, pageExpr)

	details := make([]tableselection.IdcTable, 0)
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.Errorf("select idc failed, err: %v, sql: %s, rid: %s", err, sql, kt.Rid)
		return nil, err
	}

	return &types.ListResult[tableselection.IdcTable]{Count: 0, Details: details}, nil
}

// DeleteWithTx idc with tx.
func (dao IdcDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.CloudSelectionIdcTable, whereExpr)
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete idc failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
