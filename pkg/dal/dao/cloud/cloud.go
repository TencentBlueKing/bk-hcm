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

// Package cloud ...
package cloud

import (
	"fmt"
	"strings"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/dal/dao/orm"
	"hcm/pkg/dal/dao/tools"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/dal/table"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Cloud only used for cloud common operation.
type Cloud interface {
	ListResourceBasicInfo(kt *kit.Kit, resType enumor.CloudResourceType, ids []string, fields ...string) (
		[]types.CloudResourceBasicInfo, error)
	ListResourceIDs(kt *kit.Kit, resType enumor.CloudResourceType, expr *filter.Expression) ([]string, error)
	AssignResourceToBiz(kt *kit.Kit, tx *sqlx.Tx, resType enumor.CloudResourceType, expr *filter.Expression,
		bizID int64) error
}

var _ Cloud = new(CloudDao)

// CloudDao cloud dao.
type CloudDao struct {
	Orm orm.Interface
}

// ListResourceBasicInfo list cloud resource basic info.
func (dao CloudDao) ListResourceBasicInfo(kt *kit.Kit, resType enumor.CloudResourceType, ids []string,
	fields ...string) ([]types.CloudResourceBasicInfo, error) {

	tableName, err := resType.ConvTableName()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if len(ids) == 0 {
		return nil, errf.New(errf.InvalidParameter, "ids is required")
	}

	// if fields are not set, select common fields.
	if len(fields) == 0 {
		fields = types.CommonBasicInfoFields
	}

	// select cloud resource basic infos.
	sql := fmt.Sprintf("select %s from %s where id in (:ids)", strings.Join(fields, ", "), tableName)
	if tableName == table.AccountTable {
		sql = fmt.Sprintf("select id, vendor, id as account_id from %s where id in (:ids)", tableName)
	}

	list := make([]types.CloudResourceBasicInfo, 0)
	args := map[string]interface{}{
		"ids": ids,
	}
	if err := dao.Orm.Do().Select(kt.Ctx, &list, sql, args); err != nil {
		logs.Errorf("select resource vendor failed, err: %v, table: %s, id: %v, rid: %s", err, resType, ids, kt.Rid)
		return nil, err
	}

	return list, nil
}

// ListResourceIDs list cloud resource ids.
func (dao CloudDao) ListResourceIDs(kt *kit.Kit, resType enumor.CloudResourceType, expr *filter.Expression) ([]string,
	error) {

	tableName, err := resType.ConvTableName()
	if err != nil {
		return nil, errf.NewFromErr(errf.InvalidParameter, err)
	}

	if expr == nil {
		return nil, errf.New(errf.InvalidParameter, "ids is required")
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf("select id from %s %s", tableName, whereExpr)

	list := make([]types.CloudResourceBasicInfo, 0)
	if err := dao.Orm.Do().Select(kt.Ctx, &list, sql, whereValue); err != nil {
		logs.Errorf("select %s resource id failed, err: %v, expr: %v, rid: %s", resType, err, expr, kt.Rid)
		return nil, err
	}

	ids := make([]string, len(list))
	for idx, info := range list {
		ids[idx] = info.ID
	}

	return ids, nil
}

// AssignResourceToBiz assign an account's cloud resource to biz, **only for ui**.
func (dao CloudDao) AssignResourceToBiz(kt *kit.Kit, tx *sqlx.Tx, resType enumor.CloudResourceType,
	expr *filter.Expression, bizID int64) error {

	tableName, err := resType.ConvTableName()
	if err != nil {
		return errf.NewFromErr(errf.InvalidParameter, err)
	}

	whereExpr, whereValue, err := expr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`update %s set bk_biz_id = :bk_biz_id %s`, tableName, whereExpr)

	updateData := map[string]interface{}{
		"bk_biz_id": bizID,
	}

	_, err = dao.Orm.Txn(tx).Update(kt.Ctx, sql, tools.MapMerge(updateData, whereValue))
	if err != nil {
		logs.ErrorJson("assign %s resource to biz failed, err: %v, biz: %d, filter: %+v, rid: %v", resType, err, bizID,
			expr, kt.Rid)
		return err
	}

	return nil
}
