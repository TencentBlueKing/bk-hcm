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

// Package resplan ...
package resplan

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	mtypes "hcm/pkg/dal/dao/types/meta"
	"hcm/pkg/dal/table"
	wz "hcm/pkg/dal/table/resource-plan/woa-zone"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// WoaZoneInterface only used for woa zone interface.
type WoaZoneInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []wz.WoaZoneTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *wz.WoaZoneTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*mtypes.WoaZoneListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	// GetZoneMap get zone id name mapping.
	GetZoneMap(kt *kit.Kit) (map[string]string, error)
	// GetRegionAreaMap get region area mapping.
	GetRegionAreaMap(kt *kit.Kit) (map[string]mtypes.RegionArea, error)
	// GetZoneList get zone id and name list.
	GetZoneList(kt *kit.Kit, expr *filter.Expression) ([]mtypes.ZoneElem, error)
	// GetRegionList get region id and name list.
	GetRegionList(kt *kit.Kit, expr *filter.Expression) ([]mtypes.RegionElem, error)
}

var _ WoaZoneInterface = new(WoaZoneDao)

// WoaZoneDao woa zone WoaZoneDao.
type WoaZoneDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create woa zone with tx.
func (d WoaZoneDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []wz.WoaZoneTable) ([]string, error) {
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
		wz.WoaZoneColumns.ColumnExpr(), wz.WoaZoneColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update update woa zone.
func (d WoaZoneDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *wz.WoaZoneTable) error {
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

	_, err = d.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := d.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update woa zone failed, filter: %v, err: %v, rid: %v", filterExpr, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update woa zone, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List get woa zone list.
func (d WoaZoneDao) List(kt *kit.Kit, opt *types.ListOption) (*mtypes.WoaZoneListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list woa zone options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(wz.WoaZoneColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.WoaZoneTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count woa zone failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &mtypes.WoaZoneListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, wz.WoaZoneColumns.FieldsNamedExpr(opt.Fields),
		table.WoaZoneTable, whereExpr, pageExpr)

	details := make([]wz.WoaZoneTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &mtypes.WoaZoneListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete woa zone with tx.
func (d WoaZoneDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.WoaZoneTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete woa zone failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// GetZoneMap get zone id name mapping.
func (d WoaZoneDao) GetZoneMap(kt *kit.Kit) (map[string]string, error) {
	sql := fmt.Sprintf(`SELECT DISTINCT zone_id, zone_name FROM %s`, table.WoaZoneTable)
	details := make([]wz.WoaZoneTable, 0)
	if err := d.Orm.Do().Select(kt.Ctx, &details, sql, nil); err != nil {
		return nil, err
	}

	zoneMap := make(map[string]string)
	for _, detail := range details {
		zoneMap[detail.ZoneID] = detail.ZoneName
	}

	return zoneMap, nil
}

// GetRegionAreaMap get region area mapping.
func (d WoaZoneDao) GetRegionAreaMap(kt *kit.Kit) (map[string]mtypes.RegionArea, error) {
	sql := fmt.Sprintf(`SELECT DISTINCT region_id, region_name, area_id, area_name FROM %s`, table.WoaZoneTable)
	details := make([]wz.WoaZoneTable, 0)
	if err := d.Orm.Do().Select(kt.Ctx, &details, sql, nil); err != nil {
		return nil, err
	}

	regionMap := make(map[string]mtypes.RegionArea)
	for _, detail := range details {
		regionMap[detail.RegionID] = mtypes.RegionArea{
			RegionID:   detail.RegionID,
			RegionName: detail.RegionName,
			AreaName:   detail.AreaName,
		}
	}

	return regionMap, nil
}

// GetZoneList get zone id and name list.
func (d WoaZoneDao) GetZoneList(kt *kit.Kit, expr *filter.Expression) ([]mtypes.ZoneElem, error) {
	if expr == nil {
		return nil, errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT DISTINCT zone_id, zone_name FROM %s %s`, table.WoaZoneTable, whereExpr)
	details := make([]mtypes.ZoneElem, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return details, nil
}

// GetRegionList get region id and name list.
func (d WoaZoneDao) GetRegionList(kt *kit.Kit, expr *filter.Expression) ([]mtypes.RegionElem, error) {
	if expr == nil {
		return nil, errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT DISTINCT region_id, region_name FROM %s %s`, table.WoaZoneTable, whereExpr)
	details := make([]mtypes.RegionElem, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return details, nil
}
