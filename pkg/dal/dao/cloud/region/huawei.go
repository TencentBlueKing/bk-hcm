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

// HuaWeiRegion only used for HuaWeiRegion.
type HuaWeiRegion interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []*region.HuaWeiRegionTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, model *region.HuaWeiRegionTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesregion.ListHuaWeiRegionDetails, error)
	DeleteWithTx(kt *kit.Kit, expr *filter.Expression) error
}

var _ HuaWeiRegion = new(HuaWeiRegionDao)

// HuaWeiRegionDao region dao.
type HuaWeiRegionDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// UpdateWithTx rule.
func (h HuaWeiRegionDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	region *region.HuaWeiRegionTable) error {

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

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(region, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, region.TableName(), setExpr, whereExpr)

	effected, err := h.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(
		kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update azure security group rule failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update azure security group rule, but record not found, filter: %v, rid: %v", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List HuaWeiRegion.
func (h HuaWeiRegionDao) List(kt *kit.Kit, opt *types.ListOption) (*typesregion.ListHuaWeiRegionDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list huawei region options is nil")
	}
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(region.HuaWeiRegionColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, argMap, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.HuaWeiRegionTable, whereExpr)
		count, err := h.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("count huawei region failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &typesregion.ListHuaWeiRegionDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, region.HuaWeiRegionColumns.FieldsNamedExpr(opt.Fields),
		table.HuaWeiRegionTable, whereExpr, pageExpr)

	details := make([]*region.HuaWeiRegionTable, 0)
	err = h.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, argMap)
	if err != nil {
		return nil, err
	}

	return &typesregion.ListHuaWeiRegionDetails{Count: 0, Details: details}, nil
}

// CreateWithTx huawei region with tx.
func (h HuaWeiRegionDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []*region.HuaWeiRegionTable) ([]string, error) {
	ids, err := h.IDGen.Batch(kt, table.HuaWeiRegionTable, len(regions))
	if err != nil {
		return nil, err
	}
	for index := range regions {
		regions[index].ID = ids[index]
	}

	for _, item := range regions {
		if err := item.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.HuaWeiRegionTable,
		region.HuaWeiRegionColumns.ColumnExpr(), region.HuaWeiRegionColumns.ColonNameExpr())
	err = h.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, regions)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.HuaWeiRegionTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.HuaWeiRegionTable, err)
	}

	return ids, nil
}

// DeleteWithTx huawei region with tx.
func (h HuaWeiRegionDao) DeleteWithTx(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, argMap, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.HuaWeiRegionTable, whereExpr)

	_, err = h.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err = h.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Delete(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("delete huawei region failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}
