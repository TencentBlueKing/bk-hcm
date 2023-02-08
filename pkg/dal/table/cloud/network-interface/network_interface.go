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

package networkinterface

import (
	"errors"
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// NetworkInterfaceColumns defines all the network interface table's columns.
var NetworkInterfaceColumns = utils.MergeColumns(nil, NetworkInterfaceTableColumnDescriptor)

// NetworkInterfaceTableColumnDescriptor is network interface's column descriptors.
var NetworkInterfaceTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "subnet_id", NamedC: "subnet_id", Type: enumor.String},
	{Column: "cloud_subnet_id", NamedC: "cloud_subnet_id", Type: enumor.String},
	{Column: "private_ip", NamedC: "private_ip", Type: enumor.String},
	{Column: "public_ip", NamedC: "public_ip", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "instance_id", NamedC: "instance_id", Type: enumor.String},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// NetworkInterfaceTable 网络接口表
type NetworkInterfaceTable struct {
	// ID 主键 ID
	ID string `db:"id"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"-"`
	// name 网络接口名称
	Name string `db:"name"`
	// AccountID 账号ID
	AccountID string `db:"account_id"`
	// Region 区域/地域
	Region string `db:"region"`
	// Zone 可用区
	Zone string `db:"zone"`
	// CloudID 网卡端口所属网络ID
	CloudID string `db:"cloud_id"`
	// VpcID VPC的ID
	VpcID string `db:"vpc_id"`
	// CloudVpcID 云VPC的ID
	CloudVpcID string `db:"cloud_vpc_id"`
	// SubnetID 子网ID
	SubnetID string `db:"subnet_id"`
	// CloudSubnetID 云子网ID
	CloudSubnetID string `db:"cloud_subnet_id"`
	// PrivateIP 内网IP
	PrivateIP string `db:"private_ip"`
	// PublicIP 公网IP
	PublicIP string `db:"public_ip"`
	// BkBizID 业务ID
	BkBizID int64 `db:"bk_biz_id"`
	// InstanceID 关联的实例ID
	InstanceID string `db:"instance_id"`
	// Extension 云厂商差异扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// Creator 创建者
	Creator string `db:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser"`
	// CreatedAt 创建时间
	CreatedAt *time.Time `db:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt *time.Time `db:"updated_at"`
}

// TableName return azure network interface table name.
func (a NetworkInterfaceTable) TableName() table.Name {
	return table.NetworkInterfaceTable
}

// InsertValidate network interface table when insert.
func (t NetworkInterfaceTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate network interface table when update.
func (t NetworkInterfaceTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
