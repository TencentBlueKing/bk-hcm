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

// SchemeInterface only used for scheme.
type SchemeInterface interface {
	Create(kt *kit.Kit, model *tableselection.SchemeTable) (string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableselection.SchemeTable], error)
	Delete(kt *kit.Kit, expr *filter.Expression) error
	UpdateByID(kt *kit.Kit, id string, model *tableselection.SchemeTable) error
}

var _ SchemeInterface = new(SchemeDao)

// SchemeDao scheme dao.
type SchemeDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// UpdateByID with tx.
func (dao SchemeDao) UpdateByID(kt *kit.Kit, id string, model *tableselection.SchemeTable) error {
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
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update scheme failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// Create ....
func (dao SchemeDao) Create(kt *kit.Kit, model *tableselection.SchemeTable) (string, error) {
	if err := model.InsertValidate(); err != nil {
		return "", err
	}

	id, err := dao.IDGen.One(kt, table.CloudSelectionSchemeTable)
	if err != nil {
		return "", err
	}
	model.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, model.TableName(),
		tableselection.SchemeTableColumns.ColumnExpr(), tableselection.SchemeTableColumns.ColonNameExpr())

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Insert(kt.Ctx, sql, model)
	if err != nil {
		if em := errf.GetMySQLDuplicated(err); em != nil {
			return "", errf.New(errf.RecordDuplicated, em.Message)
		}
		logs.Errorf("insert %s failed, err: %v, sql: %s, model: %+v, rid: %s",
			model.TableName(), err, sql, model, kt.Rid)
		return "", fmt.Errorf("insert %s failed, err: %v", model.TableName(), err)
	}

	return id, nil
}

// List scheme.
func (dao SchemeDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResult[tableselection.SchemeTable], error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tableselection.SchemeTableColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is dao count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.CloudSelectionSchemeTable, whereExpr)

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count scheme failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListResult[tableselection.SchemeTable]{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableselection.SchemeTableColumns.FieldsNamedExpr(opt.Fields),
		table.CloudSelectionSchemeTable, whereExpr, pageExpr)

	details := make([]tableselection.SchemeTable, 0)
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.Errorf("select scheme failed, err: %v, sql: %s, rid: %s", err, sql, kt.Rid)
		return nil, err
	}

	return &types.ListResult[tableselection.SchemeTable]{Count: 0, Details: details}, nil
}

// Delete scheme with tx.
func (dao SchemeDao) Delete(kt *kit.Kit, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.CloudSelectionSchemeTable, whereExpr)
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete scheme failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
