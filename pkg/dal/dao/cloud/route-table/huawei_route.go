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

package routetable

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
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	routetable "hcm/pkg/dal/table/cloud/route-table"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// HuaWeiRoute defines huawei route dao operations.
type HuaWeiRoute interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []routetable.HuaWeiRouteTable) ([]string, error)
	Update(kt *kit.Kit, expr *filter.Expression, model *routetable.HuaWeiRouteTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (*types.HuaWeiRouteListResult, error)
	BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
}

var _ HuaWeiRoute = new(huaweiRouteDao)

// huaweiRouteDao huawei route dao.
type huaweiRouteDao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewHuaWeiRouteDao create a huawei route dao.
func NewHuaWeiRouteDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) HuaWeiRoute {
	return &huaweiRouteDao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// BatchCreateWithTx create huawei route with transaction.
func (r *huaweiRouteDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, models []routetable.HuaWeiRouteTable) (
	[]string, error) {

	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "models to create cannot be empty")
	}

	for _, model := range models {
		if err := model.InsertValidate(); err != nil {
			return nil, err
		}
	}

	// generate huawei route id
	ids, err := r.idGen.Batch(kt, table.HuaWeiRouteTable, len(models))
	if err != nil {
		return nil, err
	}

	for idx := range models {
		models[idx].ID = ids[idx]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, models[0].TableName(),
		routetable.HuaWeiRouteColumns.ColumnExpr(), routetable.HuaWeiRouteColumns.ColonNameExpr())

	err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, models)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", models[0].TableName(), err)
	}

	if err = r.batchCreateAudit(kt, tx, models); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *huaweiRouteDao) batchCreateAudit(kt *kit.Kit, tx *sqlx.Tx, routes []routetable.HuaWeiRouteTable) error {
	rtIDMap := make(map[string]struct{}, 0)
	for _, route := range routes {
		rtIDMap[route.RouteTableID] = struct{}{}
	}

	rtIDs := make([]string, 0, len(rtIDMap))
	for id, _ := range rtIDMap {
		rtIDs = append(rtIDs, id)
	}

	idRtMap, err := listRouteTable(kt, r.orm, tx, rtIDs)
	if err != nil {
		return err
	}

	audits := make([]*tableaudit.AuditTable, 0, len(routes))
	for _, route := range routes {
		rt, exist := idRtMap[route.RouteTableID]
		if !exist {
			return errf.Newf(errf.RecordNotFound, "security group: %s not found", route.RouteTableID)
		}

		audits = append(audits, &tableaudit.AuditTable{
			ResID:      rt.ID,
			CloudResID: rt.CloudID,
			ResName:    converter.PtrToVal(rt.Name),
			ResType:    enumor.RouteTableAuditResType,
			Action:     enumor.Update,
			BkBizID:    rt.BkBizID,
			Vendor:     rt.Vendor,
			AccountID:  rt.AccountID,
			Operator:   kt.User,
			Source:     kt.GetRequestSource(),
			Rid:        kt.Rid,
			AppCode:    kt.AppCode,
			Detail: &tableaudit.BasicDetail{
				Data: &tableaudit.ChildResAuditData{
					ChildResType: enumor.RouteAuditResType,
					Action:       enumor.Create,
					ChildRes:     route,
				},
			},
		})
	}

	if err = r.audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return nil
}

// Update huawei routes.
func (r *huaweiRouteDao) Update(kt *kit.Kit, filterExpr *filter.Expression, model *routetable.HuaWeiRouteTable) error {
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

	_, err = r.orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update huawei route failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update huawei route, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// List huawei routes.
func (r *huaweiRouteDao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*types.HuaWeiRouteListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list huawei route options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(routetable.HuaWeiRouteColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
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
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.HuaWeiRouteTable, whereExpr)

		count, err := r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count huawei routes failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.HuaWeiRouteListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, routetable.HuaWeiRouteColumns.FieldsNamedExpr(opt.Fields),
		table.HuaWeiRouteTable, whereExpr, pageExpr)

	details := make([]routetable.HuaWeiRouteTable, 0)
	err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &types.HuaWeiRouteListResult{Details: details}, nil
}

// BatchDeleteWithTx batch delete huawei route with transaction.
func (r *huaweiRouteDao) BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.HuaWeiRouteTable, whereExpr)
	_, err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete huawei route failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
