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

package resourcegroup

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
	resourcegroup "hcm/pkg/dal/table/cloud/resource-group"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AzureRG only used for Azure resource group.
type AzureRG interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []*resourcegroup.AzureRGTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, model *resourcegroup.AzureRGTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesregion.ListAzureRGDetails, error)
	DeleteWithTx(kt *kit.Kit, expr *filter.Expression) error
}

var _ AzureRG = new(AzureRGDao)

// AzureRGDao region dao.
type AzureRGDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// UpdateWithTx rule.
func (a AzureRGDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
	rg *resourcegroup.AzureRGTable) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := rg.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(rg, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, rg.TableName(), setExpr, whereExpr)

	effected, err := a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(
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

// List Azure resource group.
func (a AzureRGDao) List(kt *kit.Kit, opt *types.ListOption) (*typesregion.ListAzureRGDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list azure resource group options is nil")
	}
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(resourcegroup.AzureRGColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, argMap, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AzureRGTable, whereExpr)
		count, err := a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("count azure resource group failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &typesregion.ListAzureRGDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, resourcegroup.AzureRGColumns.FieldsNamedExpr(opt.Fields),
		table.AzureRGTable, whereExpr, pageExpr)

	details := make([]*resourcegroup.AzureRGTable, 0)
	err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, argMap)
	if err != nil {
		return nil, err
	}

	return &typesregion.ListAzureRGDetails{Count: 0, Details: details}, nil
}

// CreateWithTx azure region with tx.
func (a AzureRGDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []*resourcegroup.AzureRGTable) ([]string, error) {

	ids, err := a.IDGen.Batch(kt, table.AzureRGTable, len(regions))
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

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AzureRGTable,
		resourcegroup.AzureRGColumns.ColumnExpr(), resourcegroup.AzureRGColumns.ColonNameExpr())
	err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, regions)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.AzureRGTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.AzureRGTable, err)
	}

	return ids, nil
}

// DeleteWithTx zure region with tx.
func (a AzureRGDao) DeleteWithTx(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, argMap, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AzureRGTable, whereExpr)

	_, err = a.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Delete(kt.Ctx, sql, argMap)
		if err != nil {
			logs.ErrorJson("delete azure resource grouop failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}
