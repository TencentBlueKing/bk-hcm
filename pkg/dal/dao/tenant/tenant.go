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

// Package tenant ...
package tenant

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	tenanttype "hcm/pkg/dal/dao/types/tenant"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/tenant"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Tenant defines tenant dao operations.
type Tenant interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, tenant []tenant.TenantTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, tenant *tenant.TenantTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*tenanttype.ListTenants, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Tenant = new(TenantDao)

// TenantDao tenant dao.
type TenantDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewTenantDao create a tenant dao.
func NewTenantDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) Tenant {
	return &TenantDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// CreateWithTx create tenant with transaction.
func (d *TenantDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, tenants []tenant.TenantTable) ([]string, error) {
	if len(tenants) == 0 {
		return nil, errf.New(errf.InvalidParameter, "tenants to create cannot be empty")
	}

	ids, err := d.idGen.Batch(kt, table.TenantTable, len(tenants))
	if err != nil {
		return nil, err
	}

	for idx := range tenants {
		tenants[idx].ID = ids[idx]
		tenants[idx].Creator = kt.User
		if err = tenants[idx].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.TenantTable, tenant.TenantColumns.ColumnExpr(),
		tenant.TenantColumns.ColonNameExpr())

	err = d.orm.Txn(tx).BulkInsert(kt.Ctx, sql, tenants)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.TenantTable, err)
	}

	return ids, nil
}

// UpdateWithTx update tenant with transaction.
func (d *TenantDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, tenant *tenant.TenantTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	tenant.Reviser = kt.User
	if err := tenant.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.TenantDefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(tenant, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, tenant.TableName(), setExpr, whereExpr)

	effected, err := d.orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))

	if err != nil {
		logs.ErrorJson("update tenant failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update tenant, but data not found, filter: %v, rid: %s", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List tenant.
func (d *TenantDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*tenanttype.ListTenants, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list tenant options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tenant.TenantColumns.ColumnTypes())),
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.TenantTable, whereExpr)

		count, err := d.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count tenant failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &tenanttype.ListTenants{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tenant.TenantColumns.FieldsNamedExpr(opt.Fields), table.TenantTable,
		whereExpr, pageExpr)

	tenants := make([]tenant.TenantTable, 0)
	if err = d.orm.Do().Select(kt.Ctx, &tenants, sql, whereValue); err != nil {
		return nil, err
	}

	return &tenanttype.ListTenants{Tenants: tenants}, nil
}

// DeleteWithTx delete tenant with transaction.
func (d *TenantDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.TenantTable, whereExpr)
	if _, err = d.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete tenant failed, err: %v, filter: %v, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
