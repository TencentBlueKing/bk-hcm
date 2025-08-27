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

// Package loadbalancer 负载均衡四层/七层规则的Package
package loadbalancer

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
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

// LbTCloudUrlRuleInterface only used for lb tcloud url rule.
type LbTCloudUrlRuleInterface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.TCloudLbUrlRuleTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.TCloudLbUrlRuleTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, model *tablelb.TCloudLbUrlRuleTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLbUrlRuleDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListJoinListener(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLbUrlRuleWithListenerDetails, error)
}

var _ LbTCloudUrlRuleInterface = new(LbTCloudUrlRuleDao)

// LbTCloudUrlRuleDao lb tcloud url rule dao.
type LbTCloudUrlRuleDao struct {
	Orm   orm.Interface
	IDGen idgen.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx lb url rule.
func (dao LbTCloudUrlRuleDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []*tablelb.TCloudLbUrlRuleTable) (
	[]string, error) {

	tableName := table.TCloudLbUrlRuleTable
	ids, err := dao.IDGen.Batch(kt, tableName, len(models))
	if err != nil {
		return nil, err
	}
	lbIds := make([]string, 0, len(models))
	for index, model := range models {
		if err = model.InsertValidate(); err != nil {
			return nil, err
		}
		model.ID = ids[index]
		lbIds = append(lbIds, model.LbID)
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		tablelb.TCloudLbUrlRuleColumns.ColumnExpr(), tablelb.TCloudLbUrlRuleColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	// 创建审计
	lbMap, err := ListLbByIDs(kt, dao.Orm, lbIds)
	if err != nil {
		logs.Errorf("fail to list lb for audit, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	// rule create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, rule := range models {
		lb, ok := lbMap[rule.LbID]
		if !ok {
			return nil, fmt.Errorf("fail to find lb(%s) of rule(%s)", rule.LbID, rule.URL)
		}
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      rule.ID,
			CloudResID: rule.CloudID,
			ResName:    rule.Name,
			ResType:    enumor.UrlRuleAuditResType,
			Action:     enumor.Create,
			BkBizID:    lb.BkBizID,
			Vendor:     enumor.TCloud,
			AccountID:  lb.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: rule,
			},
		})
	}
	if err = dao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	return ids, nil
}

// Update lb url rule.
func (dao LbTCloudUrlRuleDao) Update(kt *kit.Kit, expr *filter.Expression, model *tablelb.TCloudLbUrlRuleTable) error {
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
			logs.Errorf("update load balancer url rule failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effect == 0 {
			logs.Infof("update load balancer url rule, but record not found, sql: %s, rid: %v", sql, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx lb url rule.
func (dao LbTCloudUrlRuleDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string,
	model *tablelb.TCloudLbUrlRuleTable) error {

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
		logs.Errorf("update load balancer url rule failed, id: %s, err: %v, rid: %v", id, err, kt.Rid)
		return err
	}

	return nil
}

// List lb url rule.
func (dao LbTCloudUrlRuleDao) List(kt *kit.Kit, opt *types.ListOption) (*typeslb.ListLbUrlRuleDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	expr := filter.NewExprOption(
		filter.RuleFields(tablelb.TCloudLbUrlRuleColumns.ColumnTypes()),
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.TCloudLbUrlRuleTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer url rule failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typeslb.ListLbUrlRuleDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tablelb.TCloudLbUrlRuleColumns.FieldsNamedExpr(opt.Fields),
		table.TCloudLbUrlRuleTable, whereExpr, pageExpr)

	details := make([]tablelb.TCloudLbUrlRuleTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typeslb.ListLbUrlRuleDetails{Details: details}, nil
}

// DeleteWithTx lb url rule.
func (dao LbTCloudUrlRuleDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.TCloudLbUrlRuleTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete load balancer url rule failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// ListJoinListener lb url rule.
func (dao LbTCloudUrlRuleDao) ListJoinListener(kt *kit.Kit, opt *types.ListOption) (
	*typeslb.ListLbUrlRuleWithListenerDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	expr := filter.NewExprOption(
		filter.RuleFields(tablelb.TCloudLbUrlRuleColumns.ColumnTypes()),
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
		sql := fmt.Sprintf(`SELECT count(*)	FROM %s AS rule LEFT JOIN
		 (select id as listener_id,name as lbl_name, protocol, port from %s) AS t 
		ON rule.lbl_id = t.listener_id %s`,
			table.TCloudLbUrlRuleTable, table.LoadBalancerListenerTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count load balancer url rule failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typeslb.ListLbUrlRuleWithListenerDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s, lbl_name, protocol, port FROM %s AS rule LEFT JOIN
		 (select id as listener_id,name as lbl_name, protocol, port from %s) AS t 
		ON rule.lbl_id = t.listener_id %s %s`,
		tablelb.TCloudLbUrlRuleColumns.FieldsNamedExpr(opt.Fields),
		table.TCloudLbUrlRuleTable, table.LoadBalancerListenerTable, whereExpr, pageExpr)

	details := make([]tablelb.TCloudLbUrlRuleWithListener, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &typeslb.ListLbUrlRuleWithListenerDetails{Details: details}, nil
}
