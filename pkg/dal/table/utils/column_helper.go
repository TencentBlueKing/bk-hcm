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

package utils

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"hcm/pkg/criteria/enumor"
)

// Columns defines the column's details prepared for
// ORM operation usage.
type Columns struct {
	descriptors []ColumnDescriptor
	// columns defines all the table's columns
	columns []string
	// columnType defines a table's column and it's type
	columnType map[string]enumor.ColumnType
	// columnExpr is the joined columns with comma
	columnExpr string
	// namedExpr is the joined 'named' columns with comma.
	// which column may have the 'prefix'.
	// such as App.Spec.ImportMode named columns should be:
	// "spec.name"
	namedExpr string
	// fieldsNamedExpr is the field to it's namedExpr map
	fieldsNamedExpr map[string]string
	// ColonNameExpr is the joined 'named' columns with comma.
	// which column may have the 'prefix' and each column is
	// prefix with a colon.
	// such as App.Spec.ImportMode named columns should be:
	// ":spec.name"
	colonNameExpr string
}

// Columns returns all the db columns
func (col Columns) Columns() []string {
	copied := make([]string, len(col.columns))
	for idx := range col.columns {
		copied[idx] = col.columns[idx]
	}
	return copied
}

// ColumnTypes returns each column and it's data type
func (col Columns) ColumnTypes() map[string]enumor.ColumnType {
	copied := make(map[string]enumor.ColumnType)
	for k, v := range col.columnType {
		copied[k] = v
	}

	return copied
}

// ColumnExpr returns the joined columns with comma
func (col Columns) ColumnExpr() string {
	return col.columnExpr
}

// NamedExpr returns the joined 'named' columns with comma
// like: "name as spec.name"
func (col Columns) NamedExpr() string {
	return col.namedExpr
}

// FieldsNamedExpr returns the joined 'named' columns in fields with comma
// like: "name as spec.name"
func (col Columns) FieldsNamedExpr(fields []string) string {
	if len(fields) == 0 {
		return col.namedExpr
	}

	namedExpr := make([]string, 0)
	for _, field := range fields {
		expr, exists := col.fieldsNamedExpr[field]
		if exists {
			namedExpr = append(namedExpr, expr)
		}
	}
	return strings.Join(namedExpr, ",")
}

// FieldsNamedExprWithout 返回排除掉 fields 之外的字段。
func (col Columns) FieldsNamedExprWithout(fields []string) string {
	if len(fields) == 0 {
		return col.namedExpr
	}

	withoutMap := make(map[string]struct{}, len(fields))
	for _, one := range fields {
		withoutMap[one] = struct{}{}
	}

	namedExpr := make([]string, 0)
	for field := range col.fieldsNamedExpr {
		if _, exist := withoutMap[field]; !exist {
			namedExpr = append(namedExpr, field)
		}
	}

	return strings.Join(namedExpr, ",")
}

// ColonNameExpr returns the joined 'named' columns with comma and
// prefixed with colon, like: ":spec.name"
func (col Columns) ColonNameExpr() string {
	return col.colonNameExpr
}

// WithoutColumn remove one or more columns from the 'origin' columns.
func (col Columns) WithoutColumn(column ...string) map[string]enumor.ColumnType {
	if len(column) == 0 {
		return col.ColumnTypes()
	}

	reminder := make(map[string]bool)
	for _, one := range column {
		reminder[one] = true
	}

	copied := make(map[string]enumor.ColumnType)
	for col, typ := range col.columnType {
		if reminder[col] {
			continue
		}
		copied[col] = typ
	}

	return copied
}

var InsertWithoutPrimaryID = &mergeColumnOption{
	insertWithoutColumn: []string{"id"},
}

// mergeColumnOption defines merge column option.
type mergeColumnOption struct {
	insertWithoutColumn []string
}

var timeFieldFilter = map[string]bool{
	"created_at":     true,
	"rel_created_at": true,
	"updated_at":     true,
}

// MergeColumns merge table columns together.
func MergeColumns(opt *mergeColumnOption, all ...ColumnDescriptors) *Columns {
	tc := &Columns{
		descriptors: make([]ColumnDescriptor, 0),
		columns:     make([]string, 0),
		columnType:  make(map[string]enumor.ColumnType),
		columnExpr:  "",
		namedExpr:   "",
	}
	if len(all) == 0 {
		return tc
	}

	insertWithout := make(map[string]bool, 0)
	if opt != nil && len(opt.insertWithoutColumn) != 0 {
		for _, col := range opt.insertWithoutColumn {
			insertWithout[col] = true
		}
	}

	namedExpr := make([]string, 0)
	fieldsNamedExpr := make(map[string]string, 0)
	colonExpr := make([]string, 0)
	columns := make([]string, 0)
	for _, nc := range all {
		for _, col := range nc {
			tc.descriptors = append(tc.descriptors, col)
			tc.columnType[col.Column] = col.Type
			tc.columns = append(tc.columns, col.Column)

			if _, exist := insertWithout[col.Column]; !exist {
				columns = append(columns, col.Column)

				if _, yes := timeFieldFilter[col.Column]; yes {
					colonExpr = append(colonExpr, "now()")
				} else {
					colonExpr = append(colonExpr, col.NamedC)
				}
			}

			colNamedExpr := ""
			if col.Column == col.NamedC {
				colNamedExpr = col.Column
			} else {
				colNamedExpr = fmt.Sprintf("%s as '%s'", col.Column, col.NamedC)
			}
			namedExpr = append(namedExpr, colNamedExpr)
			fieldsNamedExpr[col.Column] = colNamedExpr
		}
	}

	tc.columnExpr = strings.Join(columns, ", ")
	tc.namedExpr = strings.Join(namedExpr, ", ")
	tc.fieldsNamedExpr = fieldsNamedExpr
	tc.colonNameExpr = strings.ReplaceAll(":"+strings.Join(colonExpr, ", :"), ":now()", "now()")
	return tc
}

// ColumnDescriptor defines a table's column related information.
type ColumnDescriptor struct {
	// Column is column's name
	Column string
	// NamedC is named column's name
	NamedC string
	// Type is this column's data type.
	Type enumor.ColumnType
	_    struct{}
}

// ColumnDescriptors is a collection of ColumnDescriptor
type ColumnDescriptors []ColumnDescriptor

// MergeColumnDescriptors merge column descriptors to one map.
func MergeColumnDescriptors(prefix string, namedC ...ColumnDescriptors) ColumnDescriptors {
	if len(namedC) == 0 {
		return make([]ColumnDescriptor, 0)
	}

	merged := make([]ColumnDescriptor, 0)
	if len(prefix) == 0 {
		for _, one := range namedC {
			merged = append(merged, one...)
		}
	} else {
		for _, one := range namedC {
			for _, col := range one {
				col.NamedC = prefix + "." + col.NamedC
				merged = append(merged, col)
			}
		}
	}

	return merged
}

// RearrangeSQLDataWithOption parse a *struct into a sql expression, and
// returned with the update sql expression and the to be updated data.
//  1. the input FieldOption only works for the returned 'expr', not controls
//     the returned 'toUpdate', so the returned 'toUpdate' contains all the
//     flatted tagged 'db' field and value.
//  2. Obviously, a data field need to be updated if the field value
//     is not blank(as is not "ZERO"),
//  3. If the field is defined in the blank options deliberately, then
//     update it to blank value as required.
//  4. see the test case to know the exact data returned.
func RearrangeSQLDataWithOption(data interface{}, opts *FieldOption) (
	expr string, toUpdate map[string]interface{}, err error,
) {
	if data == nil {
		return "", nil, errors.New("parse sql expr fields, but data is nil")
	}

	if opts == nil {
		return "", nil, errors.New("parse sql expr fields, but field options is nil")
	}

	var setFields []string
	toUpdate = make(map[string]interface{})
	taggedKV, err := RecursiveGetTaggedFieldValues(data)
	if err != nil {
		return "", nil, fmt.Errorf("get recursively tagged db kv faield, err: %v", err)
	}

	for tag, value := range taggedKV {
		if opts.NeedIgnored(tag) {
			// this is a field which is need to be ignored,
			// which means do not need to be updated.
			continue
		}

		if opts.NeedBlanked(tag) {
			if isBasicValue(value) && !reflect.ValueOf(value).IsNil() {
				toUpdate[tag] = value
				setFields = append(setFields, fmt.Sprintf("%s = :%s", tag, tag))
				continue
			}

			if !isBasicValue(value) {
				toUpdate[tag] = value
				setFields = append(setFields, fmt.Sprintf("%s = :%s", tag, tag))
			}

			continue
		}

		if !isBlank(reflect.ValueOf(value)) {
			toUpdate[tag] = value
			setFields = append(setFields, fmt.Sprintf("%s = :%s", tag, tag))
		}
	}

	setFields = append(setFields, "updated_at = now()")
	expr = strings.Join(setFields, ", ")

	return "set " + expr, toUpdate, nil
}

// RecursiveGetTaggedFieldValues get all the tagged db kv
// in the struct to a flat map except ptr and struct tag.
// Note:
//  1. if the embedded tag is same, then it will be overlapped.
//  2. use this function carefully, it not supports all the type,
//     such as array, slice, map is not supported.
//  3. see the test case to know the output data example.
func RecursiveGetTaggedFieldValues(v interface{}) (map[string]interface{}, error) {
	if v == nil {
		return map[string]interface{}{}, nil
	}

	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return map[string]interface{}{}, nil
		}

		return RecursiveGetTaggedFieldValues(value.Elem().Interface())

	case reflect.Struct:
		kv := make(map[string]interface{})

		for i := 0; i < value.NumField(); i++ {
			name := value.Type().Field(i).Name
			tag := value.Type().Field(i).Tag.Get("db")
			if tag == "" {
				return nil, fmt.Errorf("field: %s do not have a 'db' tag", name)
			}

			val := value.FieldByName(name).Interface()

			if isBasicValue(val) {
				kv[tag] = val

				// handle next field.
				continue
			}

			// 如果结构体实现了Sql序列化函数（Scan），不再解析嵌套结构体中的字段，将整个结构体当成一个Sql字段
			_, exist := reflect.TypeOf(val).MethodByName("Scan")
			if exist {
				kv[tag] = val

				// handle next field.
				continue
			}

			// this is not a basic value, then do get tags again recursively.
			mapper, err := RecursiveGetTaggedFieldValues(val)
			if err != nil {
				return nil, err
			}

			for k, v := range mapper {
				kv[k] = v
			}

		}

		return kv, nil

	default:
		return nil, fmt.Errorf("unsupported struct db tagged value type: %s", value.Kind())
	}
}

var timeType = reflect.TypeOf(time.Time{})

func isBasicValue(value interface{}) bool {
	v := reflect.ValueOf(value)
	if v.Type() == timeType {
		return true
	}

	kind := reflect.ValueOf(value).Kind()
	if kind == reflect.Ptr {
		return isBasicKind(reflect.TypeOf(value).Elem().Kind())
	}

	return isBasicKind(kind)
}

func isBasicKind(kind reflect.Kind) bool {
	switch kind {
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
		reflect.String,
		reflect.Slice:
		return true
	default:
		return false
	}
}

func isBlank(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.String:
		return value.Len() == 0
	case reflect.Bool:
		return !value.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return value.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return value.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return value.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return value.IsNil()
	}

	return reflect.DeepEqual(value.Interface(), reflect.Zero(value.Type()).Interface())
}

// FieldOption is to define which field need to be:
//  1. updated to blank(as is ZERO) value.
//  2. be ignored, which means not be updated even its value
//     is not blank(as is not ZERO).
//
// NOTE:
// 1. A field can not in the blanked and ignore fields at the
// same time. if a field does, then it will be ignored without
// being updated.
// 2. The map's key is the structs' 'db' tag of that field.
type FieldOption struct {
	blanked map[string]struct{}
	ignored map[string]struct{}
}

// NewFieldOptions create a blank option instances for add keys
// to be updated when update data.
func NewFieldOptions() *FieldOption {
	return &FieldOption{
		blanked: make(map[string]struct{}),
		ignored: make(map[string]struct{}),
	}
}

// NeedBlanked check if this field need to be updated with blank
func (f *FieldOption) NeedBlanked(field string) bool {
	_, ok := f.blanked[field]
	return ok
}

// NeedIgnored check if this field does not need to be updated.
func (f *FieldOption) NeedIgnored(field string) bool {
	_, ok := f.ignored[field]
	return ok
}

// AddBlankedFields add fields to be updated to blank values.
func (f *FieldOption) AddBlankedFields(fields ...string) *FieldOption {
	for _, one := range fields {
		f.blanked[one] = struct{}{}
	}

	return f
}

// AddIgnoredFields add fields which do not need to be updated even it
// do has a value.
func (f *FieldOption) AddIgnoredFields(fields ...string) *FieldOption {
	for _, one := range fields {
		f.ignored[one] = struct{}{}
	}

	return f
}
