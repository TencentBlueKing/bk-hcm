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
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	corecloud "hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/dal/table"
	"hcm/pkg/dal/table/types"
	"hcm/pkg/dal/table/utils"
)

// GcpFirewallRuleColumns defines all the gcp firewall rule table's columns.
var GcpFirewallRuleColumns = utils.MergeColumns(nil, GcpFirewallRuleTableColumnDescriptor)

// GcpFirewallRuleTableColumnDescriptor is gcp firewall rule table's column descriptors.
var GcpFirewallRuleTableColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "priority", NamedC: "priority", Type: enumor.Numeric},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "cloud_vpc_id", NamedC: "cloud_vpc_id", Type: enumor.String},
	{Column: "vpc_self_link", NamedC: "vpc_self_link", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "vpc_id", NamedC: "vpc_id", Type: enumor.String},
	{Column: "source_ranges", NamedC: "source_ranges", Type: enumor.Json},
	{Column: "destination_ranges", NamedC: "destination_ranges", Type: enumor.Json},
	{Column: "source_tags", NamedC: "source_tags", Type: enumor.Json},
	{Column: "target_tags", NamedC: "target_tags", Type: enumor.Json},
	{Column: "source_service_accounts", NamedC: "source_service_accounts", Type: enumor.Json},
	{Column: "target_service_accounts", NamedC: "target_service_accounts", Type: enumor.Json},
	{Column: "denied", NamedC: "denied", Type: enumor.Json},
	{Column: "allowed", NamedC: "allowed", Type: enumor.Json},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "type", NamedC: "type", Type: enumor.String},
	{Column: "log_enable", NamedC: "log_enable", Type: enumor.Boolean},
	{Column: "disabled", NamedC: "disabled", Type: enumor.Boolean},
	{Column: "self_link", NamedC: "self_link", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// GcpFirewallRuleTable define gcp firewall rule table.
type GcpFirewallRuleTable struct {
	ID                    string            `db:"id" validate:"lte=64" json:"id"`
	CloudID               string            `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	AccountID             string            `db:"account_id" validate:"lte=64" json:"account_id"`
	Name                  string            `db:"name" validate:"lte=65" json:"name"`
	Priority              int64             `db:"priority" json:"priority"`
	Memo                  string            `db:"memo" validate:"lte=2048" json:"memo"`
	CloudVpcID            string            `db:"cloud_vpc_id" validate:"lte=255" json:"cloud_vpc_id"`
	VpcSelfLink           string            `db:"vpc_self_link" validate:"lte=255" json:"cloud_vpc_self_link"`
	VpcID                 string            `db:"vpc_id" validate:"lte=64" json:"vpc_id"`
	SourceRanges          types.StringArray `db:"source_ranges" json:"source_ranges"`
	DestinationRanges     types.StringArray `db:"destination_ranges" json:"destination_ranges"`
	SourceTags            types.StringArray `db:"source_tags" json:"source_tags"`
	TargetTags            types.StringArray `db:"target_tags" json:"target_tags"`
	SourceServiceAccounts types.StringArray `db:"source_service_accounts" json:"source_service_accounts"`
	TargetServiceAccounts types.StringArray `db:"target_service_accounts" json:"target_service_accounts"`
	Denied                GcpProtocolSets   `db:"denied" json:"denied"`
	Allowed               GcpProtocolSets   `db:"allowed" json:"allowed"`
	BkBizID               int64             `db:"bk_biz_id" json:"bk_biz_id"`
	Type                  string            `db:"type" validate:"lte=20" json:"type"`
	LogEnable             bool              `db:"log_enable" json:"log_enable"`
	Disabled              bool              `db:"disabled" json:"disabled"`
	SelfLink              string            `db:"self_link" validate:"lte=255" json:"self_link"`
	Creator               string            `db:"creator" validate:"lte=64" json:"creator"`
	Reviser               string            `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt             types.Time        `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt             types.Time        `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
}

// TableName return gcp firewall rule table name.
func (t GcpFirewallRuleTable) TableName() table.Name {
	return table.GcpFirewallRuleTable
}

// InsertValidate gcp firewall rule table when insert.
func (t GcpFirewallRuleTable) InsertValidate() error {
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

	if t.BkBizID == 0 {
		return errors.New("bk_biz_id is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.CloudVpcID) == 0 {
		return errors.New("cloud vpc id is required")
	}

	if len(t.VpcSelfLink) == 0 {
		return errors.New("cloud vpc self link is required")
	}

	if len(t.VpcID) == 0 {
		return errors.New("vpc id is required")
	}

	if len(t.Type) == 0 {
		return errors.New("type is required")
	}

	if len(t.SelfLink) == 0 {
		return errors.New("self_link is required")
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.Reviser) == 0 {
		return errors.New("reviser is required")
	}

	return nil
}

// UpdateValidate gcp firewall rule table when update.
func (t GcpFirewallRuleTable) UpdateValidate() error {
	// length validate.
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	return nil
}

// GcpProtocolSets define gcp protocol set.
type GcpProtocolSets []corecloud.GcpProtocolSet

// Scan is used to decode raw message which is read from db into GcpProtocolSets.
func (set *GcpProtocolSets) Scan(raw interface{}) error {
	if set == nil || raw == nil {
		return nil
	}

	switch v := raw.(type) {
	case []byte:
		if err := json.Unmarshal(v, &set); err != nil {
			return fmt.Errorf("decode into gcp protocol sets failed, err: %v", err)
		}
		return nil

	case string:
		if err := json.Unmarshal([]byte(v), &set); err != nil {
			return fmt.Errorf("decode into gcp protocol sets failed, err: %v", err)
		}
		return nil

	default:
		return fmt.Errorf("unsupported gcp protocol sets raw type: %T", v)
	}
}

// Value encode the GcpProtocolSets to a json raw, so that it can be stored to db with json raw.
func (set GcpProtocolSets) Value() (driver.Value, error) {
	if set == nil {
		return nil, nil
	}

	return json.Marshal(set)
}
