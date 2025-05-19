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

package sgcomrel

import (
	"fmt"
	"strings"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud"
	cvmtable "hcm/pkg/dal/table/cloud/cvm"
	lbtable "hcm/pkg/dal/table/cloud/load-balancer"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Interface only used for security group and common relation.
type Interface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupCommonRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupCommonRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListJoinSecurityGroup(kt *kit.Kit, sgIDs, resIDs []string, resType enumor.CloudResourceType) (
		*types.ListSGCommonRelsJoinSGDetails, error)
	ListJoinCVM(kt *kit.Kit, sgIDs []string, opt *types.ListOption) (*types.ListSGCommonRelJoinCVMDetails, error)
	ListJoinLoadBalancer(kt *kit.Kit, sgIDs []string, opt *types.ListOption) (
		*types.ListSGCommonRelJoinLBDetails, error)
}

var _ Interface = new(Dao)

// Dao define security group and common relation dao.
type Dao struct {
	Orm orm.Interface
}

// ListJoinSecurityGroup rels with security groups.
func (dao Dao) ListJoinSecurityGroup(kt *kit.Kit, sgIDs, resIDs []string, resType enumor.CloudResourceType) (
	*types.ListSGCommonRelsJoinSGDetails, error) {

	if len(resIDs) == 0 && len(sgIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "res_ids or sg_ids is required")
	}

	var withoutFields = []string{"vendor", "reviser", "updated_at"}
	withoutFields = append(withoutFields, types.DefaultRelJoinWithoutField...)

	whereExprMaps := make([]string, 0)
	updateMap := make(map[string]interface{})
	if len(resIDs) > 0 {
		whereExprMaps = append(whereExprMaps, " res_id IN (:res_ids) AND res_type = :res_type ")
		updateMap["res_ids"] = resIDs
		updateMap["res_type"] = resType
	}

	if len(sgIDs) > 0 {
		whereExprMaps = append(whereExprMaps, " sg.id IN (:sg_ids) ")
		updateMap["sg_ids"] = sgIDs
	}

	sql := fmt.Sprintf(`SELECT %s, %s, sg.vendor AS vendor,sg.reviser AS reviser,sg.updated_at AS updated_at,
		rel.res_type,rel.priority FROM %s AS rel LEFT JOIN %s AS sg ON rel.security_group_id = sg.id 
		WHERE %s`,
		cloud.SecurityGroupColumns.FieldsNamedExprWithout(withoutFields),
		tools.BaseRelJoinSqlBuild("rel", "sg", "id", "res_id"),
		table.SecurityGroupCommonRelTable, table.SecurityGroupTable,
		strings.Join(whereExprMaps, " AND "))

	details := make([]types.SecurityGroupWithCommonID, 0)
	if err := dao.Orm.Do().Select(kt.Ctx, &details, sql, updateMap); err != nil {
		logs.Errorf("select sg common rels join sg failed, err: %v, sql: (%s), sgIDs: %v, resIDs: %v, "+
			"resType: %s, rid: %s", err, sql, sgIDs, resIDs, resType, kt.Rid)
		return nil, err
	}

	return &types.ListSGCommonRelsJoinSGDetails{Details: details}, nil
}

// ListJoinCVM rels with cvm.
func (dao Dao) ListJoinCVM(kt *kit.Kit, sgIDs []string, opt *types.ListOption) (*types.ListSGCommonRelJoinCVMDetails,
	error) {

	columnTypes := cvmtable.TableColumns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	joinFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			opt.Filter,
			tools.RuleIn("security_group_id", sgIDs),
			tools.RuleEqual("res_type", enumor.CvmCloudResType),
		},
	}

	var withoutFields = []string{"vendor", "reviser", "updated_at"}
	withoutFields = append(withoutFields, types.DefaultRelJoinWithoutField...)

	whereExpr, whereValue, err := joinFilter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s AS rel LEFT JOIN %s AS t ON rel.res_id = t.id %s`,
			table.SecurityGroupCommonRelTable, table.CvmTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count sg common rel join cvm failed, err: %v, filter: %s, rid: %s", err, joinFilter,
				kt.Rid)
			return nil, err
		}

		return &types.ListSGCommonRelJoinCVMDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s, %s, t.vendor AS vendor,t.reviser AS reviser,t.updated_at AS updated_at 
        FROM %s AS rel LEFT JOIN %s AS t ON rel.res_id = t.id %s %s`,
		cvmtable.TableColumns.FieldsNamedExprWithout(withoutFields),
		tools.BaseRelJoinSqlBuild("rel", "t", "id", "security_group_id"),
		table.SecurityGroupCommonRelTable, table.CvmTable, whereExpr, pageExpr)

	details := make([]types.SGCommonRelWithCVM, 0)
	if err := dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.Errorf("select sg common rel join cvm failed, err: %v, sql: (%s), whereValue: %+v, rid: %s",
			err, sql, whereValue, kt.Rid)
		return nil, err
	}

	return &types.ListSGCommonRelJoinCVMDetails{Details: details}, nil
}

// ListJoinLoadBalancer rels with load balancer.
func (dao Dao) ListJoinLoadBalancer(kt *kit.Kit, sgIDs []string, opt *types.ListOption) (
	*types.ListSGCommonRelJoinLBDetails, error) {

	columnTypes := lbtable.LoadBalancerColumns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	joinFilter := &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			opt.Filter,
			tools.RuleIn("security_group_id", sgIDs),
			tools.RuleEqual("res_type", enumor.LoadBalancerCloudResType),
		},
	}

	var withoutFields = []string{"vendor", "reviser", "updated_at"}
	withoutFields = append(withoutFields, types.DefaultRelJoinWithoutField...)

	whereExpr, whereValue, err := joinFilter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s AS rel LEFT JOIN %s AS t ON rel.res_id = t.id %s`,
			table.SecurityGroupCommonRelTable, table.LoadBalancerTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count sg common rel join load balancer failed, err: %v, filter: %s, rid: %s", err,
				joinFilter, kt.Rid)
			return nil, err
		}

		return &types.ListSGCommonRelJoinLBDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s, %s, t.vendor AS vendor,t.reviser AS reviser,t.updated_at AS updated_at 
        FROM %s AS rel LEFT JOIN %s AS t ON rel.res_id = t.id %s %s`,
		lbtable.LoadBalancerColumns.FieldsNamedExprWithout(withoutFields),
		tools.BaseRelJoinSqlBuild("rel", "t", "id", "security_group_id"),
		table.SecurityGroupCommonRelTable, table.LoadBalancerTable, whereExpr, pageExpr)

	details := make([]types.SGCommonRelWithLB, 0)
	if err := dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.Errorf("select sg common rel join load balancer failed, err: %v, sql: (%s), whereValue: %+v, rid: %s",
			err, sql, whereValue, kt.Rid)
		return nil, err
	}

	return &types.ListSGCommonRelJoinLBDetails{Details: details}, nil
}

// BatchCreateWithTx rels.
func (dao Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupCommonRelTable) error {

	tableName := table.SecurityGroupCommonRelTable
	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		cloud.SecurityGroupCommonRelColumns.ColumnExpr(), cloud.SecurityGroupCommonRelColumns.ColonNameExpr())

	if err := dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return nil
}

// List rels.
func (dao Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupCommonRelDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.SecurityGroupCommonRelColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SecurityGroupCommonRelTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count security group cvm rels failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSecurityGroupCommonRelDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.SecurityGroupCommonRelColumns.FieldsNamedExpr(opt.Fields),
		table.SecurityGroupCommonRelTable, whereExpr, pageExpr)

	details := make([]cloud.SecurityGroupCommonRelTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select security group common rels failed, err: %v, filter: %s, rid: %s",
			err, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListSecurityGroupCommonRelDetails{Details: details}, nil
}

// DeleteWithTx rels.
func (dao Dao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SecurityGroupCommonRelTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete security group common rels failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
