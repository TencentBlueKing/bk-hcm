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
	"hcm/pkg/adaptor/types/core"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/validator"
	cvt "hcm/pkg/tools/converter"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
)

// -------------------------- List Exclusive Cluster --------------------------

// TCloudExclusiveClusterListOption defines options to list tcloud load balancer exclusive cluster instances.
type TCloudExclusiveClusterListOption struct {
	Region string           `json:"region" validate:"required"`
	Page   *core.TCloudPage `json:"page" validate:"omitempty"`
	// CloudIDs 按集群ID过滤，为空表示不过滤
	CloudIDs []string `json:"cloud_ids" validate:"omitempty,max=20"`
	// Network 按集群网络类型过滤，为空表示不过滤
	Network enumor.ClusterNetwork `json:"network" validate:"omitempty"`
	// ClusterTypes 按集群类型过滤，为空表示不过滤
	ClusterTypes []enumor.ClusterType `json:"cluster_types" validate:"omitempty"`
}

// Validate tcloud load balancer exclusive cluster list option.
func (opt TCloudExclusiveClusterListOption) Validate() error {
	if err := validator.Validate.Struct(opt); err != nil {
		return err
	}

	if err := opt.Network.Validate(); err != nil {
		return err
	}

	for _, one := range opt.ClusterTypes {
		if err := one.Validate(); err != nil {
			return err
		}
	}

	if opt.Page != nil {
		if err := opt.Page.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// TCloudExclusiveCluster for load balancer exclusive cluster instance.
type TCloudExclusiveCluster struct {
	*tclb.Cluster
}

// GetCloudID get cloud id.
func (c TCloudExclusiveCluster) GetCloudID() string {
	return cvt.PtrToVal(c.ClusterId)
}
