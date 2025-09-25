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

// TCloudLbUrlRuleColumns defines all the tcloud_lb_url_rule table's columns.
var TCloudLbUrlRuleColumns = utils.MergeColumns(nil, TCloudLbUrlRuleColumnsDescriptor)

// TCloudLbUrlRuleColumnsDescriptor is tcloud_lb_url_rule's column descriptors.
var TCloudLbUrlRuleColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "rule_type", NamedC: "rule_type", Type: enumor.String},

	{Column: "lb_id", NamedC: "lb_id", Type: enumor.String},
	{Column: "cloud_lb_id", NamedC: "cloud_lb_id", Type: enumor.String},
	{Column: "lbl_id", NamedC: "lbl_id", Type: enumor.String},
	{Column: "cloud_lbl_id", NamedC: "cloud_lbl_id", Type: enumor.String},
	{Column: "target_group_id", NamedC: "target_group_id", Type: enumor.String},
	{Column: "cloud_target_group_id", NamedC: "cloud_target_group_id", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "domain", NamedC: "domain", Type: enumor.String},
	{Column: "url", NamedC: "url", Type: enumor.String},
	{Column: "scheduler", NamedC: "scheduler", Type: enumor.String},
	{Column: "session_type", NamedC: "session_type", Type: enumor.String},
	{Column: "session_expire", NamedC: "session_expire", Type: enumor.Numeric},
	{Column: "health_check", NamedC: "health_check", Type: enumor.Json},
	{Column: "certificate", NamedC: "certificate", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// TCloudLbUrlRuleTable 腾讯云负载均衡四层/七层规则表
type TCloudLbUrlRuleTable struct {
	ID       string          `db:"id" validate:"lte=64" json:"id"`
	CloudID  string          `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name     string          `db:"name" validate:"lte=255" json:"name"`
	RuleType enumor.RuleType `db:"rule_type" validate:"lte=64" json:"rule_type"`

	LbID               string          `db:"lb_id" validate:"lte=255" json:"lb_id"`
	CloudLbID          string          `db:"cloud_lb_id" validate:"lte=255" json:"cloud_lb_id"`
	LblID              string          `db:"lbl_id" validate:"lte=255" json:"lbl_id"`
	CloudLBLID         string          `db:"cloud_lbl_id" validate:"lte=255" json:"cloud_lbl_id"`
	TargetGroupID      string          `db:"target_group_id" validate:"lte=255" json:"target_group_id"`
	CloudTargetGroupID string          `db:"cloud_target_group_id" validate:"lte=255" json:"cloud_target_group_id"`
	Region             string          `db:"region" validate:"lte=20" json:"region"`
	Domain             string          `db:"domain" json:"domain"`
	URL                string          `db:"url" json:"url"`
	Scheduler          string          `db:"scheduler" json:"scheduler"`
	SessionType        string          `db:"session_type" json:"session_type"`
	SessionExpire      int64           `db:"session_expire" json:"session_expire"`
	HealthCheck        types.JsonField `db:"health_check" json:"health_check"`
	Certificate        types.JsonField `db:"certificate" json:"certificate"`
	Memo               *string         `db:"memo" json:"memo"`
	BkBizID            int64           `db:"bk_biz_id" json:"bk_biz_id"`
	AccountID          string          `db:"account_id" validate:"lte=64" json:"account_id"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return tcloud_lb_url_rule table name.
func (tlbur TCloudLbUrlRuleTable) TableName() table.Name {
	return table.TCloudLbUrlRuleTable
}

// InsertValidate tcloud_lb_url_rule table when insert.
func (tlbur TCloudLbUrlRuleTable) InsertValidate() error {
	if err := validator.Validate.Struct(tlbur); err != nil {
		return err
	}

	if len(tlbur.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(tlbur.LblID) == 0 {
		return errors.New("lbl_id is required")
	}

	if len(tlbur.Creator) == 0 {
		return errors.New("creator is required")
	}
	return nil
}

// UpdateValidate tcloud_lb_url_rule table when update.
func (tlbur TCloudLbUrlRuleTable) UpdateValidate() error {
	if err := validator.Validate.Struct(tlbur); err != nil {
		return err
	}

	if len(tlbur.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(tlbur.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}

// TCloudLbUrlRuleWithListener define tcloud_lb_url_rule with listener.
type TCloudLbUrlRuleWithListener struct {
	TCloudLbUrlRuleTable `json:",inline"`
	LblName              string `db:"lbl_name" json:"lbl_name"`
	Protocol             string `db:"protocol" json:"protocol"`
	Port                 int    `db:"port" json:"port"`
}
