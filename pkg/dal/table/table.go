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
	"strings"

	"hcm/pkg/dal/dao/types"
	"hcm/pkg/runtime/filter"
	"hcm/pkg/tools/reflectx"
	"hcm/pkg/tools/slice"
)

const UpdateTimeField = "updated_at"

type JsonField string

type Table interface {
	TableName() string
	GenerateInsertSQL() string
	GenerateUpdateSQL(expr *filter.Expression) (string, error)
	GenerateUpdateFieldKV() map[string]interface{}
	GenerateListSQL(opt *types.ListOption) (string, error)
	GenerateDeleteSQL(expr *filter.Expression) (string, error)
}

// TableManager 作为 XXTable 的字段, 可以用来描述"插入"或者"更新"时的一些设置
type TableManager struct {
	// InsertFields 存放需要插入的 column name. 不指定表示所有字段
	InsertFields []string
	// UpdateFields 存放需要更新的 column name. 不指定表示不更新任何有效字段, 仅更新 updated_at 字段
	UpdateFields []string
}

// generateInsertSQL 生成 insert sql
func (tm *TableManager) generateInsertSQL(t Table) string {
	insertFields := tm.listInsertFields(t)
	// 插入操作, 排除自增的主键 id
	insertFields = slice.Remove(insertFields, "id")

	var fieldsWithColon []string
	for _, field := range insertFields {
		fieldsWithColon = append(fieldsWithColon, ":"+field)
	}

	return fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		t.TableName(),
		strings.Join(insertFields, ", "),
		strings.Join(fieldsWithColon, ", "),
	)
}

// GenerateInsertSQL 生成 insert sql
func (tm *TableManager) GenerateInsertSQL(t Table) string {
	insertFields := tm.listInsertFields(t)
	// 插入操作, 排除自增的主键 id
	insertFields = slice.Remove(insertFields, "id")

	var fieldsWithColon []string
	for _, field := range insertFields {
		fieldsWithColon = append(fieldsWithColon, ":"+field)
	}

	return fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES (%s)`,
		t.TableName(),
		strings.Join(insertFields, ", "),
		strings.Join(fieldsWithColon, ", "),
	)
}

// GenerateUpdateSQL 生成 update sql
func (tm *TableManager) GenerateUpdateSQL(t Table, expr *filter.Expression) (string, error) {
	whereExpr, err := GenerateWhereExpr(expr)
	if err != nil {
		return "", err
	}

	var setFields []string
	for field := range tm.GenerateUpdateFieldKV(t) {
		setFields = append(setFields, fmt.Sprintf("%s = :%s", field, field))
	}

	if slice.StringInSlice(UpdateTimeField, tm.listModelFields(t)) {
		setFields = append(setFields, fmt.Sprintf("%s = now()", UpdateTimeField))
	}

	sql := fmt.Sprintf(`UPDATE %s %s %s`, t.TableName(), "set "+strings.Join(setFields, ", "), whereExpr)
	return sql, nil
}

func (tm *TableManager) GenerateUpdateFieldKV(t Table) map[string]interface{} {
	kv := make(map[string]interface{})
	modelFields := tm.listModelFields(t)

	mValue := reflectx.ReflectValue(t)

	// 获取 {db tag: struct field} 键值对
	fieldNames := make(map[string]string)
	mType := mValue.Type()
	for j := 0; j < mType.NumField(); j++ {
		if dbField := mType.Field(j).Tag.Get("db"); dbField != "" {
			fieldNames[dbField] = mType.Field(j).Name
		}
	}

	// 移除可能的 update_at 字段, 该字段在更新时单独set update_at = now() 处理
	updateFields := slice.Remove(tm.UpdateFields, UpdateTimeField)

	for _, field := range updateFields {
		if !slice.StringInSlice(field, modelFields) {
			panic(fmt.Sprintf("field %s not in %s db tag", field, mValue.Type().Name()))
		}
		kv[field] = mValue.FieldByName(fieldNames[field]).Interface()
	}

	return kv
}

func (tm *TableManager) GenerateListSQL(t Table, opt *types.ListOption) (string, error) {
	whereExpr, err := GenerateWhereExpr(opt.FilterExpr)
	if err != nil {
		return "", err
	}

	var pageExpr string
	if opt.Page != nil {
		pageExpr, err = opt.Page.SQLExpr(&types.PageSQLOption{Sort: types.SortOption{Sort: "id", IfNotPresent: true}})
		if err != nil {
			return "", err
		}
	}

	sql := fmt.Sprintf(`SELECT %s FROM %s %s %s`, strings.Join(opt.Fields, ", "),
		t.TableName(), whereExpr, pageExpr)
	return sql, nil
}

func (tm *TableManager) GenerateDeleteSQL(t Table, expr *filter.Expression) (string, error) {
	whereExpr, err := GenerateWhereExpr(expr)
	if err != nil {
		return "", err
	}
	sql := fmt.Sprintf(`DELETE FROM %s %s`, t.TableName(), whereExpr)
	return sql, nil
}

// listInsertFields 生成 insert sql 中的 [column1, column2, column3, ...]
func (tm *TableManager) listInsertFields(t Table) []string {
	if len(tm.InsertFields) == 0 {
		return tm.listModelFields(t)
	}

	var insertFields []string
	modelFields := tm.listModelFields(t)
	// TODO 性能优化
	for _, field := range tm.InsertFields {
		if !slice.StringInSlice(field, modelFields) {
			fmt.Println(fmt.Sprintf("field %s not in %s db tag", field, reflectx.ReflectValue(t).Type().Name()))
			panic(fmt.Sprintf("field %s not in %s db tag", field, reflectx.ReflectValue(t).Type().Name()))
		}
		insertFields = append(insertFields, field)
	}

	return insertFields
}

// listModelFields 列举 Model 中带 db tag 的 fields
func (tm *TableManager) listModelFields(t Table) []string {
	return ListTableFields(t)
}

// GenerateWhereExpr ...
func GenerateWhereExpr(expr *filter.Expression) (whereExpr string, err error) {
	sqlOpt := &filter.SQLWhereOption{
		Priority: filter.Priority{"id"},
	}
	whereExpr, err = expr.SQLWhereExpr(sqlOpt)
	return
}

// ListTableFields ...
func ListTableFields(i interface{}) []string {
	var fields []string

	mType := reflectx.ReflectValue(i).Type()
	for j := 0; j < mType.NumField(); j++ {
		if dbField := mType.Field(j).Tag.Get("db"); dbField != "" {
			fields = append(fields, dbField)
		}
	}
	return fields
}
