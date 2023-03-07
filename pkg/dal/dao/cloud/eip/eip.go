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

package eip

import (
	"fmt"

	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/dao/types/cloud"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/cloud/eip"
	"hcm/pkg/kit"

	"github.com/jmoiron/sqlx"

	"hcm/pkg/api/core"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"
)

// EipDao ...
type EipDao struct {
	*dao.ObjectDaoManager
}

var _ dao.ObjectDao = new(EipDao)

// Name 返回 Dao 描述对象的表名
func (eipDao *EipDao) Name() table.Name {
	return eip.TableName
}

// BatchCreateWithTx ...
func (eipDao *EipDao) BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, eips []*eip.EipModel) ([]string, error) {
	if len(eips) == 0 {
		return nil, errf.New(errf.InvalidParameter, "eip model data is required")
	}
	for _, i := range eips {
		if err := i.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, eipDao.Name(), eip.EipColumns.ColumnExpr(),
		eip.EipColumns.ColonNameExpr(),
	)

	ids, err := eipDao.IDGen().Batch(kt, eipDao.Name(), len(eips))
	if err != nil {
		return nil, err
	}

	for idx, d := range eips {
		d.ID = ids[idx]
	}

	err = eipDao.Orm().Txn(tx).BulkInsert(kt.Ctx, sql, eips)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", eipDao.Name(), err)
	}

	return ids, nil
}

// List ...
func (eipDao *EipDao) List(kt *kit.Kit, opt *types.ListOption) (*cloud.EipListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list eip options is nil")
	}

	columnTypes := eip.EipColumns.ColumnTypes()
	if err := opt.Validate(filter.NewExprOption(filter.RuleFields(columnTypes)),
		core.DefaultPageOption); err != nil {
		return nil, err
	}

	whereOpt := tools.DefaultSqlWhereOption
	whereExpr, whereValue, err := opt.Filter.SQLWhereExpr(whereOpt)
	if err != nil {
		return nil, err
	}

	if opt.Page.Count {
		// this is a count request, then do count operation only.
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, eipDao.Name(), whereExpr)
		count, err := eipDao.Orm().Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count eip failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.EipListResult{Count: &count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(
		`SELECT %s FROM %s %s %s`,
		eip.EipColumns.FieldsNamedExpr(opt.Fields),
		eipDao.Name(),
		whereExpr,
		pageExpr,
	)
	details := make([]*eip.EipModel, 0)
	if err = eipDao.Orm().Do().Select(kt.Ctx, &details, sql, whereValue); err != nil {
		return nil, err
	}

	result := &cloud.EipListResult{Details: details}
	return result, nil
}

// UpdateByIDWithTx ...
func (eipDao *EipDao) UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, eipID string, updateData *eip.EipModel) error {
	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, eipDao.Name(), setExpr)

	toUpdate["id"] = eipID
	_, err = eipDao.Orm().Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update eip failed, err: %v, id: %s, rid: %v", err, eipID, kt.Rid)
		return err
	}

	return nil
}

// Update ...
func (eipDao *EipDao) Update(kt *kit.Kit, filterExpr *filter.Expression, updateData *eip.EipModel) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is nil")
	}

	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, eipDao.Name(), setExpr, whereExpr)

	_, err = eipDao.Orm().AutoTxn(kt, func(txn *sqlx.Tx, opt *orm.TxnOption) (interface{}, error) {
		effected, err := eipDao.Orm().Txn(txn).Update(kt.Ctx, sql, tools.MapMerge(toUpdate, whereValue))
		if err != nil {
			logs.ErrorJson("update eip failed, err: %v, filter: %s, rid: %v", err, filterExpr, kt.Rid)
			return nil, err
		}

		if effected == 0 {
			logs.ErrorJson("update eip, but record not found, filter: %v, rid: %v", filterExpr, kt.Rid)
			return nil, errf.New(errf.RecordNotFound, orm.ErrRecordNotFound.Error())
		}

		return nil, nil
	})
	if err != nil {
		return err
	}

	return nil
}

// DeleteWithTx ...
func (eipDao *EipDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, eipDao.Name(), whereExpr)
	if _, err = eipDao.Orm().Txn(tx).Delete(kt.Ctx, sql, whereValue); err != nil {
		logs.ErrorJson("delete eip failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}

// ListByIDs ...
func ListByIDs(kt *kit.Kit, orm orm.Interface, ids []string) (map[string]eip.EipModel, error) {
	sql := fmt.Sprintf(`SELECT %s FROM %s where id in (:ids)`, eip.EipColumns.FieldsNamedExpr(nil), eip.TableName)
	eips := make([]eip.EipModel, 0)
	if err := orm.Do().Select(kt.Ctx, &eips, sql, map[string]interface{}{"ids": ids}); err != nil {
		return nil, err
	}

	idToEipMap := make(map[string]eip.EipModel, len(ids))
	for _, d := range eips {
		idToEipMap[d.ID] = d
	}
	return idToEipMap, nil
}
