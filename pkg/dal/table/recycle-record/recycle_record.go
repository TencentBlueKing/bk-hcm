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
	"errors"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// RecycleRecordColumns defines all the recycle record table's columns.
var RecycleRecordColumns = utils.MergeColumns(nil, RecycleRecordColumnDescriptor)

// RecycleRecordColumnDescriptor is RecycleRecordTable's column descriptors.
var RecycleRecordColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "task_id", NamedC: "task_id", Type: enumor.String},
	{Column: "recycle_type", NamedC: "recycle_type", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "cloud_res_id", NamedC: "cloud_res_id", Type: enumor.String},
	{Column: "res_name", NamedC: "res_name", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "detail", NamedC: "detail", Type: enumor.Json},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
	{Column: "recycled_at", NamedC: "recycled_at", Type: enumor.Time},
}

// RecycleRecordTable is used to save resource's recycle record information.
type RecycleRecordTable struct {
	// ID 自增ID
	ID string `db:"id" json:"id" validate:"lte=64"`
	// TaskID 单次回收一批资源的任务ID
	TaskID string `db:"task_id" json:"task_id" validate:"lte=64"`
	// RecycleType 记录类型 用来标记关联资源类型，留空为普通类型
	RecycleType enumor.RecycleType `db:"recycle_type" json:"recycle_type" validate:"lte=64"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" json:"vendor" validate:"lte=32"`
	// ResType 资源类型
	ResType enumor.CloudResourceType `db:"res_type" json:"res_type" validate:"lte=64"`
	// ResID 资源ID
	ResID string `db:"res_id" json:"res_id" validate:"lte=64"`
	// CloudResID 资源的云上ID
	CloudResID string `db:"cloud_res_id" json:"cloud_res_id" validate:"lte=255"`
	// ResName 资源名称
	ResName string `db:"res_name" json:"res_name" validate:"lte=255"`
	// BkBizID 回收前所在业务ID
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id" validate:"min=-1"`
	// AccountID 账号ID
	AccountID string `db:"account_id" json:"account_id" validate:"lte=64"`
	// Region 地域
	Region string `db:"region" json:"region" validate:"lte=255"`
	// Detail 回收详情
	Detail types.JsonField `db:"detail" json:"detail" validate:"omitempty"`
	// Detail 回收状态
	Status string `db:"status" validate:"lte=32" json:"status"`
	// Creator 创建者
	Creator string `db:"creator" validate:"max=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"max=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"isdefault" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"isdefault" json:"updated_at"`
	// RecycledAt 回收时间
	RecycledAt time.Time `db:"recycled_at" json:"recycled_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// InsertValidate validate recycle record on insertion.
func (r RecycleRecordTable) InsertValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.ID) == 0 {
		return errors.New("id can not be empty")
	}

	if len(r.TaskID) == 0 {
		return errors.New("task id can not be empty")
	}

	if err := r.Vendor.Validate(); err != nil {
		return err
	}

	if len(r.ResType) == 0 {
		return errors.New("resource type can not be empty")
	}

	if len(r.ResID) == 0 {
		return errors.New("resource id can not be empty")
	}

	if len(r.CloudResID) == 0 {
		return errors.New("cloud resource id can not be empty")
	}

	if r.BkBizID == 0 {
		return errors.New("biz id can not be empty")
	}

	if len(r.AccountID) == 0 {
		return errors.New("account id can not be empty")
	}

	if len(r.Region) == 0 {
		return errors.New("region can not be empty")
	}

	if len(r.Creator) == 0 {
		return errors.New("creator can not be empty")
	}

	return nil
}

// TableName is the recycleRecord's database table name.
func (r RecycleRecordTable) TableName() table.Name {
	return table.RecycleRecordTable
}

// UpdateValidate validate recycle record on update.
func (r RecycleRecordTable) UpdateValidate() error {
	if err := validator.Validate.Struct(r); err != nil {
		return err
	}

	if len(r.Status) == 0 && len(r.Detail) == 0 {
		return errors.New("one of the update fields must be set")
	}

	if len(r.TaskID) != 0 {
		return errors.New("task id can not update")
	}

	if len(r.Vendor) != 0 {
		return errors.New("vendor can not update")
	}

	if len(r.ResType) != 0 {
		return errors.New("resource type can not update")
	}

	if len(r.ResID) != 0 {
		return errors.New("resource id can not update")
	}

	if len(r.CloudResID) != 0 {
		return errors.New("cloud resource id can not update")
	}

	if len(r.ResName) != 0 {
		return errors.New("resource name can not update")
	}

	if r.BkBizID != 0 {
		return errors.New("biz id can not update")
	}

	if len(r.AccountID) != 0 {
		return errors.New("account id can not update")
	}

	if len(r.Region) != 0 {
		return errors.New("region can not update")
	}

	if len(r.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(r.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(r)
}
