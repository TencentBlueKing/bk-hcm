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

package securitygroup

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

// AwsSGRule only used for aws security group rule.
type AwsSGRule interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rules []*cloud.AwsSecurityGroupRuleTable) ([]string, error)
	UpdateWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, rule *cloud.AwsSecurityGroupRuleTable) error
	List(kt *kit.Kit, opt *types.SGRuleListOption) (*types.ListAwsSGRuleDetails, error)
	Delete(kt *kit.Kit, expr *filter.Expression) error
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	CountBySecurityGroupIDs(kt *kit.Kit, expr *filter.Expression) (map[string]int64, error)
}

var _ AwsSGRule = new(AwsSGRuleDao)

// AwsSGRuleDao aws security group rule dao.
type AwsSGRuleDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
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

	if err = dao.batchCreateAudit(kt, tx, rules); err != nil {
		return nil, err
	}

	return ids, nil
}

func (dao *AwsSGRuleDao) batchCreateAudit(kt *kit.Kit, tx *sqlx.Tx, rules []*cloud.AwsSecurityGroupRuleTable) error {
	sgIDMap := make(map[string]bool, 0)
	for _, rule := range rules {
		sgIDMap[rule.SecurityGroupID] = true
	}

	sgIDs := make([]string, 0, len(sgIDMap))
	for id, _ := range sgIDMap {
		sgIDs = append(sgIDs, id)
	}

	idSgMap, err := ListSecurityGroup(kt, dao.Orm, sgIDs)
	if err != nil {
		return err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(rules))
	for _, rule := range rules {
		sg, exist := idSgMap[rule.SecurityGroupID]
		if !exist {
			return errf.Newf(errf.RecordNotFound, "security group: %s not found", rule.SecurityGroupID)
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      sg.ID,
			CloudResID: sg.CloudID,
			ResName:    sg.Name,
			ResType:    enumor.SecurityGroupAuditResType,
			Action:     enumor.Update,
			BkBizID:    sg.BkBizID,
			Vendor:     sg.Vendor,
			AccountID:  sg.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.ChildResAuditData{
					ChildResType: enumor.SecurityGroupRuleAuditResType,
					Action:       enumor.Create,
					ChildRes:     rule,
				},
			},
		})
	}

	if err = dao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
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

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(rule, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, rule.TableName(), setExpr, whereExpr)

	effected, err := dao.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
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

	columnTypes := cloud.AwsSGRuleColumns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
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
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.AwsSecurityGroupRuleTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
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
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListAwsSGRuleDetails{Details: details}, nil
}

// Delete rule.
func (dao *AwsSGRuleDao) Delete(kt *kit.Kit, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AwsSecurityGroupRuleTable, whereExpr)

	_, err = dao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		if _, err = dao.Orm.Txn(txn).Delete(kt.Ctx, sql, whereValue); err != nil {
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

// DeleteWithTx rule with tx.
func (dao *AwsSGRuleDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.AwsSecurityGroupRuleTable, whereExpr)

	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete aws security group rule failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// CountBySecurityGroupIDs count rules by security group ids.
func (dao *AwsSGRuleDao) CountBySecurityGroupIDs(kt *kit.Kit, expr *filter.Expression) (map[string]int64, error) {

	if expr == nil {
		return nil, errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}
	sql := fmt.Sprintf(`SELECT security_group_id, COUNT(*) AS rule_count FROM %s %s GROUP BY security_group_id;`,
		table.AwsSecurityGroupRuleTable, whereExpr)

	details := make([]struct {
		SecurityGroupID string `db:"security_group_id"`
		RuleCount       int64  `db:"rule_count"`
	}, 0)

	err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int64)
	for _, detail := range details {
		result[detail.SecurityGroupID] = detail.RuleCount
	}

	return result, nil
}
