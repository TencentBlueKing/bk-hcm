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

// AzureSGRuleColumns defines all the azure security group rule table's columns.
var AzureSGRuleColumns = utils.MergeColumns(nil, AzureSGRuleColumnDescriptor)

// AzureSGRuleColumnDescriptor is Azure Security Group Rule's column descriptors.
var AzureSGRuleColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "cloud_security_group_id", NamedC: "cloud_security_group_id", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "provisioning_state", NamedC: "provisioning_state", Type: enumor.String},
	{Column: "etag", NamedC: "etag", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "destination_address_prefix", NamedC: "destination_address_prefix", Type: enumor.String},
	{Column: "destination_address_prefixes", NamedC: "destination_address_prefixes", Type: enumor.Json},
	{Column: "cloud_destination_app_security_group_ids", NamedC: "cloud_destination_app_security_group_ids",
		Type: enumor.Json},
	{Column: "destination_port_range", NamedC: "destination_port_range", Type: enumor.String},
	{Column: "destination_port_ranges", NamedC: "destination_port_ranges", Type: enumor.Json},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "source_address_prefix", NamedC: "source_address_prefix", Type: enumor.String},
	{Column: "source_address_prefixes", NamedC: "source_address_prefixes", Type: enumor.Json},
	{Column: "cloud_source_app_security_group_ids", NamedC: "cloud_source_app_security_group_ids", Type: enumor.Json},
	{Column: "source_port_range", NamedC: "source_port_range", Type: enumor.String},
	{Column: "source_port_ranges", NamedC: "source_port_ranges", Type: enumor.Json},
	{Column: "priority", NamedC: "priority", Type: enumor.Numeric},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "access", NamedC: "access", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// AzureSecurityGroupRuleTable define azure security group rule table.
type AzureSecurityGroupRuleTable struct {
	ID                                  string            `db:"id" validate:"lte=64" json:"id"`
	CloudID                             string            `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	CloudSecurityGroupID                string            `db:"cloud_security_group_id" validate:"lte=255" json:"cloud_security_group_id"`
	AccountID                           string            `db:"account_id" validate:"lte=64" json:"account_id"`
	SecurityGroupID                     string            `db:"security_group_id" validate:"lte=64" json:"security_group_id"`
	Type                                string            `db:"type" validate:"lte=20" json:"type"`
	ProvisioningState                   string            `db:"provisioning_state" validate:"lte=20" json:"provisioning_state"`
	Etag                                *string           `db:"etag" validate:"omitempty,lte=255" json:"etag"`
	Name                                string            `db:"name" validate:"lte=255" json:"name"`
	Memo                                *string           `db:"memo" validate:"omitempty,lte=140" json:"memo"`
	Region                              string            `db:"region" validate:"lte=20" json:"region"`
	DestinationAddressPrefix            *string           `db:"destination_address_prefix" validate:"omitempty,lte=255" json:"destination_address_prefix"`
	DestinationAddressPrefixes          types.StringArray `db:"destination_address_prefixes" json:"destination_address_prefixes"`
	CloudDestinationAppSecurityGroupIDs types.StringArray `db:"cloud_destination_app_security_group_ids" json:"cloud_destination_app_security_group_ids"`
	DestinationPortRange                *string           `db:"destination_port_range" validate:"omitempty,lte=255" json:"destination_port_range"`
	DestinationPortRanges               types.StringArray `db:"destination_port_ranges" json:"destination_port_ranges"`
	Protocol                            string            `db:"protocol" validate:"lte=10" json:"protocol"`
	SourceAddressPrefix                 *string           `db:"source_address_prefix" validate:"omitempty,lte=255" json:"source_address_prefix"`
	SourceAddressPrefixes               types.StringArray `db:"source_address_prefixes" json:"source_address_prefixes"`
	CloudSourceAppSecurityGroupIDs      types.StringArray `db:"cloud_source_app_security_group_ids" json:"cloud_source_app_security_group_ids"`
	SourcePortRange                     *string           `db:"source_port_range" validate:"omitempty,lte=255" json:"source_port_range"`
	SourcePortRanges                    types.StringArray `db:"source_port_ranges" json:"source_port_ranges"`
	Priority                            int32             `db:"priority" json:"priority"`
	Access                              string            `db:"access" validate:"lte=20" json:"access"`
	Creator                             string            `db:"creator" validate:"lte=64" json:"creator"`
	Reviser                             string            `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt                           types.Time        `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt                           types.Time        `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return azure security group rule table name.
func (t AzureSecurityGroupRuleTable) TableName() table.Name {
	return table.AzureSecurityGroupRuleTable
}

// InsertValidate azure security group rule table when insert.
func (t AzureSecurityGroupRuleTable) InsertValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) == 0 {
		return errors.New("id is required")
	}

	if len(t.Region) == 0 {
		return errors.New("region is required")
	}

	if len(t.CloudID) == 0 {
		return errors.New("cloud id is required")
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

// UpdateValidate azure security group rule table when update.
func (t AzureSecurityGroupRuleTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
