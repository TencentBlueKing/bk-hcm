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

package image

import (
	"hcm/pkg/rest"
)

// ImageExtListResp ...
type ImageExtListResp[T ImageExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          *ImageExtListResult[T] `json:"data"`
}

// ImageExtListResult ...
type ImageExtListResult[T ImageExtensionResult] struct {
	Count   *uint64              `json:"count,omitempty"`
	Details []*ImageExtResult[T] `json:"details"`
}

// ImageExtResult ...
type ImageExtResult[T ImageExtensionResult] struct {
	ID           string `json:"id,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	CloudID      string `json:"cloud_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Platform     string `json:"platform,omitempty"`
	State        string `json:"state,omitempty"`
	Type         string `json:"type,omitempty"`
	Extension    *T     `json:"extension,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Reviser      string `json:"reviser,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// GetID ...
func (image ImageExtResult[T]) GetID() string {
	return image.ID
}

// GetCloudID ...
func (image ImageExtResult[T]) GetCloudID() string {
	return image.CloudID
}

// ImageExtensionResult ...
type ImageExtensionResult interface {
	TCloudImageExtensionResult | AwsImageExtensionResult | GcpImageExtensionResult | HuaWeiImageExtensionResult | AzureImageExtensionResult
}

// ImageListResp ...
type ImageListResp struct {
	rest.BaseResp `json:",inline"`
	Data          *ImageListResult `json:"data"`
}

// ImageResult 查询公共镜像列表时的单个公共镜像数据
type ImageResult struct {
	ID           string `json:"id,omitempty"`
	Vendor       string `json:"vendor,omitempty"`
	CloudID      string `json:"cloud_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Architecture string `json:"architecture,omitempty"`
	Platform     string `json:"platform,omitempty"`
	State        string `json:"state,omitempty"`
	Type         string `json:"type,omitempty"`
	Creator      string `json:"creator,omitempty"`
	Reviser      string `json:"reviser,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// ImageListResult ...
type ImageListResult struct {
	Count   *uint64        `json:"count,omitempty"`
	Details []*ImageResult `json:"details"`
}

// ImageExtRetrieveResp 返回单个公共镜像详情
type ImageExtRetrieveResp[T ImageExtensionResult] struct {
	rest.BaseResp `json:",inline"`
	Data          *ImageExtResult[T] `json:"data"`
}
