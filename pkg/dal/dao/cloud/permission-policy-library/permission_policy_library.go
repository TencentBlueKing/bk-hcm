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

// Package permissionpolicylibrary defines DAO for permission_policy_library.
package permissionpolicylibrary

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
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// PermissionPolicyLibrary only used for permission_policy_library.
type PermissionPolicyLibrary interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablecloud.PermissionPolicyLibraryTable) ([]string, error)
	BatchUpdate(kt *kit.Kit, models []tablecloud.PermissionPolicyLibraryTable) error
	BatchDelete(kt *kit.Kit, expr *filter.Expression) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListPermissionPolicyLibraryDetails, error)
}

var _ PermissionPolicyLibrary = new(PermissionPolicyLibraryDao)

// PermissionPolicyLibraryDao permission policy library dao.
type PermissionPolicyLibraryDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx batch create permission_policy_library with transaction.
func (dao *PermissionPolicyLibraryDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []tablecloud.PermissionPolicyLibraryTable) ([]string, error) {

	ids, err := dao.IDGen.Batch(kt, table.PermissionPolicyLibraryTable, len(models))
	if err != nil {
		return nil, err
	}

	for index := range models {
		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`,
		table.PermissionPolicyLibraryTable,
		tablecloud.PermissionPolicyLibraryColumns.ColumnExpr(),
		tablecloud.PermissionPolicyLibraryColumns.ColonNameExpr())

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s",
			table.PermissionPolicyLibraryTable, err, sql, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.PermissionPolicyLibraryTable, err)
	}

	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, one := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:    one.ID,
			ResName:  one.Name,
			ResType:  enumor.PermissionPolicyLibraryAuditResType,
			Action:   enumor.Create,
			Vendor:   one.Vendor,
			Operator: kt.User,
			Source:   kt.GetRequestSource(),
			Rid:      kt.Rid,
			AppCode:  kt.AppCode,
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

// BatchUpdate batch update permission_policy_library by ID.
func (dao *PermissionPolicyLibraryDao) BatchUpdate(kt *kit.Kit,
	models []tablecloud.PermissionPolicyLibraryTable) error {

	if len(models) == 0 {
		return nil
	}

	_, err := dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		for _, model := range models {
			if err := model.UpdateValidate(); err != nil {
				return nil, err
			}

			opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
			setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(&model, opts)
			if err != nil {
				return nil, fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
			}

			sql := fmt.Sprintf(`UPDATE %s %s WHERE id = :id`, model.TableName(), setExpr)
			toUpdate["id"] = model.ID

			_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(kt.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update permission_policy_library failed, err: %v, id: %s, rid: %s",
					err, model.ID, kt.Rid)
				return nil, err
			}
		}
		return nil, nil
	})

	return err
}

// BatchDelete batch delete permission_policy_library.
func (dao *PermissionPolicyLibraryDao) BatchDelete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.PermissionPolicyLibraryTable, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Delete(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("delete permission_policy_library failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})

	return err
}

// List list permission_policy_library.
func (dao *PermissionPolicyLibraryDao) List(kt *kit.Kit, opt *types.ListOption) (
	*types.ListPermissionPolicyLibraryDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	columnTypes := tablecloud.PermissionPolicyLibraryColumns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.PermissionPolicyLibraryTable, whereExpr)
		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count permission_policy_library failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &types.ListPermissionPolicyLibraryDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablecloud.PermissionPolicyLibraryColumns.FieldsNamedExpr(opt.Fields),
		table.PermissionPolicyLibraryTable, whereExpr, pageExpr)

	details := make([]tablecloud.PermissionPolicyLibraryTable, 0)
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.ErrorJson("select permission_policy_library failed, err: %v, sql: %s, filter: %v, rid: %s",
			err, sql, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListPermissionPolicyLibraryDetails{Count: 0, Details: details}, nil
}
