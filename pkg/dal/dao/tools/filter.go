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

package tools

import (
	"fmt"

	"hcm/pkg/runtime/filter"
)

// EqualExpression 生成资源字段等于查询的过滤条件，即fieldName=value
func EqualExpression(fieldName string, value interface{}) *filter.Expression {
	return &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: fieldName, Op: filter.Equal.Factory(), Value: value},
		},
	}
}

// EqualWithOpExpression 生成查询条件
func EqualWithOpExpression(op filter.LogicOperator, fields map[string]interface{}) *filter.Expression {
	rules := make([]filter.RuleFactory, 0)
	for name, value := range fields {
		rules = append(rules, filter.AtomRule{Field: name, Op: filter.Equal.Factory(), Value: value})
	}
	return &filter.Expression{
		Op:    op,
		Rules: rules,
	}
}

// ContainersExpression 生成资源字段包含的过滤条件，即fieldName in (1,2,3)
func ContainersExpression[T any](fieldName string, values []T) *filter.Expression {
	return &filter.Expression{
		Op: filter.And,
		Rules: []filter.RuleFactory{
			filter.AtomRule{Field: fieldName, Op: filter.In.Factory(), Value: values},
		},
	}
}

// AllExpression 生成全量查询filter。
func AllExpression() *filter.Expression {
	return &filter.Expression{
		Op:    filter.And,
		Rules: make([]filter.RuleFactory, 0),
	}
}

// DefaultSqlWhereOption define sql where option.
var DefaultSqlWhereOption = &filter.SQLWhereOption{
	Priority: filter.Priority{"id"},
}

// And merge expressions using 'and' operation.
func And(rules ...filter.RuleFactory) (*filter.Expression, error) {
	if len(rules) == 0 {
		return nil, fmt.Errorf("rules are not set")
	}

	andRules := make([]filter.RuleFactory, 0)
	for _, rule := range rules {
		switch rule.WithType() {
		case filter.AtomType:
			andRules = append(andRules, rule)
		case filter.ExpressionType:
			expr, ok := rule.(*filter.Expression)
			if !ok {
				return nil, fmt.Errorf("rule type is not expression")
			}
			if expr.Op == filter.And {
				andRules = append(andRules, expr.Rules...)
				continue
			}
			andRules = append(andRules, expr)
		default:
			return nil, fmt.Errorf("rule type %s is invalid", rule.WithType())
		}
	}

	return &filter.Expression{
		Op:    filter.And,
		Rules: andRules,
	}, nil
}

// RuleEqual 生成资源字段等于查询的AtomRule，即fieldName=value
func RuleEqual(fieldName string, value any) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.Equal.Factory(), Value: value}
}

// RuleIn 生成资源字段等于查询的AtomRule，即fieldName in values
func RuleIn[T any](fieldName string, values []T) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.In.Factory(), Value: values}
}

// RuleNotIn 生成资源字段等于查询的AtomRule，即fieldName nin values
func RuleNotIn[T any](fieldName string, values []T) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.NotIn.Factory(), Value: values}
}

// RuleGreaterThan 生成资源字段等于查询的AtomRule，即fieldName > values
func RuleGreaterThan(fieldName string, values any) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.GreaterThan.Factory(), Value: values}
}

// RuleJSONEqual 生成资源字段等于查询的AtomRule，即fieldName=value
func RuleJSONEqual(fieldName string, value any) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.JSONEqual.Factory(), Value: value}
}

// RuleJSONNotEqual 生成资源字段等于查询的AtomRule，即fieldName!=value
func RuleJSONNotEqual(fieldName string, value any) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.JSONNotEqual.Factory(), Value: value}
}

// RuleJsonIn 生成资源字段等于查询的AtomRule，即fieldName in values
func RuleJsonIn[T any](fieldName string, values []T) *filter.AtomRule {
	return &filter.AtomRule{Field: fieldName, Op: filter.JSONIn.Factory(), Value: values}
}

// ExpressionAnd expression with op and
func ExpressionAnd(rules ...*filter.AtomRule) *filter.Expression {
	// for type transformation
	var factories = make([]filter.RuleFactory, len(rules))
	for i, rule := range rules {
		factories[i] = rule
	}
	return &filter.Expression{
		Op:    filter.And,
		Rules: factories,
	}
}

// ExpressionOr expression with op or
func ExpressionOr(rules ...*filter.AtomRule) *filter.Expression {
	// for type transformation
	var factories = make([]filter.RuleFactory, len(rules))
	for i, rule := range rules {
		factories[i] = rule
	}
	return &filter.Expression{
		Op:    filter.Or,
		Rules: factories,
	}
}
