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

package recommend

import "hcm/pkg/cc"

// AlgorithmInput ...
type AlgorithmInput struct {
	CountryRate     map[string]float64            `json:"COUNTRY_RATE_ORIGIN"`
	CoverRate       float64                       `json:"COVER_RATE"`
	CoverPing       int                           `json:"COVER_PING"`
	PingInfo        map[string]map[string]float64 `json:"PING_INFO"`
	IdcPrice        map[string]float64            `json:"IDC_PRICE"`
	IdcList         []string                      `json:"IDC_LIST"`
	CoverPingRanges []cc.ThreshHoldRanges         `json:"COVER_PING_RANGES"`
	IDCPriceRanges  []cc.ThreshHoldRanges         `json:"IDC_PRICE_RANGES"`
	BanIdcList      []string                      `json:"BAN_IDC_LIST"`
	PickIdcList     []string                      `json:"PICK_IDC_LIST"`
}

// Solution 算法解
type Solution struct {
	Idc       []string `json:"IDC"`
	F1Score   float64  `json:"F1_SCORE"`
	F2Score   float64  `json:"F2_SCORE"`
	CoverRate float64  `json:"COVER_RATE"`
}

// AlgorithmOutput ...
type AlgorithmOutput struct {
	ParetoList []Solution `json:"PARETO_LIST"`
}
