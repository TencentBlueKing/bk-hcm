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

package excel

import (
	"math"
	"testing"

	toolsjson "hcm/pkg/tools/json"
)

func float64Ptr(v float64) *float64 { return &v }

func TestValidateCellValue_Required(t *testing.T) {
	h := Header{Name: "预算卡数", Type: "int", Required: true}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"empty value", "", 1},
		{"whitespace only", "   ", 1},
		{"valid value", "100", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_NonRequiredEmpty(t *testing.T) {
	h := Header{Name: "备注", Type: "int", Required: false}
	errs := ValidateCellValue("", h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for non-required empty, got %v", errs)
	}
}

func TestValidateCellValue_Int(t *testing.T) {
	h := Header{Name: "预算卡数", Type: "int"}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"valid integer", "2026", 0},
		{"negative integer", "-5", 0},
		{"decimal value", "12.5", 1},
		{"non-numeric string", "abc", 1},
		{"zero", "0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_IntRange(t *testing.T) {
	h := Header{
		Name: "预算卡数", Type: "int",
		GTE: float64Ptr(0), LTE: float64Ptr(1000),
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"within range", "500", 0},
		{"at gte boundary", "0", 0},
		{"at lte boundary", "1000", 0},
		{"below gte", "-1", 1},
		{"above lte", "1500", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_IntGTLT(t *testing.T) {
	h := Header{
		Name: "预算卡数", Type: "int",
		GT: float64Ptr(0), LT: float64Ptr(1000),
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"within open range", "500", 0},
		{"at gt boundary (fail)", "0", 1},
		{"at lt boundary (fail)", "1000", 1},
		{"above gt", "1", 0},
		{"below lt", "999", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_IntSingleBound(t *testing.T) {
	tests := []struct {
		name    string
		header  Header
		val     string
		wantLen int
	}{
		// GTE only: boundary inclusive
		{"gte only: at boundary", Header{Name: "n", Type: "int", GTE: float64Ptr(0)}, "0", 0},
		{"gte only: above boundary", Header{Name: "n", Type: "int", GTE: float64Ptr(0)}, "1", 0},
		{"gte only: below boundary", Header{Name: "n", Type: "int", GTE: float64Ptr(0)}, "-1", 1},
		// GT only: boundary exclusive
		{"gt only: at boundary (fail)", Header{Name: "n", Type: "int", GT: float64Ptr(0)}, "0", 1},
		{"gt only: above boundary", Header{Name: "n", Type: "int", GT: float64Ptr(0)}, "1", 0},
		{"gt only: below boundary", Header{Name: "n", Type: "int", GT: float64Ptr(0)}, "-1", 1},
		// LTE only: boundary inclusive
		{"lte only: at boundary", Header{Name: "n", Type: "int", LTE: float64Ptr(100)}, "100", 0},
		{"lte only: below boundary", Header{Name: "n", Type: "int", LTE: float64Ptr(100)}, "99", 0},
		{"lte only: above boundary", Header{Name: "n", Type: "int", LTE: float64Ptr(100)}, "101", 1},
		// LT only: boundary exclusive
		{"lt only: at boundary (fail)", Header{Name: "n", Type: "int", LT: float64Ptr(100)}, "100", 1},
		{"lt only: below boundary", Header{Name: "n", Type: "int", LT: float64Ptr(100)}, "99", 0},
		{"lt only: above boundary", Header{Name: "n", Type: "int", LT: float64Ptr(100)}, "101", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, tt.header)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

// TestValidateCellValue_IntExtremeValues 验证极端整数值不会引发 panic 或误判。
// int64 转 float64 在接近 math.MaxInt64 时有精度衰减（约 ±1），
// 但约束值同样经过 float64 存储，比较结果在实际使用范围内是正确的。
func TestValidateCellValue_IntExtremeValues(t *testing.T) {
	tests := []struct {
		name    string
		header  Header
		val     string
		wantLen int
	}{
		// 大正数满足 LTE 约束
		{
			"large positive within lte",
			Header{Name: "n", Type: "int", LTE: float64Ptr(1e15)},
			"999999999999999", 0,
		},
		// 大正数违反 LTE 约束
		{
			"large positive above lte",
			Header{Name: "n", Type: "int", LTE: float64Ptr(1e15)},
			"1000000000000001", 1,
		},
		// 大负数满足 GTE 约束
		{
			"large negative within gte",
			Header{Name: "n", Type: "int", GTE: float64Ptr(-1e15)},
			"-999999999999999", 0,
		},
		// 大负数违反 GTE 约束
		{
			"large negative below gte",
			Header{Name: "n", Type: "int", GTE: float64Ptr(-1e15)},
			"-1000000000000001", 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, tt.header)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_Float(t *testing.T) {
	h := Header{Name: "实际使用TPM", Type: "float"}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"valid float", "12.5", 0},
		{"valid integer as float", "100", 0},
		{"negative float", "-3.14", 0},
		{"non-numeric string", "abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_FloatRange(t *testing.T) {
	h := Header{
		Name: "实际使用TPM", Type: "float",
		GTE: float64Ptr(0), LTE: float64Ptr(99.9),
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"within range", "50.5", 0},
		{"at gte boundary", "0", 0},
		{"at lte boundary", "99.9", 0},
		{"below gte", "-0.1", 1},
		{"above lte", "100", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_FloatGTLT(t *testing.T) {
	h := Header{
		Name: "实际使用TPM", Type: "float",
		GT: float64Ptr(0), LT: float64Ptr(99.9),
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"within open range", "50.5", 0},
		{"at gt boundary (fail)", "0", 1},
		{"at lt boundary (fail)", "99.9", 1},
		{"above gt", "0.1", 0},
		{"below lt", "99.8", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_FloatSingleBound(t *testing.T) {
	tests := []struct {
		name    string
		header  Header
		val     string
		wantLen int
	}{
		// GTE only
		{"gte only: at boundary", Header{Name: "f", Type: "float", GTE: float64Ptr(0)}, "0", 0},
		{"gte only: above boundary", Header{Name: "f", Type: "float", GTE: float64Ptr(0)}, "0.1", 0},
		{"gte only: below boundary", Header{Name: "f", Type: "float", GTE: float64Ptr(0)}, "-0.1", 1},
		// GT only
		{"gt only: at boundary (fail)", Header{Name: "f", Type: "float", GT: float64Ptr(0)}, "0", 1},
		{"gt only: above boundary", Header{Name: "f", Type: "float", GT: float64Ptr(0)}, "0.001", 0},
		{"gt only: below boundary", Header{Name: "f", Type: "float", GT: float64Ptr(0)}, "-0.001", 1},
		// LTE only
		{"lte only: at boundary", Header{Name: "f", Type: "float", LTE: float64Ptr(1.5)}, "1.5", 0},
		{"lte only: below boundary", Header{Name: "f", Type: "float", LTE: float64Ptr(1.5)}, "1.4", 0},
		{"lte only: above boundary", Header{Name: "f", Type: "float", LTE: float64Ptr(1.5)}, "1.6", 1},
		// LT only
		{"lt only: at boundary (fail)", Header{Name: "f", Type: "float", LT: float64Ptr(1.5)}, "1.5", 1},
		{"lt only: below boundary", Header{Name: "f", Type: "float", LT: float64Ptr(1.5)}, "1.499", 0},
		{"lt only: above boundary", Header{Name: "f", Type: "float", LT: float64Ptr(1.5)}, "1.501", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, tt.header)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

// TestValidateCellValue_FloatExtremeValues 验证 float 极端值（math.MaxFloat64 / +Inf / NaN）的行为。
// +Inf / -Inf 作为值：所有有限约束均被违反（Inf > any lte）。
// NaN 作为值：所有约束均不触发（NaN 的所有比较结果为 false），属已知的 float64 语义。
func TestValidateCellValue_FloatExtremeValues(t *testing.T) {
	tests := []struct {
		name    string
		header  Header
		val     string
		wantLen int
	}{
		// math.MaxFloat64 满足宽松 LTE（约束值 = MaxFloat64，值 = MaxFloat64）
		{
			"MaxFloat64 within lte MaxFloat64",
			Header{Name: "f", Type: "float", LTE: float64Ptr(math.MaxFloat64)},
			"1.7976931348623157e+308", 0,
		},
		// math.MaxFloat64 违反常规 LTE
		{
			"MaxFloat64 above lte 1e308",
			Header{Name: "f", Type: "float", LTE: float64Ptr(1e308)},
			"1.7976931348623157e+308", 1,
		},
		// +Inf 违反任意有限 LTE
		{
			"+Inf above lte",
			Header{Name: "f", Type: "float", LTE: float64Ptr(1e10)},
			"+Inf", 1,
		},
		// -Inf 违反任意有限 GTE
		{
			"-Inf below gte",
			Header{Name: "f", Type: "float", GTE: float64Ptr(-1e10)},
			"-Inf", 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, tt.header)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_EnumString(t *testing.T) {
	h := Header{
		Name:  "卡型",
		Type:  "enum",
		Value: []interface{}{"H20", "L20"},
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"match first", "H20", 0},
		{"match second", "L20", 0},
		{"no match", "A100", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_EnumNumeric(t *testing.T) {
	h := Header{
		Name:  "年份",
		Type:  "enum",
		Value: []interface{}{float64(2026), float64(2027), float64(2028)},
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"match integer", "2026", 0},
		{"no match integer", "2025", 1},
		{"non-numeric string", "一月", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_EnumNumericTypeMismatch(t *testing.T) {
	h := Header{
		Name:  "月份",
		Type:  "enum",
		Value: []interface{}{float64(1), float64(2), float64(3)},
	}

	errs := ValidateCellValue("一月", h)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %v", errs)
	}

	want := "月份: 值'一月'类型不匹配，应为数字"
	if errs[0] != want {
		t.Errorf("got %q, want %q", errs[0], want)
	}
}

func TestValidateCellValue_EnumEmpty(t *testing.T) {
	h := Header{Name: "卡型", Type: "enum", Value: []interface{}{"H20"}, Required: false}
	errs := ValidateCellValue("", h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for non-required empty enum, got %v", errs)
	}
}

func TestValidateCellValue_StringLength(t *testing.T) {
	h := Header{
		Name: "使用场景", Type: "string",
		GTE: float64Ptr(2), LTE: float64Ptr(10),
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"within range (Chinese)", "文生图场景", 0},
		{"at gte boundary", "短句", 0},
		{"at lte boundary", "H20卡型说明abc", 0},
		{"below gte", "短", 1},
		{"above lte", "这是一个非常长的描述文字超长", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_StringGTLT(t *testing.T) {
	h := Header{
		Name: "使用场景", Type: "string",
		GT: float64Ptr(1), LT: float64Ptr(10),
	}

	tests := []struct {
		name    string
		val     string
		wantLen int
	}{
		{"within open range", "文生图场景", 0},
		{"at gt boundary (fail)", "短", 1},
		{"at lt boundary (fail)", "H20卡型说明abc", 1},
		{"above gt", "短句", 0},
		{"below lt", "H20卡型说明ab", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_StringSingleBound(t *testing.T) {
	tests := []struct {
		name    string
		header  Header
		val     string
		wantLen int
	}{
		// GTE only: length >= 2
		{"gte only: at boundary", Header{Name: "s", Type: "string", GTE: float64Ptr(2)}, "ab", 0},
		{"gte only: above boundary", Header{Name: "s", Type: "string", GTE: float64Ptr(2)}, "abc", 0},
		{"gte only: below boundary", Header{Name: "s", Type: "string", GTE: float64Ptr(2)}, "a", 1},
		// GT only: length > 2
		{"gt only: at boundary (fail)", Header{Name: "s", Type: "string", GT: float64Ptr(2)}, "ab", 1},
		{"gt only: above boundary", Header{Name: "s", Type: "string", GT: float64Ptr(2)}, "abc", 0},
		{"gt only: below boundary", Header{Name: "s", Type: "string", GT: float64Ptr(2)}, "a", 1},
		// LTE only: length <= 5
		{"lte only: at boundary", Header{Name: "s", Type: "string", LTE: float64Ptr(5)}, "abcde", 0},
		{"lte only: below boundary", Header{Name: "s", Type: "string", LTE: float64Ptr(5)}, "abcd", 0},
		{"lte only: above boundary", Header{Name: "s", Type: "string", LTE: float64Ptr(5)}, "abcdef", 1},
		// LT only: length < 5
		{"lt only: at boundary (fail)", Header{Name: "s", Type: "string", LT: float64Ptr(5)}, "abcde", 1},
		{"lt only: below boundary", Header{Name: "s", Type: "string", LT: float64Ptr(5)}, "abcd", 0},
		{"lt only: above boundary", Header{Name: "s", Type: "string", LT: float64Ptr(5)}, "abcdef", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCellValue(tt.val, tt.header)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateCellValue_StringNoConstraints(t *testing.T) {
	h := Header{Name: "备注", Type: "string"}
	errs := ValidateCellValue("任意长度的字符串", h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for unconstrained string, got %v", errs)
	}
}

func TestValidateCellValue_StringEmptyNonRequired(t *testing.T) {
	h := Header{Name: "备注", Type: "string", GTE: float64Ptr(1), Required: false}
	errs := ValidateCellValue("", h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for non-required empty string, got %v", errs)
	}
}

func TestValidateCellValue_UnknownType(t *testing.T) {
	h := Header{Name: "未知", Type: "unknown"}
	errs := ValidateCellValue("anything", h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for unknown type, got %v", errs)
	}
}

// --- ValidateTypedValue tests ---

func TestValidateTypedValue_Required(t *testing.T) {
	h := Header{Name: "预算卡数", Type: "int", Required: true}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"nil value", nil, 1},
		{"empty string", "", 1},
		{"whitespace string", "   ", 1},
		{"valid value", float64(100), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_NonRequiredNil(t *testing.T) {
	h := Header{Name: "备注", Type: "string", Required: false}
	errs := ValidateTypedValue(nil, h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for non-required nil, got %v", errs)
	}
}

func TestValidateTypedValue_Int(t *testing.T) {
	h := Header{Name: "预算卡数", Type: "int"}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"integer as float64", float64(100), 0},
		{"zero", float64(0), 0},
		{"negative integer", float64(-5), 0},
		{"decimal float64", float64(12.5), 1},
		{"parseable string", "100", 0},
		{"non-parseable string", "abc", 1},
		{"decimal string", "12.5", 1},
		{"bool value", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_IntRange(t *testing.T) {
	h := Header{
		Name: "预算卡数", Type: "int",
		GTE: float64Ptr(0), LTE: float64Ptr(1000),
	}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"within range", float64(500), 0},
		{"at gte", float64(0), 0},
		{"at lte", float64(1000), 0},
		{"below gte", float64(-1), 1},
		{"above lte", float64(1500), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_Float(t *testing.T) {
	h := Header{Name: "实际使用TPM", Type: "float"}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"valid float", float64(12.5), 0},
		{"integer float", float64(100), 0},
		{"parseable string", "12.5", 0},
		{"non-parseable string", "abc", 1},
		{"bool value", true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_FloatRange(t *testing.T) {
	h := Header{
		Name: "实际使用TPM", Type: "float",
		GTE: float64Ptr(0), LTE: float64Ptr(99.9),
	}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"within range", float64(50.5), 0},
		{"at gte", float64(0), 0},
		{"at lte", float64(99.9), 0},
		{"below gte", float64(-0.1), 1},
		{"above lte", float64(100), 1},
		{"string within range", "50.5", 0},
		{"string below gte", "-1.5", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_EnumNumeric(t *testing.T) {
	h := Header{
		Name:  "年份",
		Type:  "enum",
		Value: []interface{}{float64(2026), float64(2027)},
	}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"match value", float64(2026), 0},
		{"no match", float64(2025), 1},
		{"parseable string match", "2026", 0},
		{"parseable string no match", "2025", 1},
		{"non-numeric string", "abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_EnumString(t *testing.T) {
	h := Header{
		Name:  "卡型",
		Type:  "enum",
		Value: []interface{}{"H20", "L20"},
	}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"match", "H20", 0},
		{"no match", "A100", 1},
		{"numeric type", float64(100), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_String(t *testing.T) {
	h := Header{
		Name: "使用场景", Type: "string",
		GTE: float64Ptr(2), LTE: float64Ptr(10),
	}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"within range (Chinese)", "文生图场景", 0},
		{"at gte", "短句", 0},
		{"below gte", "短", 1},
		{"above lte", "这是一个非常长的描述文字超长", 1},
		{"non-string type", float64(123), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_UnknownType(t *testing.T) {
	h := Header{Name: "未知", Type: "unknown"}
	errs := ValidateTypedValue("anything", h)
	if len(errs) != 0 {
		t.Errorf("expected no errors for unknown type, got %v", errs)
	}
}

// --- ValidateExtension tests ---

func TestValidateExtension_AllValid(t *testing.T) {
	headers := []Header{
		{Name: "卡型", Type: "enum", Value: []interface{}{"H20", "L20"}},
		{Name: "使用场景", Type: "string", GTE: float64Ptr(1), LTE: float64Ptr(20)},
	}
	values := []interface{}{"H20", "文生图"}

	errs := ValidateExtension(values, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateExtension_MixedErrors(t *testing.T) {
	headers := []Header{
		{Name: "卡型", Type: "enum", Value: []interface{}{"H20", "L20"}, Required: true},
		{Name: "数量", Type: "int", Required: true},
	}
	values := []interface{}{"A100", nil}

	errs := ValidateExtension(values, headers)
	if len(errs) != 2 {
		t.Errorf("expected 2 errors, got %d; errs=%v", len(errs), errs)
	}
}

func TestValidateExtension_ShorterValues(t *testing.T) {
	headers := []Header{
		{Name: "字段1", Type: "string"},
		{Name: "字段2", Type: "int", Required: true},
	}
	values := []interface{}{"hello"}

	errs := ValidateExtension(values, headers)
	if len(errs) != 1 {
		t.Errorf("expected 1 error (required field missing), got %d; errs=%v", len(errs), errs)
	}
}

func TestValidateExtension_Empty(t *testing.T) {
	headers := []Header{
		{Name: "可选", Type: "string"},
	}

	errs := ValidateExtension(nil, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors for nil values with non-required headers, got %v", errs)
	}
}

// unmarshalJSONValue uses toolsjson (UseNumber: true) to deserialize a JSON literal,
// returning the actual runtime type that production code sees (json.Number for numbers).
func unmarshalJSONValue(raw string) interface{} {
	var v interface{}
	if err := toolsjson.Unmarshal([]byte(raw), &v); err != nil {
		panic("bad test literal: " + raw)
	}
	return v
}

// unmarshalJSONSlice uses toolsjson to deserialize a JSON array into []interface{}.
func unmarshalJSONSlice(raw string) []interface{} {
	var v []interface{}
	if err := toolsjson.Unmarshal([]byte(raw), &v); err != nil {
		panic("bad test literal: " + raw)
	}
	return v
}

func TestNormalizeJSONNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantVal interface{}
	}{
		{"json.Number integer", unmarshalJSONValue("42"), float64(42)},
		{"json.Number float", unmarshalJSONValue("3.14"), float64(3.14)},
		{"json.Number negative", unmarshalJSONValue("-100"), float64(-100)},
		{"json.Number zero", unmarshalJSONValue("0"), float64(0)},
		{"plain float64 passthrough", float64(99), float64(99)},
		{"string passthrough", "hello", "hello"},
		{"nil passthrough", nil, nil},
		{"bool passthrough", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeJSONNumber(tt.input)
			if got != tt.wantVal {
				t.Errorf("normalizeJSONNumber(%v [%T]) = %v [%T], want %v [%T]",
					tt.input, tt.input, got, got, tt.wantVal, tt.wantVal)
			}
		})
	}
}

// TestNormalizeJSONNumber_Precision verifies that normalizeJSONNumber preserves
// exact values within float64's safe integer range (2^53) and documents the
// known precision loss boundary beyond it.
func TestNormalizeJSONNumber_Precision(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantF64 float64
		exact   bool
	}{
		// float64 安全整数范围内（<=2^53），精度无损
		{"year 2026", "2026", 2026, true},
		{"month 12", "12", 12, true},
		{"gpu_num 10000", "10000", 10000, true},
		{"qpm_max 999999", "999999", 999999, true},
		{"max safe integer (2^53)", "9007199254740992", 9007199254740992, true},
		{"negative safe integer", "-9007199254740992", -9007199254740992, true},
		// 超出 float64 安全整数范围（>2^53），精度会丢失，此为 float64 的已知限制
		{"beyond safe integer (2^53+1)", "9007199254740993", 9007199254740992, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := unmarshalJSONValue(tt.raw)
			got := normalizeJSONNumber(val)
			fv, ok := got.(float64)
			if !ok {
				t.Fatalf("expected float64, got %T", got)
			}

			if tt.exact {
				if fv != tt.wantF64 {
					t.Errorf("precision loss: raw=%s, got=%.0f, want=%.0f", tt.raw, fv, tt.wantF64)
				}
			} else {
				if fv == tt.wantF64 {
					t.Logf("expected precision loss confirmed: raw=%s → float64=%.0f", tt.raw, fv)
				}
			}
		})
	}
}

// TestValidateTypedValue_JSONNumber_IntPrecision verifies that integer validation
// produces correct results for values within the business range after json.Number
// normalization, ensuring no precision-related false positives or negatives.
func TestValidateTypedValue_JSONNumber_IntPrecision(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		header  Header
		wantLen int
	}{
		{
			"year boundary exact match",
			"2026",
			Header{Name: "年份", Type: "int", GTE: float64Ptr(2020), LTE: float64Ptr(2030)},
			0,
		},
		{
			"month 12 at lte boundary",
			"12",
			Header{Name: "月份", Type: "int", GTE: float64Ptr(1), LTE: float64Ptr(12)},
			0,
		},
		{
			"large gpu count within safe range",
			"999999999",
			Header{Name: "卡数", Type: "int", GTE: float64Ptr(0), LTE: float64Ptr(1e10)},
			0,
		},
		{
			"max safe integer at exact boundary",
			"9007199254740992",
			Header{Name: "n", Type: "int", LTE: float64Ptr(9007199254740992)},
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := unmarshalJSONValue(tt.raw)
			errs := ValidateTypedValue(val, tt.header)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

// TestValidateTypedValue_JSONNumber_EnumPrecision verifies that numeric enum matching
// works correctly for values within the business range after normalization.
func TestValidateTypedValue_JSONNumber_EnumPrecision(t *testing.T) {
	tests := []struct {
		name      string
		enumRaw   string
		valRaw    string
		wantMatch bool
	}{
		{"year enum exact match", "[2025, 2026, 2027]", "2026", true},
		{"month enum exact match", "[1, 6, 12]", "12", true},
		{"month enum no match", "[1, 6, 12]", "13", false},
		{"large enum id exact match", "[1000000001, 1000000002]", "1000000002", true},
		{"large enum id no match", "[1000000001, 1000000002]", "1000000003", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enumValues := unmarshalJSONSlice(tt.enumRaw)
			h := Header{Name: "e", Type: "enum", Value: enumValues}
			val := unmarshalJSONValue(tt.valRaw)

			errs := ValidateTypedValue(val, h)
			matched := len(errs) == 0
			if matched != tt.wantMatch {
				t.Errorf("enum match=%v, want %v; errs=%v", matched, tt.wantMatch, errs)
			}
		})
	}
}

func TestValidateTypedValue_JSONNumber_Int(t *testing.T) {
	h := Header{Name: "预算卡数", Type: "int", GTE: float64Ptr(0), LTE: float64Ptr(1000)}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"json.Number valid integer", unmarshalJSONValue("500"), 0},
		{"json.Number zero", unmarshalJSONValue("0"), 0},
		{"json.Number at upper bound", unmarshalJSONValue("1000"), 0},
		{"json.Number above upper bound", unmarshalJSONValue("1500"), 1},
		{"json.Number below lower bound", unmarshalJSONValue("-1"), 1},
		{"json.Number decimal (not integer)", unmarshalJSONValue("12.5"), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_JSONNumber_Float(t *testing.T) {
	h := Header{Name: "实际使用TPM", Type: "float", GTE: float64Ptr(0), LTE: float64Ptr(99.9)}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"json.Number valid float", unmarshalJSONValue("50.5"), 0},
		{"json.Number zero", unmarshalJSONValue("0"), 0},
		{"json.Number at upper bound", unmarshalJSONValue("99.9"), 0},
		{"json.Number above upper bound", unmarshalJSONValue("100"), 1},
		{"json.Number negative", unmarshalJSONValue("-0.1"), 1},
		{"json.Number integer as float", unmarshalJSONValue("42"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_JSONNumber_EnumNumeric(t *testing.T) {
	enumValues := unmarshalJSONSlice("[2026, 2027, 2028]")
	h := Header{Name: "年份", Type: "enum", Value: enumValues}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"json.Number match", unmarshalJSONValue("2026"), 0},
		{"json.Number no match", unmarshalJSONValue("2025"), 1},
		{"float64 match against json.Number enum", float64(2027), 0},
		{"float64 no match against json.Number enum", float64(2025), 1},
		{"string match against json.Number enum", "2028", 0},
		{"string no match against json.Number enum", "2025", 1},
		{"non-numeric string", "abc", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateTypedValue_JSONNumber_Required(t *testing.T) {
	h := Header{Name: "数量", Type: "int", Required: true}

	tests := []struct {
		name    string
		val     interface{}
		wantLen int
	}{
		{"json.Number satisfies required", unmarshalJSONValue("1"), 0},
		{"nil fails required", nil, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateTypedValue(tt.val, h)
			if len(errs) != tt.wantLen {
				t.Errorf("got %d errors, want %d; errs=%v", len(errs), tt.wantLen, errs)
			}
		})
	}
}

func TestValidateExtension_JSONNumber(t *testing.T) {
	enumValues := unmarshalJSONSlice("[1, 2, 3]")
	headers := []Header{
		{Name: "月份", Type: "enum", Value: enumValues, Required: true},
		{Name: "数量", Type: "int", GTE: float64Ptr(0), Required: true},
		{Name: "使用率", Type: "float", GTE: float64Ptr(0), LTE: float64Ptr(100)},
	}
	values := unmarshalJSONSlice("[2, 500, 85.5]")

	errs := ValidateExtension(values, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateFixedFields_JSONNumber(t *testing.T) {
	headers := []Header{
		{Name: "年份", Type: "int", DBField: "demand_year", GTE: float64Ptr(2020), LTE: float64Ptr(2030)},
		{Name: "月份", Type: "int", DBField: "demand_month", GTE: float64Ptr(1), LTE: float64Ptr(12)},
		{Name: "卡数", Type: "int", DBField: "gpu_num", GTE: float64Ptr(0)},
	}

	var fieldMap map[string]interface{}
	raw := `{"demand_year": 2026, "demand_month": 6, "gpu_num": 100}`
	if err := toolsjson.Unmarshal([]byte(raw), &fieldMap); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	errs := ValidateFixedFields(fieldMap, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
}

func TestValidateFixedFields_JSONNumber_OutOfRange(t *testing.T) {
	headers := []Header{
		{Name: "月份", Type: "int", DBField: "demand_month", GTE: float64Ptr(1), LTE: float64Ptr(12)},
	}

	var fieldMap map[string]interface{}
	raw := `{"demand_month": 13}`
	if err := toolsjson.Unmarshal([]byte(raw), &fieldMap); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	errs := ValidateFixedFields(fieldMap, headers)
	if len(errs) != 1 {
		t.Errorf("expected 1 error, got %d; errs=%v", len(errs), errs)
	}
}
