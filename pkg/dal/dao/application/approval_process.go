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

package application

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/application"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ApprovalProcess ...
type ApprovalProcess interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *application.ApprovalProcessTable) (string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *application.ApprovalProcessTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListApprovalProcessDetails, error)
}

var _ ApprovalProcess = new(ApprovalProcessDao)

// ApprovalProcessDao approval process dao.
type ApprovalProcessDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
}

// CreateWithTx ...
func (a *ApprovalProcessDao) CreateWithTx(
	kt *kit.Kit, tx *sqlx.Tx, model *application.ApprovalProcessTable,
) (string, error) {
	if err := model.InsertValidate(); err != nil {
		return "", err
	}

	// generate approval process id
	id, err := a.IDGen.One(kt, table.ApprovalProcessTable)
	if err != nil {
		return "", err
	}
	model.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`,
		model.TableName(), application.ApprovalProcessColumns.ColumnExpr(),
		application.ApprovalProcessColumns.ColonNameExpr(),
	)

	err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Insert(kt.Ctx, sql, model)
	if err != nil {
		return "", fmt.Errorf("insert %s failed, err: %v", model.TableName(), err)
	}
	return id, nil
}

// Update ...
func (a *ApprovalProcessDao) Update(
	kt *kit.Kit, filterExpr *filter.Expression, model *application.ApprovalProcessTable,
) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = a.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update approval process failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update approval process, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List ...
func (a *ApprovalProcessDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListApprovalProcessDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list approval process options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(application.ApprovalProcessColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ApprovalProcessTable, whereExpr)

		count, err := a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count approval process failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListApprovalProcessDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, application.ApprovalProcessColumns.FieldsNamedExpr(opt.Fields),
		table.ApprovalProcessTable, whereExpr, pageExpr)

	details := make([]*application.ApprovalProcessTable, 0)
	err = a.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.ListApprovalProcessDetails{Count: 0, Details: details}, nil
}
