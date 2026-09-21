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

package cslb

import (
	"errors"
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
)

// -------------------------- List --------------------------

// ListExclusiveClusterResult defines list load balancer exclusive cluster result, extension is a tcloud
// strong-typed object since this is currently the only supported vendor.
type ListExclusiveClusterResult = core.ListResultT[corelb.ExclusiveCluster[corelb.TCloudExclusiveClusterExtension]]

// -------------------------- Assign --------------------------

// AssignExclusiveClusterToBizReq define assign load balancer exclusive cluster to biz req.
type AssignExclusiveClusterToBizReq struct {
	ClusterIDs []string `json:"cluster_ids" validate:"required,min=1"`
	BkBizID    int64    `json:"bk_biz_id" validate:"required,min=0"`
}

// Validate assign load balancer exclusive cluster to biz request.
func (req *AssignExclusiveClusterToBizReq) Validate() error {
	if len(req.ClusterIDs) == 0 {
		return errors.New("cluster_ids is required")
	}

	if len(req.ClusterIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cluster_ids should <= %d", constant.BatchOperationMaxLimit)
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should > 0")
	}

	return validator.Validate.Struct(req)
}

// -------------------------- List Tags --------------------------

// ListExclusiveClusterTagsReq define list biz load balancer exclusive cluster tags req, used by purchase page
// to render cluster_tag/cluster dropdowns for exclusive load balancer spec. bk_biz_id is taken from the path
// parameter only, the request body does not accept a bk_biz_id field.
// Zones and BackZones decide the DB zone filter: one zone with empty back_zones
// matches the top-level zone column; two zones or any back_zones match
// extension.clusters_zone arrays.
type ListExclusiveClusterTagsReq struct {
	AccountID   string             `json:"account_id" validate:"required"`
	Region      string             `json:"region" validate:"required"`
	Isp         enumor.ClusterIsp  `json:"isp" validate:"required"`
	Zones       []string           `json:"zones" validate:"omitempty,dive,min=1"`
	BackZones   []string           `json:"back_zones" validate:"omitempty,dive,min=1"`
	ClusterType enumor.ClusterType `json:"cluster_type" validate:"omitempty"`
}

// Validate list biz load balancer exclusive cluster tags request.
func (req *ListExclusiveClusterTagsReq) Validate() error {
	if len(req.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(req.Region) == 0 {
		return errors.New("region is required")
	}

	switch req.Isp {
	case enumor.BGPClusterIsp, enumor.CMCCClusterIsp, enumor.CUCCClusterIsp, enumor.CTCCClusterIsp:
	default:
		return fmt.Errorf("isp must be one of BGP/CMCC/CUCC/CTCC, got: %s", req.Isp)
	}

	switch req.ClusterType {
	case "", enumor.TGWClusterType, enumor.STGWClusterType:
	default:
		return fmt.Errorf("cluster_type must be TGW or STGW if set, got: %s", req.ClusterType)
	}

	return validator.Validate.Struct(req)
}

// ListExclusiveClusterTagsResult define list biz load balancer exclusive cluster tags result.
type ListExclusiveClusterTagsResult struct {
	Details []ExclusiveClusterTagGroup `json:"details"`
}

// ExclusiveClusterTagGroup define a load balancer exclusive cluster group aggregated by cluster_tag+cluster_type.
type ExclusiveClusterTagGroup struct {
	ClusterTag  string                    `json:"cluster_tag"`
	ClusterType enumor.ClusterType        `json:"cluster_type"`
	Clusters    []ExclusiveClusterTagItem `json:"clusters"`
}

// ExclusiveClusterTagItem define a single load balancer exclusive cluster info within a tag group.
type ExclusiveClusterTagItem struct {
	CloudClusterID string                            `json:"cloud_cluster_id"`
	ClusterID      string                            `json:"cluster_id"`
	ClusterName    string                            `json:"cluster_name"`
	Egress         string                            `json:"egress"`
	Isp            enumor.ClusterIsp                 `json:"isp"`
	ClusterZone    corelb.TCloudExclusiveClusterZone `json:"cluster_zone"`
}

// -------------------------- List Idle Vips --------------------------

// ListExclusiveClusterIdleVipsReq define list biz load balancer exclusive cluster idle vips req, used by purchase
// page to render the "specify ip" dropdown after a TGW(layer-4) cluster is chosen. bk_biz_id is taken from the
// path parameter only, the request body does not accept a bk_biz_id field.
type ListExclusiveClusterIdleVipsReq struct {
	AccountID      string `json:"account_id" validate:"required"`
	Region         string `json:"region" validate:"required"`
	CloudClusterID string `json:"cloud_cluster_id" validate:"required"`
}

// Validate list biz load balancer exclusive cluster idle vips request.
func (req *ListExclusiveClusterIdleVipsReq) Validate() error {
	if len(req.AccountID) == 0 {
		return errors.New("account_id is required")
	}

	if len(req.Region) == 0 {
		return errors.New("region is required")
	}

	if len(req.CloudClusterID) == 0 {
		return errors.New("cloud_cluster_id is required")
	}

	return validator.Validate.Struct(req)
}

// ListExclusiveClusterIdleVipsResult define list biz load balancer exclusive cluster idle vips result. this is a
// real-time query result, not persisted, and may become stale immediately after the response is returned.
type ListExclusiveClusterIdleVipsResult struct {
	Count   uint64   `json:"count"`
	Details []string `json:"details"`
}
