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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/criteria/enumor"
	"hcm/pkg/criteria/errf"
	"hcm/pkg/logs"
	"hcm/pkg/tools/assert"

	"github.com/tidwall/gjson"
)

const (
	// DefaultMaxInLimit defines the default max in limit
	DefaultMaxInLimit = uint(500)
	// DefaultMaxNotInLimit defines the default max nin limit
	DefaultMaxNotInLimit = uint(500)
	// DefaultMaxRuleLimit defines the default max number of rules limit
	DefaultMaxRuleLimit = uint(10)
)

// ExprOption defines how to validate an
// expression.
type ExprOption struct {
	// RuleFields:
	// 1. used to test if all the expression rule's field
	//    is in the RuleFields' key restricts.
	// 2. all the expression's rule filed should be a sub-set
	//    of the RuleFields' key.
	RuleFields map[string]enumor.ColumnType
	// MaxInLimit defines the max element of the in operator
	// If not set, then use default value: DefaultMaxInLimit
	MaxInLimit uint
	// MaxNotInLimit defines the max element of the nin operator
	// If not set, then use default value: DefaultMaxNotInLimit
	MaxNotInLimit uint
	// MaxRulesLimit defines the max number of rules an expression allows.
	// If not set, then use default value: DefaultMaxRuleLimit
	MaxRulesLimit uint
}

// ExprOptionFunc expr option func defines.
type ExprOptionFunc func(opt *ExprOption)

// RuleFields set rule fields func.
func RuleFields(fields map[string]enumor.ColumnType) ExprOptionFunc {
	return func(opt *ExprOption) {
		opt.RuleFields = fields
	}
}

// MaxInLimit set max in limit func.
func MaxInLimit(limit uint) ExprOptionFunc {
	return func(opt *ExprOption) {
		opt.MaxInLimit = limit
	}
}

// MaxNotInLimit set max not in limit func.
func MaxNotInLimit(limit uint) ExprOptionFunc {
	return func(opt *ExprOption) {
		opt.MaxNotInLimit = limit
	}
}

// MaxRulesLimit set max rule limit func.
func MaxRulesLimit(limit uint) ExprOptionFunc {
	return func(opt *ExprOption) {
		opt.MaxRulesLimit = limit
	}
}

// NewExprOption new expr option.
// ExprOptionFunc: RuleFields、MaxInLimit、MaxNotInLimit、MaxRulesLimit
func NewExprOption(opts ...ExprOptionFunc) *ExprOption {
	exprOpt := new(ExprOption)
	for _, opt := range opts {
		opt(exprOpt)
	}

	return exprOpt
}

// Expression is to build a query expression
type Expression struct {
	Op    LogicOperator `json:"op"`
	Rules []RuleFactory `json:"rules"`
}

// Validate the expression is valid or not.
func (exp Expression) Validate(opt *ExprOption) (hitErr error) {
	defer func() {
		if hitErr != nil {
			hitErr = errf.New(errf.InvalidParameter, hitErr.Error())
		}
	}()

	if exp.IsEmpty() {
		return nil
	}

	if err := exp.Op.Validate(); err != nil {
		return err
	}

	if len(exp.Rules) == 0 {
		return nil
	}

	maxRules := DefaultMaxRuleLimit
	if opt != nil {
		if opt.MaxRulesLimit > 0 {
			maxRules = opt.MaxRulesLimit
		}
	}

	if len(exp.Rules) > int(maxRules) {
		return fmt.Errorf("rules elements number is overhead, it at most have %d rules", maxRules)
	}

	fieldsReminder := make(map[string]bool)
	exprCountReminder := 0
	for _, r := range exp.Rules {
		switch r.WithType() {
		case AtomType:
			fieldsReminder[r.RuleField()] = true
		case ExpressionType:
			exprCountReminder++
		default:
			return fmt.Errorf("unknown rule type: %s", r.WithType())
		}
	}

	if opt != nil {
		reminder := make(map[string]bool)
		for col := range opt.RuleFields {
			reminder[col] = true
		}

		// all the rule's filed should exist in the reminder.
		for one := range fieldsReminder {
			if exist := reminder[one]; !exist {
				return fmt.Errorf("expression rules filed(%s) should not exist(not supported)", one)
			}
		}
	}

	var valOpt *ExprOption
	if opt != nil {
		valOpt = opt
	}

	for _, one := range exp.Rules {
		if err := one.Validate(valOpt); err != nil {
			return err
		}
	}

	return nil
}

// IsEmpty when rules is empty or filter is null
func (exp *Expression) IsEmpty() bool {
	return exp == nil || len(exp.Rules) == 0
}

// WithType return this expression rule's tye.
func (exp Expression) WithType() RuleType {
	return ExpressionType
}

// RuleField implementing RuleFactory requires，it do not work.
func (exp Expression) RuleField() string {
	return ""
}

// SQLExprAndValue convert this expression rule to a mysql's sub query expression, and field's value.
// Implement RuleFactory, which is used to generate SQL expressions, and
// is not used externally.
func (exp *Expression) SQLExprAndValue(opt *SQLWhereOption) (string, map[string]interface{}, error) {
	if exp == nil {
		return "", nil, errors.New("expression is nil")
	}

	if opt == nil {
		return "", nil, errors.New("SQLWhereOption is nil")
	}

	if err := opt.Validate(); err != nil {
		return "", nil, err
	}

	expr, value, err := doSoloSQLWhereExpr(opt, exp.Op, exp.Rules, opt.Priority)
	if err != nil {
		return "", nil, err
	}

	return fmt.Sprintf("(%s)", expr), value, nil
}

// SQLWhereExpr convert this expression and crowned rules to the mysql's WHERE
// expression automatically.
// the generated SQL Where expression depends on various options:
//  1. the Expression itself.
//  2. the crowned rules.
//  3. the priority fields which is corresponding to the db's indexes order.
//     the position of Expression's expression and crowned rules' expression is
//     determined by the first 'field' occurred in the SQLWhereOption.Priority.
//     For example, if the first hit field in the SQLWhereOption.Priority is found
//     in the Expression's rule then the Expression's expression is ahead of the
//     crowned rule's expression in the final generated SQL WHERE expression.
//
// Note:
//  1. if the expression is NULL, then return an empty string "" as the expression
//     directly without "WHERE" keyword.
//  2. if the expression is not NULL, then return the expression prefixed with "WHERE"
//     keyword.
func (exp *Expression) SQLWhereExpr(opt *SQLWhereOption) (where string, value map[string]interface{}, err error) {
	defer func() {
		if err != nil {
			err = errf.NewFromErr(errf.InvalidParameter, err)
		}
	}()

	if opt == nil {
		return "", nil, errors.New("SQLWhereOption is nil")
	}

	if err := opt.Validate(); err != nil {
		return "", nil, err
	}

	if opt.CrownedOption == nil || (opt.CrownedOption != nil && len(opt.CrownedOption.Rules) == 0) {
		if len(exp.Rules) == 0 {
			return "", nil, nil
		}

		// no crowned option is configured, then generate SQL where expression directly.
		expr, value, err := doSoloSQLWhereExpr(opt, exp.Op, exp.Rules, opt.Priority)
		if err != nil {
			return "", nil, err
		}

		return "WHERE " + expr, value, nil
	}

	// generate SQL where expression depends on mixed logic operator.
	var expr string
	switch exp.Op {
	case And:
		switch opt.CrownedOption.CrownedOp {
		case And:
			// both expression rules and crowned rules need to do logic 'AND', so put them
			// together and generate SQL expression directly.
			mergedRules := append(exp.Rules, opt.CrownedOption.Rules...)
			expr, value, err = doSoloSQLWhereExpr(opt, And, mergedRules, opt.Priority)

		case Or:
			expr, value, err = doMixedSQLWhereExpr(opt, exp.Op, exp.Rules, opt.CrownedOption.CrownedOp,
				opt.CrownedOption.Rules, opt.Priority)

		default:
			return "", nil, fmt.Errorf("unsupported crown operator: %s", opt.CrownedOption.CrownedOp)
		}

	case Or:
		switch opt.CrownedOption.CrownedOp {
		case And:
			expr, value, err = doMixedSQLWhereExpr(opt, exp.Op, exp.Rules, opt.CrownedOption.CrownedOp,
				opt.CrownedOption.Rules, opt.Priority)

		case Or:
			// although both expression's op and crowned op is OR, but rules in the crowned rules is still
			// use AND operator.
			expr, value, err = doMixedSQLWhereExpr(opt, exp.Op, exp.Rules, opt.CrownedOption.CrownedOp,
				opt.CrownedOption.Rules, opt.Priority)

		default:
			return "", nil, fmt.Errorf("unsupported crown operator: %s", opt.CrownedOption.CrownedOp)
		}

	default:
		return "", nil, fmt.Errorf("unsupported expression operator: %s", exp.Op)
	}
	if err != nil {
		return "", nil, err
	}

	return "WHERE " + expr, value, nil
}

// UnmarshalJSON unmarshal a json raw to this expression
func (exp *Expression) UnmarshalJSON(raw []byte) error {
	parsed := gjson.GetManyBytes(raw, "op", "rules")
	op := LogicOperator(parsed[0].String())
	rules := parsed[1]
	rules.Raw = strings.TrimSpace(rules.Raw)

	if len(op) == 0 {
		// both op and raw is empty, then it's an empty expression json.
		if len(rules.Raw) == 0 {
			return nil
		}

		return errors.New("invalid expression, operator field is empty, but have none empty rules")
	}

	exp.Op = op
	if err := op.Validate(); err != nil {
		return err
	}

	if rules.Raw == "null" {
		return nil
	}

	if !rules.IsArray() {
		return errors.New("rules should be an array")
	}

	if rules.Raw == "[]" {
		return nil
	}

	if strings.TrimSpace(rules.Raw) == "[]" {
		return nil
	}

	for _, value := range rules.Array() {
		if isAtomType(value) {
			atom := new(AtomRule)
			if err := json.Unmarshal([]byte(value.Raw), &atom); err != nil {
				return err
			}

			exp.Rules = append(exp.Rules, atom)
			continue
		}

		if isExpressionType(value) {
			expr := new(Expression)
			if err := json.Unmarshal([]byte(value.Raw), &expr); err != nil {
				return err
			}

			exp.Rules = append(exp.Rules, expr)
			continue
		}

		return fmt.Errorf("unknown expression rule type: %s", value.Raw)
	}

	return nil
}

func isAtomType(value gjson.Result) bool {
	parsed := gjson.GetMany(value.Raw, "field", "op", "value")
	if !parsed[0].Exists() || !parsed[1].Exists() || !parsed[2].Exists() {
		return false
	}

	return true
}

func isExpressionType(value gjson.Result) bool {
	parsed := gjson.GetMany(value.Raw, "op", "rules")
	if !parsed[0].Exists() || !parsed[1].Exists() {
		return false
	}

	return true
}

// LogMarshal marshal Expression to string for log print.
func (exp *Expression) LogMarshal() string {
	return logs.ObjectEncode(exp)
}

// RuleFactory defines an expression's basic rule.
// which is used to filter the resources.
type RuleFactory interface {
	// WithType get a rule's type
	WithType() RuleType
	// Validate this rule is valid or not
	Validate(opt *ExprOption) error
	// RuleField get this rule's filed
	RuleField() string
	// SQLExprAndValue convert this rule to a mysql's sub query expression, and field's value
	SQLExprAndValue(opt *SQLWhereOption) (string, map[string]interface{}, error)
}

var _ RuleFactory = new(AtomRule)

var _ RuleFactory = new(Expression)

// AtomRule is the basic query rule.
type AtomRule struct {
	Field string      `json:"field"`
	Op    OpFactory   `json:"op"`
	Value interface{} `json:"value"`
}

// WithType return this atom rule's tye.
func (ar AtomRule) WithType() RuleType {
	return AtomType
}

// Validate this atom rule is valid or not
// Note: opt can be nil, check it before use it.
func (ar AtomRule) Validate(opt *ExprOption) error {
	if len(ar.Field) == 0 {
		return errors.New("filed is empty")
	}

	// validate operator
	if err := ar.Op.Validate(); err != nil {
		return err
	}

	if ar.Value == nil {
		return errors.New("rule value can not be nil")
	}

	if opt != nil {
		typ, exist := opt.RuleFields[ar.Field]
		if !exist {
			return fmt.Errorf("rule field: %s is not exist in the expr option", ar.Field)
		}

		if err := validateFieldValue(ar.Value, typ); err != nil {
			return fmt.Errorf("invalid %s's value, %v", ar.Field, err)
		}
	}
	if strings.HasPrefix(string(ar.Op), "id_") && ar.Field != "id" {
		return fmt.Errorf("operator %s field only support id field", ar.Op)
	}

	// validate the operator's value
	if err := ar.Op.Operator().ValidateValue(ar.Value, opt); err != nil {
		return fmt.Errorf("%s validate failed, %v", ar.Field, err)
	}

	return nil
}

func validateFieldValue(v interface{}, typ enumor.ColumnType) error {
	switch reflect.TypeOf(v).Kind() {
	case reflect.Array, reflect.Slice:
		return validateSliceElements(v, typ)
	default:
	}

	switch typ {
	case enumor.String:
		if reflect.ValueOf(v).Type().Kind() != reflect.String {
			return errors.New("value should be a string")
		}

	case enumor.Numeric:
		if !assert.IsNumeric(v) {
			return errors.New("value should be a numeric")
		}

	case enumor.Boolean:
		if reflect.ValueOf(v).Type().Kind() != reflect.Bool {
			return errors.New("value should be a boolean")
		}

	case enumor.Time:
		valOf := reflect.ValueOf(v)
		if valOf.Type().Kind() != reflect.String {
			return fmt.Errorf("value should be a string time format like: %s", constant.TimeStdFormat)
		}

		if !constant.TimeStdRegexp.MatchString(valOf.String()) {
			return fmt.Errorf("invalid time format, should be like: %s", constant.TimeStdFormat)
		}

		_, err := time.Parse(constant.TimeStdFormat, valOf.String())
		if err != nil {
			return fmt.Errorf("parse time from value failed, err: %v", err)
		}
	case enumor.Json:
		// json字段的类型任意都行，无法进行校验

	default:
		return fmt.Errorf("unsupported value type format: %s", typ)
	}

	return nil
}

func validateSliceElements(v interface{}, typ enumor.ColumnType) error {
	value := reflect.ValueOf(v)
	length := value.Len()
	if length == 0 {
		return nil
	}

	// validate each slice's element data type
	for i := 0; i < length; i++ {
		if err := validateFieldValue(value.Index(i).Interface(), typ); err != nil {
			return err
		}
	}

	return nil
}

// RuleField get atom rule's filed
func (ar AtomRule) RuleField() string {
	return ar.Field
}

// SQLExprAndValue convert this atom rule to a mysql's sub query expression, and field's value.
func (ar AtomRule) SQLExprAndValue(opt *SQLWhereOption) (string, map[string]interface{}, error) {
	expr, value, err := ar.Op.Operator().SQLExprAndValue(ar.Field, ar.Value)
	if err != nil {
		return "", nil, err
	}

	return expr, value, nil
}

type broker struct {
	Field string          `json:"field"`
	Op    OpFactory       `json:"op"`
	Value json.RawMessage `json:"value"`
}

// UnmarshalJSON unmarshal the json raw to AtomRule
func (ar *AtomRule) UnmarshalJSON(raw []byte) error {
	br := new(broker)
	err := json.Unmarshal(raw, br)
	if err != nil {
		return err
	}

	ar.Field = br.Field
	ar.Op = br.Op
	if br.Op == OpFactory(In) || br.Op == OpFactory(NotIn) {
		// in and nin operator's value should be an array.
		array := make([]interface{}, 0)
		if err := json.Unmarshal(br.Value, &array); err != nil {
			return fmt.Errorf("unmarshal in/not_in value to []interface{} failed, err: %v", err)
		}

		ar.Value = array

		return nil
	}

	to := new(interface{})
	if err := json.Unmarshal(br.Value, to); err != nil {
		return err
	}
	ar.Value = *to

	return nil
}
