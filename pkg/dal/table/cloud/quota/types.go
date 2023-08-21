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

package tablequota

import (
	"database/sql/driver"

	"hcm/pkg/criteria/enumor"
	"hcm/pkg/dal/table"
)

// Level define levels.
type Level struct {
	Name  enumor.BizQuotaLevelName `json:"name"`
	Value string                   `json:"value"`
}

// Levels define levels.
type Levels []Level

// Scan is used to decode raw message which is read from db into Levels.
func (set *Levels) Scan(raw interface{}) error {
	return table.SqlScan(set, raw)
}

// Value encode the Levels to a json raw, so that it can be stored to db with json raw.
func (set Levels) Value() (driver.Value, error) {
	return table.SqlValue(set)
}

// Dimension define dimensions.
type Dimension struct {
	Type       enumor.DimensionType `json:"type"`
	TotalQuota int                  `json:"total_quota"`
	UsedQuota  int                  `json:"used_quota"`
}

// Dimensions define Dimensions.
type Dimensions []Dimension

// Scan is used to decode raw message which is read from db into Levels.
func (set *Dimensions) Scan(raw interface{}) error {
	return table.SqlScan(set, raw)
}

// Value encode the Levels to a json raw, so that it can be stored to db with json raw.
func (set Dimensions) Value() (driver.Value, error) {
	return table.SqlValue(set)
}
