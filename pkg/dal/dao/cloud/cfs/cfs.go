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

// Package cfs 如下:
// orm 接口
package cfs

import (
	"fmt"
	"github.com/pkg/errors"
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
	tablecfs "hcm/pkg/dal/table/cloud/cfs"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Interface only used for cfs.
type Interface interface {
	// BatchCreateWithTx cfs
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecfs.Table) ([]string, error)

	// CreateWithTx cfs
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models *tablecfs.Table) (string, error)

	// DeleteWithTx cfs
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error

	// Update cfs
	Update(kt *kit.Kit, expr *filter.Expression, model *tablecfs.Table) error
	// UpdateByIDWithTx cfs
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablecfs.Table) error

	// List cfs
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListCfsDetails, error)
	// ListWithTx cfs
	ListWithTx(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*types.ListCfsDetails, error)
}

var _ Interface = new(Dao)

// Dao cfs dao.
type Dao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx cfs
func (dao Dao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models *tablecfs.Table) (string, error) {
	id, err := dao.IDGen.One(kt, table.CfsTable)
	if err != nil {
		logs.Errorf("create cfs id failed, err: %v, rid: %s", err.Error(), kt.Rid)
		return "", errors.Wrap(err, "create cfs id failed")
	}
	models.ID = id
	if err = models.InsertValidate(); err != nil {
		logs.Errorf("insert cfs %s failed, id: %s, err: %v, rid: %s", models.Name, models.ID, err.Error(), kt.Rid)
		return "", errors.Wrapf(err, "insert cfs %s failed, id: %s", models.Name, models.ID)
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.CfsTable, tablecfs.TableColumns.ColumnExpr(),
		tablecfs.TableColumns.ColonNameExpr())

	//logs.Infof("sql: %s, rid: %s", sql, kt.Rid) // note: debug log

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Insert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.CfsTable, err, kt.Rid)
		return "", errors.Wrapf(err, "insert %s failed, rid: %s", table.CfsTable, kt.Rid)
	}

	// create audit.
	auditModel := &tableaudit.AuditTable{
		ResID:      models.ID,
		CloudResID: models.CloudID,
		ResName:    models.Name,
		ResType:    enumor.CfsAuditResType,
		Action:     enumor.Create,
		BkBizID:    models.BkBizID,
		Vendor:     models.Vendor,
		AccountID:  models.AccountID,
		Operator:   kt.User,
		Source:     kt.GetRequestSource(),
		Rid:        kt.Rid,
		AppCode:    kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: models,
		},
	}
	if err = dao.Audit.Create(kt, auditModel); err != nil {
		logs.Errorf("create audit failed, err: %v, rid: %s", err, kt.Rid)
		return "", errors.Wrapf(err, "create audit failed, rid: %s", kt.Rid)
	}

	return id, nil
}

// BatchCreateWithTx cfs
func (dao Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablecfs.Table) ([]string, error) {
	ids, err := dao.IDGen.Batch(kt, table.CfsTable, len(models))
	if err != nil {
		return nil, errors.Wrap(err, "batch create cfs id failed")
	}

	for index, model := range models {
		model.ID = ids[index]
		if err := model.InsertValidate(); err != nil {
			return nil, errors.Wrapf(err, "insert cfs %s failed, id: %s", model.Name, model.ID)
		}
	}
	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.CfsTable, tablecfs.TableColumns.ColumnExpr(),
		tablecfs.TableColumns.ColonNameExpr())

	//logs.Infof("sql: %s, rid: %s", sql, kt.Rid) // note: debug log

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.CfsTable, err, kt.Rid)
		return nil, errors.Wrapf(err, "insert %s failed, rid: %s", table.CfsTable, kt.Rid)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, one := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ID,
			CloudResID: one.CloudID,
			ResName:    one.Name,
			ResType:    enumor.CfsAuditResType,
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
		return nil, errors.Wrapf(err, "batch create audit failed, rid: %s", kt.Rid)
	}

	return ids, nil
}

// DeleteWithTx cfs.
func (dao Dao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return errors.Wrap(err, "prepare parsed sql filter expr failed")
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.CfsTable, whereExpr)

	//logs.Infof("sql: %s, whereValue: %s, rid: %s", sql, whereValue, kt.Rid) // note: debug log

	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete cfs failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return errors.Wrapf(err, "delete cfs failed, filter: %s, rid: %s", expr, kt.Rid)
	}

	return nil
}

// Update cfs.
func (dao Dao) Update(kt *kit.Kit, expr *filter.Expression, model *tablecfs.Table) error {
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
		effected, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update cfs failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.Infof("update cfs, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx cfs.
func (dao Dao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablecfs.Table) error {
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

	_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update cfs failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// List cfs.
func (dao Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListCfsDetails, error) {
	if opt == nil {
		logs.Errorf("list cfs failed, opt is nil, rid: %s", kt.Rid)
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}
	columnTypes := tablecfs.TableColumns.ColumnTypes()
	//columnTypes["extension.resource_group_name"] = enumor.String // note: 待定
	//columnTypes["extension.zones"] = enumor.Json

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		logs.Errorf("validate list option failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
		return nil, errors.Wrapf(err, "validate list option failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, errors.Wrap(err, "prepare parsed sql filter expr failed")
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.CfsTable, whereExpr)

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count cfs failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, errors.Wrapf(err, "count cfs failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
		}

		return &types.ListCfsDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		logs.Errorf("prepare parsed sql page expr failed, err: %v, rid: %s", err, kt.Rid)
		return nil, errors.Wrap(err, "prepare parsed sql page expr failed")
	}

	details := make([]tablecfs.Table, 0)
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablecfs.TableColumns.FieldsNamedExpr(opt.Fields), table.CfsTable,
		whereExpr, pageExpr)

	//logs.Infof("sql: %s, whereValue: %s, rid: %s", sql, whereValue, kt.Rid) // note: debug log

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.ErrorJson("select cfs failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
		return nil, errors.Wrapf(err, "select cfs failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
	}

	//logs.Infof("select cfs details: %v, count: %d, rid: %s", details, len(details), kt.Rid) // note: debug log

	return &types.ListCfsDetails{Details: details, Count: uint64(len(details))}, nil
}

// ListWithTx cfs with tx.
func (dao Dao) ListWithTx(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*types.ListCfsDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}
	columnTypes := tablecfs.TableColumns.ColumnTypes()
	//columnTypes["extension.resource_group_name"] = enumor.String // note: 待定
	//columnTypes["extension.zones"] = enumor.Json

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, errors.Wrapf(err, "validate list option failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, errors.Wrap(err, "prepare parsed sql filter expr failed")
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.CfsTable, whereExpr)

		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count cfs failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, errors.Wrapf(err, "count cfs failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
		}

		return &types.ListCfsDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, errors.Wrap(err, "prepare parsed sql page expr failed")
	}

	details := make([]tablecfs.Table, 0)
	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablecfs.TableColumns.FieldsNamedExpr(opt.Fields), table.CfsTable,
		whereExpr, pageExpr)

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, errors.Wrapf(err, "select cfs failed, filter: %s, rid: %s", opt.Filter, kt.Rid)
	}

	return &types.ListCfsDetails{Details: details, Count: uint64(len(details))}, nil
}

//// ListCfs TODO: 考虑之后这种跨表查询是否可以直接引用对象的 List 函数，而不是再写一个。
//func ListCfs(kt *kit.Kit, ormi orm.Interface, ids []string) (map[string]tablecfs.Table, error) {
//	sql := fmt.Sprintf(`SELECT %s FROM %s where id in (:ids)`, tablecfs.TableColumns.FieldsNamedExpr(nil),
//		table.CfsTable)
//
//	cfss := make([]tablecfs.Table, 0)
//	if err := ormi.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &cfss, sql,
//		map[string]interface{}{"ids": ids}); err != nil {
//		return nil, err
//	}
//
//	idCfsMap := make(map[string]tablecfs.Table, len(ids))
//	for _, one := range cfss {
//		idCfsMap[one.ID] = one
//	}
//
//	return idCfsMap, nil
//}
