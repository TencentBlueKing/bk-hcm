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
	"hcm/pkg/criteria/enumor"
)

// Region define region.
type Region interface {
	TCloudRegion | AwsRegion | GcpRegion
}

// TCloudRegion define tcloud region.
type TCloudRegion struct {
	ID         string        `json:"id"`
	Vendor     enumor.Vendor `json:"vendor"`
	RegionID   string        `json:"region_id"`
	RegionName string        `json:"region_name"`
	Status     string        `json:"status"`
	Creator    string        `json:"creator"`
	Reviser    string        `json:"reviser"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// GetID ...
func (region TCloudRegion) GetID() string {
	return region.ID
}

// GetCloudID ...
func (region TCloudRegion) GetCloudID() string {
	return region.RegionID
}

// AwsRegion define aws region.
type AwsRegion struct {
	ID         string        `json:"id"`
	Vendor     enumor.Vendor `json:"vendor"`
	RegionID   string        `json:"region_id"`
	RegionName string        `json:"region_name"`
	Status     string        `json:"status"`
	Endpoint   string        `json:"endpoint"`
	Creator    string        `json:"creator"`
	Reviser    string        `json:"reviser"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// GetID ...
func (region AwsRegion) GetID() string {
	return region.ID
}

// GetCloudID ...
func (region AwsRegion) GetCloudID() string {
	return region.RegionID
}

// GcpRegion define gcp region.
type GcpRegion struct {
	ID         string        `json:"id"`
	Vendor     enumor.Vendor `json:"vendor"`
	RegionID   string        `json:"region_id"`
	RegionName string        `json:"region_name"`
	Status     string        `json:"status"`
	Creator    string        `json:"creator"`
	Reviser    string        `json:"reviser"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// GetID ...
func (region GcpRegion) GetID() string {
	return region.ID
}

// GetCloudID ...
func (region GcpRegion) GetCloudID() string {
	return region.RegionID
}
