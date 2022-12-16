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

// SecurityGroup only used for security group.
type SecurityGroup interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, sg *cloud.SecurityGroupTable) (string, error)
	Update(kt *kit.Kit, expr *filter.Expression, sg *cloud.SecurityGroupTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ SecurityGroup = new(SecurityGroupDao)

// SecurityGroupDao security group dao.
type SecurityGroupDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx sg with tx.
func (s SecurityGroupDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, sg *cloud.SecurityGroupTable) (string, error) {
	// generate account id
	id, err := s.IDGen.One(kt, table.SecurityGroupTable)
	if err != nil {
		return "", err
	}
	sg.ID = id

	if err := sg.InsertValidate(); err != nil {
		return "", err
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, sg.TableName(), cloud.SecurityGroupColumns.ColumnExpr(),
		cloud.SecurityGroupColumns.ColonNameExpr())

	if err = s.Orm.Txn(tx).Insert(kt.Ctx, sql, sg); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", sg.TableName(), err, kt.Rid)
		return "", fmt.Errorf("insert %s failed, err: %v", sg.TableName(), err)
	}

	return id, nil
}

// Update sg.
func (s SecurityGroupDao) Update(kt *kit.Kit, expr *filter.Expression, sg *cloud.SecurityGroupTable) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := sg.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
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
		effected, err := s.Orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
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

// List sgs.
func (s SecurityGroupDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListSecurityGroupDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list security group options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.SecurityGroupColumns.ColumnTypes())),
		types.DefaultPageOption); err != nil {
		return nil, err
	}

	whereExpr, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SecurityGroupTable, whereExpr)

		count, err := s.Orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			logs.ErrorJson("count security group failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListSecurityGroupDetails{Count: count}, nil
	}

	pageExpr, err := opt.Page.SQLExpr(types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.SecurityGroupColumns.FieldsNamedExpr(opt.Fields),
		table.SecurityGroupTable, whereExpr, pageExpr)

	details := make([]cloud.SecurityGroupTable, 0)
	if err = s.Orm.Do().Select(kt.Ctx, &details, sql); err != nil {
		return nil, err
	}

	return &types.ListSecurityGroupDetails{Details: details}, nil
}

// DeleteWithTx sg with filter.
func (s SecurityGroupDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SecurityGroupTable, whereExpr)
	if err = s.Orm.Txn(tx).Delete(kt.Ctx, sql); err != nil {
		logs.ErrorJson("delete security group failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}
