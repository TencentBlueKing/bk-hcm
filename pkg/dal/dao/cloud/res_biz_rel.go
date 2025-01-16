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

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// ResUsageBizRel cloud resource biz relation
type ResUsageBizRel interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*cloud.ResBizRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResBizRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
	// ListUsageBizs 查询指定资源关联业务id，保证返回的数组和传入resIDs的顺序以及数量一致
	ListUsageBizs(kt *kit.Kit, resType enumor.CloudResourceType, resIDs []string) ([]types.ResBizInfo, error)
}

var _ ResUsageBizRel = new(ResUsageBizRelDao)

// ResUsageBizRelDao cloud resource biz relation dao.
type ResUsageBizRelDao struct {
	Orm orm.Interface
}

// BatchCreateWithTx ResUsageBizRel with tx.
func (a ResUsageBizRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*cloud.ResBizRelTable) error {
	if len(rels) == 0 {
		return errf.New(errf.InvalidParameter, "res_biz_rel is required")
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ResBizRelTable,
		cloud.ResBizRelColumns.ColumnExpr(), cloud.ResBizRelColumns.ColonNameExpr())

	err := a.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels)
	if err != nil {
		return fmt.Errorf("insert %s failed, err: %v", table.ResBizRelTable, err)
	}

	return nil
}

// List ResUsageBizRel list.
func (a ResUsageBizRelDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResBizRelDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list res_biz_rel options is nil")
	}

	exprOpt := filter.NewExprOption(filter.RuleFields(cloud.ResBizRelColumns.ColumnTypes()))
	if err := opt.Validate(exprOpt, core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ResBizRelTable, whereExpr)

		count, err := a.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count res_biz_rel failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListResBizRelDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.ResBizRelColumns.FieldsNamedExpr(opt.Fields),
		table.ResBizRelTable, whereExpr, pageExpr)

	details := make([]cloud.ResBizRelTable, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListResBizRelDetails{Count: 0, Details: details}, nil
}

// DeleteWithTx ResUsageBizRel with tx.
func (a ResUsageBizRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ResBizRelTable, whereExpr)
	if _, err := a.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete res_biz_rel failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// ListUsageBizs 查询指定资源关联业务id，保证返回的数组和传入resIDs的顺序以及数量一致
func (a ResUsageBizRelDao) ListUsageBizs(kt *kit.Kit, resType enumor.CloudResourceType, resIDs []string) (
	[]types.ResBizInfo, error) {

	sql := fmt.Sprintf(`SELECT * FROM %s WHERE res_type = :res_type and res_id IN (:res_ids)`, table.ResBizRelTable)
	relTables := make([]cloud.ResBizRelTable, 0)
	args := map[string]interface{}{"res_type": resType, "res_ids": resIDs}
	if err := a.Orm.Do().Select(kt.Ctx, &relTables, sql, args); err != nil {
		logs.Errorf("delete res_biz_rel failed, err: %v, res_type: %s, res_id: %s, rid: %s",
			err, resType, resIDs, kt.Rid)
		return nil, err
	}
	resBizMap := make(map[string][]int64)
	for i := range relTables {
		resBizMap[relTables[i].ResID] = append(resBizMap[relTables[i].ResID], relTables[i].UsageBizID)
	}

	resBizInfos := make([]types.ResBizInfo, len(resIDs))
	for i := range resIDs {
		resBizInfos[i] = types.ResBizInfo{
			ResType: resType,
			ResID:   resIDs[i],
			BizIDs:  resBizMap[resIDs[i]],
		}
	}
	return resBizInfos, nil
}
