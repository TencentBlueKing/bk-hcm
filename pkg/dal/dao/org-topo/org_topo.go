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

// Package orgtopo implements the data service for org topo.
package orgtopo

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/errf"
	idgen "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	orgtable "hcm/pkg/dal/table/org-topo"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/times"

	"github.com/bluele/gcache"
	"github.com/jmoiron/sqlx"
)

// Interface holds all the supported operations for the org topo.
type Interface interface {
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListOrgTopoResult, error)
	ListByDeptIDs(kt *kit.Kit, deptIDs []string) (*types.ListOrgTopoResult, error)
	BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []orgtable.OrgTopo) ([]string, error)
	BatchUpdate(kt *kit.Kit, tx *sqlx.Tx, models []orgtable.OrgTopo) (int64, error)
	BatchDelete(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) (int64, error)
	ListAllDepartment(kt *kit.Kit) (map[string]orgtable.OrgTopo, error)
}

// New create instance.
func New(oi orm.Interface, idGen idgen.IDGenInterface) Interface {
	cache := gcache.New(100).LRU().Expiration(time.Duration(1) * time.Hour).Build()
	return &OrgTopoDao{
		oi:    oi,
		idGen: idGen,
		cache: cache,
	}
}

var _ Interface = new(OrgTopoDao)

// OrgTopoDao org topo dao
type OrgTopoDao struct {
	oi    orm.Interface
	idGen idgen.IDGenInterface
	cache gcache.Cache
}

// List list org topo.
func (otd *OrgTopoDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListOrgTopoResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list org topo but options is nil")
	}

	exprOpt := filter.NewExprOption(
		filter.RuleFields(orgtable.OrgTopoColumns.ColumnTypes()),
	)
	if err := opt.Validate(exprOpt, &core.PageOption{EnableUnlimitedLimit: true}); err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.OrgTopoTable, whereExpr)
		count, cErr := otd.oi.Do().Count(kt.Ctx, sql, whereValue)
		if cErr != nil {
			logs.Errorf("count org topo failed, sql: %s, err: %v, rid: %s", sql, cErr, kt.Rid)
			return nil, cErr
		}
		return &types.ListOrgTopoResult{Count: count}, nil
	}
	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, orgtable.OrgTopoColumns.FieldsNamedExpr(opt.Fields),
		table.OrgTopoTable, whereExpr, pageExpr)
	details := make([]orgtable.OrgTopo, 0)
	if err = otd.oi.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.Errorf("org topo list db failed, sql: %s, error: %v, rid: %s", sql, err, kt.Rid)
		return nil, err
	}

	return &types.ListOrgTopoResult{Details: details}, nil
}

// ListByDeptIDs list by dept ids items.
func (otd *OrgTopoDao) ListByDeptIDs(kt *kit.Kit, deptIDs []string) (*types.ListOrgTopoResult, error) {
	if len(deptIDs) == 0 {
		return &types.ListOrgTopoResult{}, nil
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s WHERE dept_id IN (:ids)`,
		orgtable.OrgTopoColumns.ColumnExpr(), table.OrgTopoTable)

	details := make([]orgtable.OrgTopo, 0)
	if err := otd.oi.Do().Select(kt.Ctx, &details, sql, map[string]any{"ids": deptIDs}); err != nil {
		logs.Errorf("select org topo by deptids failed, sql: %s, err: %v, rid: %s", sql, err, kt.Rid)
		return nil, err
	}

	return &types.ListOrgTopoResult{Details: details}, nil
}

// BatchCreate batch create org topo.
func (otd *OrgTopoDao) BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []orgtable.OrgTopo) ([]string, error) {
	if len(models) == 0 {
		return nil, errf.New(errf.InvalidParameter, "org topo models to create cannot be empty")
	}

	tableName := models[0].TableName()
	ids, err := otd.idGen.Batch(kt, tableName, len(models))
	if err != nil {
		return nil, err
	}

	for index, model := range models {
		if err = model.ValidateInsert(); err != nil {
			return nil, err
		}

		models[index].ID = ids[index]
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, tableName,
		orgtable.OrgTopoColumns.ColumnExpr(), orgtable.OrgTopoColumns.ColonNameExpr())

	if err = otd.oi.Txn(tx).BulkInsert(kt.Ctx, sql, models); err != nil {
		logs.Errorf("insert %s failed, err: %v, rid: %s", tableName, err, kt.Rid)
		return nil, fmt.Errorf("insert %s failed, err: %v", tableName, err)
	}

	return ids, nil
}

// BatchUpdate batch update org topo.
func (otd *OrgTopoDao) BatchUpdate(kt *kit.Kit, tx *sqlx.Tx, models []orgtable.OrgTopo) (int64, error) {
	if len(models) == 0 {
		return 0, errors.New("to be updated org topo is empty")
	}

	for _, one := range models {
		if err := one.ValidateUpdate(); err != nil {
			return 0, errf.NewFromErr(errf.InvalidParameter, err)
		}
	}

	opts := utils.NewFieldOptions().AddBlankedFields("memo").AddIgnoredFields(types.DefaultIgnoredFields...)
	for _, one := range models {
		setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(one, opts)
		if err != nil {
			return 0, fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
		}
		sql := fmt.Sprintf("UPDATE %s %s WHERE id = '%s'", one.TableName(), setExpr, one.ID)
		_, err = otd.oi.Txn(tx).Update(kt.Ctx, sql, toUpdate)
		if err != nil {
			logs.Errorf("batch update org topo failed, model: %s, err: %v, sql: %s, rid: %s", one, err, sql, kt.Rid)
			return 0, err
		}
	}

	return int64(len(models)), nil
}

// BatchDelete batch delete org topo.
func (otd *OrgTopoDao) BatchDelete(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) (int64, error) {
	if expr == nil {
		return 0, errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return 0, err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.OrgTopoTable, whereExpr)
	affectNum, err := otd.oi.Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.Errorf("batch delete org topo failed, err: %v, expr: %v, rid: %s", err, expr, kt.Rid)
		return 0, err
	}

	return affectNum, err
}

// ListAllDepartment 获取所有部门信息
func (otd *OrgTopoDao) ListAllDepartment(kt *kit.Kit) (map[string]orgtable.OrgTopo, error) {
	var allDeptMap = make(map[string]orgtable.OrgTopo)
	val, err := otd.cache.GetIFPresent("cache_hcm_all_org_map")
	if err == nil {
		var ok bool
		allDeptMap, ok = val.(map[string]orgtable.OrgTopo)
		if !ok {
			logs.Errorf("unsupported dept list cache value, type: %v, rid: %s", reflect.TypeOf(val).String(), kt.Rid)
			return nil, fmt.Errorf("unsupported dept list cache value type: %v", reflect.TypeOf(val).String())
		}
		return allDeptMap, nil
	}

	offset := uint32(0)
	limit := uint(constant.DeptQueryUserMgrMaxNum)
	for {
		orgReq := &types.ListOption{
			Filter: tools.ExpressionAnd(tools.RuleNotEqual("dept_id", "")),
			Page: &core.BasePage{
				Start: offset,
				Limit: limit,
				Sort:  "id",
				Order: core.Ascending,
			},
		}
		list, oErr := otd.List(kt, orgReq)
		if oErr != nil {
			logs.Errorf("list all dept from db failed, err: %v, rid: %s", oErr, kt.Rid)
			return nil, oErr
		}

		for _, item := range list.Details {
			allDeptMap[item.DeptID] = item
		}

		if len(list.Details) == 0 {
			break
		}

		offset += uint32(limit)
	}

	err = otd.cache.SetWithExpire("cache_hcm_all_org_map", allDeptMap,
		time.Duration(times.GetTodayRemainDuration())*time.Second)
	if err != nil {
		// 仅记录Warning日志，不影响已获取的DB数据展示
		logs.Warnf("list all dept from db set cache failed, err: %v, rid: %s", err, kt.Rid)
	}

	return allDeptMap, nil
}
