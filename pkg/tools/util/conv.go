/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package util ...
package util

import (
	"fmt"
	"reflect"
)

// GetStrByInterface interface to string
func GetStrByInterface(a interface{}) string {
	if nil == a {
		return ""
	}

	typeOf := reflect.TypeOf(a)
	for typeOf.Kind() == reflect.Ptr {
		typeOf = typeOf.Elem()
	}

	switch typeOf.Kind() {
	case reflect.String:
		return a.(string)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return fmt.Sprintf("%d", a)

	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%f", a)
	case reflect.Slice, reflect.Array:
		return fmt.Sprintf("%v", a)
	}

	return ""
}
