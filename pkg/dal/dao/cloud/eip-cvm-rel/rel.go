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

package eipcvmrel

import (
	"fmt"

	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/audit"
	"hcm/pkg/dal/dao/cloud/cvm"
	"hcm/pkg/dal/dao/cloud/eip"
	idgenerator "hcm/pkg/dal/dao/id-generator"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/dal/table"
	tablecloud "hcm/pkg/dal/table/cloud"
	tableeip "hcm/pkg/dal/table/cloud/eip"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/converter"

	"github.com/jmoiron/sqlx"
)

// EipCvmRel only used for EipCvmRel.
type EipCvmRel interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*tablecloud.EipCvmRelModel) error
	List(kt *kit.Kit, opt *types.ListOption) (*cloud.EipCvmRelListResult, error)
	ListJoinEip(kt *kit.Kit, cvmIDs []string) (*cloud.EipCvmRelJoinEipListResult, error)
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
	ListEipLeftJoinRel(kt *kit.Kit, opt *types.ListOption) (
		*cloud.EipLeftJoinEipCvmRelResult, error)
	insertValidate(kt *kit.Kit, rels []*tablecloud.EipCvmRelModel) error
}

var _ EipCvmRel = new(EipCvmRelDao)

// EipCvmRelDao EipCvmRelDao dao.
type EipCvmRelDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx ...
func (relDao EipCvmRelDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, rels []*tablecloud.EipCvmRelModel) error {
	if err := relDao.insertValidate(kt, rels); err != nil {
		return err
	}

	sql := fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES(%s)`,
		table.EipCvmRelTableName,
		tablecloud.EipCvmRelColumns.ColumnExpr(),
		tablecloud.EipCvmRelColumns.ColonNameExpr(),
	)
	if err := relDao.Orm.Txn(tx).BulkInsert(kt.Ctx, sql, rels); err != nil {
		logs.Errorf("batch create eip cvm rels failed, err: %v, rels: %v, rid: %s", err, rels, kt.Rid)
		return fmt.Errorf("insert %s failed, err: %v", table.EipCvmRelTableName, err)
	}

	return nil
}

// List ...
func (relDao EipCvmRelDao) List(kt *kit.Kit, opt *types.ListOption) (*cloud.EipCvmRelListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list eip cvm rel options is nil")
	}

	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(tablecloud.EipCvmRelColumns.ColumnTypes())),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf(
			"gen where expr for list eip cvm rels failed, err: %v, filter: %s, rid: %s",
			err,
			opt.Filter,
			kt.Rid,
		)
		return nil, err
	}

	if opt.Page.Count {
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.EipCvmRelTableName, whereExpr)
		count, err := relDao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.Errorf("count eip cvm rels failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.EipCvmRelListResult{Count: &count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		logs.Errorf(
			"gen page expr for list eip cvm rels failed, err: %v, filter: %s, rid: %s",
			err,
			opt.Filter,
			kt.Rid,
		)
		return nil, err
	}

	sql := fmt.Sprintf(
		`SELECT %s FROM %s %s %s`,
		tablecloud.EipCvmRelColumns.FieldsNamedExpr(opt.Fields),
		table.EipCvmRelTableName,
		whereExpr,
		pageExpr,
	)

	details := make([]*tablecloud.EipCvmRelModel, 0)
	if err = relDao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.Errorf("list eip cvm rels failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
		return nil, err
	}
	return &cloud.EipCvmRelListResult{Details: details}, nil
}

// ListJoinEip ...
func (relDao EipCvmRelDao) ListJoinEip(kt *kit.Kit, cvmIDs []string) (*cloud.EipCvmRelJoinEipListResult, error) {
	if len(cvmIDs) == 0 {
		return nil, errf.Newf(errf.InvalidParameter, "cvm ids is required")
	}

	sql := fmt.Sprintf(
		`SELECT %s, %s FROM %s as rel left join %s as eip on rel.eip_id = eip.id where cvm_id in (:cvm_ids)`,
		tableeip.EipColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		tools.BaseRelJoinSqlBuild(
			"rel",
			"eip",
			"id",
			"cvm_id",
		),
		table.EipCvmRelTableName,
		table.EipTable,
	)

	details := make([]*cloud.EipWithCvmID, 0)
	if err := relDao.Orm.Do().Select(kt.Ctx, &details, sql, map[string]interface{}{"cvm_ids": cvmIDs}); err != nil {
		logs.ErrorJson("select eip cvm rels join eip failed, err: %v, sql: (%s), rid: %s", err, sql, kt.Rid)
		return nil, err
	}

	return &cloud.EipCvmRelJoinEipListResult{Details: details}, nil
}

// DeleteWithTx ...
func (relDao EipCvmRelDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.EipCvmRelTableName, whereExpr)
	if _, err = relDao.Orm.Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.Errorf("delete eip cvm rels failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// insertValidate 校验待创建的关联关系表中, 对应的 Eip 和 CVM 是否存在
func (relDao EipCvmRelDao) insertValidate(kt *kit.Kit, rels []*tablecloud.EipCvmRelModel) error {
	relCount := len(rels)

	eipIDs := make([]string, relCount)
	cvmIDs := make([]string, relCount)
	for idx, rel := range rels {
		eipIDs[idx] = rel.EipID
		cvmIDs[idx] = rel.CvmID
	}

	idToEipMap, err := eip.ListByIDs(kt, relDao.Orm, eipIDs)
	if err != nil {
		logs.Errorf("list eip by ids failed, err: %v, ids: %v, rid: %s", err, eipIDs, kt.Rid)
		return err
	}
	if len(idToEipMap) != len(converter.StringSliceToMap(eipIDs)) {
		// TODO 将不存在的 ID 记录到错误信息中
		return fmt.Errorf("some eip does not exists")
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

// ListEipLeftJoinRel ...
func (relDao EipCvmRelDao) ListEipLeftJoinRel(kt *kit.Kit, opt *types.ListOption) (
	*cloud.EipLeftJoinEipCvmRelResult, error) {

	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list eip cvm rel options is nil")
	}

	columnTypes := tableeip.EipColumns.ColumnTypes()
	columnTypes["extension.resource_group_name"] = enumor.String
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.NewDefaultPageOption()); err != nil {
		return nil, err
	}

	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		logs.Errorf(
			"gen where expr for list eip cvm rels failed, err: %v, filter: %s, rid: %s",
			err,
			opt.Filter,
			kt.Rid,
		)
		return nil, err
	}

	whereExpr += " and rel.cvm_id is NULL"

	if opt.Page.Count {
		sql := fmt.Sprintf(
			`SELECT count(distinct(eip.id)) FROM %s as eip left join %s as rel on eip.id = rel.eip_id %s`,
			table.EipTable, table.EipCvmRelTableName, whereExpr)

		count, err := relDao.Orm.Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count eip left join eip_cvm_rel failed, err: %v, filter: %s, rid: %s", err,
				opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.EipLeftJoinEipCvmRelResult{Count: count}, nil
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
		`SELECT eip.id as id, %s FROM %s as eip left join %s as rel on eip.id = rel.eip_id %s group by eip.id %s`,
		tableeip.EipColumns.FieldsNamedExprWithout(types.DefaultRelJoinWithoutField),
		table.EipTable,
		table.EipCvmRelTableName,
		whereExpr,
		pageExpr,
	)

	details := make([]cloud.EipLeftJoinEipCvmRel, 0)
	if err := relDao.Orm.Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		logs.ErrorJson("select eip left join eip_cvm_rel failed, err: %v, filter: %s, rid: %s", err,
			opt.Filter, kt.Rid)
		return nil, err
	}

	return &cloud.EipLeftJoinEipCvmRelResult{Details: details}, nil
}
