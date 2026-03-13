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

// Package resplan ...
package resplan

import (
	"fmt"

	"hcm/pkg/api/core"
	rpproto "hcm/pkg/api/data-service/resource-plan"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	rpdgs "hcm/pkg/dal/table/resource-plan/res-plan-demand-gpu-suborder"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ResPlanDemandGpuSubOrderInterface only used for resource plan demand gpu sub order interface.
type ResPlanDemandGpuSubOrderInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []rpdgs.ResPlanDemandGpuSubOrderTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression,
		model *rpdgs.ResPlanDemandGpuSubOrderTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*rpproto.ResPlanDemandGpuSubOrderListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ ResPlanDemandGpuSubOrderInterface = new(ResPlanDemandGpuSubOrderDao)

// ResPlanDemandGpuSubOrderDao resource plan demand gpu sub order dao.
type ResPlanDemandGpuSubOrderDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create resource plan demand gpu sub order with tx.
func (d ResPlanDemandGpuSubOrderDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
	models []rpdgs.ResPlanDemandGpuSubOrderTable) ([]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := d.IDGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for index := range models {
		models[index].ID = ids[index]

		if err = models[index].InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s) VALUES(%s)`, models[0].TableName(),
		rpdgs.ResPlanDemandGpuSubOrderColumns.ColumnExpr(), rpdgs.ResPlanDemandGpuSubOrderColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// UpdateWithTx update resource plan demand gpu sub order.
func (d ResPlanDemandGpuSubOrderDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression,
	model *rpdgs.ResPlanDemandGpuSubOrderTable) error {

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

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)
	effected, err := d.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.Errorf("update resource plan demand gpu sub order failed, filter: %v, err: %v, rid: %s",
			filterExpr, err, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.Errorf("update resource plan demand gpu sub order, but record not found, filter: %v, rid: %s",
			filterExpr, kt.Rid)
	}

	return nil
}

// List get resource plan demand gpu sub order list.
func (d ResPlanDemandGpuSubOrderDao) List(kt *kit.Kit, opt *types.ListOption) (
	*rpproto.ResPlanDemandGpuSubOrderListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list res plan demand gpu sub order options is nil")
	}

	exprOpt := filter.NewExprOption(
		filter.RuleFields(rpdgs.ResPlanDemandGpuSubOrderColumns.ColumnTypes()),
		filter.MaxInLimit(constant.BkBizIDMaxLimit),
	)
	if err := opt.Validate(exprOpt, core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ResPlanDemandGpuSubOrderTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count res plan demand gpu sub order failed, err: %v, filter: %v, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rpproto.ResPlanDemandGpuSubOrderListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`,
		rpdgs.ResPlanDemandGpuSubOrderColumns.FieldsNamedExpr(opt.Fields),
		table.ResPlanDemandGpuSubOrderTable, whereExpr, pageExpr)

	details := make([]rpdgs.ResPlanDemandGpuSubOrderTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &rpproto.ResPlanDemandGpuSubOrderListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete resource plan demand gpu sub order with tx.
func (d ResPlanDemandGpuSubOrderDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ResPlanDemandGpuSubOrderTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete resource plan demand gpu sub order failed, err: %v, filter: %v, rid: %s",
			err, expr, kt.Rid)
		return err
	}

	return nil
}
