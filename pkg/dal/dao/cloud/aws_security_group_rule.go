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

// AwsSGRule only used for aws security group rule.
type AwsSGRule interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []*cloud.AwsSecurityGroupRuleTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, rule *cloud.AwsSecurityGroupRuleTable) error
	List(kt *kit.Kit, opt *types.SGRuleListOption) (*types.ListAwsSGRuleDetails, error)
	Delete(kt *kit.Kit, expr *filter.Expression) error
}

var _ AwsSGRule = new(AwsSGRuleDao)

// AwsSGRuleDao aws security group rule dao.
type AwsSGRuleDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// BatchCreateWithTx rule.
func (dao *AwsSGRuleDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []*cloud.AwsSecurityGroupRuleTable) (
	[]string, error) {

	// generate account id
	ids, err := dao.IDGen.Batch(kt, table.AwsSecurityGroupRuleTable, len(rules))
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

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.AwsSecurityGroupRuleTable,
		cloud.AwsSGRuleColumns.ColumnExpr(), cloud.AwsSGRuleColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rules); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.AwsSecurityGroupRuleTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.AwsSecurityGroupRuleTable, err)
	}

	return ids, nil
}

// UpdateWithTx rule.
func (dao *AwsSGRuleDao) UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, rule *cloud.
	AwsSecurityGroupRuleTable) error {

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
		logs.ErrorJson("update aws security group rule failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update aws security group rule, but record not found, filter: %v, rid: %v", expr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List rules.
func (dao *AwsSGRuleDao) List(kt *kit.Kit, opt *types.SGRuleListOption) (*types.ListAwsSGRuleDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.AwsSGRuleColumns.ColumnTypes())),
		core.DefaultPageOption); err != nil {
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AwsSecurityGroupRuleTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			logs.ErrorJson("count aws security group rule failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListAwsSGRuleDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.AwsSGRuleColumns.FieldsNamedExpr(opt.Fields),
		table.AwsSecurityGroupRuleTable, whereExpr, pageExpr)

	details := make([]cloud.AwsSecurityGroupRuleTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql); err != nil {
		return nil, err
	}

	return &types.ListAwsSGRuleDetails{Details: details}, nil
}

// Delete rule.
func (dao *AwsSGRuleDao) Delete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AwsSecurityGroupRuleTable, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if err = dao.Orm.Txn(txn).Delete(kt.Ctx, sql); err != nil {
			logs.ErrorJson("delete aws security group rule failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}
