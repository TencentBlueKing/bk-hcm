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

package daoquota

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
	tablequota "hcm/pkg/dal/table/cloud/quota"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// BizQuota defines quota dao operations.
type BizQuota interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tablequota.BizQuotaTable) (string, error)
	Update(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablequota.BizQuotaTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*types.ListBizQuotaDetails, error)
	BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ BizQuota = new(bizQuotaDao)

// bizQuotaDao quota dao.
type bizQuotaDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewBizQuotaDao create a quota dao.
func NewBizQuotaDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) BizQuota {
	return &bizQuotaDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// CreateWithTx create quota with transaction.
func (dao *bizQuotaDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tablequota.BizQuotaTable) (
	string, error) {

	var err error
	if err = model.InsertValidate(); err != nil {
		return "", err
	}

	// generate quota id
	model.ID, err = dao.idGen.One(kt, table.BizQuotaTable)
	if err != nil {
		return "", err
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, model.TableName(), tablequota.BizQuotaColumns.ColumnExpr(),
		tablequota.BizQuotaColumns.ColonNameExpr())

	err = dao.orm.Txn(tx).BulkInsert(kt.Ctx, sql, model)
	if err != nil {
		return "", fmt.Errorf("insert %s failed, err: %v", model.TableName(), err)
	}

	// create audit.
	audits := &tableaudit.AuditTable{
		ResID:     model.ID,
		ResType:   enumor.BizQuotaAuditResType,
		Action:    enumor.Create,
		BkBizID:   model.BkBizID,
		Vendor:    model.Vendor,
		AccountID: model.AccountID,
		Operator:  kt.User,
		Source:    kt.GetRequestSource(),
		Rid:       kt.Rid,
		AppCode:   kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: model,
		},
	}
	if err = dao.audit.Create(kt, audits); err != nil {
		logs.Errorf("create audit failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return model.ID, nil
}

// Update quotas.
func (dao *bizQuotaDao) Update(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablequota.BizQuotaTable) error {
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
	effected, err := dao.orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, toUpdate))
	if err != nil {
		logs.ErrorJson("update biz quota failed, err: %v, id: %s, model: %+v, rid: %s", err, id, model, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update biz quota, but record not found, id: %s, rid: %s", id, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List quotas.
func (dao *bizQuotaDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*types.ListBizQuotaDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list biz quota options is nil")
	}

	columnTypes := tablequota.BizQuotaColumns.ColumnTypes()
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.BizQuotaTable, whereExpr)

		count, err := dao.orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count quotas failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListBizQuotaDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablequota.BizQuotaColumns.FieldsNamedExpr(opt.Fields),
		table.BizQuotaTable, whereExpr, pageExpr)

	details := make([]tablequota.BizQuotaTable, 0)
	if err = dao.orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListBizQuotaDetails{Details: details}, nil
}

// BatchDeleteWithTx batch delete quota with transaction.
func (dao *bizQuotaDao) BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.BizQuotaTable, whereExpr)
	if _, err = dao.orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete biz quota failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
