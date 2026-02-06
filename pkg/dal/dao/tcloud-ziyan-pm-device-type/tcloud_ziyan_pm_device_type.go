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

// Package tcloudziyanpmdevicetype ...
package tcloudziyanpmdevicetype

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	ziyanpmdtdal "hcm/pkg/dal/dao/types/tcloud-ziyan-pm-device-type"
	"hcm/pkg/dal/table"
	ziyanpmdt "hcm/pkg/dal/table/tcloud-ziyan-pm-device-type"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// TCloudZiyanPmDeviceType defines tcloud ziyan pm device type dao operations.
type TCloudZiyanPmDeviceType interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, deviceTypes []ziyanpmdt.TCloudZiyanPmDeviceTypeTable) (
		[]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
		deviceType *ziyanpmdt.TCloudZiyanPmDeviceTypeTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
		*ziyanpmdtdal.ListTCloudZiyanPmDeviceTypes, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ TCloudZiyanPmDeviceType = new(TCloudZiyanPmDeviceTypeDao)

// TCloudZiyanPmDeviceTypeDao tcloud ziyan pm device type dao.
type TCloudZiyanPmDeviceTypeDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewTCloudZiyanPmDeviceTypeDao create a tcloud ziyan pm device type dao.
func NewTCloudZiyanPmDeviceTypeDao(orm orm.Interface, idGen idgenerator.IDGenInterface,
	audit audit.Interface) TCloudZiyanPmDeviceType {
	return &TCloudZiyanPmDeviceTypeDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// CreateWithTx create tcloud ziyan pm device type with transaction.
func (d *TCloudZiyanPmDeviceTypeDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	deviceTypes []ziyanpmdt.TCloudZiyanPmDeviceTypeTable) ([]string, error) {
	if len(deviceTypes) == 0 {
		return nil, errf.New(errf.InvalidParameter, "device types to create cannot be empty")
	}

	ids, err := d.idGen.Batch(kt, table.TCloudZiyanPmDeviceTypeTable, len(deviceTypes))
	if err != nil {
		return nil, err
	}

	for idx := range deviceTypes {
		deviceTypes[idx].ID = ids[idx]
		if err = deviceTypes[idx].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.TCloudZiyanPmDeviceTypeTable,
		ziyanpmdt.TCloudZiyanPmDeviceTypeColumns.ColumnExpr(),
		ziyanpmdt.TCloudZiyanPmDeviceTypeColumns.ColonNameExpr())

	err = d.orm.Txn(tx).BulkInsert(kt.Ctx, sql, deviceTypes)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.TCloudZiyanPmDeviceTypeTable, err)
	}

	return ids, nil
}

// UpdateWithTx update tcloud ziyan pm device type with transaction.
func (d *TCloudZiyanPmDeviceTypeDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	deviceType *ziyanpmdt.TCloudZiyanPmDeviceTypeTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	deviceType.Reviser = kt.User
	if err := deviceType.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(deviceType, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, deviceType.TableName(), setExpr, whereExpr)

	effected, err := d.orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))

	if err != nil {
		logs.ErrorJson("update tcloud ziyan pm device type failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update tcloud ziyan pm device type, but data not found, filter: %v, rid: %s", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List tcloud ziyan pm device type.
func (d *TCloudZiyanPmDeviceTypeDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*ziyanpmdtdal.ListTCloudZiyanPmDeviceTypes, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list tcloud ziyan pm device type options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(
		ziyanpmdt.TCloudZiyanPmDeviceTypeColumns.ColumnTypes())),
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.TCloudZiyanPmDeviceTypeTable, whereExpr)

		count, err := d.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count tcloud ziyan pm device type failed, err: %v, filter: %s, rid: %s", err, opt.Filter,
				kt.Rid)
			return nil, err
		}

		return &ziyanpmdtdal.ListTCloudZiyanPmDeviceTypes{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		ziyanpmdt.TCloudZiyanPmDeviceTypeColumns.FieldsNamedExpr(opt.Fields),
		table.TCloudZiyanPmDeviceTypeTable, whereExpr, pageExpr)

	deviceTypes := make([]ziyanpmdt.TCloudZiyanPmDeviceTypeTable, 0)
	if err = d.orm.Do().Select(kt.Ctx, &deviceTypes, sql, whereValue); err != nil {
		return nil, err
	}

	return &ziyanpmdtdal.ListTCloudZiyanPmDeviceTypes{DeviceTypes: deviceTypes}, nil
}

// DeleteWithTx delete tcloud ziyan pm device type with transaction.
func (d *TCloudZiyanPmDeviceTypeDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.TCloudZiyanPmDeviceTypeTable, whereExpr)
	if _, err = d.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete tcloud ziyan pm device type failed, err: %v, filter: %v, rid: %s", err, filterExpr,
			kt.Rid)
		return err
	}

	return nil
}
