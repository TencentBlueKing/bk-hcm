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

package validator

import (
	"reflect"

	gvalidator "github.com/go-playground/validator/v10"
	"hcm/pkg/tools"
)

var Validate = gvalidator.New()

// ExtractValidFields
func ExtractValidFields(i interface{}) []string {
	v := tools.ReflectValue(i)
	t := v.Type()
	var fields []string

	for j := 0; j < t.NumField(); j++ {
		tags := v.Type().Field(j).Tag
		if jsonTag := tags.Get("json"); jsonTag != "" && jsonTag != "filter_expr" {
			name := v.Type().Field(j).Name
			if isValidField(v.FieldByName(name).Interface()) {
				fields = append(fields, jsonTag)
			}
		}
	}
	return fields
}

// isValidField ...
// int 和 string 等基础类型, 通过指针方式可以区分是否传递和做零值判断
func isValidField(i interface{}) bool {
	return !reflect.ValueOf(i).IsZero()
}
