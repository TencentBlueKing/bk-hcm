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

package export

import (
	"reflect"

	"github.com/TencentBlueKing/gopkg/conv"
)

// Table excel表结构抽象
type Table interface {
	// GetHeaderValues 根据表头字段顺序解析数据
	GetHeaderValues() ([]string, error)
	// GetHeaders 获取表头列
	GetHeaders() ([]string, error)
}

func parseHeaderFields(obj interface{}) ([]string, error) {
	rt := reflect.TypeOf(obj)
	rv := reflect.ValueOf(obj)

	var headers []string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.Tag.Get("header") != "" && !field.Anonymous {
			value := rv.Field(i)
			headers = append(headers, conv.ToString(value.Interface()))
		}
	}
	return headers, nil
}

func parseHeader(obj interface{}) ([]string, error) {
	rt := reflect.TypeOf(obj)

	var headers []string
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		tag := field.Tag.Get("header")
		if tag != "" && !field.Anonymous {
			headers = append(headers, tag)
		}
	}
	return headers, nil
}
