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
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	"hcm/pkg/dal/table/cloud/cvm"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// ListCvmLeftJoinRel ...
func (relDao *DiskCvmRelDao) ListCvmLeftJoinRel(kt *kit.Kit, opt *types.ListOption, notEqualDiskID string) (
	*cloud.CvmLeftJoinDiskCvmRelResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list disk cvm rel options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(cvm.TableColumns.ColumnTypes())),
		core.DefaultPageOption); err != nil {
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
			`SELECT count(*) FROM %s as cvm left join %s as rel on cvm.id = rel.cvm_id %s`,
			table.CvmTable, tablecloud.DiskCvmRelTableName, whereExpr)

		count, err := relDao.Orm().Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count cvm left join disk_cvm_rel failed, err: %v, filter: %s, diskID: %s, rid: %s", err,
				opt.Filter, notEqualDiskID, kt.Rid)
			return nil, err
		}
		return &cloud.CvmLeftJoinDiskCvmRelResult{Count: count}, nil
	}

	sql := fmt.Sprintf(
		`SELECT %s, %s FROM %s as cvm left join %s as rel on cvm.id = rel.cvm_id %s`,
		cvm.TableColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		tools.BaseRelJoinSqlBuild(
			"rel",
			"cvm",
			"id",
			"disk_id",
		),
		table.CvmTable,
		tablecloud.DiskCvmRelTableName,
		whereExpr,
	)

	details := make([]cloud.CvmLeftJoinDiskCvmRel, 0)
	if err := relDao.Orm().Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select cvm left join disk_cvm_rel failed, err: %v, filter: %s, diskID: %s, rid: %s", err,
			opt.Filter, notEqualDiskID, kt.Rid)
		return nil, err
	}

	return &cloud.CvmLeftJoinDiskCvmRelResult{Details: details}, nil
}
