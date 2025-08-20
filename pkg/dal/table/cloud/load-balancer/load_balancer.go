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

// Package tablelb table definition of load balancer related resources
package tablelb

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// LoadBalancerColumns defines all the load_balancer table's columns.
var LoadBalancerColumns = utils.MergeColumns(nil, LoadBalancerColumnsDescriptor)

// LoadBalancerColumnsDescriptor is load_balancer's column descriptors.
var LoadBalancerColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},

	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zones", NamedC: "zones", Type: enumor.Json},
	{Column: "backup_zones", NamedC: "backup_zones", Type: enumor.Json},
	{Column: "lb_type", NamedC: "lb_type", Type: enumor.String},
	{Column: "ip_version", NamedC: "ip_version", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "subnet_id", NamedC: "subnet_id", Type: enumor.String},
	{Column: "cloud_subnet_id", NamedC: "cloud_subnet_id", Type: enumor.String},
	{Column: "private_ipv4_addresses", NamedC: "private_ipv4_addresses", Type: enumor.Json},
	{Column: "private_ipv6_addresses", NamedC: "private_ipv6_addresses", Type: enumor.Json},
	{Column: "public_ipv4_addresses", NamedC: "public_ipv4_addresses", Type: enumor.Json},
	{Column: "public_ipv6_addresses", NamedC: "public_ipv6_addresses", Type: enumor.Json},
	{Column: "domain", NamedC: "domain", Type: enumor.String},
	{Column: "status", NamedC: "status", Type: enumor.String},
	{Column: "isp", NamedC: "isp", Type: enumor.String},
	{Column: "band_width", NamedC: "band_width", Type: enumor.Numeric},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "cloud_created_time", NamedC: "cloud_created_time", Type: enumor.String},
	{Column: "cloud_status_time", NamedC: "cloud_status_time", Type: enumor.String},
	{Column: "cloud_expired_time", NamedC: "cloud_expired_time", Type: enumor.String},
	{Column: "sync_time", NamedC: "sync_time", Type: enumor.String},
	{Column: "tags", NamedC: "tags", Type: enumor.Json},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// LoadBalancerTable 负载均衡表
type LoadBalancerTable struct {
	ID                   string            `db:"id" validate:"lte=64" json:"id"`
	CloudID              string            `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name                 string            `db:"name" validate:"lte=255" json:"name"`
	Vendor               enumor.Vendor     `db:"vendor" validate:"lte=16"  json:"vendor"`
	AccountID            string            `db:"account_id" validate:"lte=64" json:"account_id"`
	BkBizID              int64             `db:"bk_biz_id" json:"bk_biz_id"`
	Region               string            `db:"region" validate:"lte=20" json:"region"`
	Zones                types.StringArray `db:"zones" validate:"lte=20" json:"zones"`
	BackupZones          types.StringArray `db:"backup_zones" json:"backup_zones"`
	LBType               string            `db:"lb_type" json:"lb_type"`
	IPVersion            string            `db:"ip_version"  json:"ip_version"`
	VpcID                string            `db:"vpc_id" json:"vpc_id"`
	CloudVpcID           string            `db:"cloud_vpc_id" json:"cloud_vpc_id"`
	SubnetID             string            `db:"subnet_id" json:"subnet_id"`
	CloudSubnetID        string            `db:"cloud_subnet_id" validate:"-" json:"cloud_subnet_id"`
	PrivateIPv4Addresses types.StringArray `db:"private_ipv4_addresses" validate:"-" json:"private_ipv4_addresses"`
	PrivateIPv6Addresses types.StringArray `db:"private_ipv6_addresses" validate:"-" json:"private_ipv6_addresses"`
	PublicIPv4Addresses  types.StringArray `db:"public_ipv4_addresses" json:"public_ipv4_addresses"`
	PublicIPv6Addresses  types.StringArray `db:"public_ipv6_addresses" json:"public_ipv6_addresses"`
	Domain               string            `db:"domain" json:"domain"`
	Status               string            `db:"status" json:"status"`
	Memo                 *string           `db:"memo" json:"memo"`
	CloudCreatedTime     string            `db:"cloud_created_time" json:"cloud_created_time"`
	CloudStatusTime      string            `db:"cloud_status_time" json:"cloud_status_time"`
	CloudExpiredTime     string            `db:"cloud_expired_time" json:"cloud_expired_time"`
	SyncTime             string            `db:"sync_time" json:"sync_time"`
	Tags                 types.StringMap   `db:"tags" json:"tags"`
	Extension            types.JsonField   `db:"extension" json:"extension"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
	// TenantID 租户ID
	TenantID  string `db:"tenant_id" json:"tenant_id"`
	BandWidth int64  `db:"band_width" json:"band_width"` // 带宽
	Isp       string `db:"isp" json:"isp"`               // 运营商
}

// TableName return load_balancer table name.
func (lb LoadBalancerTable) TableName() table.Name {
	return table.LoadBalancerTable
}

// InsertValidate load_balancer table when insert.
func (lb LoadBalancerTable) InsertValidate() error {
	if err := validator.Validate.Struct(lb); err != nil {
		return err
	}

	if len(lb.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(lb.Name) == 0 {
		return errors.New("name is required")
	}

	if len(lb.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(lb.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(lb.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate load_balancer table when update.
func (lb LoadBalancerTable) UpdateValidate() error {
	if err := validator.Validate.Struct(lb); err != nil {
		return err
	}

	if len(lb.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(lb.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
