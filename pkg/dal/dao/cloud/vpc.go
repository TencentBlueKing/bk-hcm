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

// Vpc defines vpc dao operations.
type Vpc interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []cloud.VpcTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *cloud.VpcTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*types.VpcListResult, error)
	BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Vpc = new(vpcDao)

// vpcDao vpc dao.
type vpcDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewVpcDao create a vpc dao.
func NewVpcDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) Vpc {
	return &vpcDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// BatchCreateWithTx create vpc with transaction.
func (v *vpcDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []cloud.VpcTable) ([]string, error) {
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	for _, model := range models {
		if err := model.InsertValidate(); err != nil {
			return nil, err
		}
	}

	// generate vpc id
	ids, err := v.idGen.Batch(kt, table.VpcTable, len(models))
	if err != nil {
		return nil, err
	}

	for idx := range models {
		models[idx].ID = ids[idx]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(), cloud.VpcColumns.ColumnExpr(),
		cloud.VpcColumns.ColonNameExpr())

	err = v.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, one := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ID,
			CloudResID: one.CloudID,
			ResName:    *one.Name,
			ResType:    enumor.VpcCloudAuditResType,
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
	if err = v.audit.BatchCreate(kt, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update vpcs.
func (v *vpcDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *cloud.VpcTable) error {
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

	opts := utils.NewFieldOptions().AddBlankedFields("name", "memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = v.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := v.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update vpc failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update vpc, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List vpcs.
func (v *vpcDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*types.VpcListResult,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list vpc options is nil")
	}

	columnTypes := cloud.VpcColumns.ColumnTypes()
	columnTypes["extension.self_link"] = enumor.String
	columnTypes["extension.resource_group_name"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereOpt := tools.DefaultSqlWhereOption
	if len(whereOpts) != 0 && whereOpts[0] != nil {
		err := whereOpts[0].Validate()
		if err != nil {
			return nil, err
		}
		whereOpt = whereOpts[0]
	}
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.VpcTable, whereExpr)

		count, err := v.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count vpcs failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.VpcListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.VpcColumns.FieldsNamedExpr(opt.Fields), table.VpcTable,
		whereExpr, pageExpr)

	details := make([]cloud.VpcTable, 0)
	err = v.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.VpcListResult{Details: details}, nil
}

// BatchDeleteWithTx batch delete vpc with transaction.
func (v *vpcDao) BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.VpcTable, whereExpr)
	_, err = v.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete vpc failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// ListVpc TODO: 考虑之后这种跨表查询是否可以直接引用对象的 List 函数，而不是再写一个。
func ListVpc(kt *kit.Kit, ormi orm.Interface, ids []string) (map[string]cloud.VpcTable, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s where id in (:ids)`, cloud.VpcColumns.FieldsNamedExpr(nil), table.VpcTable)

	vpcs := make([]cloud.VpcTable, 0)
	if err := ormi.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(
		kt.Ctx, &vpcs, sql, map[string]interface{}{"ids": ids}); err != nil {
		return nil, err
	}

	idVpcMap := make(map[string]cloud.VpcTable, len(ids))
	for _, one := range vpcs {
		idVpcMap[one.ID] = one
	}

	return idVpcMap, nil
}
