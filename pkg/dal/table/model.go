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

package table

import (
	"fmt"
	"reflect"
	"strings"

	"hcm/pkg/tools/slice"
)

type JsonField string

type Model interface {
	TableName() string
	GenerateInsertSQL() string
}

// ModelManager 作为 XXModel 的字段, 可以用来描述"插入"或者"更新"时的一些设置
type ModelManager struct {
	// InsertFields 存放需要插入的 column name. 不指定表示所有字段
	InsertFields []string
	// UpdateFields 存放需要更新的 column name. 不指定表示所有可更新字段
	UpdateFields []string
}

// GenerateInsertSQL 生成插入 sql 语句
func (manager *ModelManager) GenerateInsertSQL(m Model) string {
	insertFields := manager.listInsertFields(m)
	// 插入操作, 排除自增的主键 id
	insertFields = slice.Remove(insertFields, "id")

	var fieldsWithColon []string
	for _, field := range insertFields {
		fieldsWithColon = append(fieldsWithColon, ":"+field)
	}

	return fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		m.TableName(),
		strings.Join(insertFields, ", "),
		strings.Join(fieldsWithColon, ", "),
	)
}

// listInsertFields 生成 insert sql 中的 [column1, column2, column3, ...]
func (manager *ModelManager) listInsertFields(m Model) []string {
	if len(manager.InsertFields) == 0 {
		return listModelFields(m)
	}

	var insertFields []string
	fields := listModelFields(m)
	// TODO 性能优化
	for _, field := range fields {
		if slice.StringInSlice(field, manager.InsertFields) {
			insertFields = append(insertFields, field)
		}
	}

	return insertFields
}

// listModelFields 列举 Model 中带 db tag 的 fields
func listModelFields(m Model) []string {
	value := reflect.ValueOf(m)

	var i any
	if value.Kind() == reflect.Ptr {
		i = value.Elem().Interface()
	} else {
		i = reflect.ValueOf(&m).Elem().Interface()
	}

	var fields []string
	v := reflect.Indirect(reflect.ValueOf(i))
	s := v.Type()
	for j := 0; j < s.NumField(); j++ {
		if tag := s.Field(j).Tag.Get("db"); tag != "" {
			fields = append(fields, tag)
		}
	}
	return fields
}
