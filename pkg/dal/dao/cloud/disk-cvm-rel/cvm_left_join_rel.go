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
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud/cvm"
	"hcm/pkg/dal/table/cloud/disk"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// ListCvmIDLeftJoinRel ...
func (relDao DiskCvmRelDao) ListCvmIDLeftJoinRel(kt *kit.Kit, opt *types.ListOption, notEqualDiskID string) (
	*cloud.CvmLeftJoinDiskCvmRelResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list disk cvm rel options is nil")
	}

	columnTypes := cvm.TableColumns.ColumnTypes()
	columnTypes["extension.zones"] = enumor.Json
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)), core.NewDefaultPageOption()); err != nil {
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

	if len(notEqualDiskID) != 0 {
		whereValue["disk_id"] = notEqualDiskID
		whereExpr += " and disk_id != :disk_id"
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(
			`SELECT count(distinct(cvm.id)) FROM %s as cvm left join %s as rel on cvm.id = rel.cvm_id %s`,
			table.CvmTable, table.DiskCvmRelTableName, whereExpr)

		count, err := relDao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count cvm left join disk_cvm_rel failed, err: %v, filter: %s, diskID: %s, rid: %s", err,
				opt.Filter, notEqualDiskID, kt.Rid)
			return nil, err
		}
		return &cloud.CvmLeftJoinDiskCvmRelResult{Count: count}, nil
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
		`SELECT cvm.id as id, %s FROM %s as cvm left join %s as rel on cvm.id = rel.cvm_id %s group by cvm.id %s`,
		cvm.TableColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		table.CvmTable,
		table.DiskCvmRelTableName,
		whereExpr,
		pageExpr,
	)

	details := make([]cvm.Table, 0)
	if err := relDao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select cvm left join disk_cvm_rel failed, err: %v, filter: %s, diskID: %s, rid: %s", err,
			opt.Filter, notEqualDiskID, kt.Rid)
		return nil, err
	}

	return &cloud.CvmLeftJoinDiskCvmRelResult{Details: details}, nil
}

// ListDiskLeftJoinRel ...
func (relDao *DiskCvmRelDao) ListDiskLeftJoinRel(kt *kit.Kit, opt *types.ListOption) (
	*cloud.DiskLeftJoinDiskCvmRelResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list disk cvm rel options is nil")
	}

	columnTypes := disk.DiskColumns.ColumnTypes()
	columnTypes["extension.resource_group_name"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
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

	whereExpr += " and rel.cvm_id is NULL"

	if opt.Page.Count {
		sql := fmt.Sprintf(
			`SELECT count(distinct(disk.id)) FROM %s as disk left join %s as rel on disk.id = rel.disk_id %s`,
			table.DiskTable, table.DiskCvmRelTableName, whereExpr)

		count, err := relDao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count cvm left join disk_cvm_rel failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.DiskLeftJoinDiskCvmRelResult{Count: count}, nil
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
		`SELECT disk.id as id, %s FROM %s as disk left join %s as rel on disk.id = rel.disk_id %s group by disk.id %s`,
		disk.DiskColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		table.DiskTable,
		table.DiskCvmRelTableName,
		whereExpr,
		pageExpr,
	)

	details := make([]cloud.DiskLeftJoinDiskCvmRel, 0)
	if err := relDao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select disk left join disk_cvm_rel failed, err: %v, filter: %s, rid: %s", err,
			opt.Filter, kt.Rid)
		return nil, err
	}

	return &cloud.DiskLeftJoinDiskCvmRelResult{Details: details}, nil
}
