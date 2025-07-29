/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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

package table

import (
	"reflect"
	"strings"

	"github.com/TencentBlueKing/gopkg/conv"
)

// Table excel表结构抽象
type Table interface {
	// GetHeaders 获取表头列
	GetHeaders() ([][]string, error)
	// GetValuesByHeader 根据表头字段顺序解析数据
	GetValuesByHeader() ([]string, error)
}

// GetHeaders 获取表头列
func GetHeaders(obj interface{}) ([][]string, error) {
	rt := reflect.TypeOf(obj)

	fieldTags := make([][]string, 0)
	fieldNum := rt.NumField()
	rowNum := 0

	for i := 0; i < fieldNum; i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("header")
		if tag == "" || field.Anonymous {
			continue
		}

		splitTags := strings.Split(tag, ";")
		if len(splitTags) > rowNum {
			rowNum = len(splitTags)
		}
		fieldTags = append(fieldTags, splitTags)
	}

	headers := make([][]string, rowNum)
	for i := range headers {
		headers[i] = make([]string, fieldNum)
	}
	for col, tags := range fieldTags {
		for row, tag := range tags {
			headers[row][col] = tag
		}
	}

	return headers, nil
}

// GetValuesByHeader 根据表头字段顺序解析数据
func GetValuesByHeader(obj interface{}) ([]string, error) {
	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)

	var headers []string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Tag.Get("header") != "" && !field.Anonymous {
			value := rv.Field(i)
			// 处理指针类型的字段
			if value.Kind() == reflect.Ptr {
				if value.IsNil() {
					// 如果指针为nil，则添加空字符串
					headers = append(headers, "")
					continue
				}
				// 获取指针指向的实际值
				value = value.Elem()
			}

			headers = append(headers, conv.ToString(value.Interface()))
		}
	}
	return headers, nil
}
