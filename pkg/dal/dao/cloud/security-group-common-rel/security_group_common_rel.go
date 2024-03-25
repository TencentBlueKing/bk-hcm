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

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/cloud/load-balancer"
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

// Interface only used for security group and common relation.
type Interface interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupCommonRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupCommonRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	ListJoinSecurityGroup(kt *kit.Kit, resIDs []string, resType enumor.CloudResourceType) (
		*types.ListSGCommonRelsJoinSGDetails, error)
}

var _ Interface = new(Dao)

// Dao define security group and common relation dao.
type Dao struct {
	Orm orm.Interface
}

// ListJoinSecurityGroup rels with security groups.
func (dao Dao) ListJoinSecurityGroup(kt *kit.Kit, resIDs []string, resType enumor.CloudResourceType) (
	*types.ListSGCommonRelsJoinSGDetails, error) {

	if len(resIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "res ids is required")
	}

	var withoutFields = []string{"vendor", "reviser", "updated_at"}
	withoutFields = append(withoutFields, types.DefaultRelJoinWithoutField...)
	sql := fmt.Sprintf(`SELECT %s, %s, sg.vendor AS vendor,sg.reviser AS reviser,sg.updated_at AS updated_at,
		rel.res_type,rel.priority FROM %s AS rel LEFT JOIN %s AS sg ON rel.security_group_id = sg.id 
		WHERE res_id IN (:res_ids) AND res_type = :res_type`,
		cloud.SecurityGroupColumns.FieldsNamedExprWithout(withoutFields),
		tools.BaseRelJoinSqlBuild("rel", "sg", "id", "res_id"),
		table.SecurityGroupCommonRelTable, table.SecurityGroupTable)

	details := make([]types.SecurityGroupWithCommonID, 0)
	updateMap := map[string]interface{}{"res_ids": resIDs, "res_type": resType}
	if err := dao.Orm.Do().Select(kt.Ctx, &details, sql, updateMap); err != nil {
		logs.Errorf("select sg common rels join sg failed, err: %v, sql: (%s), resIDs: %v, resType: %s, rid: %s",
			err, resIDs, resType, sql, kt.Rid)
		return nil, err
	}

	return &types.ListSGCommonRelsJoinSGDetails{Details: details}, nil
}

// BatchCreateWithTx rels.
func (dao Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupCommonRelTable) error {
	// 校验关联资源是否存在
	sgIDs := make([]string, 0)
	resIDs := make([]string, 0)
	for _, rel := range rels {
		sgIDs = append(sgIDs, rel.SecurityGroupID)
		resIDs = append(resIDs, rel.ResID)
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

	resMap, err := loadbalancer.ListLbByIDs(kt, dao.Orm, resIDs)
	if err != nil {
		logs.Errorf("list clb by ids failed, err: %v, sgIDs: %v, resIDs: %v, rid: %s", err, sgIDs, resIDs, kt.Rid)
		return err
	}

	if len(resMap) != len(converter.StringSliceToMap(resIDs)) {
		logs.Errorf("get clb count not right, err: %v, ids: %v, count: %d, rid: %s", err, resIDs, len(resMap), kt.Rid)
		return fmt.Errorf("get clb count not right")
	}

	tableName := table.SecurityGroupCommonRelTable
	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		cloud.SecurityGroupCommonRelColumns.ColumnExpr(), cloud.SecurityGroupCommonRelColumns.ColonNameExpr())

	if err = dao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels); err != nil {
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
