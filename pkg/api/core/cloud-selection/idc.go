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

import (
	"hcm/pkg/api/core"
	"hcm/pkg/criteria/enumor"
)

// Idc idc表
type Idc struct {
	ID            string        `json:"id"`
	BkBizID       int64         `json:"bk_biz_id"`
	Name          string        `json:"name"`
	Vendor        enumor.Vendor `json:"vendor"`
	Country       string        `json:"country"`
	Region        string        `json:"region"`
	core.Revision `json:",inline"`
}

// IdcWithPrice with unit price
type IdcWithPrice struct {
	Idc   `json:",inline"`
	Price float64 `json:"price"`
}
