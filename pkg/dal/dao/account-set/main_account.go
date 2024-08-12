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

package accountset

import (
	"fmt"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	tableaccountset "hcm/pkg/dal/table/account-set"
	tableaudit "hcm/pkg/dal/table/audit"
	tabletype "hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// MainAccount only used for main account.
type MainAccount interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, account *tableaccountset.MainAccountTable) (string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tableaccountset.MainAccountTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListMainAccountDetails, error)
}

var _ MainAccount = new(MainAccountDao)

// MainAccountDao main
type MainAccountDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx create main account with tx.
func (a MainAccountDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, model *tableaccountset.MainAccountTable) (string,
	error) {
	if err := model.InsertValidate(); err != nil {
		return "", err
	}

	id, err := a.IDGen.One(kt, table.MainAccountTable)
	if err != nil {
		return "", err
	}
	model.ID = id

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`,
		model.TableName(), tableaccountset.MainAccountColumns.ColumnExpr(),
		tableaccountset.MainAccountColumns.ColonNameExpr())

	err = a.Orm.Txn(tx).Insert(kt.Ctx, sql, model)
	if err != nil {
		return "", fmt.Errorf("insert %s failed, err: %v", model.TableName(), err)
	}

	// create audit.
	extension := tools.MainAccountExtensionRemoveSecretKey(string(model.Extension))
	model.Extension = tabletype.JsonField(extension)

	auditInfo := &tableaudit.AuditTable{
		ResID:     model.CloudID,
		ResName:   model.Name,
		ResType:   enumor.MainAccountAuditResType,
		Action:    enumor.Create,
		Vendor:    enumor.Vendor(model.Vendor),
		AccountID: model.ID,
		Operator:  kt.User,
		Source:    kt.GetRequestSource(),
		Rid:       kt.Rid,
		AppCode:   kt.AppCode,
		Detail: &tableaudit.BasicDetail{
			Data: model,
		},
	}
	if err = a.Audit.Create(kt, auditInfo); err != nil {
		logs.Errorf("create account audit failed, err: %v, rid: %s", err, kt.Rid)
		return "", err
	}

	return id, nil
}

// Update accounts.
func (a MainAccountDao) Update(kt *kit.Kit, filterExpr *filter.Expression,
	model *tableaccountset.MainAccountTable) error {
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

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(model, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, model.TableName(), setExpr, whereExpr)

	_, err = a.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := a.Orm.Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update account failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update account, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List list main accounts.
func (ma MainAccountDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListMainAccountDetails, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list account options is nil")
	}
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.MainAccountTable, whereExpr)

		count, err := ma.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count accounts failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListMainAccountDetails{Count: count}, nil
	}
	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableaccountset.MainAccountColumns.FieldsNamedExpr(opt.Fields),
		table.MainAccountTable, whereExpr, pageExpr)

	details := make([]*tableaccountset.MainAccountTable, 0)
	if err = ma.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListMainAccountDetails{Count: 0, Details: details}, nil
}
