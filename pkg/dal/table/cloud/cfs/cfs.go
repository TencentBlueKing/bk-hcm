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

// Package cfs 文件存储
package cfs

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// TableColumns defines all the cfs table's columns.
var TableColumns = utils.MergeColumns(nil, TableColumnDescriptor)

// TableColumnDescriptor is cfs table column descriptors.
var TableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "size_limit", NamedC: "size_limit", Type: enumor.Numeric},
	{Column: "size_byte", NamedC: "size_byte", Type: enumor.Numeric},
	{Column: "avail_capacity", NamedC: "avail_capacity", Type: enumor.Numeric},
	{Column: "bandwidth_limit", NamedC: "bandwidth_limit", Type: enumor.Numeric},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "storage_type", NamedC: "storage_type", Type: enumor.String},
	{Column: "encrypted", NamedC: "encrypted", Type: enumor.Boolean},
	{Column: "crypt_key_id", NamedC: "crypt_key_id", Type: enumor.String},
	{Column: "cloud_vpc_ids", NamedC: "cloud_vpc_ids", Type: enumor.Json},
	{Column: "cloud_subnet_ids", NamedC: "cloud_subnet_ids", Type: enumor.Json},
	{Column: "vpc_ids", NamedC: "vpc_ids", Type: enumor.Json},
	{Column: "subnet_ids", NamedC: "subnet_ids", Type: enumor.Json},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "cloud_created_time", NamedC: "cloud_created_time", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// Table define cfs table.
type Table struct {
	// ID 主键
	ID string `db:"id" validate:"lte=64" json:"id"`
	// CloudID 云上资源id
	CloudID string `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	// Name 资源名称
	Name string `db:"name" validate:"lte=255" json:"name"`
	// Vendor 各个云具体名称
	Vendor enumor.Vendor `db:"vendor" validate:"lte=16" json:"vendor"`

	// BkBizID 业务id
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// AccountID 账号id
	AccountID string `db:"account_id" validate:"lte=64" json:"account_id"`

	// Region 所属地域
	Region string `db:"region" validate:"lte=20" json:"region"`
	// Zone 所属可用区
	Zone string `db:"zone" validate:"lte=20" json:"zone"`
	// SizeLimit 文件系统最大空间限制(单位:GiB, 示例值：50)
	SizeLimit uint64 `db:"size_limit" json:"size_limit"`
	// SizeByte 文件系统已使用容量.单位：Byte; 示例值：10
	SizeByte uint64 `db:"size_byte" json:"size_byte"`
	// AvailCapacity 文件系统剩余容量. 单位：Byte. 示例值：10
	AvailCapacity uint64 `db:"avail_capacity" json:"avail_capacity"`
	// BandwidthLimit 文件系统吞吐上限，吞吐上限是根据文件系统当前已使用存储量、绑定的存储资源包以及吞吐资源包一同确定. 单位MiB/s
	BandwidthLimit float64 `db:"bandwidth_limit" json:"bandwidth_limit" `
	// Protocol 文件系统协议类型, 支持 NFS,CIFS,TURBO; 示例值：NFS
	Protocol string `db:"protocol" json:"protocol" `
	// StorageType 文件系统存储类型. HP：通用性能型；SD：通用标准型；TP:turbo性能型；TB：turbo标准型；THP：吞吐型; 示例值：HP
	StorageType string `db:"storage_type" json:"storage_type" `
	// Encrypted 文件系统是否加密
	Encrypted bool `db:"encrypted" json:"encrypted"`
	// CryptKeyId 加密所使用的密钥，可以为密钥的 ID 或者 ARN
	CryptKeyId string `db:"crypt_key_id" json:"crypt_key_id"`

	// CloudVpcIDs 云上vpc id
	CloudVpcIDs types.StringArray `db:"cloud_vpc_ids" json:"cloud_vpc_ids"`
	// CloudSubnetIDs 云上子网 id
	CloudSubnetIDs types.StringArray `db:"cloud_subnet_ids" json:"cloud_subnet_ids"`
	// VpcIDs vpc id
	VpcIDs types.StringArray `db:"vpc_ids" json:"vpc_ids"`
	// SubnetIDs 子网 id
	SubnetIDs types.StringArray `db:"subnet_ids" json:"subnet_ids"`

	// Status 资源状态
	Status string `db:"status" validate:"lte=64" json:"status"`
	// Memo 备注字段
	Memo *string `db:"memo" json:"memo"`
	// Extension 多云差异化字段
	// note: 挂载点放到Extension字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// CloudCreatedTime 云上资源创建时间
	CloudCreatedTime string `db:"cloud_created_time" json:"cloud_created_time"`
	// Creator 创建人
	Creator string `db:"creator" validate:"lte=64" json:"creator"`
	// Reviser 审核人
	Reviser string `db:"reviser" validate:"lte=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return cfs table name.
func (t Table) TableName() table.Name {
	return table.CfsTable
}

var skipPartialFieldValidateVendor = map[enumor.Vendor]struct{}{
	enumor.Other: {},
}

// InsertValidate cfs table when insert.
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

// UpdateValidate cfs table when update.
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
