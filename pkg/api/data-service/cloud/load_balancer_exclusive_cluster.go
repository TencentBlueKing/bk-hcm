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
	"fmt"

	"hcm/pkg/api/core"
	corelb "hcm/pkg/api/core/cloud/load-balancer"
	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// -------------------------- List --------------------------

// ExclusiveClusterListResult define load balancer exclusive cluster list result, extension is returned as raw
// json since this list interface is not vendor-scoped.
type ExclusiveClusterListResult = core.ListResultT[corelb.ExclusiveClusterRaw]

// -------------------------- Create --------------------------

// ExclusiveClusterBatchCreateReq load balancer exclusive cluster batch create req.
type ExclusiveClusterBatchCreateReq[T corelb.ExclusiveClusterExtension] struct {
	Clusters []ExclusiveClusterCreate[T] `json:"clusters" validate:"required,min=1"`
}

// TCloudExclusiveClusterBatchCreateReq batch create tcloud load balancer exclusive cluster req.
type TCloudExclusiveClusterBatchCreateReq = ExclusiveClusterBatchCreateReq[corelb.TCloudExclusiveClusterExtension]

// ExclusiveClusterCreate define load balancer exclusive cluster create, bk_biz_id is not included, it is fixed to
// constant.UnassignedBiz by the handler.
type ExclusiveClusterCreate[T corelb.ExclusiveClusterExtension] struct {
	CloudID     string                `json:"cloud_id" validate:"required"`
	Name        string                `json:"name" validate:"required"`
	AccountID   string                `json:"account_id" validate:"required"`
	Region      string                `json:"region" validate:"required"`
	Zone        string                `json:"zone"`
	ClusterType enumor.ClusterType    `json:"cluster_type" validate:"required"`
	ClusterTag  string                `json:"cluster_tag"`
	Network     enumor.ClusterNetwork `json:"network"`
	Isp         enumor.ClusterIsp     `json:"isp"`
	Egress      string                `json:"egress"`
	IPVersion   string                `json:"ip_version"`

	MaxConn          *int64  `json:"max_conn"`
	ClbResourceCount int64   `json:"clb_resource_count"`
	Memo             *string `json:"memo"`

	Extension *T `json:"extension"`
}

// Validate load balancer exclusive cluster batch create request.
func (req *ExclusiveClusterBatchCreateReq[T]) Validate() error {
	if len(req.Clusters) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("clusters count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Update --------------------------

// ExclusiveClusterUpdate define load balancer exclusive cluster update, bk_biz_id is not included, it is only
// updated through ExclusiveClusterBatchUpdateBizIDReq.
type ExclusiveClusterUpdate[T corelb.ExclusiveClusterExtension] struct {
	ID string `json:"id" validate:"required"`

	Name        string                `json:"name"`
	Zone        string                `json:"zone"`
	ClusterType enumor.ClusterType    `json:"cluster_type"`
	ClusterTag  string                `json:"cluster_tag"`
	Network     enumor.ClusterNetwork `json:"network"`
	Isp         enumor.ClusterIsp     `json:"isp"`
	Egress      string                `json:"egress"`
	IPVersion   string                `json:"ip_version"`

	MaxConn          *int64  `json:"max_conn"`
	ClbResourceCount int64   `json:"clb_resource_count"`
	Memo             *string `json:"memo"`

	Extension *T `json:"extension"`
}

// ExclusiveClusterBatchUpdateReq load balancer exclusive cluster batch update req, used by sync logic to update
// cloud attributes, it does NOT contain bk_biz_id field.
type ExclusiveClusterBatchUpdateReq[T corelb.ExclusiveClusterExtension] struct {
	Clusters []ExclusiveClusterUpdate[T] `json:"clusters" validate:"required,min=1"`
}

// TCloudExclusiveClusterBatchUpdateReq batch update tcloud load balancer exclusive cluster cloud attributes req.
type TCloudExclusiveClusterBatchUpdateReq = ExclusiveClusterBatchUpdateReq[corelb.TCloudExclusiveClusterExtension]

// Validate load balancer exclusive cluster batch update request.
func (req *ExclusiveClusterBatchUpdateReq[T]) Validate() error {
	if len(req.Clusters) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("clusters count should <= %d", constant.BatchOperationMaxLimit)
	}

	return validator.Validate.Struct(req)
}

// ExclusiveClusterBatchUpdateBizIDReq load balancer exclusive cluster batch update bk_biz_id req(assign to biz).
// This interface does NOT do "whether already assigned" business pre-check, authentication or audit, it is only
// supposed to be called internally by cloud-server, and it reuses the same DAO BatchUpdateWithTx as cloud
// attribute update, isolated purely by the narrow field set of this request body.
type ExclusiveClusterBatchUpdateBizIDReq struct {
	ClusterIDs []string `json:"cluster_ids" validate:"required,min=1,max=100"`
	BkBizID    int64    `json:"bk_biz_id" validate:"required,gt=0"`
}

// Validate load balancer exclusive cluster batch update bk_biz_id request.
func (req *ExclusiveClusterBatchUpdateBizIDReq) Validate() error {
	if len(req.ClusterIDs) == 0 {
		return errors.New("cluster_ids is required")
	}

	if len(req.ClusterIDs) > constant.BatchOperationMaxLimit {
		return fmt.Errorf("cluster_ids count should <= %d", constant.BatchOperationMaxLimit)
	}

	if req.BkBizID <= 0 {
		return errors.New("bk_biz_id should > 0")
	}

	return validator.Validate.Struct(req)
}

// -------------------------- Delete --------------------------

// ExclusiveClusterBatchDeleteReq load balancer exclusive cluster batch delete req.
type ExclusiveClusterBatchDeleteReq struct {
	Filter *filter.Expression `json:"filter" validate:"required"`
}

// Validate load balancer exclusive cluster batch delete request.
func (req *ExclusiveClusterBatchDeleteReq) Validate() error {
	return validator.Validate.Struct(req)
}
