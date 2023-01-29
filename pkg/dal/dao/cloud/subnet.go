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

// Subnet defines subnet dao operations.
type Subnet interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []cloud.SubnetTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *cloud.SubnetTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*types.SubnetListResult, error)
	BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ Subnet = new(subnetDao)

// subnetDao subnet dao.
type subnetDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
}

// NewSubnetDao create a subnet dao.
func NewSubnetDao(orm orm.Interface, idGen idgenerator.IDGenInterface) Subnet {
	return &subnetDao{
		orm:   orm,
		idGen: idGen,
	}
}

// BatchCreateWithTx create subnet with transaction.
func (v *subnetDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []cloud.SubnetTable) ([]string, error) {
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	for idx := range models {
		if models[idx].Ipv4Cidr == nil {
			models[idx].Ipv4Cidr = make([]string, 0)
		}

		if models[idx].Ipv6Cidr == nil {
			models[idx].Ipv6Cidr = make([]string, 0)
		}

		if err := models[idx].InsertValidate(); err != nil {
			return nil, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	// generate subnet id
	ids, err := v.idGen.Batch(kt, table.SubnetTable, len(models))
	if err != nil {
		return nil, err
	}

	for idx := range models {
		models[idx].ID = ids[idx]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(), cloud.SubnetColumns.ColumnExpr(),
		cloud.SubnetColumns.ColonNameExpr())

	err = v.orm.Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	return ids, nil
}

// Update subnets.
func (v *subnetDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *cloud.SubnetTable) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := model.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("name", "ipv6_cidr", "memo").
		AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = v.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := v.orm.Txn(txn).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.ErrorJson("update subnet failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update subnet, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List subnets.
func (v *subnetDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*types.SubnetListResult,
	error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list subnet options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cloud.SubnetColumns.ColumnTypes())),
		types.DefaultPageOption); err != nil {
		return nil, err
	}

	whereOpt := tools.DefaultSqlWhereOption
	if len(whereOpts) != 0 && whereOpts[0] != nil {
		err := whereOpts[0].Validate()
		if err != nil {
			return nil, err
		}
		whereOpt = whereOpts[0]
	}
	whereExpr, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.SubnetTable, whereExpr)

		count, err := v.orm.Do().Count(kt.Ctx, sql)
		if err != nil {
			logs.ErrorJson("count subnets failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.SubnetListResult{Count: count}, nil
	}

	pageExpr, err := opt.Page.SQLExpr(types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.SubnetColumns.FieldsNamedExpr(opt.Fields), table.SubnetTable,
		whereExpr, pageExpr)

	details := make([]cloud.SubnetTable, 0)
	if err = v.orm.Do().Select(kt.Ctx, &details, sql); err != nil {
		return nil, err
	}

	return &types.SubnetListResult{Details: details}, nil
}

// BatchDeleteWithTx batch delete subnet with transaction.
func (v *subnetDao) BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.SubnetTable, whereExpr)
	if err = v.orm.Txn(tx).Delete(kt.Ctx, sql); err != nil {
		logs.ErrorJson("delete subnet failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
