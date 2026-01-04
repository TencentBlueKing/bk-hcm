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
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	rtypes "hcm/pkg/dal/dao/types/resource-plan"
	"hcm/pkg/dal/table"
	rpt "hcm/pkg/dal/table/resource-plan/res-plan-ticket"
	rpts "hcm/pkg/dal/table/resource-plan/res-plan-ticket-status"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ResPlanTicketInterface only used for resource plan ticket interface.
type ResPlanTicketInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []rpt.ResPlanTicketTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *rpt.ResPlanTicketTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*rtypes.RPTicketListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	// ListWithStatus list resource plan ticket with corresponding status.
	ListWithStatus(kt *kit.Kit, opt *types.ListOption) (*rtypes.RPTicketWithStatusListRst, error)
}

var _ ResPlanTicketInterface = new(ResPlanTicketDao)

// ResPlanTicketDao resource plan ticket ResPlanTicketDao.
type ResPlanTicketDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create resource plan ticket with tx.
func (d ResPlanTicketDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []rpt.ResPlanTicketTable) ([]string, error) {
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

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		rpt.ResPlanTicketColumns.ColumnExpr(), rpt.ResPlanTicketColumns.ColonNameExpr())

	if err = d.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update update resource plan ticket.
func (d ResPlanTicketDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *rpt.ResPlanTicketTable) error {
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

	_, err = d.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := d.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update resource plan ticket failed, filter: %v, err: %v, rid: %v",
				filterExpr, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update resource plan ticket, but record not found, filter: %v, rid: %v",
				filterExpr, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List get resource plan ticket list.
func (d ResPlanTicketDao) List(kt *kit.Kit, opt *types.ListOption) (*rtypes.RPTicketListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list res plan ticket options is nil")
	}

	exprOpt := filter.NewExprOption(
		filter.RuleFields(rpt.ResPlanTicketColumns.ColumnTypes()),
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
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ResPlanTicketTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count res plan ticket failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rtypes.RPTicketListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, rpt.ResPlanTicketColumns.FieldsNamedExpr(opt.Fields),
		table.ResPlanTicketTable, whereExpr, pageExpr)

	details := make([]rpt.ResPlanTicketTable, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &rtypes.RPTicketListResult{Count: 0, Details: details}, nil
}

// DeleteWithTx delete resource plan ticket with tx.
func (d ResPlanTicketDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ResPlanTicketTable, whereExpr)

	if _, err = d.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete resource plan ticket failed, err: %v, filter: %v, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// ListWithStatus list resource plan ticket with corresponding status.
// TODO 无法用 res_plan_ticket 和 res_plan_ticket_status 的共有字段作为查询条件，例如 created_time
func (d ResPlanTicketDao) ListWithStatus(kt *kit.Kit, opt *types.ListOption) (
	*rtypes.RPTicketWithStatusListRst, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list res plan ticket options is nil")
	}

	// append status col type to col types.
	colTypes := rpt.ResPlanTicketColumns.ColumnTypes()
	statusColType := rpts.ResPlanTicketStatusColumns.ColumnTypes()["status"]
	itsmSnColType := rpts.ResPlanTicketStatusColumns.ColumnTypes()["itsm_sn"]
	crpSnColType := rpts.ResPlanTicketStatusColumns.ColumnTypes()["crp_sn"]
	colTypes["status"] = statusColType
	colTypes["itsm_sn"] = itsmSnColType
	colTypes["crp_sn"] = crpSnColType

	exprOpt := filter.NewExprOption(
		filter.RuleFields(colTypes),
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
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s rpt JOIN %s rpts ON rpt.id = rpts.ticket_id %s`,
			table.ResPlanTicketTable, table.ResPlanTicketStatusTable, whereExpr)

		count, err := d.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count res plan ticket failed, err: %v, filter: %v, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rtypes.RPTicketWithStatusListRst{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	// convert resource plan ticket columns to rpt.column.
	columns := make([]string, 0, len(rpt.ResPlanTicketColumns.Columns()))
	for _, col := range rpt.ResPlanTicketColumns.Columns() {
		columns = append(columns, "rpt."+col)
	}
	sql := fmt.Sprintf(
		`SELECT %s, rpts.status, rpts.itsm_sn, rpts.crp_sn FROM %s rpt JOIN %s rpts ON rpt.id = rpts.ticket_id %s %s`,
		strings.Join(columns, ","), table.ResPlanTicketTable, table.ResPlanTicketStatusTable, whereExpr, pageExpr)

	details := make([]rtypes.RPTicketWithStatus, 0)
	if err = d.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	// set status name.
	for idx, detail := range details {
		details[idx].StatusName = detail.Status.Name()
	}

	return &rtypes.RPTicketWithStatusListRst{Count: 0, Details: details}, nil
}
