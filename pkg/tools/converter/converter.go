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

package converter

import (
	"strconv"

	"hcm/pkg/tools/json"
)

// ValToPtr convert one value to pointer.
func ValToPtr[T any](val T) *T {
	return &val
}

// PtrToVal convert pointer to one value.
func PtrToVal[T any](ptr *T) T {
	var value T
	if ptr != nil {
		value = *ptr
	}
	return value
}

// SliceToPtr convert slice to pointer.
func SliceToPtr[T any](slice []T) []*T {
	ptrArr := make([]*T, len(slice))
	for idx := range slice {
		ptrArr[idx] = &slice[idx]
	}
	return ptrArr
}

// PtrToSlice convert pointer to slice.
func PtrToSlice[T any](slice []*T) []T {
	ptrArr := make([]T, len(slice))
	for idx, ptr := range slice {
		if ptr != nil {
			ptrArr[idx] = *ptr
		}
	}
	return ptrArr
}

// StructToMap convert struct to map.
func StructToMap(source interface{}) (map[string]interface{}, error) {
	marshal, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	if err = json.Unmarshal(marshal, &result); err != nil {
		return nil, err
	}

	return result, nil
}

// Uint64SliceToStringSlice []uint64 to []string.
func Uint64SliceToStringSlice(source []uint64) []string {
	target := make([]string, len(source))
	for index, one := range source {
		target[index] = strconv.FormatUint(one, 10)
	}

	return target
}

// StringSliceToMap []string to map[string]struct.
func StringSliceToMap(source []string) map[string]struct{} {
	target := make(map[string]struct{}, len(source))
	for _, one := range source {
		target[one] = struct{}{}
	}

	return target
}

// MapKeyToStringSlice map[string]struct{} to []string.
func MapKeyToStringSlice(source map[string]struct{}) []string {
	target := make([]string, 0, len(source))
	for key, _ := range source {
		target = append(target, key)
	}

	return target
}

// StringSliceToSliceStringPtr []string to "['id1','id2',...]" ptr.
func StringSliceToSliceStringPtr(source []string) *string {
	if len(source) <= 0 {
		return nil
	}

	target := "["
	for index, one := range source {
		target = target + "'" + one + "'"
		if index != len(source)-1 {
			target += ","
		}
	}
	target += "]"

	return &target
}
