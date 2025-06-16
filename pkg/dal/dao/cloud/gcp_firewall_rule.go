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

// GcpFirewallRule only used for gcp firewall rule.
type GcpFirewallRule interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []*cloud.GcpFirewallRuleTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, rule *cloud.GcpFirewallRuleTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, rule *cloud.GcpFirewallRuleTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListGcpFirewallRuleDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ GcpFirewallRule = new(GcpFirewallRuleDao)

// GcpFirewallRuleDao gcp firewall rule dao.
type GcpFirewallRuleDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx rule.
func (g GcpFirewallRuleDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []*cloud.GcpFirewallRuleTable) (
	[]string, error) {

	ids, err := g.IDGen.Batch(kt, table.GcpFirewallRuleTable, len(rules))
	if err != nil {
		return nil, err
	}
	for index, rule := range rules {
		rule.ID = ids[index]

		if err := rule.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.GcpFirewallRuleTable,
		cloud.GcpFirewallRuleColumns.ColumnExpr(), cloud.GcpFirewallRuleColumns.ColonNameExpr())
	err = g.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, rules)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.GcpFirewallRuleTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.GcpFirewallRuleTable, err)
	}

	audits := make([]*tableaudit.AuditTable, 0, len(rules))
	for _, rule := range rules {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      rule.ID,
			CloudResID: rule.CloudID,
			ResName:    rule.Name,
			ResType:    enumor.GcpFirewallRuleAuditResType,
			Action:     enumor.Create,
			BkBizID:    rule.BkBizID,
			Vendor:     enumor.Gcp,
			AccountID:  rule.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: rule,
			},
		})
	}
	if err = g.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update rule.
func (g GcpFirewallRuleDao) Update(kt *kit.Kit, expr *filter.Expression, rule *cloud.GcpFirewallRuleTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := rule.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(rule, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, rule.TableName(), setExpr, whereExpr)

	_, err = g.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := g.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update %s failed, err: %v, filter: %s, rid: %v", table.GcpFirewallRuleTable, err,
				expr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update %s, but record not found, filter: %v, rid: %v", table.GcpFirewallRuleTable,
				expr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx rule.
func (g GcpFirewallRuleDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, rule *cloud.
	GcpFirewallRuleTable) error {

	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := rule.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(rule, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, rule.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = g.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update %s failed, err: %v, id: %s, rid: %v", table.GcpFirewallRuleTable, err, id, kt.Rid)
		return err
	}

	return nil
}

// List rule.
func (g GcpFirewallRuleDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListGcpFirewallRuleDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.GcpFirewallRuleColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.GcpFirewallRuleTable, whereExpr)

		count, err := g.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count %s failed, err: %v, filter: %s, rid: %s", table.GcpFirewallRuleTable, err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListGcpFirewallRuleDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.GcpFirewallRuleColumns.FieldsNamedExpr(opt.Fields),
		table.GcpFirewallRuleTable, whereExpr, pageExpr)

	details := make([]cloud.GcpFirewallRuleTable, 0)
	err = g.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.ListGcpFirewallRuleDetails{Details: details}, nil
}

// DeleteWithTx rule.
func (g GcpFirewallRuleDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.GcpFirewallRuleTable, whereExpr)
	_, err = g.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete %s failed, err: %v, filter: %s, rid: %s", table.GcpFirewallRuleTable, err, expr, kt.Rid)
		return err
	}

	return nil
}
