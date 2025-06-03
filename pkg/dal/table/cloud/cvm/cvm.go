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

// Package cvm ...
package cvm

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// TableColumns defines all the cvm table's columns.
var TableColumns = utils.MergeColumns(nil, TableColumnDescriptor)

// TableColumnDescriptor is cvm table column descriptors.
var TableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "bk_host_id", NamedC: "bk_host_id", Type: enumor.Numeric},
	{Column: "bk_cloud_id", NamedC: "bk_cloud_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "cloud_vpc_ids", NamedC: "cloud_vpc_ids", Type: enumor.Json},
	{Column: "vpc_ids", NamedC: "vpc_ids", Type: enumor.Json},
	{Column: "cloud_subnet_ids", NamedC: "cloud_subnet_ids", Type: enumor.Json},
	{Column: "subnet_ids", NamedC: "subnet_ids", Type: enumor.Json},
	{Column: "cloud_image_id", NamedC: "cloud_image_id", Type: enumor.String},
	{Column: "image_id", NamedC: "image_id", Type: enumor.String},
	{Column: "os_name", NamedC: "os_name", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "recycle_status", NamedC: "recycle_status", Type: enumor.String},
	{Column: "private_ipv4_addresses", NamedC: "private_ipv4_addresses", Type: enumor.Json},
	{Column: "private_ipv6_addresses", NamedC: "private_ipv6_addresses", Type: enumor.Json},
	{Column: "public_ipv4_addresses", NamedC: "public_ipv4_addresses", Type: enumor.Json},
	{Column: "public_ipv6_addresses", NamedC: "public_ipv6_addresses", Type: enumor.Json},
	{Column: "machine_type", NamedC: "machine_type", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "cloud_created_time", NamedC: "cloud_created_time", Type: enumor.String},
	{Column: "cloud_launched_time", NamedC: "cloud_launched_time", Type: enumor.String},
	{Column: "cloud_expired_time", NamedC: "cloud_expired_time", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// Table define cvm table.
type Table struct {
	ID                   string            `db:"id" validate:"lte=64" json:"id"`
	CloudID              string            `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name                 string            `db:"name" validate:"lte=255" json:"name"`
	Vendor               enumor.Vendor     `db:"vendor" validate:"lte=16" json:"vendor"`
	BkBizID              int64             `db:"bk_biz_id" json:"bk_biz_id"`
	BkHostID             int64             `db:"bk_host_id" json:"bk_host_id"`
	BkCloudID            *int64            `db:"bk_cloud_id" json:"bk_cloud_id"`
	AccountID            string            `db:"account_id" validate:"lte=64" json:"account_id"`
	Region               string            `db:"region" validate:"lte=20" json:"region"`
	Zone                 string            `db:"zone" validate:"lte=64" json:"zone"`
	CloudVpcIDs          types.StringArray `db:"cloud_vpc_ids" json:"cloud_vpc_ids"`
	VpcIDs               types.StringArray `db:"vpc_ids" json:"vpc_ids"`
	CloudSubnetIDs       types.StringArray `db:"cloud_subnet_ids" json:"cloud_subnet_ids"`
	SubnetIDs            types.StringArray `db:"subnet_ids" json:"subnet_ids"`
	CloudImageID         string            `db:"cloud_image_id" json:"cloud_image_id"`
	ImageID              string            `db:"image_id" json:"image_id"`
	OsName               string            `db:"os_name" json:"os_name"`
	Memo                 *string           `db:"memo" json:"memo"`
	Status               string            `db:"status" validate:"lte=50" json:"status"`
	RecycleStatus        string            `db:"recycle_status" validate:"lte=32" json:"recycle_status"`
	PrivateIPv4Addresses types.StringArray `db:"private_ipv4_addresses" json:"private_ipv4_addresses"`
	PrivateIPv6Addresses types.StringArray `db:"private_ipv6_addresses" json:"private_ipv6_addresses"`
	PublicIPv4Addresses  types.StringArray `db:"public_ipv4_addresses" json:"public_ipv4_addresses"`
	PublicIPv6Addresses  types.StringArray `db:"public_ipv6_addresses" json:"public_ipv6_addresses"`
	MachineType          string            `db:"machine_type" json:"machine_type"`
	Extension            types.JsonField   `db:"extension" json:"extension"`
	CloudCreatedTime     string            `db:"cloud_created_time" json:"cloud_created_time"`
	CloudLaunchedTime    string            `db:"cloud_launched_time" json:"cloud_launched_time"`
	CloudExpiredTime     string            `db:"cloud_expired_time" json:"cloud_expired_time"`
	Creator              string            `db:"creator" validate:"lte=64" json:"creator"`
	Reviser              string            `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt            types.Time        `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt            types.Time        `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return cvm table name.
func (t Table) TableName() table.Name {
	return table.CvmTable
}

var skipPartialFieldValidateVendor = map[enumor.Vendor]struct{}{
	enumor.Other: {},
}

// InsertValidate cvm table when insert.
func (t Table) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if t.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	if len(t.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(t.Extension) == 0 {
		return errors.New("extension is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	if _, ok := skipPartialFieldValidateVendor[t.Vendor]; ok {
		return nil
	}

	if len(t.Region) == 0 {
		return errors.New("region is required")
	}

	if len(t.CloudVpcIDs) == 0 {
		return errors.New("cloud_vpc_id is required")
	}

	if len(t.CloudSubnetIDs) == 0 {
		return errors.New("cloud_subnet_id is required")
	}

	if len(t.VpcIDs) == 0 {
		return errors.New("vpc_id is required")
	}

	if len(t.SubnetIDs) == 0 {
		return errors.New("subnet_id is required")
	}

	return nil
}

// UpdateValidate cvm table when update.
func (t Table) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
