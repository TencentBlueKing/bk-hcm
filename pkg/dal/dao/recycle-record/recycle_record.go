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

package recyclerecord

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
	rrtypes "hcm/pkg/dal/dao/types/recycle-record"
	"hcm/pkg/dal/table"
	rr "hcm/pkg/dal/table/recycle-record"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// RecycleRecord defines recycle record dao operations.
type RecycleRecord interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, records []rr.RecycleRecordTable) (string, error)
	Update(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression, record *rr.RecycleRecordTable) error
	List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
		*rrtypes.RecycleRecordListResult, error)
	BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, expr *filter.Expression) error
	UpdateResource(kt *kit.Kit, tx *sqlx.Tx, opt *rrtypes.ResourceUpdateOptions) error
	ListResourceInfo(kt *kit.Kit, resType enumor.CloudResourceType, ids []string) ([]rrtypes.RecycleResourceInfo, error)
}

var _ RecycleRecord = new(Dao)

// Dao recycle record dao.
type Dao struct {
	orm   orm.Interface
	idGen idgenerator.IDGenInterface
	audit audit.Interface
}

// NewRecycleRecordDao create a recycle record dao.
func NewRecycleRecordDao(orm orm.Interface, idGen idgenerator.IDGenInterface, audit audit.Interface) RecycleRecord {
	return &Dao{
		orm:   orm,
		idGen: idGen,
		audit: audit,
	}
}

// BatchCreateWithTx create recycle record with transaction.
func (r *Dao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, records []rr.RecycleRecordTable) (string, error) {
	if len(records) == 0 {
		return "", errf.New(errf.InvalidParameter, "records to create cannot be empty")
	}

	// generate task id
	taskID, err := r.idGen.One(kt, table.RecycleRecordTableTaskID)
	if err != nil {
		return "", err
	}

	ids, err := r.idGen.Batch(kt, table.RecycleRecordTable, len(records))
	if err != nil {
		return "", err
	}

	for idx := range records {
		records[idx].ID = ids[idx]
		records[idx].TaskID = taskID
		if err = records[idx].InsertValidate(); err != nil {
			return "", err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, records[0].TableName(),
		rr.RecycleRecordColumns.ColumnExpr(), rr.RecycleRecordColumns.ColonNameExpr())

	err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, records)
	if err != nil {
		return "", fmt.Errorf("insert %s failed, err: %v", records[0].TableName(), err)
	}

	return taskID, nil
}

// Update recycle records.
func (r *Dao) Update(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression, record *rr.RecycleRecordTable) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := record.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(record, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, record.TableName(), setExpr, whereExpr)

	effected, err := r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(
		kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
	if err != nil {
		logs.ErrorJson("update recycle record failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
		return err
	}

	if effected == 0 {
		logs.ErrorJson("update recycle record, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// List recycle records.
func (r *Dao) List(kt *kit.Kit, opt *types.ListOption, whereOpts ...*filter.SQLWhereOption) (
	*rrtypes.RecycleRecordListResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list recycle record options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(rr.RecycleRecordColumns.ColumnTypes())),
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.RecycleRecordTable, whereExpr)

		count, err := r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count recycle records failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}

		return &rrtypes.RecycleRecordListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, rr.RecycleRecordColumns.FieldsNamedExpr(opt.Fields),
		table.RecycleRecordTable, whereExpr, pageExpr)

	details := make([]rr.RecycleRecordTable, 0)
	err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql, whereValue)
	if err != nil {
		return nil, err
	}

	return &rrtypes.RecycleRecordListResult{Details: details}, nil
}

// BatchDeleteWithTx batch delete recycle record with transaction.
func (r *Dao) BatchDeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.RecycleRecordTable, whereExpr)
	_, err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete recycle record failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// UpdateResource 更新资源回收相关信息，目前只更新 recycle_status
func (r *Dao) UpdateResource(kt *kit.Kit, tx *sqlx.Tx, opt *rrtypes.ResourceUpdateOptions) error {
	tableName, err := opt.ResType.ConvTableName()
	if err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	sql := fmt.Sprintf(`update %s set recycle_status = :recycle_status where id in (:id)`,
		tableName)
	updateData := map[string]interface{}{
		"recycle_status": opt.Status,
	}
	whereValue := map[string]interface{}{
		"id": opt.IDs,
	}

	effected, err := r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(
		kt.Ctx, sql, tools.MapMerge(updateData, whereValue))
	if err != nil {
		logs.ErrorJson("update resource failed, err: %v, table: %s, ids: %+v, rid: %v", err, tableName, opt.IDs, kt.Rid)
		return err
	}

	if effected != int64(len(opt.IDs)) {
		logs.ErrorJson("update count %d is invalid, err: %v, table: %s, ids: %+v, rid: %v", effected, err, tableName,
			opt.IDs, kt.Rid)
		return errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
	}

	return nil
}

// ListResourceInfo list recycle resource info to generate recycle record.
func (r *Dao) ListResourceInfo(kt *kit.Kit, resType enumor.CloudResourceType, ids []string) (
	[]rrtypes.RecycleResourceInfo, error) {

	tableName, err := resType.ConvTableName()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(ids) == 0 {
		return nil, errf.New(errf.InvalidParameter, "ids is required")
	}

	sql := fmt.Sprintf("select vendor, id, cloud_id, name, bk_biz_id, account_id, region from %s where id in (:id)",
		tableName)

	info := make([]rrtypes.RecycleResourceInfo, 0)
	args := map[string]interface{}{
		"id": ids,
	}
	err = r.orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &info, sql, args)
	if err != nil {
		logs.Errorf("list recycle resource info failed, err: %v, type: %s, ids: %v, rid: %s", err, resType, ids, kt.Rid)
		return nil, err
	}

	if len(info) != len(ids) {
		return nil, errf.New(errf.InvalidParameter, "not all recycle resource exists")
	}

	return info, nil
}
