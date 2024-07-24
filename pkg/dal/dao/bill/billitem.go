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

// Package bill ...
package bill

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typesbill "hcm/pkg/dal/dao/types/bill"
	"hcm/pkg/dal/table"
	tablebill "hcm/pkg/dal/table/bill"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// AccountBillItem only used for interface.
type AccountBillItem interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, commonOpt *typesbill.ItemCommonOpt,
		items []*tablebill.AccountBillItem) ([]string, error)
	List(kt *kit.Kit, commonOpt *typesbill.ItemCommonOpt, opt *types.ListOption) (
		*typesbill.ListAccountBillItemDetails, error)

	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, commonOpt *typesbill.ItemCommonOpt, billID string,
		updateData *tablebill.AccountBillItem) error

	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, commonOpt *typesbill.ItemCommonOpt, filterExpr *filter.Expression) error
}

// AccountBillItemDao account bill item dao
type AccountBillItemDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx create account bill item with tx.
func (a AccountBillItemDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, commonOpt *typesbill.ItemCommonOpt,
	models []*tablebill.AccountBillItem) ([]string, error) {

	if commonOpt == nil {
		return nil, errf.New(errf.InvalidParameter, "common options is nil")
	}
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	tableName := table.AccountBillItemTable

	ids, err := a.IDGen.Batch(kt, table.Name(tableName), len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		models[index].ID = ids[index]

		if err = model.InsertValidate(); err != nil {
			return nil, err
		}
	}
	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tablebill.AccountBillItemColumns.ColumnExpr(), tablebill.AccountBillItemColumns.ColonNameExpr())

	shardingOpt, err := convertShardingOpt(tableName, commonOpt)
	if err != nil {
		return nil, err
	}
	if err = a.Orm.TableSharding(shardingOpt).Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, shardingOpt: %v, rid: %s", tableName, err, shardingOpt, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return ids, nil
}

// List get account bill item list.
func (a AccountBillItemDao) List(kt *kit.Kit, commonOpt *typesbill.ItemCommonOpt, opt *types.ListOption) (
	*typesbill.ListAccountBillItemDetails, error) {

	if commonOpt == nil {
		return nil, errf.New(errf.InvalidParameter, "common options is nil")
	}
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account bill item options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablebill.AccountBillItemColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	tableName := table.AccountBillItemTable
	shardingOpt, err := convertShardingOpt(tableName, commonOpt)
	if err != nil {
		return nil, err
	}
	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, tableName, whereExpr)
		count, err := a.Orm.TableSharding(shardingOpt).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count account bill item failed, err: %v, shardingOpt: %v, opt: %+v, rid: %s",
				err, shardingOpt, opt, kt.Rid)
			return nil, err
		}

		return &typesbill.ListAccountBillItemDetails{Count: count}, nil
	}

	opt.Page.Order = ""
	opt.Page.Sort = ""
	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	idSql := fmt.Sprintf(`SELECT id FROM %s %s %s`, tableName, whereExpr, pageExpr)
	preDetails := make([]tablebill.AccountBillItem, 0)
	if err = a.Orm.TableSharding(shardingOpt).Do().Select(kt.Ctx, &preDetails, idSql, whereValue); err != nil {
		logs.Errorf("fail to select id for bill item, err: %v, table: %s, opt: %+v, rid: %s",
			err, tableName, opt, kt.Rid)
		return nil, err
	}
	detailIDs := make([]string, 0, len(preDetails))
	for _, detail := range preDetails {
		detailIDs = append(detailIDs, detail.ID)
	}
	if len(detailIDs) == 0 {
		return &typesbill.ListAccountBillItemDetails{Details: make([]tablebill.AccountBillItem, 0)}, nil
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE id IN (:ids)`,
		tablebill.AccountBillItemColumns.FieldsNamedExpr(opt.Fields), tableName)
	details := make([]tablebill.AccountBillItem, 0)

	err = a.Orm.TableSharding(shardingOpt).Do().Select(kt.Ctx, &details, sql, map[string]any{"ids": detailIDs})
	if err != nil {
		logs.Errorf("fail to select bill item by ids, err: %v, shardingOpt: %v, opt: %+v, ids: %v, rid: %s",
			err, shardingOpt.String(), opt, detailIDs, kt.Rid)
		return nil, err
	}
	return &typesbill.ListAccountBillItemDetails{Details: details}, nil
}

// UpdateByIDWithTx update account bill item.
func (a AccountBillItemDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, commonOpt *typesbill.ItemCommonOpt,
	id string, updateData *tablebill.AccountBillItem) error {

	if commonOpt == nil {
		return errf.New(errf.InvalidParameter, "common options is nil")
	}
	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	tableName := table.AccountBillItemTable
	shardingOpt, err := convertShardingOpt(tableName, commonOpt)
	if err != nil {
		return err
	}
	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, tableName, setExpr)

	toUpdate["id"] = id
	_, err = a.Orm.TableSharding(shardingOpt).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update account bill item failed, err: %v, shardingOpt: %v, id: %s, rid: %v",
			err, shardingOpt, id, kt.Rid)
		return err
	}

	return nil
}

// DeleteWithTx delete account bill item with tx.
func (a AccountBillItemDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, commonOpt *typesbill.ItemCommonOpt,
	expr *filter.Expression) error {

	if commonOpt == nil {
		return errf.New(errf.InvalidParameter, "common options is nil")
	}
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	tableName := table.AccountBillItemTable
	shardingOpt, err := convertShardingOpt(tableName, commonOpt)
	if err != nil {
		return err
	}
	sql := fmt.Sprintf(`DELETE FROM %s %s`, tableName, whereExpr)

	if _, err = a.Orm.TableSharding(shardingOpt).Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete account bill item failed, err: %v, shardingOpt: %s, filter: %s, rid: %s",
			err, shardingOpt, expr, kt.Rid)
		return err
	}

	return nil
}

func convertShardingOpt(tableName string, commonOpt *typesbill.ItemCommonOpt) (*orm.TableSuffixShardingOpt, error) {
	if commonOpt == nil {
		return nil, errors.New("common opt is required")
	}
	shardingOpt := orm.NewTableSuffixShardingOpt(tableName,
		[]string{fmt.Sprintf("%s_%d%02d", commonOpt.Vendor, commonOpt.Year, commonOpt.Month)})
	return shardingOpt, nil
}
