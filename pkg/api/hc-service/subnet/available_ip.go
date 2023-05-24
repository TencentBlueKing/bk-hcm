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

package subnet

import (
	"hcm/pkg/criteria/validator"
	"hcm/pkg/rest"
)

// ListCountIPReq count subnet available ip.
type ListCountIPReq struct {
	Region    string   `json:"region" validate:"required"`
	AccountID string   `json:"account_id" validate:"required"`
	IDs       []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate count subnet available ip.
func (req *ListCountIPReq) Validate() error {
	return validator.Validate.Struct(req)
}

// ListAvailIPResp list count subnet available ips response.
type ListAvailIPResp struct {
	rest.BaseResp `json:",inline"`
	Data          map[string]AvailIPResult `json:"data"`
}

// GetAvailIPResp get avail ip resp..
type GetAvailIPResp struct {
	rest.BaseResp `json:",inline"`
	Data          *AvailIPResult `json:"data"`
}

// ListAzureCountIPReq list azure count subnet available ip.
type ListAzureCountIPReq struct {
	ResourceGroupName string   `json:"resource_group_name" validate:"required"`
	VpcID             string   `json:"vpc_id" validate:"required"`
	AccountID         string   `json:"account_id" validate:"required"`
	IDs               []string `json:"ids" validate:"required,min=1,max=100"`
}

// Validate list azure count subnet available ip.
func (req *ListAzureCountIPReq) Validate() error {
	return validator.Validate.Struct(req)
}

// AvailIPResult count subnet available ips result.
type AvailIPResult struct {
	AvailableIPCount uint64 `json:"available_ip_count"`
	TotalIPCount     uint64 `json:"total_ip_count"`
	UsedIPCount      uint64 `json:"used_ip_count"`
}
