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

// Package zone ...
package zone

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typeszone "hcm/pkg/dal/dao/types/zone"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud/zone"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Zone only used for zone.
type Zone interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, zones []*zone.ZoneTable) ([]string, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, zone *zone.ZoneTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typeszone.ListZoneDetails, error)
	Delete(kt *kit.Kit, expr *filter.Expression) error
}

var _ Zone = new(ZoneDao)

// ZoneDao zone dao.
type ZoneDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create zone with tx
func (z ZoneDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, zones []*zone.ZoneTable) ([]string, error) {

	ids, err := z.IDGen.Batch(kt, table.ZoneTable, len(zones))
	if err != nil {
		return nil, err
	}
	for index, zone := range zones {
		zone.ID = ids[index]

		if err := zone.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ZoneTable,
		zone.ZoneColumns.ColumnExpr(), zone.ZoneColumns.ColonNameExpr())
	err = z.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, zones)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.ZoneTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.ZoneTable, err)
	}

	return ids, nil
}

// UpdateByIDWithTx update zone by id
func (z ZoneDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, zone *zone.ZoneTable) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := zone.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(zone, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, zone.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = z.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update zone failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// List list zone
func (z ZoneDao) List(kt *kit.Kit, opt *types.ListOption) (*typeszone.ListZoneDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list zone options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(zone.ZoneColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ZoneTable, whereExpr)

		count, err := z.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count zone failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typeszone.ListZoneDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, zone.ZoneColumns.FieldsNamedExpr(opt.Fields),
		table.ZoneTable, whereExpr, pageExpr)

	details := make([]zone.ZoneTable, 0)
	err = z.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &typeszone.ListZoneDetails{Details: details}, nil
}

// Delete delete zone
func (z ZoneDao) Delete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, argMap, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ZoneTable, whereExpr)

	_, err = z.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err = z.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Delete(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("delete zone failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}
