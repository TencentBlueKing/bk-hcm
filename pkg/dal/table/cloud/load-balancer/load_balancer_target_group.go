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

package tablelb

import (
	"errors"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// LoadBalancerTargetGroupColumns defines all the load_balancer_target_group table's columns.
var LoadBalancerTargetGroupColumns = utils.MergeColumns(nil, LoadBalancerTargetGroupColumnsDescriptor)

// LoadBalancerTargetGroupColumnsDescriptor is load_balancer_target_group's column descriptors.
var LoadBalancerTargetGroupColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "target_group_type", NamedC: "target_group_type", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "port", NamedC: "port", Type: enumor.Numeric},
	{Column: "weight", NamedC: "weight", Type: enumor.Numeric},
	{Column: "health_check", NamedC: "health_check", Type: enumor.Json},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// LoadBalancerTargetGroupTable 负载均衡目标组表
type LoadBalancerTargetGroupTable struct {
	ID              string                 `db:"id" validate:"lte=64" json:"id"`
	CloudID         string                 `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name            string                 `db:"name" validate:"lte=255" json:"name"`
	Vendor          enumor.Vendor          `db:"vendor" validate:"lte=16"  json:"vendor"`
	AccountID       string                 `db:"account_id" validate:"lte=64" json:"account_id"`
	BkBizID         int64                  `db:"bk_biz_id" json:"bk_biz_id"`
	TargetGroupType enumor.TargetGroupType `db:"target_group_type" validate:"lte=64" json:"target_group_type"`
	VpcID           string                 `db:"vpc_id" json:"vpc_id"`
	CloudVpcID      string                 `db:"cloud_vpc_id" json:"cloud_vpc_id"`
	Region          string                 `db:"region" validate:"lte=20" json:"region"`
	Protocol        enumor.ProtocolType    `db:"protocol" json:"protocol"`
	Port            int64                  `db:"port" json:"port"`
	Weight          *int64                 `db:"weight" json:"weight"`
	HealthCheck     types.JsonField        `db:"health_check" json:"health_check"`
	Extension       types.JsonField        `db:"extension" json:"extension"`
	Memo            *string                `db:"memo" json:"memo"`
	Creator         string                 `db:"creator" validate:"lte=64" json:"creator"`
	Reviser         string                 `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt       types.Time             `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt       types.Time             `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return load_balancer_target_group table name.
func (lbtg LoadBalancerTargetGroupTable) TableName() table.Name {
	return table.LoadBalancerTargetGroupTable
}

// InsertValidate load_balancer_target_group table when insert.
func (lbtg LoadBalancerTargetGroupTable) InsertValidate() error {
	if len(lbtg.Name) == 0 {
		return errors.New("name is required")
	}

	if len(lbtg.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(lbtg.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(lbtg.VpcID) == 0 {
		return errors.New("vpc_id is required")
	}

	if len(lbtg.Region) == 0 {
		return errors.New("region is required")
	}

	if len(lbtg.Protocol) == 0 {
		return errors.New("protocol is required")
	}

	if len(lbtg.Creator) == 0 {
		return errors.New("creator is required")
	}

	return validator.Validate.Struct(lbtg)
}

// UpdateValidate load_balancer_target_group table when update.
func (lbtg LoadBalancerTargetGroupTable) UpdateValidate() error {
	if len(lbtg.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(lbtg.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return validator.Validate.Struct(lbtg)
}
