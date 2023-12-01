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

package coreselection

// CountryInfo ...
type CountryInfo struct {
	Country string `json:"country"`
}

// AreaInfo ...
type AreaInfo struct {
	Name     string     `json:"name,omitempty" validate:"required"`
	Children []AreaInfo `json:"children,omitempty"`
}

// AreaValue ...
type AreaValue[V any] struct {
	Name     string         `json:"name,omitempty" validate:"required"`
	Value    V              `json:"value,omitempty"`
	Children []AreaValue[V] `json:"children,omitempty"`
}

// ProvinceToIDCLatency 省份到IDC 延迟
type ProvinceToIDCLatency struct {
	Province string  `json:"province,omitempty"`
	IDCName  string  `json:"idc_name,omitempty"`
	Latency  float64 `json:"latency,omitempty"`
}

// UserDistribution 用户分布
type UserDistribution struct {
	Country  string  `json:"country,omitempty"`
	Province string  `json:"province,omitempty"`
	Count    float64 `json:"count,omitempty"`
}

// FlatAreaInfo ...
type FlatAreaInfo struct {
	CountryName    string  `json:"country_name,omitempty"`
	ProvinceName   string  `json:"province_name,omitempty"`
	NetworkLatency float64 `json:"network_latency,omitempty"`
}

// IdcServiceAreaRel ...
type IdcServiceAreaRel struct {
	IdcID        string         `json:"idc_id,omitempty"`
	AvgLatency   float64        `json:"avg_latency"`
	ServiceAreas []FlatAreaInfo `json:"service_areas,omitempty"`
}
