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

package daosubaccountsecret

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
	tablesubas "hcm/pkg/dal/table/cloud/sub-account-secret"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// SubAccountSecret only used for sub account secret.
type SubAccountSecret interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablesubas.Table) ([]string, error)
	BatchUpdate(kt *kit.Kit, models []tablesubas.Table) error
	BatchDelete(kt *kit.Kit, expr *filter.Expression) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSubAccountSecretDetails, error)
}

var _ SubAccountSecret = new(SubAccountSecretDao)

// SubAccountSecretDao sub account secret dao.
type SubAccountSecretDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx sub account secret with tx.
func (dao *SubAccountSecretDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablesubas.Table) (
	[]string, error) {

	ids, err := dao.IDGen.Batch(kt, table.SubAccountSecretTable, len(models))
	if err != nil {
		return nil, err
	}

	for index := range models {
		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}

		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, table.SubAccountSecretTable,
		tablesubas.Columns.ColumnExpr(), tablesubas.Columns.ColonNameExpr())

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s", table.SubAccountSecretTable, err, sql, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.SubAccountSecretTable, err)
	}

	// create audit
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, model := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:    model.ID,
			ResType:  enumor.SubAccountSecretAuditResType,
			Action:   enumor.Create,
			Vendor:   model.Vendor,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: model,
			},
		})
	}
	if err = dao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create sub account secret audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// BatchUpdate sub account secrets.
func (dao *SubAccountSecretDao) BatchUpdate(kt *kit.Kit, models []tablesubas.Table) error {
	if len(models) == 0 {
		return nil
	}

	_, err := dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, model := range models {
			if err := model.UpdateValidate(); err != nil {
				return nil, err
			}

			opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
			setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(&model, opts)
			if err != nil {
				return nil, fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
			}

			sql := fmt.Sprintf(`UPDATE %s %s WHERE id = :id`, model.TableName(), setExpr)
			toUpdate["id"] = model.ID

			_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(kt.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update sub account secret failed, err: %v, id: %s, rid: %v", err, model.ID, kt.Rid)
				return nil, err
			}
		}
		return nil, nil
	})

	return err
}

// BatchDelete sub account secrets.
func (dao *SubAccountSecretDao) BatchDelete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SubAccountSecretTable, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Delete(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("delete sub account secret failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})

	return err
}

// List sub account secrets.
func (dao *SubAccountSecretDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSubAccountSecretDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list sub account secret options is nil")
	}

	columnTypes := tablesubas.Columns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SubAccountSecretTable, whereExpr)
		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count sub account secrets failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &types.ListSubAccountSecretDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablesubas.Columns.FieldsNamedExpr(opt.Fields),
		table.SubAccountSecretTable, whereExpr, pageExpr)

	details := make([]tablesubas.Table, 0)
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.ErrorJson("select sub account secret failed, err: %v, sql: %s, filter: %v, rid: %s", err, sql, opt.Filter,
			kt.Rid)
		return nil, err
	}

	return &types.ListSubAccountSecretDetails{Count: 0, Details: details}, nil
}
