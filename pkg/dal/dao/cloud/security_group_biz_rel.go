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

	"github.com/jmoiron/sqlx"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// SecurityGroupBizRel only used for security_group and biz rel.
type SecurityGroupBizRel interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []cloud.SecurityGroupBizRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupBizRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
}

var _ SecurityGroupBizRel = new(SecurityGroupBizRelDao)

// SecurityGroupBizRelDao security_group and biz relation dao.
type SecurityGroupBizRelDao struct {
	Orm orm.Interface
}

// BatchCreateWithTx SecurityGroupBizRel with tx.
func (a SecurityGroupBizRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx,
		rels []cloud.SecurityGroupBizRelTable) error {

	if len(rels) == 0 {
		return errf.New(errf.InvalidParameter, "SecurityGroupBizRelTables is required")
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.SecurityGroupBizRelTable,
		cloud.SecurityGroupBizRelColumns.ColumnExpr(), cloud.SecurityGroupBizRelColumns.ColonNameExpr())

	err := a.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels)
	if err != nil {
		return fmt.Errorf("insert %s failed, err: %v", table.SecurityGroupBizRelTable, err)
	}

	return nil
}

// List SecurityGroupBizRel list.
func (a SecurityGroupBizRelDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupBizRelDetails,
		error) {
	
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.SecurityGroupBizRelColumns.ColumnTypes())),
		types.DefaultPageOption); err != nil {
		return nil, err
	}

	whereExpr, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SecurityGroupBizRelTable, whereExpr)

		count, err := a.Orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			logs.ErrorJson("count security_group_biz_rel failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSecurityGroupBizRelDetails{Count: count}, nil
	}

	pageExpr, err := opt.Page.SQLExpr(types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.SecurityGroupBizRelColumns.FieldsNamedExpr(opt.Fields),
		table.SecurityGroupBizRelTable, whereExpr, pageExpr)

	details := make([]cloud.SecurityGroupBizRelTable, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &details, sql); err != nil {
		return nil, err
	}

	return &types.ListSecurityGroupBizRelDetails{Count: 0, Details: details}, nil
}

// DeleteWithTx SecurityGroupBizRel with tx.
func (a SecurityGroupBizRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SecurityGroupBizRelTable, whereExpr)
	if err := a.Orm.Txn(tx).Delete(kt.Ctx, sql); err != nil {
		logs.ErrorJson("delete security_group_biz_rel failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
