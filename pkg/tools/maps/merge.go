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

package maps

// MapMerge merge map, return new map.
func MapMerge[T any](m1 map[string]T, m2 map[string]T) map[string]T {
	result := make(map[string]T, len(m1))

	for key, value := range m1 {
		result[key] = value
	}

	for key, value := range m2 {
		result[key] = value
	}

	return result
}

// MapAppend append m2 to m1.
func MapAppend[T any](m1 map[string]T, m2 map[string]T) map[string]T {
	for key, value := range m2 {
		m1[key] = value
	}

	return m1
}
