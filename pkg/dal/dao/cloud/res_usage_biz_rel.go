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
	"hcm/pkg/criteria/constant"
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
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*cloud.ResUsageBizRelTable) error
	List(kt *kit.Kit, opt *types.ListOption) (*types.ListResUsageBizRelDetails, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
	// ListUsageBizs 查询指定资源关联业务id，保证返回的数组和传入resIDs的顺序以及数量一致
	ListUsageBizs(kt *kit.Kit, resType enumor.CloudResourceType, resIDs []string) ([]types.ResBizInfo, error)
	// UpsertUsageBizs 当指定的关联业务不存在时，新增关联；否则不进行任何操作
	UpsertUsageBizs(kt *kit.Kit, tx *sqlx.Tx, resType enumor.CloudResourceType, resID string,
		resVendor enumor.Vendor, resCloudID string, upsertBizIDs []int64) error
}

var _ ResUsageBizRel = new(ResUsageBizRelDao)

// ResUsageBizRelDao cloud resource biz relation dao.
type ResUsageBizRelDao struct {
	Orm orm.Interface
}

// BatchCreateWithTx ResUsageBizRel with tx.
func (a ResUsageBizRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*cloud.ResUsageBizRelTable) error {
	if len(rels) == 0 {
		return errf.New(errf.InvalidParameter, "res_biz_rel is required")
	}

	for i := range rels {
		if rels[i] == nil {
			return fmt.Errorf("res_biz_rel is nil at index %d", i)
		}
		if err := rels[i].InsertValidate(); err != nil {
			return fmt.Errorf("validate res_biz_rel failed at idx %d, err: %w", i, err)
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ResUsageBizRelTable,
		cloud.ResUsageBizRelColumns.ColumnExpr(), cloud.ResUsageBizRelColumns.ColonNameExpr())

	err := a.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels)
	if err != nil {
		return fmt.Errorf("insert %s failed, err: %v", table.ResUsageBizRelTable, err)
	}

	return nil
}

// List ResUsageBizRel list.
func (a ResUsageBizRelDao) List(kt *kit.Kit, opt *types.ListOption) (*types.ListResUsageBizRelDetails, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list res_biz_rel options is nil")
	}

	exprOpt := filter.NewExprOption(filter.RuleFields(cloud.ResUsageBizRelColumns.ColumnTypes()))
	if err := opt.Validate(exprOpt, core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ResUsageBizRelTable, whereExpr)

		count, err := a.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count res_biz_rel failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &types.ListResUsageBizRelDetails{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, cloud.ResUsageBizRelColumns.FieldsNamedExpr(opt.Fields),
		table.ResUsageBizRelTable, whereExpr, pageExpr)

	details := make([]cloud.ResUsageBizRelTable, 0)
	if err = a.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	return &types.ListResUsageBizRelDetails{Count: 0, Details: details}, nil
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

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ResUsageBizRelTable, whereExpr)
	if _, err := a.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete res_biz_rel failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// ListUsageBizs 查询指定资源关联业务id，保证返回的数组和传入resIDs的顺序以及数量一致
func (a ResUsageBizRelDao) ListUsageBizs(kt *kit.Kit, resType enumor.CloudResourceType, resIDs []string) (
	[]types.ResBizInfo, error) {

	sql := fmt.Sprintf(`SELECT * FROM %s WHERE res_type = :res_type and res_id IN (:res_ids) ORDER BY rel_id`,
		table.ResUsageBizRelTable)
	relTables := make([]cloud.ResUsageBizRelTable, 0)
	args := map[string]interface{}{"res_type": resType, "res_ids": resIDs}
	if err := a.Orm.Do().Select(kt.Ctx, &relTables, sql, args); err != nil {
		logs.Errorf("list res_biz_rel failed, err: %v, res_type: %s, res_id: %s, rid: %s",
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

// UpsertUsageBizs 不存在时新增指定资源的关联业务
func (a ResUsageBizRelDao) UpsertUsageBizs(kt *kit.Kit, tx *sqlx.Tx, resType enumor.CloudResourceType, resID string,
	resVendor enumor.Vendor, resCloudID string, upsertBizIDs []int64) error {

	sql := fmt.Sprintf(`SELECT * FROM %s WHERE res_type = :res_type and res_id = :res_id`,
		table.ResUsageBizRelTable)
	relTables := make([]cloud.ResUsageBizRelTable, 0)
	args := map[string]interface{}{"res_type": resType, "res_id": resID}
	if err := a.Orm.Txn(tx).Select(kt.Ctx, &relTables, sql, args); err != nil {
		logs.Errorf("list res_biz_rel failed, err: %v, res_type: %s, res_id: %s, rid: %s",
			err, resType, resID, kt.Rid)
		return err
	}

	existsBizIDs := make(map[int64]interface{})
	for i := range relTables {
		// 使用业务包含全部业务时，跳过
		if relTables[i].UsageBizID == constant.AttachedAllBiz {
			return nil
		}
		existsBizIDs[relTables[i].UsageBizID] = struct{}{}
	}

	insertRels := make([]cloud.ResUsageBizRelTable, 0)
	for _, bizID := range upsertBizIDs {
		if _, ok := existsBizIDs[bizID]; !ok {
			insertRels = append(insertRels, cloud.ResUsageBizRelTable{
				UsageBizID: bizID,
				ResID:      resID,
				ResType:    resType,
				ResVendor:  resVendor,
				ResCloudID: resCloudID,
				RelCreator: kt.User,
			})
		}
	}

	if len(insertRels) == 0 {
		return nil
	}

	insertSql := fmt.Sprintf(`INSERT INTO %s (%s)  VALUES(%s)`, table.ResUsageBizRelTable,
		cloud.ResUsageBizRelColumns.ColumnExpr(), cloud.ResUsageBizRelColumns.ColonNameExpr())

	err := a.Orm.Txn(tx).BulkInsert(kt.Ctx, insertSql, insertRels)
	if err != nil {
		return fmt.Errorf("upsert %s failed, err: %v", table.ResUsageBizRelTable, err)
	}

	return nil
}
