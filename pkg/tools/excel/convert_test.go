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
	"testing"
)

func TestConvertCellValue_EmptyAndWhitespace(t *testing.T) {
	h := Header{Name: "数量", Type: "int"}

	tests := []struct {
		name string
		val  string
		want interface{}
	}{
		{"empty string", "", ""},
		{"whitespace only", "   ", "   "},
		{"tab", "\t", "\t"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCellValue(tt.val, h)
			if got != tt.want {
				t.Errorf("ConvertCellValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertCellValue_Int(t *testing.T) {
	h := Header{Name: "数量", Type: "int"}

	tests := []struct {
		name string
		val  string
		want interface{}
	}{
		{"valid positive", "100", int64(100)},
		{"valid negative", "-5", int64(-5)},
		{"zero", "0", int64(0)},
		{"decimal fallback", "12.5", "12.5"},
		{"non-numeric fallback", "abc", "abc"},
		{"leading space trimmed", " 42 ", int64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCellValue(tt.val, h)
			if got != tt.want {
				t.Errorf("ConvertCellValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertCellValue_Float(t *testing.T) {
	h := Header{Name: "权重", Type: "float"}

	tests := []struct {
		name string
		val  string
		want interface{}
	}{
		{"valid decimal", "12.5", float64(12.5)},
		{"integer as float", "100", float64(100)},
		{"negative", "-3.14", float64(-3.14)},
		{"non-numeric fallback", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCellValue(tt.val, h)
			if got != tt.want {
				t.Errorf("ConvertCellValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertCellValue_FloatPrefix(t *testing.T) {
	h := Header{Name: "TPM", Type: "float64"}

	got := ConvertCellValue("3.14", h)
	want := float64(3.14)
	if got != want {
		t.Errorf("ConvertCellValue with type 'float64' = %v(%T), want %v(%T)",
			got, got, want, want)
	}
}

func TestConvertCellValue_EnumString(t *testing.T) {
	h := Header{Name: "卡型", Type: "enum", Value: []interface{}{"H20", "L20"}}

	tests := []struct {
		name string
		val  string
		want interface{}
	}{
		{"match first", "H20", "H20"},
		{"match second", "L20", "L20"},
		{"no match passthrough", "A100", "A100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCellValue(tt.val, h)
			if got != tt.want {
				t.Errorf("ConvertCellValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertCellValue_EnumNumeric(t *testing.T) {
	h := Header{
		Name:  "年份",
		Type:  "enum",
		Value: []interface{}{float64(2026), float64(2027)},
	}

	tests := []struct {
		name string
		val  string
		want interface{}
	}{
		{"integer string to int64", "2026", int64(2026)},
		{"float string to float64", "12.5", float64(12.5)},
		{"non-numeric passthrough", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCellValue(tt.val, h)
			if got != tt.want {
				t.Errorf("ConvertCellValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestConvertCellValue_EnumEmptyValues(t *testing.T) {
	h := Header{Name: "空枚举", Type: "enum", Value: []interface{}{}}
	got := ConvertCellValue("hello", h)
	if got != "hello" {
		t.Errorf("got %v(%T), want %q", got, got, "hello")
	}
}

func TestConvertCellValue_StringAndUnknown(t *testing.T) {
	tests := []struct {
		name     string
		header   Header
		val      string
		want     interface{}
	}{
		{"string type passthrough", Header{Name: "备注", Type: "string"}, "hello", "hello"},
		{"unknown type passthrough", Header{Name: "未知", Type: "custom"}, "world", "world"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCellValue(tt.val, tt.header)
			if got != tt.want {
				t.Errorf("ConvertCellValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

// --- convertEnumValue tests ---

func Test_convertEnumValue_EmptyValues(t *testing.T) {
	got := convertEnumValue("anything", nil)
	if got != "anything" {
		t.Errorf("got %v, want %q", got, "anything")
	}

	got = convertEnumValue("test", []interface{}{})
	if got != "test" {
		t.Errorf("got %v, want %q", got, "test")
	}
}

func Test_convertEnumValue_NumericEnum(t *testing.T) {
	enumVals := []interface{}{float64(2026), float64(2027)}

	tests := []struct {
		name string
		val  string
		want interface{}
	}{
		{"integer string", "2026", int64(2026)},
		{"float string", "12.5", float64(12.5)},
		{"non-numeric", "abc", "abc"},
		{"negative int", "-1", int64(-1)},
		{"zero", "0", int64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertEnumValue(tt.val, enumVals)
			if got != tt.want {
				t.Errorf("convertEnumValue(%q) = %v(%T), want %v(%T)",
					tt.val, got, got, tt.want, tt.want)
			}
		})
	}
}

func Test_convertEnumValue_StringEnum(t *testing.T) {
	enumVals := []interface{}{"H20", "L20"}

	got := convertEnumValue("A100", enumVals)
	if got != "A100" {
		t.Errorf("got %v, want %q", got, "A100")
	}

	got = convertEnumValue("H20", enumVals)
	if got != "H20" {
		t.Errorf("got %v, want %q", got, "H20")
	}
}
