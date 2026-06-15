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

// Package permissiontemplate defines DAO for permission_template.
package permissiontemplate

import (
	"encoding/json"
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabletypes "hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"
	"hcm/pkg/tools/slice"

	"github.com/jmoiron/sqlx"
)

// PermissionTemplate only used for permission_template.
type PermissionTemplate interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tablecloud.PermissionTemplateTable) ([]string, error)
	BatchUpdate(kt *kit.Kit, models []tablecloud.PermissionTemplateTable) error
	BatchDelete(kt *kit.Kit, expr *filter.Expression) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListPermissionTemplateDetails, error)
	ListJoinSubAccount(kt *kit.Kit, opt *types.ListPermTmplJoinOption) (
		*types.ListPermissionTmplJoinDetails, error)
}

var _ PermissionTemplate = new(PermissionTemplateDao)

// PermissionTemplateDao permission template dao.
type PermissionTemplateDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx batch create permission_template with transaction.
func (dao *PermissionTemplateDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []tablecloud.PermissionTemplateTable) ([]string, error) {

	ids, err := dao.IDGen.Batch(kt, table.PermissionTemplateTable, len(models))
	if err != nil {
		return nil, err
	}

	accountIDs := make([]string, 0)
	for index := range models {
		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
		models[index].ID = ids[index]
		accountIDs = append(accountIDs, models[index].AccountID)
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`,
		table.PermissionTemplateTable,
		tablecloud.PermissionTemplateColumns.ColumnExpr(),
		tablecloud.PermissionTemplateColumns.ColonNameExpr())

	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, sql: %s, rid: %s",
			table.PermissionTemplateTable, err, sql, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.PermissionTemplateTable, err)
	}

	accountBizIDMap, err := dao.findAccountBizID(kt, accountIDs)
	if err != nil {
		logs.Errorf("find account biz id failed, err: %v, accountIDs: %v, rid: %s", err, accountIDs, kt.Rid)
		return nil, err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, one := range models {
		bizID, ok := accountBizIDMap[one.AccountID]
		if !ok {
			logs.Errorf("account biz id not found, accountID: %s, rid: %s", one.AccountID, kt.Rid)
			return nil, fmt.Errorf("account biz id not found, accountID: %s", one.AccountID)
		}
		audits = append(audits, &tableaudit.AuditTable{
			ResID:     one.ID,
			ResName:   one.Name,
			ResType:   enumor.PermissionTemplateAuditResType,
			Action:    enumor.Create,
			BkBizID:   bizID,
			Vendor:    one.Vendor,
			AccountID: one.AccountID,
			Operator:  kt.User,
			Source:    kt.GetRequestSource(),
			Rid:       kt.Rid,
			AppCode:   kt.AppCode,
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

// findAccountBizID queries bk_biz_id for given account IDs in batches of 500.
func (dao *PermissionTemplateDao) findAccountBizID(kt *kit.Kit, accountIDs []string) (map[string]int64, error) {
	result := make(map[string]int64, len(accountIDs))

	for _, batch := range slice.Split(accountIDs, int(filter.DefaultMaxInLimit)) {
		rows := make([]tablecloud.AccountTable, 0, len(batch))
		sql := fmt.Sprintf(`SELECT id, bk_biz_id FROM %s WHERE id IN (:id)`, table.AccountTable)
		whereValue := map[string]interface{}{"id": batch}

		err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &rows, sql, whereValue)
		if err != nil {
			logs.Errorf("select account biz id failed, err: %v, accountIDs: %v, rid: %s", err, batch, kt.Rid)
			return nil, err
		}

		for _, row := range rows {
			result[row.ID] = row.BkBizID
		}
	}

	return result, nil
}

// BatchUpdate batch update permission_template by ID.
func (dao *PermissionTemplateDao) BatchUpdate(kt *kit.Kit, models []tablecloud.PermissionTemplateTable) error {
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

			_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).
				Update(kt.Ctx, sql, toUpdate)
			if err != nil {
				logs.Errorf("update permission_template failed, err: %v, id: %s, rid: %s",
					err, model.ID, kt.Rid)
				return nil, err
			}
		}
		return nil, nil
	})

	return err
}

// BatchDelete batch delete permission_template.
func (dao *PermissionTemplateDao) BatchDelete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.PermissionTemplateTable, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		_, err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).
			Delete(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("delete permission_template failed, err: %v, filter: %s, rid: %s",
				err, expr, kt.Rid)
			return nil, err
		}
		return nil, nil
	})

	return err
}

// List list permission_template.
func (dao *PermissionTemplateDao) List(kt *kit.Kit, opt *types.ListOption) (
	*types.ListPermissionTemplateDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	columnTypes := tablecloud.PermissionTemplateColumns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.PermissionTemplateTable, whereExpr)
		count, err := dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().
			Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count permission_template failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &types.ListPermissionTemplateDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		tablecloud.PermissionTemplateColumns.FieldsNamedExpr(opt.Fields),
		table.PermissionTemplateTable, whereExpr, pageExpr)

	details := make([]tablecloud.PermissionTemplateTable, 0)
	err = dao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().
		Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.ErrorJson("select permission_template failed, err: %v, sql: %s, filter: %v, rid: %s",
			err, sql, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListPermissionTemplateDetails{Count: 0, Details: details}, nil
}

// ListJoinSubAccount lists permission_template rows with associated sub_account count.
// pt=permission_template、sa=sub_account
func (dao *PermissionTemplateDao) ListJoinSubAccount(kt *kit.Kit,
	opt *types.ListPermTmplJoinOption) (*types.ListPermissionTmplJoinDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list permission template join options is nil")
	}
	if opt.Page == nil {
		return nil, errf.New(errf.InvalidParameter, "page is required")
	}
	if err := opt.Page.Validate(core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereSQL, whereArgs, err := buildPermTmplJoinWhere(opt)
	if err != nil {
		return nil, err
	}

	ormOpts := orm.NewInjectTenantIDOpt(kt.TenantID)

	if opt.Page.Count {
		sqlStr := fmt.Sprintf(`SELECT COUNT(*) FROM %s AS pt %s`, table.PermissionTemplateTable, whereSQL)
		count, err := dao.Orm.ModifySQLOpts(ormOpts).Do().Count(kt.Ctx, sqlStr, whereArgs)
		if err != nil {
			logs.Errorf("count permission_template join failed, err: %v, rid: %s", err, kt.Rid)
			return nil, err
		}
		return &types.ListPermissionTmplJoinDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, &types.PageSQLOption{
		Sort: types.SortOption{Sort: opt.Page.Sort, ForceOverlap: true},
	})
	if err != nil {
		return nil, err
	}

	innerSQL := fmt.Sprintf(
		`SELECT pt.*,
			(SELECT COUNT(*) FROM %s AS sa
				WHERE JSON_CONTAINS(sa.permission_template_ids, JSON_QUOTE(pt.id))) AS associated_sub_account_count
		FROM %s AS pt
		%s`,
		table.SubAccountTable,
		table.PermissionTemplateTable,
		whereSQL,
	)
	selectSQL := fmt.Sprintf(
		`SELECT * FROM (%s) AS tmp %s`,
		innerSQL, pageExpr,
	)

	details := make([]types.PermissionTmplJoinRow, 0)
	err = dao.Orm.ModifySQLOpts(ormOpts).Do().Select(kt.Ctx, &details, selectSQL, whereArgs)
	if err != nil {
		logs.Errorf("select permission_template join failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return &types.ListPermissionTmplJoinDetails{Count: 0, Details: details}, nil
}

// buildPermTmplJoinWhere builds the WHERE clause and named args for permission_template join list.
func buildPermTmplJoinWhere(opt *types.ListPermTmplJoinOption) (
	string, map[string]interface{}, error) {

	if opt == nil {
		return "", nil, fmt.Errorf("list permission template join option is nil")
	}

	whereExprs := make([]string, 0)
	args := make(map[string]interface{})

	whereExprs = append(whereExprs, "pt.vendor = :vendor")
	args["vendor"] = string(opt.Vendor)

	if len(opt.AccountIDs) > 0 {
		whereExprs = append(whereExprs, "pt.account_id IN (:account_ids)")
		args["account_ids"] = opt.AccountIDs
	}

	if len(opt.IDs) > 0 {
		whereExprs = append(whereExprs, "pt.id IN (:ids)")
		args["ids"] = opt.IDs
	}

	if len(opt.CloudIDs) > 0 {
		whereExprs = append(whereExprs, "pt.cloud_id IN (:cloud_ids)")
		args["cloud_ids"] = opt.CloudIDs
	}

	if len(opt.Names) > 0 {
		nameConds := make([]string, 0, len(opt.Names))
		for i, n := range opt.Names {
			key := fmt.Sprintf("name_%d", i)
			nameConds = append(nameConds, fmt.Sprintf("pt.name LIKE CONCAT('%%', :%s, '%%')", key))
			args[key] = n
		}
		whereExprs = append(whereExprs, "("+strings.Join(nameConds, " OR ")+")")
	}

	if opt.Creator != "" {
		whereExprs = append(whereExprs, "pt.creator = :creator")
		args["creator"] = opt.Creator
	}

	if opt.Reviser != "" {
		whereExprs = append(whereExprs, "pt.reviser = :reviser")
		args["reviser"] = opt.Reviser
	}

	if !opt.Extension.IsEmpty() {
		var err error
		whereExprs, err = buildPermTmplExtWhere(whereExprs, args, opt.Vendor, opt.Extension)
		if err != nil {
			return "", nil, err
		}
	}

	if opt.PolicyLibraryIDIsNull != nil {
		if *opt.PolicyLibraryIDIsNull {
			whereExprs = append(whereExprs, "pt.policy_library_id IS NULL")
		} else {
			whereExprs = append(whereExprs, "pt.policy_library_id IS NOT NULL")
		}
	}
	if len(whereExprs) == 0 {
		return "", args, nil
	}

	return "WHERE " + strings.Join(whereExprs, " AND "), args, nil
}

// buildPermTmplExtWhere appends vendor-specific extension filter conditions.
func buildPermTmplExtWhere(whereExprs []string, args map[string]interface{}, vendor enumor.Vendor,
	extension tabletypes.JsonField) ([]string, error) {

	switch vendor {
	case enumor.TCloud:
		return buildPermTmplExtWhereForTCloud(whereExprs, args, extension)
	default:
		return nil, fmt.Errorf("unsupported vendor: %s", vendor)
	}
}

// buildPermTmplExtWhereForTCloud appends TCloud extension filter conditions.
func buildPermTmplExtWhereForTCloud(whereExprs []string, args map[string]interface{},
	extension tabletypes.JsonField) ([]string, error) {

	tc := new(types.TCloudPermTmplJoinExt)
	if err := json.Unmarshal([]byte(extension), tc); err != nil {
		return nil, fmt.Errorf("invalid tcloud extension json: %w", err)
	}
	if err := validator.Validate.Struct(tc); err != nil {
		return nil, err
	}

	if len(tc.CloudSubAccountIDs) > 0 {
		whereExprs = append(whereExprs,
			fmt.Sprintf(`EXISTS (SELECT 1 FROM %s AS sa WHERE sa.account_id = pt.account_id`+
				` AND sa.vendor = pt.vendor AND sa.cloud_id IN (:cloud_sub_account_ids)`+
				` AND JSON_CONTAINS(sa.permission_template_ids, JSON_QUOTE(pt.id)))`,
				table.SubAccountTable))
		args["cloud_sub_account_ids"] = tc.CloudSubAccountIDs
	}

	if tc.CloudType != nil {
		whereExprs = append(whereExprs, "JSON_EXTRACT(pt.extension, '$.cloud_type') = :cloud_type")
		args["cloud_type"] = int(converter.PtrToVal(tc.CloudType))
	}

	return whereExprs, nil
}
