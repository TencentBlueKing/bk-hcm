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

// Package dataservice defines data-service api call protocols.
package dataservice

import (
	"hcm/pkg/api/core"
	"hcm/pkg/dal/dao/types"
	"hcm/pkg/rest"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/thirdparty/api-gateway/iam"
)

// ListInstancesReq defines list instances for iam pull resource callback http request.
type ListInstancesReq struct {
	ResourceType iam.TypeID         `json:"resource_type"`
	Filter       *filter.Expression `json:"filter"`
	Page         *core.BasePage     `json:"page"`
}

// ListInstancesResp  defines list instances for iam pull resource callback result.
type ListInstancesResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ListInstancesResult `json:"data"`
}

// ListInstancesResult defines list instances for iam pull resource callback result.
type ListInstancesResult struct {
	Count   uint64                   `json:"count"`
	Details []types.InstanceResource `json:"details"`
}
