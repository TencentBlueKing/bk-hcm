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

package sgcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/cloud/cvm"
	securitygroup "hcm/pkg/dal/dao/cloud/security-group"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// Interface only used for security group and cvm relation.
type Interface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupCvmRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupCvmRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListJoinSecurityGroup(kt *kit.Kit, cvmIDs []string) (*types.ListSGCvmRelsJoinSGDetails, error)
}

var _ Interface = new(Dao)

// Dao define security group and cvm relation dao.
type Dao struct {
	Orm orm.Interface
}

// ListJoinSecurityGroup rels with security groups.
func (dao Dao) ListJoinSecurityGroup(kt *kit.Kit, cvmIDs []string) (*types.ListSGCvmRelsJoinSGDetails, error) {
	if len(cvmIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "cvm ids is required")
	}

	sql := fmt.Sprintf(`SELECT %s, %s FROM %s as rel left join %s as sg on rel.security_group_id = sg.id 
	where cvm_id in (:cvm_ids)`,
	cloud.SecurityGroupColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
	tools.BaseRelJoinSqlBuild("rel", "sg", "id", "cvm_id"),
	table.SecurityGroupCvmTable, table.SecurityGroupTable)

	details := make([]types.SecurityGroupWithCvmID, 0)
	if err := dao.Orm.Do().Select(kt.Ctx, &details, sql, map[string]interface{}{"cvm_ids": cvmIDs}); err != nil {
		logs.ErrorJson("select sg cvm rels join sg failed, err: %v, sql: (%s), rid: %s", err, sql, kt.Rid)
		return nil, err
	}

	return &types.ListSGCvmRelsJoinSGDetails{Details: details}, nil
}

// BatchCreateWithTx rels.
func (dao Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupCvmRelTable) error {
	// 校验关联资源是否存在
	sgIDs := make([]string, 0)
	cvmIDs := make([]string, 0)
	for _, rel := range rels {
		sgIDs = append(sgIDs, rel.SecurityGroupID)
		cvmIDs = append(cvmIDs, rel.CvmID)
	}

	sgMap, err := securitygroup.ListSecurityGroup(kt, dao.Orm, sgIDs)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
		return err
	}

	if len(sgMap) != len(converter.StringSliceToMap(sgIDs)) {
		logs.Errorf("get security group count not right, ids: %v, count: %d, rid: %s", sgIDs, len(sgMap), kt.Rid)
		return fmt.Errorf("get security group count not right")
	}

	cvmMap, err := cvm.ListCvm(kt, dao.Orm, cvmIDs)
	if err != nil {
		logs.Errorf("list security group failed, err: %v, ids: %v, rid: %s", err, sgIDs, kt.Rid)
		return err
	}

	if len(cvmMap) != len(converter.StringSliceToMap(cvmIDs)) {
		logs.Errorf("get cvm count not right, err: %v, ids: %v, count: %d, rid: %s", err, cvmIDs, len(cvmMap), kt.Rid)
		return fmt.Errorf("get cvm count not right")
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.SecurityGroupCvmTable,
		cloud.SecurityGroupCvmRelColumns.ColumnExpr(), cloud.SecurityGroupCvmRelColumns.ColonNameExpr())

	if err := dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", table.SecurityGroupCvmTable, err, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", table.SecurityGroupCvmTable, err)
	}

	return nil
}

// List rels.
func (dao Dao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupCvmRelDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.SecurityGroupCvmRelColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SecurityGroupCvmTable, whereExpr)

		count, err := dao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count security group cvm rels failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSecurityGroupCvmRelDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.SecurityGroupCvmRelColumns.FieldsNamedExpr(opt.Fields),
		table.SecurityGroupCvmTable, whereExpr, pageExpr)

	details := make([]cloud.SecurityGroupCvmRelTable, 0)
	if err = dao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select security group cvm rels failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
		return nil, err
	}

	return &types.ListSecurityGroupCvmRelDetails{Details: details}, nil
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

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SecurityGroupCvmTable, whereExpr)
	if _, err = dao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete security group cvm rels failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
