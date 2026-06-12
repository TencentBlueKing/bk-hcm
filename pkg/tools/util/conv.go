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

// Package util ...
package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// GetStrByInterface interface to string
func GetStrByInterface(a interface{}) string {
	if nil == a {
		return ""
	}
	return fmt.Sprintf("%v", a)
}

// GetIntByInterface interface to int
func GetIntByInterface(a interface{}) (int, error) {
	id := 0
	var err error
	switch val := a.(type) {
	case int:
		id = val
	case int32:
		id = int(val)
	case int64:
		id = int(val)
	case json.Number:
		var tmpID int64
		tmpID, err = val.Int64()
		id = int(tmpID)
	case float64:
		id = int(val)
	case float32:
		id = int(val)
	case string:
		var tmpID int64
		tmpID, err = strconv.ParseInt(a.(string), 10, 64)
		id = int(tmpID)
	default:
		err = errors.New("not numeric")

	}
	return id, err
}

// GetInt64ByInterface interface to int64
func GetInt64ByInterface(a interface{}) (int64, error) {
	typeOf := reflect.TypeOf(a)
	valueOf := reflect.ValueOf(a)
	for typeOf.Kind() == reflect.Ptr {
		typeOf = reflect.TypeOf(a).Elem()
		valueOf = reflect.ValueOf(a).Elem()
	}

	var id int64 = 0
	var err error
	switch typeOf.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64:
		id = valueOf.Int()
	case reflect.Float64, reflect.Float32:
		id = int64(valueOf.Float())
	case reflect.String:
		id, err = strconv.ParseInt(valueOf.String(), 10, 64)
	default:
		err = fmt.Errorf("not numeric, type: %v", reflect.TypeOf(a))
	}
	return id, err
}

// GetFloat64ByInterface interface to float64
func GetFloat64ByInterface(a interface{}) (float64, error) {
	switch i := a.(type) {
	case int:
		return float64(i), nil
	case int8:
		return float64(i), nil
	case int16:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case uint:
		return float64(i), nil
	case uint8:
		return float64(i), nil
	case uint16:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case string:
		return strconv.ParseFloat(i, 64)
	case json.Number:
		return i.Float64()
	default:
		return 0, errors.New("not numeric")
	}
}

// GetMapInterfaceByInerface interface to map
func GetMapInterfaceByInerface(data interface{}) ([]interface{}, error) {
	values := make([]interface{}, 0)
	switch data.(type) {
	case []int:
		vs, _ := data.([]int)
		for _, v := range vs {
			values = append(values, v)
		}
	case []int32:
		vs, _ := data.([]int32)
		for _, v := range vs {
			values = append(values, v)
		}
	case []int64:
		vs, _ := data.([]int64)
		for _, v := range vs {
			values = append(values, v)
		}
	case []string:
		vs, _ := data.([]string)
		for _, v := range vs {
			values = append(values, v)
		}
	case []interface{}:
		values = data.([]interface{})
	default:
		return nil, errors.New("params value can not be empty")
	}

	return values, nil
}

// GetStrSliceByInterface interface to []string
func GetStrSliceByInterface(data interface{}) ([]string, error) {
	values := make([]string, 0)

	if data == nil {
		return nil, fmt.Errorf("data is nil")
	}

	typeOf := reflect.TypeOf(data)
	valueOf := reflect.ValueOf(data)
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
		valueOf = valueOf.Elem()
	}

	switch typeOf.Kind() {
	case reflect.Slice, reflect.Array:
		if !valueOf.IsValid() || valueOf.IsZero() {
			return values, nil
		}

		eleType := typeOf.Elem()
		switch eleType.Kind() {
		case reflect.String:
			for i := 0; i < valueOf.Len(); i++ {
				values = append(values, valueOf.Index(i).String())
			}
		default:
			return nil, errors.New("not []string, type is: " + typeOf.Kind().String())
		}
	default:
		return nil, errors.New("not []string, type is: " + typeOf.Kind().String())
	}

	return values, nil
}

// SliceStrToInt 将字符串切片转换为整型切片
func SliceStrToInt(sliceStr []string) ([]int, error) {
	sliceInt := make([]int, 0)
	for _, idStr := range sliceStr {

		if idStr == "" {
			continue
		}

		id, err := strconv.Atoi(idStr)
		if err != nil {
			return []int{}, err
		}
		sliceInt = append(sliceInt, id)
	}
	return sliceInt, nil
}

// SliceStrToInt64 将字符串切片转换为整型切片
func SliceStrToInt64(sliceStr []string) ([]int64, error) {
	sliceInt := make([]int64, 0)
	for _, idStr := range sliceStr {

		if idStr == "" {
			continue
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return []int64{}, err
		}
		sliceInt = append(sliceInt, id)
	}
	return sliceInt, nil
}

// GetStrValsFromArrMapInterfaceByKey get []string from []map[string]interface{}, Do not consider errors
func GetStrValsFromArrMapInterfaceByKey(arrI []interface{}, key string) []string {
	ret := make([]string, 0)
	for _, row := range arrI {
		mapRow, ok := row.(map[string]interface{})
		if ok {
			val, ok := mapRow[key].(string)
			if ok {
				ret = append(ret, val)
			}
		}
	}

	return ret
}

// ConverToInterfaceSlice convert interface to []interface{}
func ConverToInterfaceSlice(value interface{}) []interface{} {
	rflVal := reflect.ValueOf(value)
	for rflVal.CanAddr() {
		rflVal = rflVal.Elem()
	}
	if rflVal.Kind() != reflect.Slice {
		return []interface{}{value}
	}

	result := make([]interface{}, 0)
	for i := 0; i < rflVal.Len(); i++ {
		if rflVal.Index(i).CanInterface() {
			result = append(result, rflVal.Index(i).Interface())
		}
	}

	return result
}

// SplitStrField    split string field, remove empty string
func SplitStrField(str, sep string) []string {
	if "" == str {
		return nil
	}
	return strings.Split(str, sep)
}

// SliceInterfaceToInt64 将interface切片转化为int64切片,且interface的真实类型可以是任何整数类型.
// 失败则返回nil,error.
func SliceInterfaceToInt64(faceSlice []interface{}) ([]int64, error) {
	// 预分配空间.
	var results = make([]int64, len(faceSlice))

	// 转化操作.
	for i, item := range faceSlice {
		switch val := item.(type) {
		case float64:
			results[i] = int64(val)
		case float32:
			results[i] = int64(val)
		case json.Number:
			v, err := val.Int64()
			if err != nil {
				return nil, err
			}
			results[i] = v
		case int64:
			results[i] = val
		case int:
			results[i] = int64(val)
		case int8:
			results[i] = int64(val)
		case int16:
			results[i] = int64(val)
		case int32:
			results[i] = int64(val)
		case uint:
			results[i] = int64(val)
		case uint8:
			results[i] = int64(val)
		case uint16:
			results[i] = int64(val)
		case uint32:
			results[i] = int64(val)
		case uint64:
			results[i] = int64(val)
		default:
			return nil, errors.New("can't convert to int64")
		}
	}
	return results, nil
}

// SliceInterfaceToString 将interface切片转化为string切片,且interface的真实类型必须是string.
// 失败则返回nil,error.
func SliceInterfaceToString(faceSlice []interface{}) ([]string, error) {
	// 预分配空间.
	var results = make([]string, len(faceSlice))

	// 转化操作.
	for i, item := range faceSlice {
		var ok bool

		// 如果转化失败则返回错误.
		if results[i], ok = item.(string); !ok {
			return nil, errors.New("can't convert to string")
		}

	}
	return results, nil
}

// SliceInterfaceToBool 将interface切片转化为bool切片,且interface的真实类型必须是bool.
// 失败则返回nil,error.
func SliceInterfaceToBool(faceSlice []interface{}) ([]bool, error) {
	// 预分配空间.
	var results = make([]bool, len(faceSlice))

	// 转化操作.
	for i, item := range faceSlice {
		var ok bool

		// 如果转化失败则返回错误.
		if results[i], ok = item.(bool); !ok {
			return nil, errors.New("can't convert to bool")
		}

	}
	return results, nil
}

// GetInt32ByInterface get int32 by interface
func GetInt32ByInterface(a interface{}) (int32, error) {
	id := int32(0)
	var err error
	switch val := a.(type) {
	case int:
		id = int32(val)
	case int32:
		id = val
	case int64:
		id = int32(val)
	case json.Number:
		var tmpID int64
		tmpID, err = val.Int64()
		id = int32(tmpID)
	case float64:
		id = int32(val)
	case float32:
		id = int32(val)
	default:
		err = errors.New("not numeric")

	}
	return id, err
}
