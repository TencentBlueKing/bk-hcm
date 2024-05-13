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

// Package loadbalancer 负载均衡目标组的Package
package loadbalancer

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	typeslb "hcm/pkg/dal/dao/types/load-balancer"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tablelb "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// TargetGroupInterface only used for target group.
type TargetGroupInterface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.LoadBalancerTargetGroupTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.LoadBalancerTargetGroupTable) error
	UpdateBatch(kt *kit.Kit, models []*tablelb.LoadBalancerTargetGroupTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablelb.LoadBalancerTargetGroupTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLbTargetGroupDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ TargetGroupInterface = new(TargetGroupDao)

// TargetGroupDao target group dao.
type TargetGroupDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx lb target group.
func (dao TargetGroupDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.LoadBalancerTargetGroupTable) (
	[]string, error) {

	tableName := table.LoadBalancerTargetGroupTable
	ids, err := dao.IDGen.Batch(kt, tableName, len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		if err = model.InsertValidate(); err != nil {
			return nil, err
		}
		model.ID = ids[index]
		if len(model.CloudID) == 0 {
			model.CloudID = model.ID
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tablelb.LoadBalancerTargetGroupColumns.ColumnExpr(), tablelb.LoadBalancerTargetGroupColumns.ColonNameExpr())
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
			ResType:    enumor.TargetGroupAuditResType,
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
		logs.Errorf("batch create target group audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update lb target group.
func (dao TargetGroupDao) Update(kt *kit.Kit, expr *filter.Expression,
	model *tablelb.LoadBalancerTargetGroupTable) error {

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

	opts := utils.NewFieldOptions().AddBlankedFields("memo", "weight").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effect, err := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.Errorf("update load balancer target group failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update load balancer target group, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateBatch lb target group.
func (dao TargetGroupDao) UpdateBatch(kt *kit.Kit, models []*tablelb.LoadBalancerTargetGroupTable) error {
	for _, model := range models {
		if len(model.ID) == 0 {
			return errors.New("id is require for tg batch update")
		}
		if err := model.UpdateValidate(); err != nil {
			return err
		}
	}
	_, err := dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {

		for _, model := range models {

			expr := tools.EqualExpression("id", model.ID)
			whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
			if err != nil {
				return nil, err
			}

			opts := utils.NewFieldOptions().AddBlankedFields("memo",
				"weight").AddIgnoredFields(types.DefaultIgnoredFields...)
			setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
			if err != nil {
				return nil, fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
			}

			sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

			effect, err := dao.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
			if err != nil {
				logs.Errorf("batch update load balancer target group failed, err: %v, filter: %s, rid: %v",
					err, expr, kt.Rid)
				return nil, err
			}

			if effect == 0 {
				logs.Infof("batch update load balancer target group, but record not found, sql: %s, rid: %v",
					sql, kt.Rid)
			}
		}
		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx lb target group.
func (dao TargetGroupDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	model *tablelb.LoadBalancerTargetGroupTable) error {

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
	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.Errorf("update load balancer target group failed, id: %s, err: %v, rid: %v", id, err, kt.Rid)
		return err
	}

	return nil
}

// List lb target group.
func (dao TargetGroupDao) List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLbTargetGroupDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	columnTypes := tablelb.LoadBalancerTargetGroupColumns.ColumnTypes()
	columnTypes["health_check.health_switch"] = enumor.Numeric
	columnTypes["health_check.check_port"] = enumor.Numeric
	columnTypes["health_check.check_type"] = enumor.String
	columnTypes["health_check.http_check_path"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(
		filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.LoadBalancerTargetGroupTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer target group failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typeslb.ListLbTargetGroupDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablelb.LoadBalancerTargetGroupColumns.FieldsNamedExpr(opt.Fields),
		table.LoadBalancerTargetGroupTable, whereExpr, pageExpr)

	details := make([]tablelb.LoadBalancerTargetGroupTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typeslb.ListLbTargetGroupDetails{Details: details}, nil
}

// DeleteWithTx target group.
func (dao TargetGroupDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.LoadBalancerTargetGroupTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete load balancer target group failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
