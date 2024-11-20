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
	"fmt"
	"strings"
	"testing"

	"hcm/pkg/criteria/enumor"
)

func TestUnmarshal(t *testing.T) {
	exprJson := `
{
	"op": "and",
	"rules": [{
			"field": "deploy_type",
			"op": "eq",
			"value": "common"
		},
		{
			"field": "creator",
			"op": "eq",
			"value": "tom"
		}
	]
}
`
	expr := new(Expression)
	err := expr.UnmarshalJSON([]byte(exprJson))
	if err != nil {
		t.Error(err)
		return
	}

	if !(expr.Op == And && len(expr.Rules) == 2 && (expr.Rules[0].(*AtomRule).Field == "deploy_type" &&
		expr.Rules[0].(*AtomRule).Op == Equal.Factory() && expr.Rules[0].(*AtomRule).Value == "common") &&
		(expr.Rules[1].(*AtomRule).Field == "creator" && expr.Rules[1].(*AtomRule).Op == "eq" &&
			expr.Rules[1].(*AtomRule).Value == "tom")) {
		t.Errorf("expression is not expected, op: %s, rules[0]: %v, rules[1]: %v", expr.Op,
			expr.Rules[0].(*AtomRule), expr.Rules[1].(*AtomRule))
		return
	}
}

func TestExpressionAnd(t *testing.T) {
	expr := &Expression{
		Op: And,
		Rules: []RuleFactory{
			&AtomRule{
				Field: "name",
				Op:    Equal.Factory(),
				Value: "hcm",
			},
			&AtomRule{
				Field: "age",
				Op:    GreaterThan.Factory(),
				Value: 18,
			},
			&AtomRule{
				Field: "age",
				Op:    LessThan.Factory(),
				Value: 30,
			},
			&AtomRule{
				Field: "servers",
				Op:    In.Factory(),
				Value: []string{"api", "web"},
			},
		},
	}

	if err := expr.Validate(nil); err != nil {
		t.Errorf("validate expression failed, err: %v", err)
		return
	}

	opt := &SQLWhereOption{Priority: []string{"servers", "age", "name"}}
	sql, value, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate expression's sql where expression failed, err: %v", err)
		return
	}

	fmt.Printf("where And expr: %s, value: %v\n", sql, value)
	if sql != `WHERE servers IN (:servers) AND age > :age AND age < :age AND name = :name` {
		t.Errorf("expression's sql where is not expected, sql: %s", sql)
		return
	}
}

func TestExpressionOr(t *testing.T) {
	expr := &Expression{
		Op: Or,
		Rules: []RuleFactory{
			&AtomRule{
				Field: "name",
				Op:    Equal.Factory(),
				Value: "hcm",
			},
			&AtomRule{
				Field: "age",
				Op:    GreaterThan.Factory(),
				Value: 18,
			},
			&AtomRule{
				Field: "age",
				Op:    LessThan.Factory(),
				Value: 30,
			},
			&AtomRule{
				Field: "servers",
				Op:    In.Factory(),
				Value: []string{"api", "web"},
			},
		},
	}

	if err := expr.Validate(nil); err != nil {
		t.Errorf("validate expression failed, err: %v", err)
		return
	}

	opt := &SQLWhereOption{Priority: []string{"servers", "age", "name"}}
	sql, value, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate expression's sql where expression failed, err: %v", err)
		return
	}

	fmt.Printf("where OR expr: %s, value: %v\n", sql, value)
	if sql != `WHERE servers IN (:servers) OR age > :age OR age < :age OR name = :name` {
		t.Errorf("expression's sql where is not expected, sql: %s", sql)
		return
	}
}

func TestExpressionValidateOption(t *testing.T) {
	expr := &Expression{
		Op: And,
		Rules: []RuleFactory{
			&AtomRule{
				Field: "name",
				Op:    Equal.Factory(),
				Value: "hcm",
			},
			&AtomRule{
				Field: "age",
				Op:    GreaterThan.Factory(),
				Value: 18,
			},
			&AtomRule{
				Field: "age",
				Op:    LessThan.Factory(),
				Value: 30,
			},
			&AtomRule{
				Field: "servers",
				Op:    In.Factory(),
				Value: []string{"api", "web"},
			},
			&AtomRule{
				Field: "asDefault",
				Op:    Equal.Factory(),
				Value: true,
			},
			&AtomRule{
				Field: "created_at",
				Op:    GreaterThan.Factory(),
				Value: "2006-01-02 15:04:05",
			},
		},
	}

	opt := &ExprOption{
		RuleFields: map[string]enumor.ColumnType{
			"name":       enumor.String,
			"age":        enumor.Numeric,
			"servers":    enumor.String,
			"asDefault":  enumor.Boolean,
			"created_at": enumor.Time,
		},
		MaxInLimit:    0,
		MaxNotInLimit: 0,
		MaxRulesLimit: 10,
	}

	if err := expr.Validate(opt); err != nil {
		t.Errorf("validate expression failed, err: %v", err)
		return
	}

	// test invalidate scenario
	opt.RuleFields["name"] = enumor.Numeric
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a numeric") {
		t.Errorf("validate numeric type failed, err: %v", err)
		return
	}
	opt.RuleFields["name"] = enumor.String

	opt.RuleFields["age"] = enumor.String
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a string") {
		t.Errorf("validate string type failed, err: %v", err)
		return
	}
	opt.RuleFields["age"] = enumor.Numeric

	opt.RuleFields["asDefault"] = enumor.Time
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a string time format") {
		t.Errorf("validate time type failed, err: %v", err)
		return
	}
	opt.RuleFields["asDefault"] = enumor.Boolean

	opt.RuleFields["created_at"] = enumor.Boolean
	if err := expr.Validate(opt); !strings.Contains(err.Error(), "value should be a boolean") {
		t.Errorf("validate boolean type failed, err: %v", err)
		return
	}
}

func TestWildcardExpressionValidateOption(t *testing.T) {
	opt := &ExprOption{
		RuleFields: map[string]enumor.ColumnType{
			"extension.name":     enumor.String,
			"extension.*.field2": enumor.String,
		},
		MaxInLimit:    0,
		MaxNotInLimit: 0,
		MaxRulesLimit: 10,
	}
	t.Run("single_dot", func(t *testing.T) {
		expr := &Expression{
			Op: And,
			Rules: []RuleFactory{
				&AtomRule{
					Field: "extension.field1",
					Op:    JSONEqual.Factory(),
					Value: "hcm",
				},
			},
		}

		if err := expr.Validate(opt); !strings.Contains(err.Error(),
			"expression rules filed(extension.field1) should not exist(not supported)") {
			t.Errorf("validate field failed, err: %v", err)
			return
		}
		opt.RuleFields["extension.*"] = enumor.String
		if err := expr.Validate(opt); err != nil {
			t.Errorf("validate wildcard expression failed, err: %v", err)
			return
		}
		return
	})

	t.Run("second_dot_wildcard", func(t *testing.T) {
		expr := &Expression{
			Op: And,
			Rules: []RuleFactory{
				&AtomRule{
					Field: "extension.field1.field2",
					Op:    JSONEqual.Factory(),
					Value: "hcm",
				},
			},
		}
		if err := expr.Validate(opt); !strings.Contains(err.Error(),
			"expression rules filed(extension.field1.field2) should not exist(not supported)") {
			t.Errorf("validate field failed, err: %v", err)
			return
		}

		opt.RuleFields["extension.field1.*"] = enumor.String
		if err := expr.Validate(opt); err != nil {
			t.Errorf("validate wildcard expression failed, err: %v", err)
			return
		}
	})

}

func TestCrownSQLWhereExpr(t *testing.T) {
	expr := &Expression{
		Op: And,
		Rules: []RuleFactory{
			&AtomRule{
				Field: "name",
				Op:    Equal.Factory(),
				Value: "hcm",
			},
			&AtomRule{
				Field: "age",
				Op:    GreaterThan.Factory(),
				Value: 18,
			},
		},
	}

	opt := &SQLWhereOption{
		Priority: []string{"biz_id", "age"},
		CrownedOption: &CrownedOption{
			CrownedOp: And,
			Rules: []RuleFactory{
				&AtomRule{
					Field: "biz_id",
					Op:    Equal.Factory(),
					Value: 20,
				},
				&AtomRule{
					Field: "created_at",
					Op:    GreaterThan.Factory(),
					Value: "2023-01-02T07:04:05+08:00",
				},
			},
		},
	}

	where, value, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where expression failed, err: %v", err)
		return
	}

	fmt.Printf("where AND-AND expr: %s, value: %v\n", where, value)
	if where != `WHERE biz_id = :biz_id AND age > :age AND name = :name AND created_at > :created_at` {
		t.Errorf("generate SQL AND-AND Where expression failed, err: %v", err)
		return
	}

	expr.Op = And
	opt.CrownedOption.CrownedOp = Or
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where AND-OR expression failed, err: %v", err)
		return
	}

	fmt.Printf("where AND-OR expr: %s, value: %v\n", where, value)
	if where != `WHERE (biz_id = :biz_id AND created_at > :created_at) OR (age > :age AND name = :name)` {
		t.Errorf("generate SQL AND-OR Where expression failed, where: %v", where)
		return
	}

	expr.Op = Or
	opt.CrownedOption.CrownedOp = Or
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where OR-OR expression failed, err: %v", err)
		return
	}

	fmt.Printf("where OR-OR expr: %s, value: %v\n", where, value)
	if where != `WHERE (biz_id = :biz_id AND created_at > :created_at) OR age > :age OR name = :name` {
		t.Errorf("generate SQL OR-OR Where expression failed, where: %v", where)
		return
	}

	opt.Priority = []string{"age", "biz_id"}
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where OR-OR expression failed, err: %v", err)
		return
	}

	// reverse the priority
	fmt.Printf("where OR-OR-PRIORITY expr: %s, value: %v\n", where, value)
	if where != `WHERE age > :age OR name = :name OR (biz_id = :biz_id AND created_at > :created_at)` {
		t.Errorf("generate SQL OR-OR-PRIORITY Where expression failed, where: %v", where)
		return
	}

	expr.Op = Or
	opt.CrownedOption.CrownedOp = And
	opt.Priority = []string{"biz_id", "age"}
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where OR-AND expression failed, err: %v", err)
		return
	}

	fmt.Printf("where OR-AND expr: %s, value: %v\n", where, value)
	if where != `WHERE (biz_id = :biz_id AND created_at > :created_at) AND (age > :age OR name = :name)` {
		t.Errorf("generate SQL OR-AND Where expression failed, where: %v", where)
		return
	}

	// test NULL crown rules
	expr.Rules = []RuleFactory{
		&AtomRule{Field: "name", Op: Equal.Factory(), Value: "hcm"},
		&AtomRule{Field: "age", Op: GreaterThan.Factory(), Value: 18},
	}
	opt.CrownedOption.Rules = make([]RuleFactory, 0)
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where NULL crown expr expression failed, err: %v", err)
		return
	}

	fmt.Printf("where crown NULL expr: %s, value: %v\n", where, value)
	if where != `WHERE age > :age OR name = :name` {
		t.Errorf("generate SQL Where NULL crown expr expression failed, where: %s", where)
		return
	}

	// test NULL Expression rules
	expr.Rules = make([]RuleFactory, 0)
	opt.CrownedOption.Rules = []RuleFactory{&AtomRule{
		Field: "age",
		Op:    Equal.Factory(),
		Value: 8,
	}}
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where NULL crown expr expression failed, err: %v", err)
		return
	}

	fmt.Printf("where crown NULL expr: %s, value: %v\n", where, value)
	if where != `WHERE age = :age` {
		t.Errorf("generate SQL Where NULL crown expr expression failed, where: %s", where)
		return
	}

	// test both Expression and crown rules is empty
	expr.Rules = make([]RuleFactory, 0)
	opt.CrownedOption.Rules = make([]RuleFactory, 0)
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where both rule is NULL failed, err: %v", err)
		return
	}

	fmt.Printf("where both NULL expr: %s, value: %v\n", where, value)
	if where != "" {
		t.Errorf("generate SQL Where both rule is NULL failed, where: %s", where)
		return
	}
}

func TestNestedSqlWhereExpr(t *testing.T) {
	expr := &Expression{
		Op: And,
		Rules: []RuleFactory{
			&AtomRule{
				Field: "name",
				Op:    Equal.Factory(),
				Value: "jim",
			},
			&AtomRule{
				Field: "age",
				Op:    GreaterThan.Factory(),
				Value: 18,
			},
			&Expression{
				Op: Or,
				Rules: []RuleFactory{
					&AtomRule{
						Field: "child_name",
						Op:    Equal.Factory(),
						Value: "jon",
					},
					&AtomRule{
						Field: "child_age",
						Op:    LessThan.Factory(),
						Value: 10,
					},
				},
			},
		},
	}

	err := expr.Validate(nil)
	if err != nil {
		t.Errorf("validate SQL Where expression failed, err: %v", err)
		return
	}

	opt := &SQLWhereOption{
		Priority: []string{"biz_id", "age"},
		CrownedOption: &CrownedOption{
			CrownedOp: And,
			Rules: []RuleFactory{
				&AtomRule{
					Field: "biz_id",
					Op:    Equal.Factory(),
					Value: 20,
				},
				&AtomRule{
					Field: "created_at",
					Op:    GreaterThan.Factory(),
					Value: "2021-01-01 08:09:10",
				},
			},
		},
	}

	where, value, err := expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where expression failed, err: %v", err)
		return
	}

	fmt.Printf("where Nested AND-AND expr: %s, value: %v\n", where, value)
	if where != "WHERE biz_id = :biz_id AND age > :age AND name = :name AND (child_name = :child_name OR "+
		"child_age < :child_age) AND created_at > :created_at" {
		t.Errorf("generate SQL Nested Where expression failed, sql: %v", where)
		return
	}

	opt.CrownedOption.CrownedOp = Or
	where, value, err = expr.SQLWhereExpr(opt)
	if err != nil {
		t.Errorf("generate SQL Where expression failed, err: %v", err)
		return
	}

	fmt.Printf("where Nested AND-OR expr: %s, value: %v\n", where, value)
	if where != "WHERE (biz_id = :biz_id AND created_at > :created_at) OR (age > :age AND name = :name AND "+
		"(child_name = :child_name OR child_age < :child_age))" {
		t.Errorf("generate SQL Nested Where expression failed, sql: %v", where)
		return
	}
}
