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

package core

// PageWithoutSort page without sort.
type PageWithoutSort struct {
	// Count describe if this query only return the total request
	// count of the resources.
	// If true, then the request will only return the total count
	// without the resource's detail infos. and start, limit must
	// be 0.
	Count bool `json:"count"`
	// Start is the start position of the queried resource's page.
	// Note:
	// 1. Start only works when the Count = false.
	// 2. Start's minimum value is 0, not 1.
	// 3. if PageOption.EnableUnlimitedLimit = true, then Start = 0
	//   and Limit = 0 means query all the resources at once.
	Start uint32 `json:"start"`
	// Limit is the total returned resources at once query.
	// Limit only works when the Count = false.
	Limit uint `json:"limit"`
}
