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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Account only used for account.
type Account interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, account *cloud.AccountTable) (string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *cloud.AccountTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListAccountDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Account = new(AccountDao)

// AccountDao account dao.
type AccountDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx account with tx.
func (a AccountDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *cloud.AccountTable) (string, error) {
	if err := model.InsertValidate(); err != nil {
		return "", err
	}

	// generate account id
	id, err := a.IDGen.One(kt, table.AccountTable)
	if err != nil {
		return "", err
	}
	model.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, model.TableName(), cloud.AccountColumns.ColumnExpr(),
		cloud.AccountColumns.ColonNameExpr())

	model.TenantID = kt.TenantID
	err = a.Orm.Txn(tx).Insert(kt.Ctx, sql, model)
	if err != nil {
		return "", fmt.Errorf("insert %s failed, err: %v", model.TableName(), err)
	}

	// create audit.
	auditInfo := &tableaudit.AuditTable{
		ResID:     model.ID,
		ResName:   model.Name,
		ResType:   enumor.AccountAuditResType,
		Action:    enumor.Create,
		Vendor:    enumor.Vendor(model.Vendor),
		AccountID: model.ID,
		Operator:  kt.User,
		Source:    kt.GetRequestSource(),
		Rid:       kt.Rid,
		AppCode:   kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: model,
		},
	}
	if err = a.Audit.Create(kt, auditInfo); err != nil {
		logs.Errorf("create account audit failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return id, nil
}

// Update accounts.
func (a AccountDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *cloud.AccountTable) error {
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

	_, err = a.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := a.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update account failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
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
func (a AccountDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListAccountDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account options is nil")
	}

	columnTypes := cloud.AccountColumns.ColumnTypes()
	columnTypes["extension.cloud_main_account_id"] = enumor.String
	columnTypes["extension.cloud_account_id"] = enumor.String
	columnTypes["extension.cloud_main_account_name"] = enumor.String
	columnTypes["extension.cloud_project_id"] = enumor.String
	columnTypes["extension.cloud_tenant_id"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.DefaultPageOption); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AccountTable, whereExpr)

		count, err := a.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count accounts failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListAccountDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.AccountColumns.FieldsNamedExpr(opt.Fields),
		table.AccountTable, whereExpr, pageExpr)

	details := make([]*cloud.AccountTable, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListAccountDetails{Count: 0, Details: details}, nil
}

// DeleteWithTx account with tx.
func (a AccountDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AccountTable, whereExpr)
	if _, err = a.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete account failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
