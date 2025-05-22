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

package networkinterface

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
	typesni "hcm/pkg/dal/dao/types/network-interface"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	tableni "hcm/pkg/dal/table/cloud/network-interface"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// NetworkInterface only used for network interface.
type NetworkInterface interface {
	CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, regions []tableni.NetworkInterfaceTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *tableni.NetworkInterfaceTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*typesni.ListNetworkInterfaceDetails, error)
	ListAssociate(kt *kit.Kit, opt *types.ListOption, isAssociate bool) (
		*types.ListCvmRelsJoinNetworkInterfaceDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ NetworkInterface = new(NetworkInterfaceDao)

// NetworkInterfaceDao network interface dao.
type NetworkInterfaceDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// CreateWithTx network interface with tx.
func (n NetworkInterfaceDao) CreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []tableni.NetworkInterfaceTable) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	ids, err := n.IDGen.Batch(kt, models[0].TableName(), len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		models[index].ID = ids[index]

		if err := model.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		tableni.NetworkInterfaceColumns.ColumnExpr(), tableni.NetworkInterfaceColumns.ColonNameExpr())
	err = n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", models[0].TableName(), err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(models))
	for _, one := range models {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ID,
			CloudResID: one.CloudID,
			ResName:    one.Name,
			ResType:    enumor.NetworkInterfaceAuditResType,
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
	if err = n.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update update network interface.
func (n NetworkInterfaceDao) Update(kt *kit.Kit, filterExpr *filter.Expression,
	model *tableni.NetworkInterfaceTable) error {

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

	_, err = n.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update network interface failed, filter: %s, err: %v, rid: %v",
				filterExpr, err, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update network interface, but record not found, filter: %v, rid: %v",
				filterExpr, kt.Rid)
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List get network interface list.
func (n NetworkInterfaceDao) List(kt *kit.Kit, opt *types.ListOption) (*typesni.ListNetworkInterfaceDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list network interface options is nil")
	}

	columnTypes := tableni.NetworkInterfaceColumns.ColumnTypes()
	columnTypes["extension.self_link"] = enumor.String
	columnTypes["extension.security_group_id"] = enumor.String
	columnTypes["extension.resource_group_name"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.NetworkInterfaceTable, whereExpr)
		count, err := n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count network interface failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &typesni.ListNetworkInterfaceDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, tableni.NetworkInterfaceColumns.FieldsNamedExpr(opt.Fields),
		table.NetworkInterfaceTable, whereExpr, pageExpr)

	details := make([]tableni.NetworkInterfaceTable, 0)
	err = n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}
	return &typesni.ListNetworkInterfaceDetails{Details: details}, nil
}

// ListAssociate get network interface associate list.
func (n NetworkInterfaceDao) ListAssociate(kt *kit.Kit, opt *types.ListOption, isAssociate bool) (
	*types.ListCvmRelsJoinNetworkInterfaceDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list network interface associate options is nil")
	}

	columnTypes := tableni.NetworkInterfaceColumns.ColumnTypes()
	columnTypes["extension.self_link"] = enumor.String
	columnTypes["extension.security_group_id"] = enumor.String
	columnTypes["extension.resource_group_name"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	// 查询未关联主机的网络接口
	joinDirection := "LEFT JOIN"
	if isAssociate == false {
		whereExpr += " AND rel.cvm_id IS NULL"
	} else {
		joinDirection = "INNER JOIN"
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(
			`SELECT COUNT(*) FROM %s AS ni %s %s AS rel ON rel.network_interface_id = ni.id %s`,
			table.NetworkInterfaceTable,
			joinDirection,
			table.NetworkInterfaceCvmRelTable,
			whereExpr,
		)
		count, err := n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count network interface failed, err: %v, filter: %s, rid: %s",
				err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListCvmRelsJoinNetworkInterfaceDetails{Count: converter.ValToPtr(count)}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, &types.PageSQLOption{
		Sort: types.SortOption{Sort: "ni.id", IfNotPresent: true}})
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(
		`SELECT %s,%s FROM %s AS ni %s %s AS rel ON rel.network_interface_id = ni.id %s %s`,
		tableni.NetworkInterfaceColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		tools.BaseRelJoinSqlBuild("rel", "ni", "id", "cvm_id"),
		table.NetworkInterfaceTable,
		joinDirection,
		table.NetworkInterfaceCvmRelTable,
		whereExpr, pageExpr,
	)

	details := make([]*types.NetworkInterfaceWithCvmID, 0)
	err = n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.ListCvmRelsJoinNetworkInterfaceDetails{Details: details}, nil
}

// DeleteWithTx network interface with tx.
func (n NetworkInterfaceDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error {
	if expr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.NetworkInterfaceTable, whereExpr)
	_, err = n.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete azure network interface failed, err: %v, filter: %s, rid: %s", err, expr, kt.Rid)
		return err
	}

	return nil
}

// ListNetworkInterface list network interface
func ListNetworkInterface(kt *kit.Kit, ormi orm.Interface, ids []string) (
	map[string]tableni.NetworkInterfaceTable, error) {

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE id IN (:ids)`,
		tableni.NetworkInterfaceColumns.FieldsNamedExpr(nil), table.NetworkInterfaceTable)

	list := make([]tableni.NetworkInterfaceTable, 0)
	err := ormi.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(
		kt.Ctx, &list, sql, map[string]interface{}{"ids": ids})
	if err != nil {
		return nil, err
	}

	idMap := make(map[string]tableni.NetworkInterfaceTable, len(ids))
	for _, sg := range list {
		idMap[sg.ID] = sg
	}

	return idMap, nil
}
