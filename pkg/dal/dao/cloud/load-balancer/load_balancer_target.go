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

// Package loadbalancer 负载均衡目标的Package
package loadbalancer

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typeslb "hcm/pkg/dal/dao/types/load-balancer"
	"hcm/pkg/dal/table"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// TargetInterface only used for target.
type TargetInterface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.LoadBalancerTargetTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.LoadBalancerTargetTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablelb.LoadBalancerTargetTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLoadBalancerTargetDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListInstInfo(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListInstInfoDetails, error)
}

var _ TargetInterface = new(TargetDao)

// TargetDao target dao.
type TargetDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx target.
func (dao TargetDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.LoadBalancerTargetTable) (
	[]string, error) {

	tableName := table.LoadBalancerTargetTable
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
		tablelb.LoadBalancerTargetColumns.ColumnExpr(), tablelb.LoadBalancerTargetColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return ids, nil
}

// Update target.
func (dao TargetDao) Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.LoadBalancerTargetTable) error {
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
		effect, err := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.Errorf("update load balancer target failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update load balancer target, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx target.
func (dao TargetDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	model *tablelb.LoadBalancerTargetTable) error {

	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo", "weight").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, model.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update load balancer target failed, id: %s, err: %v, rid: %v", id, err, kt.Rid)
		return err
	}

	return nil
}

// List target.
func (dao TargetDao) List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLoadBalancerTargetDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	expr := filter.NewExprOption(
		filter.RuleFields(tablelb.LoadBalancerTargetColumns.ColumnTypes()),
		filter.MaxInLimit(constant.CLBTopoFindInLimit),
	)
	if err := opt.Validate(expr, core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.LoadBalancerTargetTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer target failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typeslb.ListLoadBalancerTargetDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablelb.LoadBalancerTargetColumns.FieldsNamedExpr(opt.Fields),
		table.LoadBalancerTargetTable, whereExpr, pageExpr)

	details := make([]tablelb.LoadBalancerTargetTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typeslb.ListLoadBalancerTargetDetails{Details: details}, nil
}

// DeleteWithTx target.
func (dao TargetDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.LoadBalancerTargetTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete load balancer target failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// ListInstInfo list instance info.
func (dao TargetDao) ListInstInfo(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListInstInfoDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	expr := filter.NewExprOption(
		filter.RuleFields(tablelb.LoadBalancerTargetColumns.ColumnTypes()),
		filter.MaxInLimit(constant.CLBTopoFindInLimit),
	)
	if err := opt.Validate(expr, core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	subSql := fmt.Sprintf(`SELECT ip,inst_id,inst_type,inst_name,zone,cloud_vpc_ids FROM %s %s group by ip,inst_id,
inst_type,inst_name,zone,cloud_vpc_ids`, table.LoadBalancerTargetTable, whereExpr)

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM (%s) AS tmp`, subSql)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer target inst info failed, err: %v, filter: %s, rid: %s", err, opt.Filter,
				kt.Rid)
			return nil, err
		}

		return &typeslb.ListInstInfoDetails{Count: count}, nil
	}

	if opt.Page.Sort == "" {
		opt.Page.Sort = "ip"
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`%s %s`, subSql, pageExpr)
	details := make([]typeslb.ListInstInfo, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typeslb.ListInstInfoDetails{Details: details}, nil
}
