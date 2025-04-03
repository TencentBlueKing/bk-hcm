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
	"strings"

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

// SecurityGroup only used for security group.
type SecurityGroup interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, sgs []*cloud.SecurityGroupTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, sg *cloud.SecurityGroupTable) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, sg *cloud.SecurityGroupTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ SecurityGroup = new(SecurityGroupDao)

// SecurityGroupDao security group dao.
type SecurityGroupDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx sg with tx.
func (s SecurityGroupDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, sgs []*cloud.SecurityGroupTable) (
	[]string, error) {

	ids, err := s.IDGen.Batch(kt, table.SecurityGroupTable, len(sgs))
	if err != nil {
		return nil, err
	}

	for index, sg := range sgs {
		sg.ID = ids[index]

		if err := sg.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.SecurityGroupTable,
		cloud.SecurityGroupColumns.ColumnExpr(), cloud.SecurityGroupColumns.ColonNameExpr())

	if err = s.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, sgs); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.SecurityGroupTable, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", table.SecurityGroupTable, err)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(sgs))
	for _, one := range sgs {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ID,
			CloudResID: one.CloudID,
			ResName:    one.Name,
			ResType:    enumor.SecurityGroupAuditResType,
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
	if err = s.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update sg.
func (s SecurityGroupDao) Update(kt *kit.Kit, expr *filter.Expression, sg *cloud.SecurityGroupTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := sg.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(sg, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, sg.TableName(), setExpr, whereExpr)

	_, err = s.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := s.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update security group failed, err: %v, filter: %s, rid: %v", err, expr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update security group, but record not found, filter: %v, rid: %v", expr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx sg.
func (s SecurityGroupDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, id string, sg *cloud.SecurityGroupTable) error {
	if len(id) == 0 {
		return errf.New(errf.InvalidParameter, "id is required")
	}

	if err := sg.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(sg, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, sg.TableName(), setExpr)

	toUpdate["id"] = id
	_, err = s.Orm.Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update security group failed, err: %v, id: %s, rid: %v", err, id, kt.Rid)
		return err
	}

	return nil
}

// List sgs.
func (s SecurityGroupDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list security group options is nil")
	}

	// 去除field中的使用业务id字段
	newFields := make([]string, 0, len(opt.Fields))
	for i := range opt.Fields {
		if opt.Fields[i] == "usage_biz_ids" || opt.Fields[i] == "usage_biz_id" {
			continue
		}
		newFields = append(newFields, opt.Fields[i])
	}
	opt.Fields = newFields

	columnTypes := cloud.SecurityGroupColumns.ColumnTypes()
	columnTypes["extension.resource_group_name"] = enumor.String
	columnTypes["extension.vpc_id"] = enumor.String
	columnTypes["rel.res_type"] = enumor.String
	columnTypes["usage_biz_id"] = enumor.Numeric
	err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)), core.NewDefaultPageOption())
	if err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}
	if strings.Contains(whereExpr, "usage_biz_id") {
		// 处理带usage_biz_id的查询
		return s.listWithUsageBiz(kt, opt, whereExpr, err, whereValue)
	}
	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SecurityGroupTable, whereExpr)
		count, err := s.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count security group failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSecurityGroupDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s  %s %s`, cloud.SecurityGroupColumns.FieldsNamedExpr(opt.Fields),
		table.SecurityGroupTable, whereExpr, pageExpr)

	details := make([]cloud.SecurityGroupTable, 0)
	if err = s.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select security group failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListSecurityGroupDetails{Details: details}, nil
}

func (s SecurityGroupDao) listWithUsageBiz(kt *kit.Kit, opt *types.ListOption, whereExpr string, err error,
	whereValue map[string]interface{}) (*types.ListSecurityGroupDetails, error) {

	// 用户传了usage_biz_id，则需要补充res_type条件，保证筛选结果的正确性
	opt.Filter, err = tools.And(tools.EqualExpression("rel.res_type", enumor.SecurityGroupCloudResType), opt.Filter)
	if err != nil {
		logs.Errorf("fail to merge res_type expression, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}
	whereExpr, whereValue, err = opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(
			`SELECT COUNT(*) FROM (SELECT sg.id FROM %s AS sg LEFT JOIN %s AS rel ON sg.id = rel.res_id %s GROUP BY sg.id) sgid`,
			table.SecurityGroupTable, table.ResUsageBizRelTable, whereExpr)

		count, err := s.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count security group with usage biz failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSecurityGroupDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s AS sg LEFT JOIN %s AS rel ON sg.id = rel.res_id %s GROUP BY sg.id %s`,
		cloud.SecurityGroupColumns.FieldsNamedExpr(opt.Fields), table.SecurityGroupTable, table.ResUsageBizRelTable,
		whereExpr, pageExpr)

	details := make([]cloud.SecurityGroupTable, 0)
	if err = s.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select security group with usage biz failed, err: %v, filter: %s, rid: %s",
			err, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListSecurityGroupDetails{Details: details}, nil
}

// DeleteWithTx sg with filter.
func (s SecurityGroupDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SecurityGroupTable, whereExpr)
	if _, err = s.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete security group failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// ListSecurityGroup TODO: 考虑之后这种跨表查询是否可以直接引用对象的 List 函数，而不是再写一个。
func ListSecurityGroup(kt *kit.Kit, orm orm.Interface, ids []string) (map[string]cloud.SecurityGroupTable, error) {

	sql := fmt.Sprintf(`SELECT %s FROM %s where id in (:ids)`, cloud.SecurityGroupColumns.FieldsNamedExpr(nil),
		table.SecurityGroupTable)

	sgs := make([]cloud.SecurityGroupTable, 0)
	if err := orm.Do().Select(kt.Ctx, &sgs, sql, map[string]interface{}{"ids": ids}); err != nil {
		return nil, err
	}

	idSgMap := make(map[string]cloud.SecurityGroupTable, len(ids))
	for _, sg := range sgs {
		idSgMap[sg.ID] = sg
	}

	return idSgMap, nil
}
