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

package diskcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/cloud/cvm"
	"hcm/pkg/dal/dao/cloud/disk"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	tabledisk "hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// DiskCvmRel only used for DiskCvmRel.
type DiskCvmRel interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*tablecloud.DiskCvmRelModel) error
	List(kt *kit.Kit, opt *types.ListOption) (*cloud.DiskCvmRelListResult, error)
	ListJoinDisk(kt *kit.Kit, cvmIDs []string) (*cloud.DiskCvmRelJoinDiskListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
	ListCvmIDLeftJoinRel(kt *kit.Kit, opt *types.ListOption, notEqualDiskID string) (
		*cloud.CvmLeftJoinDiskCvmRelResult, error)
	ListDiskLeftJoinRel(kt *kit.Kit, opt *types.ListOption) (
		*cloud.DiskLeftJoinDiskCvmRelResult, error)
	insertValidate(kt *kit.Kit, rels []*tablecloud.DiskCvmRelModel) error
}

var _ DiskCvmRel = new(DiskCvmRelDao)

// DiskCvmRelDao DiskCvmRelDao dao.
type DiskCvmRelDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx ...
func (relDao DiskCvmRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*tablecloud.DiskCvmRelModel) error {
	if err := relDao.insertValidate(kt, rels); err != nil {
		return err
	}

	sql := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES(%s)`,
		table.DiskCvmRelTableName,
		tablecloud.DiskCvmRelColumns.ColumnExpr(),
		tablecloud.DiskCvmRelColumns.ColonNameExpr(),
	)
	if err := relDao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels); err != nil {
		logs.Errorf("batch create disk cvm rels failed, err: %v, rels: %v, rid: %s", err, rels, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", table.DiskCvmRelTableName, err)
	}

	return nil
}

// List ...
func (relDao DiskCvmRelDao) List(kt *kit.Kit, opt *types.ListOption) (*cloud.DiskCvmRelListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list disk cvm rel options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablecloud.DiskCvmRelColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf(
			"gen where expr for list disk cvm rels failed, err: %v, filter: %s, rid: %s",
			err,
			opt.Filter,
			kt.Rid,
		)
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.DiskCvmRelTableName, whereExpr)
		count, err := relDao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count disk cvm rels failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.DiskCvmRelListResult{Count: &count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		logs.Errorf(
			"gen page expr for list disk cvm rels failed, err: %v, filter: %s, rid: %s",
			err,
			opt.Filter,
			kt.Rid,
		)
		return nil, err
	}

	sql := fmt.Sprintf(
		`SELECT %s FROM %s %s %s`,
		tablecloud.DiskCvmRelColumns.FieldsNamedExpr(opt.Fields),
		table.DiskCvmRelTableName,
		whereExpr,
		pageExpr,
	)

	details := make([]*tablecloud.DiskCvmRelModel, 0)
	if err = relDao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.Errorf("list disk cvm rels failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
		return nil, err
	}
	return &cloud.DiskCvmRelListResult{Details: details}, nil
}

// ListJoinDisk ...
func (relDao DiskCvmRelDao) ListJoinDisk(kt *kit.Kit, cvmIDs []string) (*cloud.DiskCvmRelJoinDiskListResult, error) {
	if len(cvmIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "cvm ids is required")
	}

	sql := fmt.Sprintf(
		`SELECT %s, %s FROM %s as rel left join %s as disk on rel.disk_id = disk.id where cvm_id in (:cvm_ids)`,
		tabledisk.DiskColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		tools.BaseRelJoinSqlBuild(
			"rel",
			"disk",
			"id",
			"cvm_id",
		),
		table.DiskCvmRelTableName,
		table.DiskTable,
	)

	details := make([]*cloud.DiskWithCvmID, 0)
	if err := relDao.Orm.Do().Select(kt.Ctx, &details, sql, map[string]interface{}{"cvm_ids": cvmIDs}); err != nil {
		logs.ErrorJson("select disk cvm rels join disk failed, err: %v, sql: (%s), rid: %s", err, sql, kt.Rid)
		return nil, err
	}

	return &cloud.DiskCvmRelJoinDiskListResult{Details: details}, nil
}

// DeleteWithTx ...
func (relDao DiskCvmRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.DiskCvmRelTableName, whereExpr)
	if _, err = relDao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete disk cvm rels failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// insertValidate 校验待创建的关联关系表中, 对应的云盘和 CVM 是否存在
func (relDao DiskCvmRelDao) insertValidate(kt *kit.Kit, rels []*tablecloud.DiskCvmRelModel) error {
	relCount := len(rels)

	diskIDs := make([]string, relCount)
	cvmIDs := make([]string, relCount)
	for idx, rel := range rels {
		diskIDs[idx] = rel.DiskID
		cvmIDs[idx] = rel.CvmID
	}

	idToDiskMap, err := disk.ListByIDs(kt, relDao.Orm, diskIDs)
	if err != nil {
		logs.Errorf("list disk by ids failed, err: %v, ids: %v, rid: %s", err, diskIDs, kt.Rid)
		return err
	}

	if len(idToDiskMap) != len(converter.StringSliceToMap(diskIDs)) {
		// TODO 将不存在的 ID 记录到错误信息中
		return fmt.Errorf("some disk does not exists")
	}

	idToCvmMap, err := cvm.ListCvm(kt, relDao.Orm, cvmIDs)
	if err != nil {
		logs.Errorf("list cvm by ids failed, err: %v, ids: %v, rid: %s", err, cvmIDs, kt.Rid)
		return err
	}

	if len(idToCvmMap) != len(converter.StringSliceToMap(cvmIDs)) {
		// TODO 将不存在的 ID 记录到错误信息中
		return fmt.Errorf("some cvm does not exists")
	}

	return nil
}
