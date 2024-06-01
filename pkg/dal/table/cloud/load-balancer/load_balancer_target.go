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

// LoadBalancerTargetColumns defines all the load_balancer_target table's columns.
var LoadBalancerTargetColumns = utils.MergeColumns(nil, LoadBalancerTargetColumnsDescriptor)

// LoadBalancerTargetColumnsDescriptor is load_balancer_target's column descriptors.
var LoadBalancerTargetColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},

	{Column: "inst_type", NamedC: "inst_type", Type: enumor.String},
	{Column: "inst_id", NamedC: "inst_id", Type: enumor.String},
	{Column: "cloud_inst_id", NamedC: "cloud_inst_id", Type: enumor.String},
	{Column: "inst_name", NamedC: "inst_name", Type: enumor.String},
	{Column: "target_group_id", NamedC: "target_group_id", Type: enumor.String},
	{Column: "cloud_target_group_id", NamedC: "cloud_target_group_id", Type: enumor.String},
	{Column: "port", NamedC: "port", Type: enumor.Numeric},
	{Column: "weight", NamedC: "weight", Type: enumor.Numeric},
	{Column: "private_ip_address", NamedC: "private_ip_address", Type: enumor.Json},
	{Column: "public_ip_address", NamedC: "public_ip_address", Type: enumor.Json},
	{Column: "cloud_vpc_ids", NamedC: "cloud_vpc_ids", Type: enumor.Json},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// LoadBalancerTargetTable 负载均衡目标表
type LoadBalancerTargetTable struct {
	ID        string `db:"id" validate:"lte=64" json:"id"`
	AccountID string `db:"account_id" validate:"lte=64" json:"account_id"`

	InstType           enumor.InstType   `db:"inst_type" validate:"lte=255" json:"inst_type"`
	InstID             string            `db:"inst_id" validate:"lte=255" json:"inst_id"`
	CloudInstID        string            `db:"cloud_inst_id" validate:"lte=255" json:"cloud_inst_id"`
	InstName           string            `db:"inst_name" validate:"lte=255" json:"inst_name"`
	TargetGroupID      string            `db:"target_group_id" validate:"lte=255" json:"target_group_id"`
	CloudTargetGroupID string            `db:"cloud_target_group_id" validate:"lte=255" json:"cloud_target_group_id"`
	Port               int64             `db:"port" json:"port"`
	Weight             *int64            `db:"weight" json:"weight"`
	PrivateIPAddress   types.StringArray `db:"private_ip_address" json:"private_ip_address"`
	PublicIPAddress    types.StringArray `db:"public_ip_address" json:"public_ip_address"`
	CloudVpcIDs        types.StringArray `db:"cloud_vpc_ids" json:"cloud_vpc_ids"`
	Zone               string            `db:"zone" json:"zone"`
	Memo               *string           `db:"memo" json:"memo"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return load_balancer_target table name.
func (lbt LoadBalancerTargetTable) TableName() table.Name {
	return table.LoadBalancerTargetTable
}

// InsertValidate load_balancer_target table when insert.
func (lbt LoadBalancerTargetTable) InsertValidate() error {
	if err := validator.Validate.Struct(lbt); err != nil {
		return err
	}

	if len(lbt.CloudInstID) == 0 {
		return errors.New("cloud_inst_id is required")
	}

	if len(lbt.CloudTargetGroupID) == 0 {
		return errors.New("cloud_target_group_id is required")
	}

	if len(lbt.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate load_balancer_target table when update.
func (lbt LoadBalancerTargetTable) UpdateValidate() error {
	if err := validator.Validate.Struct(lbt); err != nil {
		return err
	}

	if len(lbt.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(lbt.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
