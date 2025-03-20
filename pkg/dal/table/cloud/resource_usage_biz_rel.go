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

package cloud

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// ResUsageBizRelColumns defines resource biz relation table's columns.
var ResUsageBizRelColumns = utils.MergeColumns(nil, ResUsageBizRelColumnDescriptor)

// ResUsageBizRelColumnDescriptor defines resource biz relation table's column descriptors.
var ResUsageBizRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "rel_id", NamedC: "rel_id", Type: enumor.Numeric},
	{Column: "res_id", NamedC: "res_id", Type: enumor.String},
	{Column: "res_type", NamedC: "res_type", Type: enumor.String},
	{Column: "usage_biz_id", NamedC: "usage_biz_id", Type: enumor.String},
	{Column: "res_vendor", NamedC: "res_vendor", Type: enumor.String},
	{Column: "res_cloud_id", NamedC: "res_cloud_id", Type: enumor.String},
	{Column: "rel_creator", NamedC: "rel_creator", Type: enumor.String},
	{Column: "rel_created_at", NamedC: "rel_created_at", Type: enumor.Time},
}

// ResUsageBizRelTable define resource usage biz relation table.
type ResUsageBizRelTable struct {
	RelID        uint64                   `db:"rel_id" json:"rel_id"`
	ResType      enumor.CloudResourceType `db:"res_type" validate:"lte=64" json:"res_type"`
	ResID        string                   `db:"res_id" validate:"lte=64" json:"res_id"`
	UsageBizID   int64                    `db:"usage_biz_id" validate:"" json:"usage_biz_id"`
	ResVendor    enumor.Vendor            `db:"res_vendor" validate:"lte=64" json:"res_vendor"`
	ResCloudID   string                   `db:"res_cloud_id" validate:"lte=255" json:"res_cloud_id"`
	RelCreator   string                   `db:"rel_creator" validate:"lte=64" json:"rel_creator"`
	RelCreatedAt types.Time               `db:"rel_created_at" json:"rel_created_at"`
}

// TableName return resource biz relation table name.
func (t ResUsageBizRelTable) TableName() table.Name {
	return table.ResUsageBizRelTable
}

// InsertValidate resource biz relation table when inserted.
func (t ResUsageBizRelTable) InsertValidate() error {

	if t.UsageBizID == 0 {
		return errors.New("usage_biz_id is required")
	}

	if len(t.ResID) == 0 {
		return errors.New("res_id is required")
	}

	if len(t.ResType) == 0 {
		return errors.New("res_type is required")
	}

	if _, err := t.ResType.ConvTableName(); err != nil {
		return err
	}

	if len(t.RelCreator) == 0 {
		return errors.New("rel_creator is required")
	}

	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}
	return nil
}

// UpdateValidate resource biz relation table when updated.
func (t ResUsageBizRelTable) UpdateValidate() error {

	if len(t.RelCreator) != 0 {
		return errors.New("rel_creator can not update")
	}
	if len(t.RelCreatedAt) > 0 {
		return errors.New("rel_created_at can not update")
	}
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}
	return nil
}
