/*
 *
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package cos cos dao.
package cos

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecos "hcm/pkg/dal/table/cloud/cos"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// CosInterface cos dao interface.
type CosInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tablecos.CosTable) (string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablecos.CosTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablecos.CosTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListCosDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filter *filter.Expression) error
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecos.CosTable) ([]string, error)
}

// CosDao cos dao.
type CosDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create cos with tx.
func (dao *CosDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tablecos.CosTable) (string, error) {
	tableName := table.CosTable
	id, err := dao.IDGen.Batch(kt, tableName, 1)
	if err != nil {
		return "", err
	}
	model.ID = id[0]

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tablecos.CosColumns.ColumnExpr(), tablecos.CosColumns.ColonNameExpr())
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, model)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return "", fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	// cos create audit.
	audits := make([]*tableaudit.AuditTable, 0, 1)

	audits = append(audits, &tableaudit.AuditTable{
		ResID:      model.ID,
		CloudResID: model.CloudID,
		ResName:    model.Name,
		ResType:    enumor.CosAuditResType,
		Action:     enumor.Create,
		BkBizID:    model.BkBizID,
		Vendor:     model.Vendor,
		AccountID:  model.AccountID,
		Operator:   kt.User,
		Source:     kt.GetRequestSource(),
		Rid:        kt.Rid,
		AppCode:    kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: model,
		},
	})

	if err = dao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}
	return id[0], nil
}

// Update update cos.
func (dao *CosDao) Update(kt *kit.Kit, expr *filter.Expression, model *tablecos.CosTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}
	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
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
		effect, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.Errorf("update cos failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update cos, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateByIDWithTx update cos by id with tx.
func (dao *CosDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablecos.CosTable) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo", "tags").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, model.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update cos failed, id: %s, err: %v, rid: %v", id, err, kt.Rid)
		return err
	}

	return nil
}

// List list cos.
func (dao *CosDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListCosDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	columnTypes := tablecos.CosColumns.ColumnTypes()
	columnTypes["tags.*"] = enumor.String

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.LoadBalancerTable, whereExpr)

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count cos failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListCosDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,tablecos.CosColumns.FieldsNamedExpr(opt.Fields),
		table.LoadBalancerTable, whereExpr, pageExpr)

	details := make([]tablecos.CosTable, 0)
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.ListCosDetails{Details: details}, nil
}

// DeleteWithTx delete cos.
func (dao *CosDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.LoadBalancerTable, whereExpr)
	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.Errorf("delete cos failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// BatchCreateWithTx batch create cos with tx.
func (dao *CosDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecos.CosTable) ([]string, error) {
	result := make([]string, 0)
	for _, model := range models {
		id, err := dao.CreateWithTx(kt, tx, model)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}