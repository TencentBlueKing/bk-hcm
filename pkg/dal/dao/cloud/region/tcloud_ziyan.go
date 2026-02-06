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

package region

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud/region"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// TCloudZiyanRegion 腾讯自研云region操作接口定义
type TCloudZiyanRegion interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []region.TCloudZiyanRegionTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *region.TCloudZiyanRegionTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
		*types.ListResult[region.TCloudZiyanRegionTable], error)
	BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
}

var _ TCloudZiyanRegion = new(tcloudZiyanRegionDao)

// tcloudZiyanRegionDao region dao.
type tcloudZiyanRegionDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
}

// NewTCloudZiyanRegionDao create a region dao.
func NewTCloudZiyanRegionDao(orm orm.Interface, idGen idgenerator.IDGenInterface) TCloudZiyanRegion {
	return &tcloudZiyanRegionDao{
		orm:   orm,
		idGen: idGen,
	}
}

// BatchCreateWithTx create region with transaction.
func (v *tcloudZiyanRegionDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []region.TCloudZiyanRegionTable) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	for _, model := range models {
		if err := model.InsertValidate(); err != nil {
			return nil, err
		}
	}

	// generate region id
	ids, err := v.idGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for idx := range models {
		models[idx].ID = ids[idx]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		region.TCloudZiyanRegionColumns.ColumnExpr(), region.TCloudZiyanRegionColumns.ColonNameExpr())

	err = v.orm.Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update update region.
func (v *tcloudZiyanRegionDao) Update(kt *kit.Kit, filterExpr *filter.Expression,
	model *region.TCloudZiyanRegionTable) error {
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

	opts := utils.NewFieldOptions().AddBlankedFields("name", "memo").
		AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = v.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := v.orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update tcloud ziyan region failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update tcloud ziyan region, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})

	if err != nil {
		return err
	}

	return nil
}

// List get region list.
func (v *tcloudZiyanRegionDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*types.ListResult[region.TCloudZiyanRegionTable], error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list tcloud ziyan region options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(region.TCloudZiyanRegionColumns.ColumnTypes())),
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
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	tableName := table.TCloudZiyanRegionTable
	if opt.Page.Count {
		// this is a count request, do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableName, whereExpr)

		count, err := v.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count tcloud ziyan region failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListResult[region.TCloudZiyanRegionTable]{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, region.TCloudZiyanRegionColumns.FieldsNamedExpr(opt.Fields),
		tableName, whereExpr, pageExpr)

	details := make([]region.TCloudZiyanRegionTable, 0)
	if err = v.orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListResult[region.TCloudZiyanRegionTable]{Details: details}, nil
}

// BatchDeleteWithTx batch delete region with transaction.
func (v *tcloudZiyanRegionDao) BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.TCloudZiyanRegionTable, whereExpr)
	if _, err = v.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete tcloud ziyan region failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
