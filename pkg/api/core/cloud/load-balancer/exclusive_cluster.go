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

package loadbalancer

import (
	"encoding/json"

	"hcm/pkg/criteria/enumor"
)

// BaseExclusiveCluster define base load balancer exclusive cluster, excluding the extension field.
type BaseExclusiveCluster struct {
	ID               string                `json:"id"`
	CloudID          string                `json:"cloud_id"`
	Name             string                `json:"name"`
	Vendor           enumor.Vendor         `json:"vendor"`
	AccountID        string                `json:"account_id"`
	BkBizID          int64                 `json:"bk_biz_id"`
	Region           string                `json:"region"`
	Zone             string                `json:"zone"`
	ClusterType      enumor.ClusterType    `json:"cluster_type"`
	ClusterTag       string                `json:"cluster_tag"`
	Network          enumor.ClusterNetwork `json:"network"`
	Isp              enumor.ClusterIsp     `json:"isp"`
	Egress           string                `json:"egress"`
	IPVersion        string                `json:"ip_version"`
	MaxConn          *int64                `json:"max_conn"`
	ClbResourceCount int64                 `json:"clb_resource_count"`
	Memo             *string               `json:"memo"`
	Creator          string                `json:"creator"`
	Reviser          string                `json:"reviser"`
	CreatedAt        string                `json:"created_at"`
	UpdatedAt        string                `json:"updated_at"`
}

// TCloudExclusiveClusterExtension define tcloud load balancer exclusive cluster extension.
type TCloudExclusiveClusterExtension struct {
	// MaxInFlow 最大入带宽，单位Mbps
	MaxInFlow int64 `json:"max_in_flow"`
	// MaxOutFlow 最大出带宽，单位Mbps
	MaxOutFlow int64 `json:"max_out_flow"`
	// MaxInPkg 最大入包量，个/秒
	MaxInPkg int64 `json:"max_in_pkg"`
	// MaxOutPkg 最大出包量，个/秒
	MaxOutPkg int64 `json:"max_out_pkg"`
	// MaxNewConn 最大新建连接数，个/秒
	MaxNewConn int64 `json:"max_new_conn"`
	// HTTPMaxNewConn http最大新建连接数，个/秒
	HTTPMaxNewConn int64 `json:"http_max_new_conn"`
	// HTTPSMaxNewConn https最大新建连接数，个/秒
	HTTPSMaxNewConn int64 `json:"https_max_new_conn"`
	// HTTPQps http QPS
	HTTPQps int64 `json:"http_qps"`
	// HTTPSQps https QPS
	HTTPSQps int64 `json:"https_qps"`
	// LoadBalanceDirectorCount 集群内转发机数目
	LoadBalanceDirectorCount int64 `json:"load_balance_director_count"`
	// ClustersVersion 集群版本
	ClustersVersion string `json:"clusters_version"`
	// DisasterRecoveryType 集群容灾类型（SINGLE-ZONE/DISASTER-RECOVERY/MUTUAL-DISASTER-RECOVERY）
	DisasterRecoveryType string `json:"disaster_recovery_type"`
	// ClustersZone 集群所在可用区
	ClustersZone TCloudExclusiveClusterZone `json:"clusters_zone"`
}

// TCloudExclusiveClusterZone define tcloud load balancer exclusive cluster zone.
type TCloudExclusiveClusterZone struct {
	// MasterZone 集群所在主可用区
	MasterZone []string `json:"master_zone"`
	// SlaveZone 集群所在备可用区
	SlaveZone []string `json:"slave_zone"`
}

// ExclusiveClusterExtension is the union of all vendor's load balancer exclusive cluster extension types.
type ExclusiveClusterExtension interface {
	TCloudExclusiveClusterExtension
}

// ExclusiveCluster define load balancer exclusive cluster with strong-typed extension.
type ExclusiveCluster[Ext ExclusiveClusterExtension] struct {
	BaseExclusiveCluster `json:",inline"`
	Extension            *Ext `json:"extension"`
}

// ExclusiveClusterRaw define load balancer exclusive cluster with raw json extension, used by list interface
// that is not vendor-scoped.
type ExclusiveClusterRaw struct {
	BaseExclusiveCluster `json:",inline"`
	Extension            json.RawMessage `json:"extension"`
}
