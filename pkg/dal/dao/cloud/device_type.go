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

package cloud

import (
	"fmt"

	"hcm/pkg/api/core"
	protocloud "hcm/pkg/api/data-service/cloud"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	dt "hcm/pkg/dal/table/cloud/device-type"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// DeviceTypeInterface only used for device type interface.
type DeviceTypeInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []dt.DeviceTypeTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, model *dt.DeviceTypeTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*protocloud.DeviceTypeTableListResult, error)
	ListDistinctDeviceType(kt *kit.Kit, opt *types.ListOption) (*protocloud.DeviceTypeTableListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ DeviceTypeInterface = new(DeviceTypeDao)

// DeviceTypeDao device type dao.
type DeviceTypeDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create device type with tx.
func (d DeviceTypeDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []dt.DeviceTypeTable) ([]string, error) {
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := d.IDGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for index := range models {
		models[index].ID = ids[index]

		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		dt.DeviceTypeColumns.ColumnExpr(), dt.DeviceTypeColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// UpdateWithTx update device type with tx.
func (d DeviceTypeDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
	model *dt.DeviceTypeTable) error {

	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	effected, err := d.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update device type failed, filter: %v, err: %v, rid: %v", filterExpr, err, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update device type, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
	}

	return nil
}

// List get device type list.
func (d DeviceTypeDao) List(kt *kit.Kit, opt *types.ListOption) (*protocloud.DeviceTypeTableListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list device type options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(dt.DeviceTypeColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.DeviceTypeTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count device type failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &protocloud.DeviceTypeTableListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, dt.DeviceTypeColumns.FieldsNamedExpr(opt.Fields),
		table.DeviceTypeTable, whereExpr, pageExpr)

	details := make([]dt.DeviceTypeTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &protocloud.DeviceTypeTableListResult{Count: 0, Details: details}, nil
}

// ListDistinctDeviceType list distinct device type.
func (d DeviceTypeDao) ListDistinctDeviceType(kt *kit.Kit, opt *types.ListOption) (
	*protocloud.DeviceTypeTableListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list distinct device type options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(dt.DeviceTypeColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(DISTINCT device_type) FROM %s %s`, table.DeviceTypeTable, whereExpr)
		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			return nil, err
		}
		return &protocloud.DeviceTypeTableListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s GROUP BY device_type %s`, dt.DeviceTypeColumns.FieldsNamedExpr(opt.Fields),
		table.DeviceTypeTable, whereExpr, pageExpr)

	details := make([]dt.DeviceTypeTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &protocloud.DeviceTypeTableListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete device type with tx.
func (d DeviceTypeDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.DeviceTypeTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete device type failed, err: %v, filter: %v, rid: %s", expr, err, kt.Rid)
		return err
	}

	return nil
}
