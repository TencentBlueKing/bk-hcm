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

// Package zone ...
package zone

import (
	"hcm/pkg/criteria/enumor"
)

// BaseZone define base zone.
type BaseZone struct {
	ID        string        `json:"id"`
	Vendor    enumor.Vendor `json:"vendor"`
	CloudID   string        `json:"cloud_id"`
	Name      string        `json:"name"`
	NameCn    string        `json:"name_cn"`
	Region    string        `json:"region"`
	State     string        `json:"state"`
	Creator   string        `json:"creator"`
	Reviser   string        `json:"reviser"`
	CreatedAt string        `json:"created_at"`
	UpdatedAt string        `json:"updated_at"`
}

// GetID ...
func (zone BaseZone) GetID() string {
	return zone.ID
}

// GetCloudID ...
func (zone BaseZone) GetCloudID() string {
	return zone.CloudID
}

// Zone define zone
type Zone[Extension ZoneExtension] struct {
	BaseZone  `json:",inline"`
	Extension *Extension `json:"extension"`
}

// ZoneExtension define zone extension.
type ZoneExtension interface {
	TCloudZoneExtension | AwsZoneExtension | HuaWeiZoneExtension | GcpZoneExtension
}

// TCloudZoneExtension define tcloud zone extension.
type TCloudZoneExtension struct {
}

// HuaWeiZoneExtension define huawei zone extension.
type HuaWeiZoneExtension struct {
	Port string `json:"port"`
}

// GcpZoneExtension define gcp zone extension.
type GcpZoneExtension struct {
	SelfLink string `json:"self_link"`
}

// AwsZoneExtension define aws zone extension.
type AwsZoneExtension struct {
}
