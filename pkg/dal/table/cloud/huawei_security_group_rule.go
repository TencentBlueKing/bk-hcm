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
	"time"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/utils"
)

// HuaWeiSGRuleColumns defines all the huawei security group rule table's columns.
var HuaWeiSGRuleColumns = utils.MergeColumns(nil, HuaWeiSGRuleColumnDescriptor)

// HuaWeiSGRuleColumnDescriptor is HuaWei Security Group Rule's column descriptors.
var HuaWeiSGRuleColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "cloud_security_group_id", NamedC: "cloud_security_group_id", Type: enumor.String},
	{Column: "security_group_id", NamedC: "security_group_id", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "cloud_project_id", NamedC: "cloud_project_id", Type: enumor.String},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "ethertype", NamedC: "ethertype", Type: enumor.String},
	{Column: "action", NamedC: "action", Type: enumor.String},
	{Column: "cloud_remote_group_id", NamedC: "cloud_remote_group_id", Type: enumor.String},
	{Column: "remote_ip_prefix", NamedC: "remote_ip_prefix", Type: enumor.String},
	{Column: "cloud_remote_address_group_id", NamedC: "cloud_remote_address_group_id", Type: enumor.String},
	{Column: "port", NamedC: "port", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "priority", NamedC: "priority", Type: enumor.Numeric},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// HuaWeiSecurityGroupRuleTable define huawei security group rule table.
type HuaWeiSecurityGroupRuleTable struct {
	ID                        string     `db:"id" validate:"lte=64" json:"id"`
	CloudID                   string     `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Type                      string     `db:"type" validate:"lte=20" json:"type"`
	CloudSecurityGroupID      string     `db:"cloud_security_group_id" validate:"lte=255" json:"cloud_security_group_id"`
	SecurityGroupID           string     `db:"security_group_id" validate:"lte=64" json:"security_group_id"`
	AccountID                 string     `db:"account_id" validate:"lte=64" json:"account_id"`
	CloudProjectID            string     `db:"cloud_project_id" validate:"lte=255" json:"cloud_project_id"`
	Memo                      *string    `db:"memo" validate:"omitempty,lte=255" json:"memo"`
	Action                    string     `db:"action" validate:"lte=10" json:"action"`
	Region                    string     `db:"region" validate:"lte=20" json:"region"`
	Protocol                  string     `db:"protocol" validate:"lte=10" json:"protocol"`
	Ethertype                 string     `db:"ethertype" validate:"lte=10" json:"ethertype"`
	CloudRemoteGroupID        string     `db:"cloud_remote_group_id" validate:"lte=255" json:"cloud_remote_group_id"`
	RemoteIPPrefix            string     `db:"remote_ip_prefix" validate:"lte=255" json:"remote_ip_prefix"`
	CloudRemoteAddressGroupID string     `db:"cloud_remote_address_group_id" validate:"lte=255" json:"cloud_remote_address_group_id"`
	Port                      string     `db:"port" validate:"lte=255" json:"port"`
	Priority                  int64      `db:"priority" json:"priority"`
	Creator                   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser                   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt                 *time.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt                 *time.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return huawei security group rule table name.
func (t HuaWeiSecurityGroupRuleTable) TableName() table.Name {
	return table.HuaWeiSecurityGroupRuleTable
}

// InsertValidate huawei security group rule table when insert.
func (t HuaWeiSecurityGroupRuleTable) InsertValidate() error {
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

	if len(t.Action) == 0 {
		return errors.New("action is required")
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

// UpdateValidate huawei security group rule table when update.
func (t HuaWeiSecurityGroupRuleTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}
