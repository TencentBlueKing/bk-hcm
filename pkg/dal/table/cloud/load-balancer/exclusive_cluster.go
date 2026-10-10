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

// LoadBalancerExclusiveClusterColumns defines all the load_balancer_exclusive_cluster table's columns.
var LoadBalancerExclusiveClusterColumns = utils.MergeColumns(nil, LoadBalancerExclusiveClusterColumnDescriptor)

// LoadBalancerExclusiveClusterColumnDescriptor is load_balancer_exclusive_cluster's column descriptors.
var LoadBalancerExclusiveClusterColumnDescriptor = utils.ColumnDescriptors{
	{Column: "id", NamedC: "id", Type: enumor.String},
	{Column: "cloud_id", NamedC: "cloud_id", Type: enumor.String},
	{Column: "name", NamedC: "name", Type: enumor.String},
	{Column: "vendor", NamedC: "vendor", Type: enumor.String},
	{Column: "account_id", NamedC: "account_id", Type: enumor.String},
	{Column: "bk_biz_id", NamedC: "bk_biz_id", Type: enumor.Numeric},
	{Column: "region", NamedC: "region", Type: enumor.String},
	{Column: "zone", NamedC: "zone", Type: enumor.String},
	{Column: "cluster_type", NamedC: "cluster_type", Type: enumor.String},
	{Column: "cluster_tag", NamedC: "cluster_tag", Type: enumor.String},
	{Column: "network", NamedC: "network", Type: enumor.String},
	{Column: "isp", NamedC: "isp", Type: enumor.String},
	{Column: "egress", NamedC: "egress", Type: enumor.String},
	{Column: "ip_version", NamedC: "ip_version", Type: enumor.String},
	{Column: "max_conn", NamedC: "max_conn", Type: enumor.Numeric},
	{Column: "clb_resource_count", NamedC: "clb_resource_count", Type: enumor.Numeric},
	{Column: "extension", NamedC: "extension", Type: enumor.Json},
	{Column: "memo", NamedC: "memo", Type: enumor.String},
	{Column: "creator", NamedC: "creator", Type: enumor.String},
	{Column: "reviser", NamedC: "reviser", Type: enumor.String},
	{Column: "created_at", NamedC: "created_at", Type: enumor.Time},
	{Column: "updated_at", NamedC: "updated_at", Type: enumor.Time},
}

// LoadBalancerExclusiveClusterTable CLB独占集群表
type LoadBalancerExclusiveClusterTable struct {
	// ID HCM主键ID
	ID string `db:"id" validate:"lte=64" json:"id"`
	// CloudID 云上集群ID
	CloudID string `db:"cloud_id" validate:"lte=64" json:"cloud_id"`
	// Name 集群名称
	Name string `db:"name" validate:"lte=255" json:"name"`
	// Vendor 云厂商
	Vendor enumor.Vendor `db:"vendor" validate:"lte=16" json:"vendor"`
	// AccountID 账号ID
	AccountID string `db:"account_id" validate:"lte=64" json:"account_id"`
	// BkBizID 业务ID，-1表示未分配
	BkBizID int64 `db:"bk_biz_id" json:"bk_biz_id"`
	// Region 地域
	Region string `db:"region" validate:"lte=20" json:"region"`
	// Zone 可用区
	Zone string `db:"zone" validate:"lte=64" json:"zone"`
	// ClusterType 集群类型：TGW四层/STGW七层/VPCGW内网
	ClusterType enumor.ClusterType `db:"cluster_type" validate:"lte=16" json:"cluster_type"`
	// ClusterTag 集群标签，空表示未打标签
	ClusterTag string `db:"cluster_tag" validate:"lte=128" json:"cluster_tag"`
	// Network 网络类型：Public/Private/Hybrid
	Network enumor.ClusterNetwork `db:"network" validate:"lte=16" json:"network"`
	// Isp 运营商：BGP/CMCC/CUCC/CTCC/INTERNAL/MIX
	Isp enumor.ClusterIsp `db:"isp" validate:"lte=16" json:"isp"`
	// Egress 网络出口，如center_egress1
	Egress string `db:"egress" validate:"lte=64" json:"egress"`
	// IPVersion IP版本
	IPVersion string `db:"ip_version" validate:"lte=16" json:"ip_version"`
	// MaxConn 最大连接数，STGW可能无值
	MaxConn *int64 `db:"max_conn" json:"max_conn"`
	// ClbResourceCount 集群内已有CLB实例数
	ClbResourceCount int64 `db:"clb_resource_count" json:"clb_resource_count"`
	// Extension 云上扩展字段
	Extension types.JsonField `db:"extension" json:"extension"`
	// Memo 备注
	Memo *string `db:"memo" json:"memo"`
	// TenantID 租户ID
	TenantID string `db:"tenant_id" json:"tenant_id"`
	// Creator 创建者
	Creator string `db:"creator" validate:"lte=64" json:"creator"`
	// Reviser 更新者
	Reviser string `db:"reviser" validate:"lte=64" json:"reviser"`
	// CreatedAt 创建时间
	CreatedAt types.Time `db:"created_at" validate:"excluded_unless" json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt types.Time `db:"updated_at" validate:"excluded_unless" json:"updated_at"`
}

// TableName return load_balancer_exclusive_cluster table name.
func (t LoadBalancerExclusiveClusterTable) TableName() table.Name {
	return table.LoadBalancerExclusiveClusterTable
}

// InsertValidate validate load_balancer_exclusive_cluster table on insert.
func (t LoadBalancerExclusiveClusterTable) InsertValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ID) != 0 {
		return errors.New("id can not set")
	}

	if len(t.CloudID) == 0 {
		return errors.New("cloud_id is required")
	}

	if len(t.Name) == 0 {
		return errors.New("name is required")
	}

	if len(t.Vendor) == 0 {
		return errors.New("vendor is required")
	}

	if len(t.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(t.Region) == 0 {
		return errors.New("region is required")
	}

	if len(t.ClusterType) == 0 {
		return errors.New("cluster_type is required")
	}

	if err := t.ClusterType.Validate(); err != nil {
		return err
	}

	if err := t.Network.Validate(); err != nil {
		return err
	}

	if err := t.Isp.Validate(); err != nil {
		return err
	}

	if len(t.Creator) == 0 {
		return errors.New("creator is required")
	}

	if len(t.CreatedAt) != 0 {
		return errors.New("created_at can not set")
	}

	if len(t.UpdatedAt) != 0 {
		return errors.New("updated_at can not set")
	}

	return nil
}

// UpdateValidate validate load_balancer_exclusive_cluster table on update.
func (t LoadBalancerExclusiveClusterTable) UpdateValidate() error {
	if err := validator.Validate.Struct(t); err != nil {
		return err
	}

	if len(t.ClusterType) != 0 {
		if err := t.ClusterType.Validate(); err != nil {
			return err
		}
	}

	if err := t.Network.Validate(); err != nil {
		return err
	}

	if err := t.Isp.Validate(); err != nil {
		return err
	}

	if len(t.CreatedAt) != 0 {
		return errors.New("created_at can not update")
	}

	if len(t.Creator) != 0 {
		return errors.New("creator can not update")
	}

	if len(t.UpdatedAt) != 0 {
		return errors.New("updated_at can not update")
	}

	return nil
}
