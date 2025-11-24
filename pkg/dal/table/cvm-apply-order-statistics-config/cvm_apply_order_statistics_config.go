/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

// Package tableapplystat cvm apply order statistics config table
package tableapplystat

import (
	"errors"
	"strings"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// CvmApplyOrderStatisticsConfigTableColumns defines all the cvm apply order statistics config table's columns.
var CvmApplyOrderStatisticsConfigTableColumns = utils.MergeColumns(nil,
	CvmApplyOrderStatisticsConfigTableColumnDescriptors)

// CvmApplyOrderStatisticsConfigTableColumnDescriptors is cvm apply order statistics config table column descriptors.
var CvmApplyOrderStatisticsConfigTableColumnDescriptors = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "stat_month", NamedC: "stat_month", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "sub_order_ids", NamedC: "sub_order_ids", Type: enumor.String},
	{Column: "start_at", NamedC: "start_at", Type: enumor.String},
	{Column: "end_at", NamedC: "end_at", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// CvmApplyOrderStatisticsConfigTable define cvm apply order statistics config table.
type CvmApplyOrderStatisticsConfigTable struct {
	// ID 主键
	ID string `db:"id" json:"id"`
	// StatMonth 年月，格式：YYYY-MM
	StatMonth string `db:"stat_month" json:"stat_month" validate:"max=16"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// SubOrderID 子订单号，多个用逗号分隔
	SubOrderIDs *string `db:"sub_order_ids" json:"sub_order_ids"`
	// StartAt 开始时间
	StartAt *string `db:"start_at" json:"start_at" validate:"omitempty,max=64"`
	// EndAt 结束时间
	EndAt *string `db:"end_at" json:"end_at" validate:"omitempty,max=64"`
	// Memo 备注
	Memo string `db:"memo" json:"memo" validate:"max=255"`
	// Extension 扩展字段
	Extension *types.JsonField `db:"extension" json:"extension"`
	// Creator 创建者
	Creator string `db:"creator" json:"creator" validate:"max=64"`
	// Reviser 更新者
	Reviser string `db:"reviser" json:"reviser" validate:"max=64"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" json:"updated_at"`
}

// Columns return cvm apply order statistics config table columns.
func (t CvmApplyOrderStatisticsConfigTable) Columns() *utils.Columns {
	return CvmApplyOrderStatisticsConfigTableColumns
}

// ColumnDescriptors define cvm apply order statistics config table column descriptor.
func (t CvmApplyOrderStatisticsConfigTable) ColumnDescriptors() utils.ColumnDescriptors {
	return CvmApplyOrderStatisticsConfigTableColumnDescriptors
}

// TableName return cvm apply order statistics config table name.
func (t CvmApplyOrderStatisticsConfigTable) TableName() table.Name {
	return table.CvmApplyOrderStatisticsConfigTable
}

// InsertValidate cvm apply order statistics config table when insert.
func (t CvmApplyOrderStatisticsConfigTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.StatMonth) == 0 {
		return errors.New("stat_month is required")
	}

	if t.BkBizID <= 0 {
		return errors.New("bk_biz_id must be greater than 0")
	}

	// 子单号和开始结束时间不能同时为空
	hasSubOrderID := t.SubOrderIDs != nil && strings.TrimSpace(*t.SubOrderIDs) != ""
	hasStartAt := t.StartAt != nil && strings.TrimSpace(*t.StartAt) != ""
	hasEndAt := t.EndAt != nil && strings.TrimSpace(*t.EndAt) != ""
	hasTimeRange := hasStartAt && hasEndAt

	if hasStartAt && !hasEndAt {
		return errors.New("end_at is required when start_at is set")
	}

	if !hasStartAt && hasEndAt {
		return errors.New("start_at is required when end_at is set")
	}

	if !hasSubOrderID && !hasTimeRange {
		return errors.New("sub_order_ids and time range cannot be empty , at least one must be provided")
	}

	if err := validator.ValidateMemo(&t.Memo, true); err != nil {

		return err
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate cvm apply order statistics config table when update.
func (t CvmApplyOrderStatisticsConfigTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not be updated")
	}

	if t.BkBizID < 0 {

		return errors.New("bk_biz_id must be greater than or equal to 0")
	}

	// 如果更新时提供了子单号或时间范围，验证它们不能同时为空
	hasSubOrderID := t.SubOrderIDs != nil && strings.TrimSpace(*t.SubOrderIDs) != ""

	hasStartAt := t.StartAt != nil && strings.TrimSpace(*t.StartAt) != ""
	hasEndAt := t.EndAt != nil && strings.TrimSpace(*t.EndAt) != ""
	hasTimeRange := hasStartAt && hasEndAt

	// 如果设置了时间范围，start_at 和 end_at 必须同时提供
	if hasStartAt && !hasEndAt {
		return errors.New("end_at is required when start_at is set")
	}

	if !hasStartAt && hasEndAt {
		return errors.New("start_at is required when end_at is set")
	}

	// 更新时允许只修改子单号、只修改时间范围，或同时修改两者
	if !hasSubOrderID && !hasTimeRange {
		return errors.New("sub_order_ids and time range cannot be empty, at least one must be provided")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not be updated")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}
