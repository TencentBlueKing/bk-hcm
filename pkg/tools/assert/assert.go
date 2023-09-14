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

// Package assert ...
package assert

import (
	"encoding/json"
	"reflect"
	"strings"

	"hcm/pkg/tools/converter"
)

// IsNumeric test if an interface is a numeric value or not.
func IsNumeric(val interface{}) bool {
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, json.Number:
		return true
	}
	return false
}

// IsBasicValue test if an interface is the basic supported
// golang type or not.
func IsBasicValue(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}

// IsString test if an interface is the string type.
func IsString(value interface{}) bool {
	v := reflect.ValueOf(value)

	switch v.Kind() {
	case reflect.String:
		return true
	default:
		return false
	}
}

// IsSameCaseNoSpaceString 判断字符串是否没有空格且大小写敏感
func IsSameCaseNoSpaceString(a string) bool {
	return a == converter.StrToLowerNoSpaceStr(a)
}

// IsSameCaseString 判断字符串是否大小写敏感
func IsSameCaseString(a string) bool {
	return a == strings.ToLower(a)
}

// IsSameCasePtrStringSlice 判断指针数组中的元素是否大小写敏感
func IsSameCasePtrStringSlice(a []*string) bool {
	if len(a) == 0 {
		return true
	}

	tmp := converter.PtrToSlice(a)
	for _, one := range tmp {
		if !IsSameCaseString(one) {
			return false
		}
	}

	return true
}

// IsPtrStringEqual 判断字符串指针是否相同
func IsPtrStringEqual(a, b *string) bool {
	if (a != nil && b != nil) && *a != *b {
		return false
	}

	if (a != nil && b != nil) && *a == *b {
		return true
	}

	if a != nil || b != nil {
		return false
	}

	return true
}

// IsPtrBoolEqual ...
func IsPtrBoolEqual(a, b *bool) bool {
	if (a != nil && b != nil) && *a != *b {
		return false
	}

	if (a != nil && b != nil) && *a == *b {
		return true
	}

	if a != nil || b != nil {
		return false
	}

	return true
}

// IsPtrInt64Equal ...
func IsPtrInt64Equal(a, b *int64) bool {
	if (a != nil && b != nil) && *a != *b {
		return false
	}

	if (a != nil && b != nil) && *a == *b {
		return true
	}

	if a != nil || b != nil {
		return false
	}

	return true
}

// IsPtrUint64Equal ...
func IsPtrUint64Equal(a, b *uint64) bool {
	if (a != nil && b != nil) && *a != *b {
		return false
	}

	if (a != nil && b != nil) && *a == *b {
		return true
	}

	if a != nil || b != nil {
		return false
	}

	return true
}

// IsPtrFloat64Equal ...
func IsPtrFloat64Equal(a, b *float64) bool {
	if (a != nil && b != nil) && *a != *b {
		return false
	}

	if (a != nil && b != nil) && *a == *b {
		return true
	}

	if a != nil || b != nil {
		return false
	}

	return true
}

// IsPtrInt32Equal ...
func IsPtrInt32Equal(a, b *int32) bool {
	if (a != nil && b != nil) && *a != *b {
		return false
	}

	if (a != nil && b != nil) && *a == *b {
		return true
	}

	if a != nil || b != nil {
		return false
	}

	return true
}

// IsPtrStringSliceEqual 判断指针数组是否相等
func IsPtrStringSliceEqual(a []*string, b []*string) bool {
	if len(a) == 0 && len(b) != 0 {
		return false
	}

	if len(a) != 0 && len(b) == 0 {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	tmp := converter.StringSliceToMap(converter.PtrToSlice(a))
	for _, one := range b {
		delete(tmp, *one)
	}

	if len(tmp) != 0 {
		return false
	}

	return true
}

// IsStringSliceEqual 判断字符串数组是否相等
func IsStringSliceEqual(a []string, b []string) bool {
	if len(a) == 0 && len(b) != 0 {
		return false
	}

	if len(a) != 0 && len(b) == 0 {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	tmp := converter.StringSliceToMap(a)
	for _, one := range b {
		delete(tmp, one)
	}

	if len(tmp) != 0 {
		return false
	}

	return true
}

// IsStringMapEqual 判断字符字典是否相等
func IsStringMapEqual(a map[string]string, b map[string]string) bool {
	if len(a) == 0 && len(b) != 0 {
		return false
	}

	if len(a) != 0 && len(b) == 0 {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	for k, v := range a {
		if _, ok := b[k]; !ok {
			return false
		} else {
			if v != b[k] {
				return false
			}
		}
	}

	return true
}
