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

package daoaccount

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tablesubaccount "hcm/pkg/dal/table/cloud/sub-account"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// SubAccount only used for sub account.
type SubAccount interface {
	Get(kt *kit.Kit, id string) (*tablesubaccount.Table, error)
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablesubaccount.Table) ([]string, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablesubaccount.Table) error
	Update(kt *kit.Kit, expr *filter.Expression, model *tablesubaccount.Table) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSubAccountDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ SubAccount = new(SubAccountDao)

// SubAccountDao account dao.
type SubAccountDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// UpdateByIDWithTx sub account.
func (dao *SubAccountDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablesubaccount.Table) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, model.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update sub account failed, err: %v, id: %s, sql: %s, rid: %v", err, id, sql, kt.Rid)
		return err
	}

	return nil
}

// Get sub account.
func (dao *SubAccountDao) Get(kt *kit.Kit, id string) (*tablesubaccount.Table, error) {
	opt := &types.ListOption{
		Filter: tools.EqualExpression("id", id),
		Page: &core.BasePage{
			Start: 0,
			Limit: 1,
		},
	}
	result, err := dao.List(kt, opt)
	if err != nil {
		logs.Errorf("list sub account failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	if len(result.Details) == 0 {
		return nil, errf.Newf(errf.RecordNotFound, "sub account: %s not found", id)
	}

	return &result.Details[0], nil
}

// BatchCreateWithTx sub account with tx.
func (dao *SubAccountDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []tablesubaccount.Table) ([]string, error) {

	ids, err := dao.IDGen.Batch(kt, table.SubAccountTable, len(models))
	if err != nil {
		return nil, err
	}
	for index := range models {
		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}

		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.SubAccountTable, tablesubaccount.Columns.ColumnExpr(),
		tablesubaccount.Columns.ColonNameExpr())

	err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s", table.SubAccountTable, err, sql, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.SubAccountTable, err)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, model := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:     model.ID,
			ResName:   model.Name,
			ResType:   enumor.SubAccountAuditResType,
			Action:    enumor.Create,
			Vendor:    model.Vendor,
			AccountID: model.AccountID,
			Operator:  kt.User,
			Source:    kt.GetRequestSource(),
			Rid:       kt.Rid,
			AppCode:   kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: model,
			},
		})
	}
	if err = dao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create sub account audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update accounts.
func (dao *SubAccountDao) Update(kt *kit.Kit, filterExpr *filter.Expression,
	model *tablesubaccount.Table) error {

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

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update sub account failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update sub account, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List accounts.
func (dao *SubAccountDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSubAccountDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list sub account options is nil")
	}

	columnTypes := tablesubaccount.Columns.ColumnTypes()
	columnTypes["extension.uid"] = enumor.Numeric
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is dao count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SubAccountTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count sub accounts failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSubAccountDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablesubaccount.Columns.FieldsNamedExpr(opt.Fields),
		table.SubAccountTable, whereExpr, pageExpr)

	details := make([]tablesubaccount.Table, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select sub account failed, err: %v, sql: %s, filter: %v, rid: %s", err, sql, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListSubAccountDetails{Count: 0, Details: details}, nil
}

// DeleteWithTx account with tx.
func (dao *SubAccountDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SubAccountTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete sub account failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
