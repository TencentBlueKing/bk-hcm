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

// LoadBalancerListenerColumns defines all the load_balancer_listener table's columns.
var LoadBalancerListenerColumns = utils.MergeColumns(nil, LoadBalancerListenerColumnsDescriptor)

// LoadBalancerListenerColumnsDescriptor is load_balancer_listener's column descriptors.
var LoadBalancerListenerColumnsDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},

	{Column: "lb_id", NamedC: "lb_id", Type: enumor.String},
	{Column: "cloud_lb_id", NamedC: "cloud_lb_id", Type: enumor.String},
	{Column: "protocol", NamedC: "protocol", Type: enumor.String},
	{Column: "port", NamedC: "port", Type: enumor.Numeric},
	{Column: "default_domain", NamedC: "default_domain", Type: enumor.String},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zones", NamedC: "zones", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "sni_switch", NamedC: "sni_switch", Type: enumor.Numeric},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},

	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// LoadBalancerListenerTable 负载均衡监听器表
type LoadBalancerListenerTable struct {
	ID        string        `db:"id" validate:"lte=64" json:"id"`
	CloudID   string        `db:"cloud_id" validate:"lte=255" json:"cloud_id"`
	Name      string        `db:"name" validate:"lte=255" json:"name"`
	Vendor    enumor.Vendor `db:"vendor" validate:"lte=16"  json:"vendor"`
	AccountID string        `db:"account_id" validate:"lte=64" json:"account_id"`
	BkBizID   int64         `db:"bk_biz_id" json:"bk_biz_id"`

	LBID          string              `db:"lb_id" validate:"lte=255" json:"lb_id"`
	CloudLBID     string              `db:"cloud_lb_id" validate:"lte=255" json:"cloud_lb_id"`
	Protocol      enumor.ProtocolType `db:"protocol" json:"protocol"`
	Port          int64               `db:"port" json:"port"`
	DefaultDomain string              `db:"default_domain" json:"default_domain"`
	Region        string              `db:"region" json:"region"`
	Zones         types.StringArray   `db:"zones" json:"zones"`
	Memo          *string             `db:"memo" json:"memo"`
	SniSwitch     enumor.SniType      `db:"sni_switch" json:"sni_switch"`
	Extension     types.JsonField     `db:"extension" json:"extension"`

	Creator   string     `db:"creator" validate:"lte=64" json:"creator"`
	Reviser   string     `db:"reviser" validate:"lte=64" json:"reviser"`
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return load_balancer_listener table name.
func (lbl LoadBalancerListenerTable) TableName() table.Name {
	return table.LoadBalancerListenerTable
}

// InsertValidate load_balancer_listener table when insert.
func (lbl LoadBalancerListenerTable) InsertValidate() error {
	if err := validator.Validate.Struct(lbl); err != nil {
		return err
	}

	if len(lbl.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(lbl.Name) == 0 {
		return errors.New("name is required")
	}

	if len(lbl.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(lbl.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(lbl.Creator) == 0 {
		return errors.New("creator is required")
	}

	return nil
}

// UpdateValidate load_balancer_listener table when update.
func (lbl LoadBalancerListenerTable) UpdateValidate() error {
	if err := validator.Validate.Struct(lbl); err != nil {
		return err
	}

	if len(lbl.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(lbl.Reviser) == 0 {
		return errors.New("reviser can not be empty")
	}

	return nil
}
