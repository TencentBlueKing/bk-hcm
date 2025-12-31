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

package meta

import (
	wz "hcm/pkg/dal/table/resource-plan/woa-zone"
)

// WoaZoneListResult is list woa zone result.
type WoaZoneListResult struct {
	Count   uint64            `json:"count"`
	Details []wz.WoaZoneTable `json:"details"`
}

// RegionArea is region and area struct.
type RegionArea struct {
	RegionID   string `db:"region_id" json:"region_id"`
	RegionName string `db:"region_name" json:"region_name"`
	AreaName   string `db:"area_name" json:"area_name"`
}

// ZoneElem is zone id and name element.
type ZoneElem struct {
	ZoneID   string `db:"zone_id" json:"zone_id"`
	ZoneName string `db:"zone_name" json:"zone_name"`
}

// RegionElem is region id and name element.
type RegionElem struct {
	RegionID   string `db:"region_id" json:"region_id"`
	RegionName string `db:"region_name" json:"region_name"`
}
