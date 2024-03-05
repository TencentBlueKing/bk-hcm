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

// Package argstpl ...
package argstpl

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
	tableargstpl "hcm/pkg/dal/table/cloud/argument-template"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Interface only used for argument template.
type Interface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tableargstpl.ArgumentTemplateTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tableargstpl.ArgumentTemplateTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, updateData *tableargstpl.ArgumentTemplateTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListArgumentTemplateDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Interface = new(Dao)

// Dao dao.
type Dao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx create argument template.
func (dao Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tableargstpl.ArgumentTemplateTable) (
	[]string, error) {

	tableName := table.ArgumentTemplateTable
	ids, err := dao.IDGen.Batch(kt, tableName, len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		if err = model.InsertValidate(); err != nil {
			return nil, err
		}

		model.ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tableargstpl.ArgumentTplTableColumns.ColumnExpr(), tableargstpl.ArgumentTplTableColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, one := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ID,
			CloudResID: one.CloudID,
			ResName:    one.Name,
			ResType:    enumor.ArgumentTemplateAuditResType,
			Action:     enumor.Create,
			BkBizID:    one.BkBizID,
			Vendor:     one.Vendor,
			AccountID:  one.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: one,
			},
		})
	}
	if err = dao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update update argument template.
func (dao Dao) Update(kt *kit.Kit, expr *filter.Expression, model *tableargstpl.ArgumentTemplateTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
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
		effected, uErr := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if uErr != nil {
			logs.Errorf("update argument template failed, sql: %s, whereValue: %+v, err: %v, rid: %v",
				sql, whereValue, uErr, kt.Rid)
			return nil, uErr
		}
		if effected == 0 {
			logs.Infof("update argument template, but record not found, sql: %s, whereValue: %+v, rid: %v",
				sql, whereValue, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx update argument template by id.
func (dao Dao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	updateData *tableargstpl.ArgumentTemplateTable) error {

	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, table.ArgumentTemplateTable, setExpr)

	toUpdate["id"] = id
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update argument template db failed, id: %s, toUpdate: %+v, err: %v, rid: %v",
			id, toUpdate, err, kt.Rid)
		return err
	}

	return nil
}

// List list argument template.
func (dao Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListArgumentTemplateDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	columnTypes := tableargstpl.ArgumentTplTableColumns.ColumnTypes()
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ArgumentTemplateTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count argument template failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListArgumentTemplateDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableargstpl.ArgumentTplTableColumns.FieldsNamedExpr(opt.Fields),
		table.ArgumentTemplateTable, whereExpr, pageExpr)

	details := make([]tableargstpl.ArgumentTemplateTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListArgumentTemplateDetails{Details: details}, nil
}

// DeleteWithTx delete argument template.
func (dao Dao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ArgumentTemplateTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete argument template failed, sql: %s, whereValue: %+v, err: %v, rid: %s",
			sql, whereValue, err, kt.Rid)
		return err
	}

	return nil
}
