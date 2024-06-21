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

// TargetGroupListenerRuleRelColumns defines all the target_group_listener_rule_rel table's columns.
var TargetGroupListenerRuleRelColumns = utils.MergeColumns(nil, TargetGroupListenerRuleRelColumnsDescriptor)

// TargetGroupListenerRuleRelColumnsDescriptor is target_group_listener_rule_rel's column descriptors.
var TargetGroupListenerRuleRelColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "listener_rule_id", NamedC: "listener_rule_id", Type: enumor.String},
	{Column: "cloud_listener_rule_id", NamedC: "cloud_listener_rule_id", Type: enumor.String},
	{Column: "listener_rule_type", NamedC: "listener_rule_type", Type: enumor.String},
	{Column: "target_group_id", NamedC: "target_group_id", Type: enumor.String},
	{Column: "cloud_target_group_id", NamedC: "cloud_target_group_id", Type: enumor.String},
	{Column: "lb_id", NamedC: "lb_id", Type: enumor.String},
	{Column: "cloud_lb_id", NamedC: "cloud_lb_id", Type: enumor.String},
	{Column: "lbl_id", NamedC: "lbl_id", Type: enumor.String},
	{Column: "cloud_lbl_id", NamedC: "cloud_lbl_id", Type: enumor.String},
	{Column: "binding_status", NamedC: "binding_status", Type: enumor.String},
	{Column: "detail", NamedC: "detail", Type: enumor.Json},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// TargetGroupListenerRuleRelTable 目标组监听器关系表
type TargetGroupListenerRuleRelTable struct {
	ID     string        `db:"id" validate:"lte=64" json:"id"`
	Vendor enumor.Vendor `db:"vendor" validate:"lte=16" json:"vendor"`

	ListenerRuleID      string `db:"listener_rule_id" validate:"lte=64" json:"listener_rule_id"`
	CloudListenerRuleID string `db:"cloud_listener_rule_id" validate:"lte=64" json:"cloud_listener_rule_id"`

	ListenerRuleType enumor.RuleType `db:"listener_rule_type" validate:"lte=64" json:"listener_rule_type"`

	TargetGroupID      string `db:"target_group_id" validate:"lte=64" json:"target_group_id"`
	CloudTargetGroupID string `db:"cloud_target_group_id" validate:"lte=64" json:"cloud_target_group_id"`
	LbID               string `db:"lb_id" validate:"lte=64" json:"lb_id"`
	CloudLbID          string `db:"cloud_lb_id" validate:"lte=64" json:"cloud_lb_id"`
	LblID              string `db:"lbl_id" validate:"lte=64" json:"lbl_id"`
	CloudLblID         string `db:"cloud_lbl_id" validate:"lte=64" json:"cloud_lbl_id"`

	BindingStatus enumor.BindingStatus `db:"binding_status" validate:"lte=64" json:"binding_status"`
	Detail        types.JsonField      `db:"detail" json:"detail"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return table name.
func (tlrr TargetGroupListenerRuleRelTable) TableName() table.Name {
	return table.TargetGroupListenerRuleRelTable
}

// InsertValidate validate table when insert.
func (tlrr TargetGroupListenerRuleRelTable) InsertValidate() error {

	if len(tlrr.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(tlrr.TargetGroupID) == 0 {
		return errors.New("target_group_id is required")
	}

	if len(tlrr.CloudTargetGroupID) == 0 {
		return errors.New("cloud_target_group_id is required")
	}

	if len(tlrr.ListenerRuleID) == 0 {
		return errors.New("listener_rule_id is required")
	}

	if len(tlrr.CloudListenerRuleID) == 0 {
		return errors.New("cloud_listener_rule_id is required")
	}

	if len(tlrr.LblID) == 0 {
		return errors.New("lbl_id is required")
	}

	if len(tlrr.CloudLblID) == 0 {
		return errors.New("cloud_lbl_id is required")
	}

	if len(tlrr.ListenerRuleType) == 0 {
		return errors.New("listener_rule_type is required")
	}

	if len(tlrr.Creator) == 0 {
		return errors.New("creator is required")
	}

	return validator.Validate.Struct(tlrr)
}

// UpdateValidate validate table when update.
func (tlrr TargetGroupListenerRuleRelTable) UpdateValidate() error {
	if err := validator.Validate.Struct(tlrr); err != nil {
		return err
	}

	if len(tlrr.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(tlrr.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
