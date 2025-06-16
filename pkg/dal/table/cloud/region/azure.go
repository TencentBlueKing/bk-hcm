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

package region

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// AzureRGColumns defines all the azure resource group table's columns.
var AzureRegionColumns = utils.MergeColumns(nil, AzureRegionTableColumnDescriptor)

// AzureRegionTableColumnDescriptor is azure region's column descriptors.
var AzureRegionTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "display_name", NamedC: "display_name", Type: enumor.String},
	{Column: "region_display_name", NamedC: "region_display_name", Type: enumor.String},
	{Column: "geography_group", NamedC: "geography_group", Type: enumor.String},
	{Column: "latitude", NamedC: "latitude", Type: enumor.String},
	{Column: "longitude", NamedC: "longitude", Type: enumor.String},
	{Column: "physical_location", NamedC: "physical_location", Type: enumor.String},
	{Column: "region_type", NamedC: "region_type", Type: enumor.String},
	{Column: "paired_region_name", NamedC: "paired_region_name", Type: enumor.String},
	{Column: "paired_region_id", NamedC: "paired_region_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AzureRegionTable azure地域表
type AzureRegionTable struct {
	ID                string     `db:"id"`
	CloudID           string     `db:"cloud_id"`
	Name              string     `db:"name"`
	Type              string     `db:"type"`
	DisplayName       string     `db:"display_name"`
	RegionDisplayName string     `db:"region_display_name"`
	GeographyGroup    string     `db:"geography_group"`
	Latitude          string     `db:"latitude"`
	Longitude         string     `db:"longitude"`
	PhysicalLocation  string     `db:"physical_location"`
	RegionType        string     `db:"region_type"`
	PairedRegionName  string     `db:"paired_region_name"`
	PairedRegionId    string     `db:"paired_region_id"`
	Creator           string     `db:"creator"`
	Reviser           string     `db:"reviser"`
	CreatedAt         types.Time `db:"created_at"`
	UpdatedAt         types.Time `db:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return azure region table name.
func (a AzureRegionTable) TableName() table.Name {
	return table.AzureRegionTable
}

// InsertValidate azure region table when insert.
func (t AzureRegionTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.Type) == 0 {
		return errors.New("type is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate azure azure region table when update.
func (t AzureRegionTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
