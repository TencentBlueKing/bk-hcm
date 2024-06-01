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
	"hcm/pkg/api/core"
	"hcm/pkg/api/core/cloud"
	"hcm/pkg/criteria/validator"
	"hcm/pkg/runtime/filter"
)

// ListVpcWithSubnetCountReq 查询vpc列表带有子网数量的请求
type ListVpcWithSubnetCountReq struct {
	Zone   string             `json:"zone" validate:"omitempty"`
	Filter *filter.Expression `json:"filter" validate:"required"`
	Page   *core.BasePage     `json:"page" validate:"required"`
}

// Validate list request.
func (req *ListVpcWithSubnetCountReq) Validate() error {
	if err := validator.Validate.Struct(req); err != nil {
		return err
	}

	pageOpt := &core.PageOption{
		EnableUnlimitedLimit: false,
		MaxLimit:             core.AggregationQueryMaxPageLimit,
		DisabledSort:         false,
	}
	if err := req.Page.Validate(pageOpt); err != nil {
		return err
	}

	return nil
}

// ListVpcWithSubnetCountResult ...
type ListVpcWithSubnetCountResult[T cloud.VpcExtension] struct {
	Count   uint64                  `json:"count"`
	Details []VpcWithSubnetCount[T] `json:"details"`
}

// VpcWithSubnetCount ...
type VpcWithSubnetCount[T cloud.VpcExtension] struct {
	cloud.Vpc[T]           `json:",inline"`
	SubnetCount            uint64 `json:"subnet_count"`
	CurrentZoneSubnetCount uint64 `json:"current_zone_subnet_count"`
}
