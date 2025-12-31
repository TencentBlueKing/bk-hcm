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

// Package devicecapacity ...
package devicecapacity

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	devicecapacitytype "hcm/pkg/dal/dao/types/device-capacity"
	"hcm/pkg/dal/table"
	devicecapacity "hcm/pkg/dal/table/device-capacity"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// DeviceCapacity defines device capacity dao operations.
type DeviceCapacity interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, deviceCapacity []devicecapacity.DeviceCapacityTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
		deviceCapacity *devicecapacity.DeviceCapacityTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
		*devicecapacitytype.ListDeviceCapacities, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListWithDeviceInfo(kt *kit.Kit, opt *types.ListOption) (*devicecapacitytype.ListCapacitiesWithDeviceInfo, error)
}

var _ DeviceCapacity = new(DeviceCapacityDao)

// DeviceCapacityDao device capacity dao.
type DeviceCapacityDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewDeviceCapacityDao create a device capacity dao.
func NewDeviceCapacityDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) DeviceCapacity {
	return &DeviceCapacityDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// CreateWithTx create device capacity with transaction.
func (d *DeviceCapacityDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	deviceCapacities []devicecapacity.DeviceCapacityTable) ([]string, error) {
	if len(deviceCapacities) == 0 {
		return nil, errf.New(errf.InvalidParameter, "device capacities to create cannot be empty")
	}

	ids, err := d.idGen.Batch(kt, table.DeviceCapacityTable, len(deviceCapacities))
	if err != nil {
		return nil, err
	}

	for idx := range deviceCapacities {
		deviceCapacities[idx].ID = ids[idx]
		if err = deviceCapacities[idx].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.DeviceCapacityTable,
		devicecapacity.DeviceCapacityColumns.ColumnExpr(),
		devicecapacity.DeviceCapacityColumns.ColonNameExpr())

	err = d.orm.Txn(tx).BulkInsert(kt.Ctx, sql, deviceCapacities)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.DeviceCapacityTable, err)
	}

	return ids, nil
}

// UpdateWithTx update device capacity with transaction.
func (d *DeviceCapacityDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	deviceCapacity *devicecapacity.DeviceCapacityTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	deviceCapacity.Reviser = kt.User
	if err := deviceCapacity.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(deviceCapacity, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, deviceCapacity.TableName(), setExpr, whereExpr)

	if _, err = d.orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue)); err != nil {
		logs.Errorf("update device capacity failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// List device capacity.
func (d *DeviceCapacityDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*devicecapacitytype.ListDeviceCapacities, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list device capacity options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(devicecapacity.DeviceCapacityColumns.ColumnTypes())),
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.DeviceCapacityTable, whereExpr)

		count, err := d.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count device capacity failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &devicecapacitytype.ListDeviceCapacities{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, devicecapacity.DeviceCapacityColumns.FieldsNamedExpr(opt.Fields),
		table.DeviceCapacityTable, whereExpr, pageExpr)

	deviceCapacities := make([]devicecapacity.DeviceCapacityTable, 0)
	if err = d.orm.Do().Select(kt.Ctx, &deviceCapacities, sql, whereValue); err != nil {
		return nil, err
	}

	return &devicecapacitytype.ListDeviceCapacities{DeviceCapacities: deviceCapacities}, nil
}

// DeleteWithTx delete device capacity with transaction.
func (d *DeviceCapacityDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.DeviceCapacityTable, whereExpr)
	if _, err = d.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete device capacity failed, err: %v, filter: %v, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// ListWithDeviceInfo list device capacity with device type details by joining woa_device_type table.
func (d *DeviceCapacityDao) ListWithDeviceInfo(kt *kit.Kit, opt *types.ListOption) (
	*devicecapacitytype.ListCapacitiesWithDeviceInfo, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list device capacity options is nil")
	}

	// Build WHERE clause for join query
	whereOpt := tools.DefaultSqlWhereOption
	if opt.Filter == nil {
		opt.Filter = tools.AllExpression()
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s AS dc LEFT JOIN %s AS dt ON dc.device_type = dt.device_type %s`,
			table.DeviceCapacityTable, table.WoaDeviceTypeTable, whereExpr)
		count, err := d.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count device capacity with details failed, err: %v, filter: %s, rid: %s", err, opt.Filter,
				kt.Rid)
			return nil, err
		}

		return &devicecapacitytype.ListCapacitiesWithDeviceInfo{Count: count}, nil
	}

	if opt.Page.Sort == "" {
		opt.Page.Sort = "dc.id"
	}
	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	// Select fields from both tables
	// dc: device_capacity table, dt: woa_device_type table
	sql := fmt.Sprintf(`SELECT dc.require_type, dc.region, dc.zone, dc.capacity, dc.device_type, dt.device_family,
       dt.memory, dt.cpu_core, dt.core_type, dt.device_type_class FROM %s AS dc LEFT JOIN %s AS dt ON dc.device_type = 
           dt.device_type %s %s`, table.DeviceCapacityTable, table.WoaDeviceTypeTable, whereExpr, pageExpr)

	details := make([]devicecapacitytype.CapacityWithDeviceInfo, 0)
	if err = d.orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("list device capacity with device info failed, err: %v, filter: %s, rid: %s", err, opt.Filter,
			kt.Rid)
		return nil, err
	}

	return &devicecapacitytype.ListCapacitiesWithDeviceInfo{Details: details}, nil
}
