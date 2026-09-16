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
