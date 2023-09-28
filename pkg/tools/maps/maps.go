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

// Package maps map操作工具类
package maps

import "golang.org/x/exp/maps"

// Clone a new map, with same key value
func Clone[K comparable, V any](m1 map[K]V) map[K]V {
	return maps.Clone(m1)
}

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

// FilterByValue 通过给定的filter函数过滤出符合条件的子map
func FilterByValue[K comparable, V any](m map[K]V, filter func(V) bool) map[K]V {
	subMap := map[K]V{}
	for k, v := range m {
		if filter(v) {
			subMap[k] = v
		}
	}
	return subMap
}
