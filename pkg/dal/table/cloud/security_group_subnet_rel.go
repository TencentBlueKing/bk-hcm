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
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// SecurityGroupSubnetRelColumns defines all the security group subnet rel table's columns.
var SecurityGroupSubnetRelColumns = utils.MergeColumns(nil, SecurityGroupSubnetRelColumnDescriptor)

// SecurityGroupSubnetRelColumnDescriptor is vpc security group rel table column descriptors.
var SecurityGroupSubnetRelColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.Numeric},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
}

// SecurityGroupSubnetRelTable define security group subnet rel table.
type SecurityGroupSubnetRelTable struct {
	ID              uint64     `db:"id" validate:"required"`
	SubnetID        string     `db:"subnet_id" validate:"required,lte=64"`
	SecurityGroupID string     `db:"security_group_id" validate:"required,lte=64"`
	Creator         string     `db:"creator" validate:"required,lte=64"`
	CreatedAt       *time.Time `db:"created_at" validate:"excluded_unless"`
}

// TableName return subnet and security group rel table name.
func (t SecurityGroupSubnetRelTable) TableName() table.Name {
	return table.SecurityGroupSubnetTable
}

// InsertValidate security group and subnet rel table when insert.
func (t SecurityGroupSubnetRelTable) InsertValidate() error {
	return validator.Validate.Struct(t)
}
