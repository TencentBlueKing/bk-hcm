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

// Package disk ...
package disk

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
	"hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/dal/table"
	tableaudit "hcm/pkg/dal/table/audit"
	"hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Disk only used for disk.
type Disk interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, disks []*disk.DiskModel) ([]string, error)
	Update(kt *kit.Kit, filterExpr *filter.Expression, updateData *disk.DiskModel) error
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, diskID string, updateData *disk.DiskModel) error
	List(kt *kit.Kit, opt *types.ListOption) (*cloud.DiskListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
	Count(kt *kit.Kit, opt *types.CountOption) (*cloud.DiskCountResult, error)
}

var _ Disk = new(DiskDao)

// DiskDao disk dao.
type DiskDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx 批量创建云盘数据
func (diskDao DiskDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, disks []*disk.DiskModel) ([]string, error) {
	if len(disks) == 0 {
		return nil, errf.New(errf.InvalidParameter, "disk model data is required")
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.DiskTable, disk.DiskColumns.ColumnExpr(),
		disk.DiskColumns.ColonNameExpr(),
	)

	ids, err := diskDao.IDGen.Batch(kt, table.DiskTable, len(disks))
	if err != nil {
		return nil, err
	}

	for idx, d := range disks {
		d.ID = ids[idx]
	}

	err = diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, disks)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.DiskTable, err)
	}

	// create audit.
	audits := make([]*tableaudit.AuditTable, 0, len(disks))
	for _, one := range disks {
		audits = append(audits, &tableaudit.AuditTable{
			ResID:      one.ID,
			CloudResID: one.CloudID,
			ResName:    one.Name,
			ResType:    enumor.DiskAuditResType,
			Action:     enumor.Create,
			BkBizID:    one.BkBizID,
			Vendor:     enumor.Vendor(one.Vendor),
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
	if err = diskDao.Audit.BatchCreateWithTx(kt, tx, audits); err != nil {
		logs.Errorf("batch create audit failed, err: %v, rid: %s", err, kt.Rid)
		return nil, err
	}

	return ids, nil
}

// Update 更新云盘信息
func (diskDao DiskDao) Update(kt *kit.Kit, filterExpr *filter.Expression, updateData *disk.DiskModel) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddBlankedFields("is_system_disk", "memo").
		AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, table.DiskTable, setExpr, whereExpr)

	_, err = diskDao.Orm.AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(txn).Update(
			kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update disk failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update disk, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// UpdateByIDWithTx 根据 ID 更新单条数据
func (diskDao DiskDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, diskID string, updateData *disk.DiskModel) error {
	opts := utils.NewFieldOptions().AddBlankedFields("is_system_disk", "memo").
		AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, table.DiskTable, setExpr)

	toUpdate["id"] = diskID
	_, err = diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update disk failed, err: %v, id: %s, rid: %v", err, diskID, kt.Rid)
		return err
	}

	return nil
}

// List 根据条件查询云盘列表
func (diskDao DiskDao) List(kt *kit.Kit, opt *types.ListOption) (*cloud.DiskListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list disk options is nil")
	}

	columnTypes := disk.DiskColumns.ColumnTypes()
	columnTypes["extension.resource_group_name"] = enumor.String
	columnTypes["extension.self_link"] = enumor.String
	columnTypes["extension.zones"] = enumor.Json
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereOpt := tools.DefaultSqlWhereOption
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.DiskTable, whereExpr)
		count, err := diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql,
			whereValue)
		if err != nil {
			logs.ErrorJson("count disk failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.DiskListResult{Count: count}, nil
	}
	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, disk.DiskColumns.FieldsNamedExpr(opt.Fields), table.DiskTable,
		whereExpr, pageExpr)

	details := make([]*disk.DiskModel, 0)
	err = diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &details, sql,
		whereValue)
	if err != nil {
		return nil, err
	}

	result := &cloud.DiskListResult{Details: details}

	return result, nil
}

// DeleteWithTx 删除云盘
func (diskDao DiskDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.DiskTable, whereExpr)
	_, err = diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete disk failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// Count 根据条件统计云盘数量
func (diskDao DiskDao) Count(kt *kit.Kit, opt *types.CountOption) (*cloud.DiskCountResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "count disk options is nil")
	}

	exprOption := filter.NewExprOption(filter.RuleFields(disk.DiskColumns.ColumnTypes()))
	if err := opt.Validate(exprOption); err != nil {
		return nil, err
	}
	whereOpt := tools.DefaultSqlWhereOption
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.DiskTable, whereExpr)
	count, err := diskDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
	if err != nil {
		return nil, err
	}
	return &cloud.DiskCountResult{Count: count}, nil
}

// ListByIDs ...
func ListByIDs(kt *kit.Kit, ormi orm.Interface, ids []string) (map[string]disk.DiskModel, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s where id in (:ids)`, disk.DiskColumns.FieldsNamedExpr(nil), table.DiskTable)
	disks := make([]disk.DiskModel, 0)
	if err := ormi.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(kt.Ctx, &disks, sql,
		map[string]interface{}{"ids": ids}); err != nil {
		return nil, err
	}

	idToDiskMap := make(map[string]disk.DiskModel, len(ids))
	for _, d := range disks {
		idToDiskMap[d.ID] = d
	}
	return idToDiskMap, nil
}
