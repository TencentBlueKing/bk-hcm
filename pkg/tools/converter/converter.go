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
	"strings"

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

// StringSliceToUint64Slice []string to []uint64.
func StringSliceToUint64Slice(source []string) []uint64 {
	target := make([]uint64, len(source))
	for index, one := range source {
		uint64Tmp, err := strconv.ParseUint(one, 10, 64)
		if err == nil {
			target[index] = uint64Tmp
		}
	}

	return target
}

// SliceToMap convert slice to map, use kvFunc to get key value pair for map.
//
//	k, v := kvFunc(one)
//	target[k] = v
func SliceToMap[T any, K comparable, V any](source []T, kvFunc func(T) (K, V)) map[K]V {
	target := make(map[K]V, len(source))
	for _, one := range source {
		k, v := kvFunc(one)
		target[k] = v
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

// StringSliceToMapBool []string to map[string]bool.
func StringSliceToMapBool(source []string) map[string]bool {
	target := make(map[string]bool, len(source))
	for _, one := range source {
		target[one] = true
	}

	return target
}

// MapKeyToStringSlice map[string]struct{} to []string.
func MapKeyToStringSlice[V any](source map[string]V) []string {
	return MapKeyToSlice(source)
}

// MapKeyToSlice map[Key]Value to []Key.
func MapKeyToSlice[K comparable, V any](source map[K]V) []K {
	target := make([]K, 0, len(source))
	for key := range source {
		target = append(target, key)
	}

	return target
}

// MapValueToSlice map[any]ValType to []ValType.
func MapValueToSlice[KeyType comparable, ValType any](source map[KeyType]ValType) []ValType {
	target := make([]ValType, 0, len(source))
	for _, val := range source {
		target = append(target, val)
	}

	return target
}

// MapToSlice 通过给定函数将map转为slice
func MapToSlice[K comparable, V any, T any](m map[K]V, mapFunc func(K, V) T) []T {
	slice := make([]T, 0, len(m))
	for k, v := range m {
		slice = append(slice, mapFunc(k, v))
	}
	return slice
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

// StrToLowerNoSpaceStr azure location need no space
func StrToLowerNoSpaceStr(str string) string {
	return strings.ToLower(strings.Replace(str, " ", "", -1))
}

// JsonStrToMap json string to map
func JsonStrToMap(jsonStr string) (map[string]string, error) {
	m := make(map[string]string)
	err := json.Unmarshal([]byte(jsonStr), &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// MapToJsonStr map to json string
func MapToJsonStr(m map[string]string) (string, error) {
	jsonByte, err := json.Marshal(m)
	if err != nil {
		return "", err
	}

	return string(jsonByte), nil
}

// StrNilPtr return pointer of string. return nil if string == ""
func StrNilPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
