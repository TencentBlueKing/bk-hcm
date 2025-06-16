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
	typesregion "hcm/pkg/dal/dao/types/region"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud/region"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AzureRegion only used for Azure region.
type AzureRegion interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []*region.AzureRegionTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, model *region.AzureRegionTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesregion.ListAzureRegionDetails, error)
	DeleteWithTx(kt *kit.Kit, expr *filter.Expression) error
}

var _ AzureRegion = new(AzureRegionDao)

// AzureRegionDao region dao.
type AzureRegionDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// UpdateWithTx rule.
func (a AzureRegionDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	region *region.AzureRegionTable) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := region.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(region, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, region.TableName(), setExpr, whereExpr)

	effected, err := a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(
		kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update azure region failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update azure region, but record not found, filter: %v, rid: %v", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List Azure resource group.
func (a AzureRegionDao) List(kt *kit.Kit, opt *types.ListOption) (*typesregion.ListAzureRegionDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list azure region options is nil")
	}
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(region.AzureRegionColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, argMap, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AzureRegionTable, whereExpr)
		count, err := a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("count azure resource group failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &typesregion.ListAzureRegionDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, region.AzureRegionColumns.FieldsNamedExpr(opt.Fields),
		table.AzureRegionTable, whereExpr, pageExpr)

	details := make([]*region.AzureRegionTable, 0)
	err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, argMap)
	if err != nil {
		return nil, err
	}

	return &typesregion.ListAzureRegionDetails{Count: 0, Details: details}, nil
}

// CreateWithTx azure region with tx.
func (a AzureRegionDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*region.AzureRegionTable) ([]string, error) {

	ids, err := a.IDGen.Batch(kt, table.AzureRegionTable, len(models))
	if err != nil {
		return nil, err
	}
	for index := range models {
		models[index].ID = ids[index]
	}

	for _, item := range models {
		if err := item.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		region.AzureRegionColumns.ColumnExpr(), region.AzureRegionColumns.ColonNameExpr())
	err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update azure region.
func (a AzureRegionDao) Update(_ *kit.Kit, _ *filter.Expression, _ *region.AzureRegionTable) error {
	return nil
}

// DeleteWithTx zure region with tx.
func (a AzureRegionDao) DeleteWithTx(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, argMap, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AzureRegionTable, whereExpr)

	_, err = a.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Delete(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("delete azure region failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}
