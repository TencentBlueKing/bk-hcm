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

	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// HuaWeiSGRule only used for huawei security group rule.
type HuaWeiSGRule interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []cloud.HuaWeiSecurityGroupRuleTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, rule *cloud.HuaWeiSecurityGroupRuleTable) error
	List(kt *kit.Kit, opt *types.SGRuleListOption) (*types.ListHuaWeiSGRuleDetails, error)
	Delete(kt *kit.Kit, expr *filter.Expression) error
}

var _ HuaWeiSGRule = new(HuaWeiSGRuleDao)

// HuaWeiSGRuleDao huawei security group rule dao.
type HuaWeiSGRuleDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// BatchCreateWithTx rule.
func (dao *HuaWeiSGRuleDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []cloud.HuaWeiSecurityGroupRuleTable) (
	[]string, error) {

	// generate account id
	ids, err := dao.IDGen.Batch(kt, table.HuaWeiSecurityGroupRuleTable, len(rules))
	if err != nil {
		return nil, err
	}
	for index := range rules {
		rules[index].ID = ids[index]
	}

	for _, rule := range rules {
		if err := rule.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.HuaWeiSecurityGroupRuleTable,
		cloud.HuaWeiSGRuleColumns.ColumnExpr(), cloud.HuaWeiSGRuleColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rules); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.HuaWeiSecurityGroupRuleTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.HuaWeiSecurityGroupRuleTable, err)
	}

	return ids, nil
}

// UpdateWithTx rule.
func (dao *HuaWeiSGRuleDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, rule *cloud.
	HuaWeiSecurityGroupRuleTable) error {

	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := rule.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(rule, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, rule.TableName(), setExpr, whereExpr)

	effected, err := dao.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update huawei security group rule failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update huawei security group rule, but record not found, filter: %v, rid: %v", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List rules.
func (dao *HuaWeiSGRuleDao) List(kt *kit.Kit, opt *types.SGRuleListOption) (*types.ListHuaWeiSGRuleDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.HuaWeiSGRuleColumns.ColumnTypes())),
		types.DefaultPageOption); err != nil {
		return nil, err
	}

	whereOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
		CrownedOption: &filter.CrownedOption{
			CrownedOp: filter.And,
			Rules: []filter.RuleFactory{
				&filter.AtomRule{
					Field: "security_group_id",
					Op:    filter.Equal.Factory(),
					Value: opt.SecurityGroupID,
				},
			},
		},
	}
	whereExpr, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.HuaWeiSecurityGroupRuleTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			logs.ErrorJson("count huawei security group rule failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListHuaWeiSGRuleDetails{Count: count}, nil
	}

	pageExpr, err := opt.Page.SQLExpr(types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.HuaWeiSGRuleColumns.FieldsNamedExpr(opt.Fields),
		table.HuaWeiSecurityGroupRuleTable, whereExpr, pageExpr)

	details := make([]cloud.HuaWeiSecurityGroupRuleTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql); err != nil {
		return nil, err
	}

	return &types.ListHuaWeiSGRuleDetails{Details: details}, nil
}

// Delete rule.
func (dao *HuaWeiSGRuleDao) Delete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.HuaWeiSecurityGroupRuleTable, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err = dao.Orm.Txn(txn).Delete(kt.Ctx, sql); err != nil {
			logs.ErrorJson("delete huawei security group rule failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}
