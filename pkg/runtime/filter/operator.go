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

package filter

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"hcm/pkg/tools/assert"
)

var opFactory map[OpFactory]Operator

func init() {
	opFactory = make(map[OpFactory]Operator)

	opFactory[Equal.Factory()] = EqualOp(Equal)
	opFactory[NotEqual.Factory()] = NotEqualOp(NotEqual)

	opFactory[GreaterThan.Factory()] = GreaterThanOp(GreaterThan)
	opFactory[GreaterThanEqual.Factory()] = GreaterThanEqualOp(GreaterThanEqual)

	opFactory[LessThan.Factory()] = LessThanOp(LessThan)
	opFactory[LessThanEqual.Factory()] = LessThanEqualOp(LessThanEqual)

	opFactory[In.Factory()] = InOp(In)
	opFactory[NotIn.Factory()] = NotInOp(NotIn)

	opFactory[ContainsSensitive.Factory()] = ContainsSensitiveOp(ContainsSensitive)
	opFactory[ContainsInsensitive.Factory()] = ContainsInsensitiveOp(ContainsInsensitive)

	opFactory[JSONEqual.Factory()] = JSONEqualOp(JSONEqual)
	opFactory[JSONIn.Factory()] = JSONInOp(JSONIn)
}

const (
	// And logic operator
	And LogicOperator = "and"
	// Or logic operator
	Or LogicOperator = "or"
	// SqlPlaceholder is sql placeholder.
	SqlPlaceholder = ":"
)

// LogicOperator defines the logic operator
type LogicOperator string

// Validate the logic operator is valid or not.
func (lo LogicOperator) Validate() error {
	switch lo {
	case And:
	case Or:
	default:
		return fmt.Errorf("unsupported expression's logic operator: %s", lo)
	}

	return nil
}

// OpFactory defines the operator's factory type.
type OpFactory string

// Operator return this operator factory's Operator
func (of OpFactory) Operator() Operator {
	op, exist := opFactory[of]
	if !exist {
		unknown := UnknownOp(Unknown)
		return &unknown
	}

	return op
}

// Validate this operator factory is valid or not.
func (of OpFactory) Validate() error {
	typ := OpType(of)
	return typ.Validate()
}

const (
	// Unknown is an unsupported operator
	Unknown OpType = "unknown"
	// Equal operator
	Equal OpType = "eq"
	// NotEqual operator
	NotEqual OpType = "neq"
	// GreaterThan operator
	GreaterThan OpType = "gt"
	// GreaterThanEqual operator
	GreaterThanEqual OpType = "gte"
	// LessThan operator
	LessThan OpType = "lt"
	// LessThanEqual operator
	LessThanEqual OpType = "lte"
	// In operator
	In OpType = "in"
	// NotIn operator
	NotIn OpType = "nin"
	// ContainsSensitive operator match the value with
	// regular expression with case-sensitive.
	ContainsSensitive OpType = "cs"
	// ContainsInsensitive operator match the value with
	// regular expression with case-insensitive.
	ContainsInsensitive OpType = "cis"

	// JSONEqual is json field equal operator.
	JSONEqual OpType = "json_eq"
	// JSONIn is json field in operator.
	JSONIn OpType = "json_in"
)

// OpType defines the operators supported by mysql.
type OpType string

// Validate test the operator is valid or not.
func (op OpType) Validate() error {
	switch op {
	case Equal, NotEqual,
		GreaterThan, GreaterThanEqual,
		LessThan, LessThanEqual,
		In, NotIn,
		ContainsSensitive, ContainsInsensitive,
		JSONEqual, JSONIn:
	default:
		return fmt.Errorf("unsupported operator: %s", op)
	}

	return nil
}

// Factory return opType's factory type.
func (op OpType) Factory() OpFactory {
	return OpFactory(op)
}

// Operator is a collection of supported query operators.
type Operator interface {
	// Name is the operator's name
	Name() OpType
	// ValidateValue validate the operator's value is valid or not
	ValidateValue(v interface{}, opt *ExprOption) error
	// SQLExprAndValue generate an operator's SQL expression with its filed
	// and value.
	SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error)
}

// UnknownOp is unknown operator
type UnknownOp OpType

// Name is equal operator
func (uo UnknownOp) Name() OpType {
	return Unknown
}

// ValidateValue validate equal's value
func (uo UnknownOp) ValidateValue(_ interface{}, _ *ExprOption) error {
	return errors.New("unknown operator")
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (uo UnknownOp) SQLExprAndValue(_ string, _ interface{}) (string, map[string]interface{}, error) {
	return "", nil, errors.New("unknown operator, can not gen sql expression")
}

// EqualOp is equal operator type
type EqualOp OpType

// Name is equal operator
func (eo EqualOp) Name() OpType {
	return Equal
}

// ValidateValue validate equal's value
func (eo EqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if !assert.IsBasicValue(v) {
		return errors.New("invalid value field")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (eo EqualOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	if !assert.IsBasicValue(value) {
		return "", nil, errors.New("invalid value field")
	}

	return fmt.Sprintf(`%s = %s%s`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// NotEqualOp is not equal operator type
type NotEqualOp OpType

// Name is not equal operator
func (ne NotEqualOp) Name() OpType {
	return NotEqual
}

// ValidateValue validate not equal's value
func (ne NotEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if !assert.IsBasicValue(v) {
		return errors.New("invalid ne operator's value field")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (ne NotEqualOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	if !assert.IsBasicValue(value) {
		return "", nil, errors.New("invalid ne operator's value field")
	}

	return fmt.Sprintf(`%s != %s%s`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// GreaterThanOp is greater than operator
type GreaterThanOp OpType

// Name is greater than operator
func (gt GreaterThanOp) Name() OpType {
	return GreaterThan
}

// ValidateValue validate greater than value
func (gt GreaterThanOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid gt operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (gt GreaterThanOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	return fmt.Sprintf(`%s > %s%s`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// GreaterThanEqualOp is greater than equal operator
type GreaterThanEqualOp OpType

// Name is greater than operator
func (gte GreaterThanEqualOp) Name() OpType {
	return GreaterThanEqual
}

// ValidateValue validate greater than value
func (gte GreaterThanEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid gte operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (gte GreaterThanEqualOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	return fmt.Sprintf(`%s >= %s%s`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// LessThanOp is less than operator
type LessThanOp OpType

// Name is less than equal operator
func (lt LessThanOp) Name() OpType {
	return LessThan
}

// ValidateValue validate less than equal value
func (lt LessThanOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid lt operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (lt LessThanOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	return fmt.Sprintf(`%s < %s%s`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// LessThanEqualOp is less than equal operator
type LessThanEqualOp OpType

// Name is less than equal operator
func (lte LessThanEqualOp) Name() OpType {
	return LessThanEqual
}

// ValidateValue validate less than equal value
func (lte LessThanEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if _, hit := isNumericOrTime(v); !hit {
		return errors.New("invalid lte operator's value, should be a numeric or time format string value")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (lte LessThanEqualOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	return fmt.Sprintf(`%s <= %s%s`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// InOp is in operator
type InOp OpType

// Name is in operator
func (io InOp) Name() OpType {
	return In
}

// ValidateValue validate in operator's value
func (io InOp) ValidateValue(v interface{}, opt *ExprOption) error {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return errors.New("in operator's value should be an array")
	}

	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return errors.New("invalid in operator's value, at least have one element")
	}

	maxInV := DefaultMaxInLimit
	if opt != nil {
		if opt.MaxInLimit > 0 {
			maxInV = opt.MaxInLimit
		}
	}

	if length > int(maxInV) {
		return fmt.Errorf("invalid in operator's value, at most have %d elements", maxInV)
	}

	// each element in the array or slice should be a basic type.
	for i := 0; i < length; i++ {
		if !assert.IsBasicValue(value.Index(i).Interface()) {
			return fmt.Errorf("invalid in operator's value: %v, each element's value should be a basic type",
				value.Index(i).Interface())
		}
	}

	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (io InOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return "", nil, errors.New("in operator's value should be an array")
	}

	return fmt.Sprintf(`%s IN (%s%s)`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// NotInOp is not in operator
type NotInOp OpType

// Name is not in operator
func (nio NotInOp) Name() OpType {
	return NotIn
}

// ValidateValue validate not in value
func (nio NotInOp) ValidateValue(v interface{}, opt *ExprOption) error {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return errors.New("nin operator's value should be an array")
	}

	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return errors.New("invalid nin operator's value, at least have one element")
	}

	maxNotInV := DefaultMaxNotInLimit
	if opt != nil {
		if opt.MaxNotInLimit > 0 {
			maxNotInV = opt.MaxNotInLimit
		}
	}

	if length > int(maxNotInV) {
		return fmt.Errorf("invalid nin operator's value, at most have %d elements", maxNotInV)
	}

	// each element in the array or slice should be a basic type.
	for i := 0; i < length; i++ {
		if !assert.IsBasicValue(value.Index(i).Interface()) {
			return fmt.Errorf("invalid nin operator's value: %v, each element's value should be a basic type",
				value.Index(i).Interface())
		}
	}

	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (nio NotInOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return "", nil, errors.New("nin operator's value should be an array")
	}

	return fmt.Sprintf(`%s NOT IN (%s%s)`, field, SqlPlaceholder, field), map[string]interface{}{field: value}, nil
}

// ContainsSensitiveOp is contains sensitive operator
type ContainsSensitiveOp OpType

// Name is 'like' expression with camel sensitive operator
func (cso ContainsSensitiveOp) Name() OpType {
	return ContainsSensitive
}

// ValidateValue validate 'like' operator's value
func (cso ContainsSensitiveOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if reflect.TypeOf(v).Kind() != reflect.String {
		return errors.New("cs operator's value should be an string")
	}

	value, ok := v.(string)
	if !ok {
		return errors.New("cs operator's value should be an string")
	}

	if len(value) == 0 {
		return errors.New("cs operator's value can not be a empty string")
	}

	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (cso ContainsSensitiveOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{},
	error) {

	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	if reflect.TypeOf(value).Kind() != reflect.String {
		return "", nil, errors.New("cs operator's value should be an string")
	}

	s, ok := value.(string)
	if !ok {
		return "", nil, errors.New("cs operator's value should be an string")
	}

	if len(s) == 0 {
		return "", nil, errors.New("cs operator's value can not be a empty string")
	}

	return fmt.Sprintf(`%s LIKE BINARY %s%s`, field, SqlPlaceholder, field),
		map[string]interface{}{field: "%" + s + "%"}, nil
}

// ContainsInsensitiveOp is contains insensitive operator
type ContainsInsensitiveOp OpType

// Name is 'like' expression with camel insensitive operator
func (cio ContainsInsensitiveOp) Name() OpType {
	return ContainsInsensitive
}

// ValidateValue validate 'like' operator's value
func (cio ContainsInsensitiveOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if reflect.TypeOf(v).Kind() != reflect.String {
		return errors.New("cis operator's value should be an string")
	}

	value, ok := v.(string)
	if !ok {
		return errors.New("cis operator's value should be an string")
	}

	if len(value) == 0 {
		return errors.New("cis operator's value can not be a empty string")
	}

	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub
// query expression.
func (cio ContainsInsensitiveOp) SQLExprAndValue(field string, value interface{}) (string,
	map[string]interface{}, error) {

	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	if reflect.TypeOf(value).Kind() != reflect.String {
		return "", nil, errors.New("cis operator's value should be an string")
	}

	s, ok := value.(string)
	if !ok {
		return "", nil, errors.New("cis operator's value should be an string")
	}

	if len(s) == 0 {
		return "", nil, errors.New("cis operator's value can not be a empty string")
	}

	return fmt.Sprintf(`%s LIKE %s%s`, field, SqlPlaceholder, field),
		map[string]interface{}{field: "%" + s + "%"}, nil
}

// JSONEqualOp is json field equal operator
type JSONEqualOp OpType

// Name is json field equal operator
func (op JSONEqualOp) Name() OpType {
	return JSONEqual
}

// ValidateValue validate json field equal's value
func (op JSONEqualOp) ValidateValue(v interface{}, opt *ExprOption) error {
	if !assert.IsBasicValue(v) {
		return errors.New("invalid value field")
	}
	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub query expression.
func (op JSONEqualOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	if !assert.IsBasicValue(value) {
		return "", nil, errors.New("invalid value field")
	}

	jsonField, err := jsonFiledSqlFormat(field)
	if err != nil {
		return "", nil, err
	}

	jsonFieldAlias := strings.ReplaceAll(field, ".", "")

	return fmt.Sprintf(`%s = %s%s`, jsonField, SqlPlaceholder, jsonFieldAlias),
		map[string]interface{}{jsonFieldAlias: value}, nil
}

// JSONInOp is json field in operator
type JSONInOp OpType

// Name is json field in operator
func (op JSONInOp) Name() OpType {
	return JSONIn
}

// ValidateValue validate json field in's value
func (op JSONInOp) ValidateValue(v interface{}, opt *ExprOption) error {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return errors.New("json in operator's value should be an array")
	}

	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return errors.New("invalid json in operator's value, at least have one element")
	}

	maxInV := DefaultMaxInLimit
	if opt != nil {
		if opt.MaxInLimit > 0 {
			maxInV = opt.MaxInLimit
		}
	}

	if length > int(maxInV) {
		return fmt.Errorf("invalid json in operator's value, at most have %d elements", maxInV)
	}

	// each element in the array or slice should be a basic type.
	for i := 0; i < length; i++ {
		if !assert.IsBasicValue(value.Index(i).Interface()) {
			return fmt.Errorf("invalid json in operator's value: %v, each element's value should be a basic type",
				value.Index(i).Interface())
		}
	}

	return nil
}

// SQLExprAndValue convert this operator's field and value to a mysql's sub query expression.
func (op JSONInOp) SQLExprAndValue(field string, value interface{}) (string, map[string]interface{}, error) {
	if len(field) == 0 {
		return "", nil, errors.New("field is empty")
	}

	switch reflect.TypeOf(value).Kind() {
	case reflect.Array:
	case reflect.Slice:
	default:
		return "", nil, errors.New("in operator's value should be an array")
	}

	jsonField, err := jsonFiledSqlFormat(field)
	if err != nil {
		return "", nil, err
	}

	jsonFieldAlias := strings.ReplaceAll(field, ".", "")

	return fmt.Sprintf(`%s IN (%s%s)`, jsonField, SqlPlaceholder, jsonFieldAlias),
		map[string]interface{}{jsonFieldAlias: value}, nil
}

// jsonFiledSqlFormat 会将用户传入的 json 字段名由 "extension.vpc_id" 转为 `extension->>"$.vpc_id"`
func jsonFiledSqlFormat(field string) (string, error) {
	if !strings.ContainsAny(field, ".") {
		return "", fmt.Errorf("feild: %s not json field format", field)
	}

	index := strings.Index(field, ".")
	return fmt.Sprintf(`%s->>"$%s"`, field[0:index], field[index:]), nil
}
