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

// Package image ...
package image

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
	"hcm/pkg/dal/table/cloud/image"
	"hcm/pkg/dal/table/utils"
	"hcm/pkg/kit"
	"hcm/pkg/logs"
	"hcm/pkg/runtime/filter"

	"github.com/jmoiron/sqlx"
)

// Image only used for image.
type Image interface {
	BatchCreateWithTx(kt *kit.Kit, tx *sqlx.Tx, images []*image.ImageModel) ([]string, error)
	List(kt *kit.Kit, opt *types.ListOption) (*cloud.ImageListResult, error)
	UpdateByIDWithTx(kt *kit.Kit, tx *sqlx.Tx, imageID string, updateData *image.ImageModel) error
	DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error
}

var _ Image = new(ImageDao)

// ImageDao image dao.
type ImageDao struct {
	Orm   orm.Interface
	IDGen idgenerator.IDGenInterface
	Audit audit.Interface
}

// BatchCreateWithTx ...
func (pImageDao ImageDao) BatchCreateWithTx(
	kt *kit.Kit,
	tx *sqlx.Tx,
	images []*image.ImageModel,
) ([]string, error) {
	if len(images) == 0 {
		return nil, errf.New(errf.InvalidParameter, "image model data is required")
	}
	for _, i := range images {
		if err := i.InsertValidate(); err != nil {
			return nil, err
		}
	}

	sql := fmt.Sprintf(`INSERT INTO %s (%s)	VALUES(%s)`, table.ImageTable, image.ImageColumns.ColumnExpr(),
		image.ImageColumns.ColonNameExpr(),
	)

	ids, err := pImageDao.IDGen.Batch(kt, table.ImageTable, len(images))
	if err != nil {
		return nil, err
	}

	for idx, d := range images {
		d.ID = ids[idx]
	}

	err = pImageDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).BulkInsert(kt.Ctx, sql, images)
	if err != nil {
		return nil, fmt.Errorf("insert %s failed, err: %v", table.ImageTable, err)
	}

	return ids, nil
}

// List ...
func (pImageDao ImageDao) List(kt *kit.Kit, opt *types.ListOption) (*cloud.ImageListResult, error) {
	if opt == nil {
		return nil, errf.New(errf.InvalidParameter, "list image options is nil")
	}

	columnTypes := image.ImageColumns.ColumnTypes()
	columnTypes["extension.region"] = enumor.String
	columnTypes["extension.project_id"] = enumor.String
	columnTypes["extension.publisher"] = enumor.String
	columnTypes["extension.offer"] = enumor.String
	columnTypes["extension.sku"] = enumor.String
	columnTypes["extension.self_link"] = enumor.String
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
		sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s %s`, table.ImageTable, whereExpr)
		count, err := pImageDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Count(kt.Ctx, sql, whereValue)
		if err != nil {
			logs.ErrorJson("count image failed, err: %v, filter: %s, rid: %s", err, opt.Filter, kt.Rid)
			return nil, err
		}
		return &cloud.ImageListResult{Count: count}, nil
	}

	pageExpr, err := types.PageSQLExpr(opt.Page, types.DefaultPageSQLOption)
	if err != nil {
		return nil, err
	}

	sql := fmt.Sprintf(
		`SELECT %s FROM %s %s %s`,
		image.ImageColumns.FieldsNamedExpr(opt.Fields),
		table.ImageTable,
		whereExpr,
		pageExpr,
	)
	details := make([]*image.ImageModel, 0)
	err = pImageDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Do().Select(
		kt.Ctx, &details, sql, whereValue)
	if err != nil {
		logs.Errorf("select image failed, err: %v, sql: %s, values: %v, rid: %s", err, sql, whereValue, kt.Rid)
		return nil, err
	}

	result := &cloud.ImageListResult{Details: details}
	return result, nil
}

// UpdateByIDWithTx ...
func (pImageDao ImageDao) UpdateByIDWithTx(
	kt *kit.Kit,
	tx *sqlx.Tx,
	imageID string,
	updateData *image.ImageModel,
) error {
	if err := updateData.UpdateValidate(); err != nil {
		return err
	}

	opts := utils.NewFieldOptions().AddIgnoredFields(types.DefaultIgnoredFields...)
	setExpr, toUpdate, err := utils.RearrangeSQLDataWithOption(updateData, opts)
	if err != nil {
		return fmt.Errorf("prepare parsed sql set filter expr failed, err: %v", err)
	}

	sql := fmt.Sprintf(`UPDATE %s %s where id = :id`, table.ImageTable, setExpr)

	toUpdate["id"] = imageID
	_, err = pImageDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Update(kt.Ctx, sql, toUpdate)
	if err != nil {
		logs.ErrorJson("update image failed, err: %v, id: %s, rid: %v", err, imageID, kt.Rid)
		return err
	}

	return nil
}

// DeleteWithTx ...
func (pImageDao ImageDao) DeleteWithTx(kt *kit.Kit, tx *sqlx.Tx, filterExpr *filter.Expression) error {
	if filterExpr == nil {
		return errf.New(errf.InvalidParameter, "filter expr is required")
	}

	whereExpr, whereValue, err := filterExpr.SQLWhereExpr(tools.DefaultSqlWhereOption)
	if err != nil {
		return err
	}

	sql := fmt.Sprintf(`DELETE FROM %s %s`, table.ImageTable, whereExpr)
	_, err = pImageDao.Orm.ModifySQLOpts(orm.NewInjectTenantIDOpt(kt.TenantID)).Txn(tx).Delete(kt.Ctx, sql, whereValue)
	if err != nil {
		logs.ErrorJson("delete image failed, err: %v, filter: %s, rid: %s", err, filterExpr, kt.Rid)
		return err
	}

	return nil
}
