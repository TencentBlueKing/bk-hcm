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

// AwsSGRuleColumns defines all the aws security group rule table's columns.
var AwsSGRuleColumns = utils.MergeColumns(nil, AwsSGRuleColumnDescriptor)

// AwsSGRuleColumnDescriptor is Aws Security Group Rule's column descriptors.
var AwsSGRuleColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "ipv4_cidr", NamedC: "ipv4_cidr", Type: enumor.String},
	{Column: "ipv6_cidr", NamedC: "ipv6_cidr", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "from_port", NamedC: "from_port", Type: enumor.Numeric},
	{Column: "to_port", NamedC: "to_port", Type: enumor.Numeric},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "cloud_prefix_list_id", NamedC: "cloud_prefix_list_id", Type: enumor.String},
	{Column: "cloud_target_security_group_id", NamedC: "cloud_target_security_group_id", Type: enumor.String},
	{Column: "cloud_security_group_id", NamedC: "cloud_security_group_id", Type: enumor.String},
	{Column: "cloud_group_owner_id", NamedC: "cloud_group_owner_id", Type: enumor.String},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AwsSecurityGroupRuleTable define aws security group rule table.
type AwsSecurityGroupRuleTable struct {
	ID                         string     `db:"id" json:"id" validate:"lte=64"`
	CloudID                    string     `db:"cloud_id" json:"cloud_id" validate:"lte=255"`
	IPv4Cidr                   *string    `db:"ipv4_cidr" json:"ipv4_cidr" validate:"omitempty,lte=255"`
	IPv6Cidr                   *string    `db:"ipv6_cidr" json:"ipv6_cidr" validate:"omitempty,lte=255"`
	Memo                       *string    `db:"memo" json:"memo" validate:"omitempty,lte=255"`
	Type                       string     `db:"type" json:"type" validate:"lte=20"`
	FromPort                   *int64     `db:"from_port" json:"from_port"`
	ToPort                     *int64     `db:"to_port" json:"to_port"`
	Protocol                   *string    `db:"protocol" json:"protocol" validate:"omitempty,lte=10"`
	CloudPrefixListID          *string    `db:"cloud_prefix_list_id" json:"cloud_prefix_list_id" validate:"omitempty,lte=255"`
	CloudTargetSecurityGroupID *string    `db:"cloud_target_security_group_id" json:"cloud_target_security_group_id" validate:"omitempty,lte=255"`
	CloudSecurityGroupID       string     `db:"cloud_security_group_id" json:"cloud_security_group_id" validate:"lte=255"`
	CloudGroupOwnerID          string     `db:"cloud_group_owner_id" json:"cloud_group_owner_id" validate:"lte=255"`
	SecurityGroupID            string     `db:"security_group_id" json:"security_group_id" validate:"lte=64"`
	AccountID                  string     `db:"account_id" json:"account_id" validate:"lte=64"`
	Region                     string     `db:"region" json:"region" validate:"lte=20"`
	Creator                    string     `db:"creator" json:"creator" validate:"lte=64"`
	Reviser                    string     `db:"reviser" json:"reviser" validate:"lte=64"`
	CreatedAt                  types.Time `db:"created_at" json:"created_at" validate:"excluded_unless"`
	UpdatedAt                  types.Time `db:"updated_at" json:"updated_at" validate:"excluded_unless"`
}

// TableName return aws security group rule table name.
func (t AwsSecurityGroupRuleTable) TableName() table.Name {
	return table.AwsSecurityGroupRuleTable
}

// InsertValidate aws security group rule table when insert.
func (t AwsSecurityGroupRuleTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.CloudID) == 0 {
		return errors.New("cloud id is required")
	}

	if len(t.Region) == 0 {
		return errors.New("region is required")
	}

	if len(t.Type) == 0 {
		return errors.New("type is required")
	}

	if len(t.CloudSecurityGroupID) == 0 {
		return errors.New("cloud security group id is required")
	}

	if len(t.SecurityGroupID) == 0 {
		return errors.New("security group id is required")
	}

	if len(t.AccountID) == 0 {
		return errors.New("account id is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate aws security group rule table when update.
func (t AwsSecurityGroupRuleTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
