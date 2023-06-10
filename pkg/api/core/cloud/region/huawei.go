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

package region

// HuaWeiRegion define huawei region.
type HuaWeiRegion struct {
	ID          string `json:"id"`
	Service     string `json:"service"`
	RegionID    string `json:"region_id"`
	Type        string `json:"type"`
	LocalesPtBr string `json:"locales_pt_br"`
	LocalesZhCn string `json:"locales_zh_cn"`
	LocalesEnUs string `json:"locales_en_us"`
	LocalesEsUs string `json:"locales_es_us"`
	LocalesEsEs string `json:"locales_es_es"`
	Creator     string `json:"creator"`
	Reviser     string `json:"reviser"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// GetID ...
func (region HuaWeiRegion) GetID() string {
	return region.ID
}

// GetCloudID ...
func (region HuaWeiRegion) GetCloudID() string {
	return region.RegionID + "|" + region.Service
}
