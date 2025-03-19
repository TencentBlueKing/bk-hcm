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
	"testing"
)

func TestEqualSQLExpr(t *testing.T) {
	// test eq
	eq := EqualOp(Equal)
	eqExpr, _, err := eq.SQLExprAndValue("name", "hcm")
	if err != nil {
		t.Errorf("test eq operator failed, err: %v", err)
		return
	}

	if eqExpr != `name = :name` {
		t.Errorf("test eq operator got wrong expr: %s", eqExpr)
		return
	}
}

func TestNotEqualSQLExpr(t *testing.T) {
	// test neq
	ne := NotEqualOp(NotEqual)
	neExpr, _, err := ne.SQLExprAndValue("name", "hcm")
	if err != nil {
		t.Errorf("test ne operator failed, err: %v", err)
		return
	}

	if neExpr != `name != :name` {
		t.Errorf("test ne operator got wrong expr: %s", neExpr)
		return
	}
}

func TestGreaterThanSQLExpr(t *testing.T) {
	// test gt
	gt := GreaterThanOp(GreaterThan)
	gtExpr, _, err := gt.SQLExprAndValue("count", 10)
	if err != nil {
		t.Errorf("test gt operator failed, err: %v", err)
		return
	}

	if gtExpr != `count > :count` {
		t.Errorf("test gt operator got wrong expr: %s", gtExpr)
		return
	}

	// test time scenario
	gtExpr, _, err = gt.SQLExprAndValue("created_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test gt operator with time failed, err: %v", err)
		return
	}

	if gtExpr != `created_at > :created_at` {
		t.Errorf("test gt operator with time got wrong expr: %s", gtExpr)
		return
	}
}

func TestGreaterThanEqualSQLExpr(t *testing.T) {
	// test gte
	gte := GreaterThanEqualOp(GreaterThanEqual)
	gteExpr, _, err := gte.SQLExprAndValue("count", 10)
	if err != nil {
		t.Errorf("test gte operator failed, err: %v", err)
		return
	}

	if gteExpr != `count >= :count` {
		t.Errorf("test gte operator got wrong expr: %s", gteExpr)
		return
	}

	// test with time scenario
	gteExpr, _, err = gte.SQLExprAndValue("created_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test gte operator with time failed, err: %v", err)
		return
	}

	if gteExpr != `created_at >= :created_at` {
		t.Errorf("test gte operator with time got wrong expr: %s", gteExpr)
		return
	}
}

func TestLessThanSQLExpr(t *testing.T) {
	// test lt
	lt := LessThanOp(LessThan)
	ltExpr, _, err := lt.SQLExprAndValue("count", 10)
	if err != nil {
		t.Errorf("test lt operator failed, err: %v", err)
		return
	}

	if ltExpr != `count < :count` {
		t.Errorf("test lt operator got wrong expr: %s", ltExpr)
		return
	}

	// test time scenario
	ltExpr, _, err = lt.SQLExprAndValue("created_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test lt operator with time failed, err: %v", err)
		return
	}

	if ltExpr != `created_at < :created_at` {
		t.Errorf("test lt operator with time got wrong expr: %s", ltExpr)
		return
	}
}

func TestLessThanEqualSQLExpr(t *testing.T) {
	// test lte
	lte := LessThanEqualOp(LessThanEqual)
	lteExpr, _, err := lte.SQLExprAndValue("count", 10)
	if err != nil {
		t.Errorf("test lte operator failed, err: %v", err)
		return
	}

	if lteExpr != `count <= :count` {
		t.Errorf("test lte operator got wrong expr: %s", lteExpr)
		return
	}

	// test time scenario
	lteExpr, _, err = lte.SQLExprAndValue("created_at", "2022-01-02 15:04:05")
	if err != nil {
		t.Errorf("test lte operator with time failed, err: %v", err)
		return
	}

	if lteExpr != `created_at <= :created_at` {
		t.Errorf("test lte operator with time got wrong expr: %s", lteExpr)
		return
	}
}

func TestInSQLExpr(t *testing.T) {
	// test in
	in := InOp(In)
	sinExpr, _, err := in.SQLExprAndValue("servers", []string{"api", "web"})
	if err != nil {
		t.Errorf("test in operator failed, err: %v", err)
		return
	}

	if sinExpr != `servers IN (:servers)` {
		t.Errorf("test in operator got wrong expr: %s", sinExpr)
		return
	}

	intInExpr, _, err := in.SQLExprAndValue("ages", []int{18, 30})
	if err != nil {
		t.Errorf("test in operator failed, err: %v", err)
		return
	}

	if intInExpr != `ages IN (:ages)` {
		t.Errorf("test in operator got wrong expr: %s", sinExpr)
		return
	}
}

func TestNotInSQLExpr(t *testing.T) {
	// test nin
	nin := NotInOp(NotIn)
	sinExpr, _, err := nin.SQLExprAndValue("servers", []string{"api", "web"})
	if err != nil {
		t.Errorf("test nin operator failed, err: %v", err)
		return
	}

	if sinExpr != `servers NOT IN (:servers)` {
		t.Errorf("test nin operator got wrong expr: %s", sinExpr)
		return
	}

	intInExpr, _, err := nin.SQLExprAndValue("ages", []int{18, 30})
	if err != nil {
		t.Errorf("test nin operator failed, err: %v", err)
		return
	}

	if intInExpr != `ages NOT IN (:ages)` {
		t.Errorf("test nin operator got wrong expr: %s", sinExpr)
		return
	}
}

func TestContainsSensitiveSQLExpr(t *testing.T) {
	// test cs
	cs := ContainsSensitiveOp(ContainsSensitive)
	csExpr, _, err := cs.SQLExprAndValue("name", "hcm-")
	if err != nil {
		t.Errorf("test cis operator failed, err: %v", err)
		return
	}

	if csExpr != `name LIKE BINARY '%:name%'` {
		t.Errorf("test cis operator got wrong expr: %s", csExpr)
		return
	}
}

func TestContainsInsensitiveSQLExpr(t *testing.T) {
	// test cis
	cis := ContainsInsensitiveOp(ContainsInsensitive)
	cisExpr, _, err := cis.SQLExprAndValue("name", "hcm-")
	if err != nil {
		t.Errorf("test cis operator failed, err: %v", err)
		return
	}

	if cisExpr != `name LIKE '%:name%'` {
		t.Errorf("test cis operator got wrong expr: %s", cisExpr)
		return
	}
}

func TestJSONEqualSQLExpr(t *testing.T) {
	jsonEqual := JSONEqualOp(JSONEqual)
	expr, valueMap, err := jsonEqual.SQLExprAndValue("extension.vpc.id", "vpc-xxx")
	if err != nil {
		t.Errorf("test json equal operator failed, err: %v", err)
		return
	}

	if expr != `extension->>"$.vpc.id" = :extensionvpcid` {
		t.Errorf("test json equal got wrong expr: %s", expr)
		return
	}

	if len(valueMap) != 1 {
		t.Errorf("test json equal got wrong value: %v", valueMap)
		return
	}

	for key, val := range valueMap {
		if key != "extensionvpcid" {
			t.Errorf("test json equal got wrong value map's key: %v", key)
			return
		}

		if val != "vpc-xxx" {
			t.Errorf("test json equal got wrong value map's value: %v", val)
			return
		}
	}
}

func TestJSONInSQLExpr(t *testing.T) {
	jsonEqual := JSONInOp(JSONIn)
	expr, valueMap, err := jsonEqual.SQLExprAndValue("extension.vpc.id", []string{"vpc-xxx"})
	if err != nil {
		t.Errorf("test json equal operator failed, err: %v", err)
		return
	}

	if expr != `extension->>"$.vpc.id" IN (:extensionvpcid)` {
		t.Errorf("test json equal got wrong expr: %s", expr)
		return
	}

	if len(valueMap) != 1 {
		t.Errorf("test json equal got wrong value: %v", valueMap)
		return
	}

	for key, val := range valueMap {
		if key != "extensionvpcid" {
			t.Errorf("test json equal got wrong value map's key: %v", key)
			return
		}

		if val.([]string)[0] != "vpc-xxx" {
			t.Errorf("test json equal got wrong value map's value: %v", val)
			return
		}
	}
}

func Test_jsonFiledSqlFormat(t *testing.T) {
	type args struct {
		field string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "multi", args: args{field: "extension.vpc.id"},
			want: `JSON_UNQUOTE(JSON_EXTRACT(extension,'$."vpc"."id"'))`},
		{name: "non-json", args: args{field: "extension"}, want: `extension`},
		{name: "single", args: args{field: "extension.id"}, want: `JSON_UNQUOTE(JSON_EXTRACT(extension,'$."id"'))`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := jsonFieldSqlFormat(tt.args.field); got != tt.want {
				t.Errorf("jsonFieldSqlFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
