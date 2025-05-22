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

// Package audit ...
package audit

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/audit"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Interface define audit interface.
type Interface interface {
	Create(kt *kit.Kit, one *audit.AuditTable) error
	BatchCreate(kt *kit.Kit, audits []*audit.AuditTable) error
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, audits []*audit.AuditTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListAuditDetails, error)
}

var _ Interface = new(Dao)

// NewAudit new audit.
func NewAudit(orm orm.Interface) Interface {
	return &Dao{
		Orm: orm,
	}
}

// Dao audit dao.
type Dao struct {
	Orm orm.Interface
}

// Create audit.
func (d Dao) Create(kt *kit.Kit, one *audit.AuditTable) error {
	return d.BatchCreate(kt, []*audit.AuditTable{one})
}

// BatchCreate batch create audit.
func (d Dao) BatchCreate(kt *kit.Kit, audits []*audit.AuditTable) error {
	for _, one := range audits {
		if err := one.CreateValidate(); err != nil {
			return err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AuditTable,
		audit.AuditColumns.ColumnExpr(), audit.AuditColumns.ColonNameExpr())
	err := d.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().BulkInsert(kt.Ctx, sql, audits)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.AuditTable, err, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", table.AuditTable, err)
	}

	return nil
}

// BatchCreateWithTx batch create audit with tx.
func (d Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, audits []*audit.AuditTable) error {
	for _, one := range audits {
		if err := one.CreateValidate(); err != nil {
			return err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AuditTable,
		audit.AuditColumns.ColumnExpr(), audit.AuditColumns.ColonNameExpr())
	err := d.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, audits)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.AuditTable, err, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", table.AuditTable, err)
	}

	return nil
}

// List audit.
func (d Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListAuditDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}
	columnTypes := audit.AuditColumns.ColumnTypes()
	columnTypes["detail.data.res_flow.flow_id"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AuditTable, whereExpr)

		count, err := d.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count audit failed, err: %v, filter: %d, rid: %d", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListAuditDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, audit.AuditColumns.FieldsNamedExpr(opt.Fields),
		table.AuditTable, whereExpr, pageExpr)

	details := make([]audit.AuditTable, 0)
	err = d.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.ListAuditDetails{Details: details}, nil
}
