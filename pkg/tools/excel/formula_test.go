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
	"bytes"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// buildExcelWithFormulas 构建含公式单元格的 Excel 文件，formulas 为 cellRef→formula 的映射。
func buildExcelWithFormulas(sheetName string, formulas map[string]string) *excelize.File {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)
	for cell, formula := range formulas {
		f.SetCellFormula(sheetName, cell, formula)
	}
	buf, _ := f.WriteToBuffer()
	result, _ := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	return result
}

// --- TestReplaceFormulaRowRefs ---

func TestReplaceFormulaRowRefs_RelativeRefs(t *testing.T) {
	tests := []struct {
		formula string
		to      int
		want    string
	}{
		{"A2+B2", 5, "A5+B5"},
		{"SUM(A3:B3)", 7, "SUM(A7:B7)"},
		{"ROUNDUP(M4*1000000000/Q4/3600/P4,0)", 6, "ROUNDUP(M6*1000000000/Q6/3600/P6,0)"},
		{"IF(A2>2,1,0)", 3, "IF(A3>2,1,0)"},
		{"A8+B8", 5, "A5+B5"},
		{"", 3, ""},
		{"100+200", 3, "100+200"},
	}

	for _, tt := range tests {
		got := replaceFormulaRowRefs(tt.formula, tt.to)
		if got != tt.want {
			t.Errorf("replaceFormulaRowRefs(%q, %d) = %q, want %q",
				tt.formula, tt.to, got, tt.want)
		}
	}
}

func TestReplaceFormulaRowRefs_AbsoluteRefs(t *testing.T) {
	tests := []struct {
		formula string
		to      int
		want    string
	}{
		{"A$2+B2", 3, "A$2+B3"},
		{"$A$2+B2", 3, "$A$2+B3"},
		{"$A2+B2", 3, "$A3+B3"},
	}

	for _, tt := range tests {
		got := replaceFormulaRowRefs(tt.formula, tt.to)
		if got != tt.want {
			t.Errorf("replaceFormulaRowRefs(%q, %d) = %q, want %q",
				tt.formula, tt.to, got, tt.want)
		}
	}
}

// buildExcelWithData 构建含值和公式的 Excel 文件。
func buildExcelWithData(
	sheetName string, values map[string]interface{}, formulas map[string]string,
) *excelize.File {
	f := excelize.NewFile()
	f.SetSheetName("Sheet1", sheetName)
	for cell, val := range values {
		_ = f.SetCellValue(sheetName, cell, val)
	}
	for cell, formula := range formulas {
		_ = f.SetCellFormula(sheetName, cell, formula)
	}
	buf, _ := f.WriteToBuffer()
	result, _ := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	return result
}

// --- TestValidateAndFillFormulas ---

func TestValidateAndFillFormulas_NoHeaders(t *testing.T) {
	f := buildExcelWithFormulas("Sheet1", nil)
	defer f.Close()

	errs := validateAndFillFormulas(f, "Sheet1", 2, nil, nil)
	if len(errs) != 0 {
		t.Errorf("expected no errors for empty headers, got %v", errs)
	}
}

func TestValidateAndFillFormulas_CorrectFormula(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{"C2": "A2+B2"})
	defer f.Close()

	headers := []Header{{Name: "合计", Field: "C", Formula: "A2+B2"}}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors for matching formula, got %v", errs)
	}
}

func TestValidateAndFillFormulas_RowAdjusted(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{"C4": "A4+B4"})
	defer f.Close()

	headers := []Header{{Name: "合计", Field: "C", Formula: "A2+B2"}}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 4, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors for correctly adjusted formula, got %v", errs)
	}
}

func TestValidateAndFillFormulas_SchemaRowDiffersFromRowStart(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{
		"H6": "ROUNDUP(M6*1000000000/Q6/3600/P6,0)",
	})
	defer f.Close()

	headers := []Header{{
		Name: "预算卡数", Field: "H",
		Formula: "ROUNDUP(M4*1000000000/Q4/3600/P4,0)",
	}}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 6, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors when schema formula row differs, got %v", errs)
	}
}

func TestValidateAndFillFormulas_Mismatch(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{"C2": "A2*B2"})
	defer f.Close()

	headers := []Header{{Name: "合计", Field: "C", Formula: "A2+B2"}}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "合计") {
		t.Errorf("error should mention column name '合计', got: %q", errs[0])
	}
}

func TestValidateAndFillFormulas_DeletedFormula(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, nil)
	defer f.Close()

	headers := []Header{{Name: "合计", Field: "C", Formula: "A2+B2"}}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for deleted formula, got %d: %v", len(errs), errs)
	}
}

func TestValidateAndFillFormulas_MultipleHeaders(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{
		"C2": "A2+B2",
		"D2": "A2*B2",
	})
	defer f.Close()

	headers := []Header{
		{Name: "合计", Field: "C", Formula: "A2+B2"},
		{Name: "乘积", Field: "D", Formula: "A2+B2"},
	}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 1 {
		t.Fatalf("expected 1 error (D2 mismatch), got %d: %v", len(errs), errs)
	}
	if !strings.Contains(errs[0], "乘积") {
		t.Errorf("error should mention '乘积', got: %q", errs[0])
	}
}

func TestValidateAndFillFormulas_CaseInsensitive(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{"C2": "sum(a2,b2)"})
	defer f.Close()

	headers := []Header{{Name: "合计", Field: "C", Formula: "SUM(A2,B2)"}}
	row := make([]string, len(headers))
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors (case insensitive match), got %v", errs)
	}
}

func TestValidateAndFillFormulas_SkipsNonFormulaHeaders(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithFormulas(sheet, map[string]string{"C2": "A2+B2"})
	defer f.Close()

	headers := []Header{
		{Name: "X", Field: "A"},
		{Name: "Y", Field: "B"},
		{Name: "合计", Field: "C", Formula: "A2+B2"},
	}
	row := []string{"1", "2", ""}
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if row[0] != "1" || row[1] != "2" {
		t.Errorf("non-formula columns should be unchanged, got %v", row)
	}
}

// --- TestValidateAndFillFormulas fill behavior ---

func TestValidateAndFillFormulas_FillMissingValue(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithData(sheet,
		map[string]interface{}{"A2": 10, "B2": 20},
		map[string]string{"C2": "A2+B2"},
	)
	defer f.Close()

	headers := []Header{
		{Name: "X", Field: "A", Type: "int"},
		{Name: "Y", Field: "B", Type: "int"},
		{Name: "合计", Field: "C", Formula: "A2+B2"},
	}
	row := []string{"10", "20", ""}
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if row[2] != "30" {
		t.Errorf("expected C2 filled with '30', got %q", row[2])
	}
}

func TestValidateAndFillFormulas_NoFillWhenMismatch(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithData(sheet,
		map[string]interface{}{"A2": 10, "B2": 20},
		map[string]string{"C2": "A2*B2"},
	)
	defer f.Close()

	headers := []Header{
		{Name: "X", Field: "A", Type: "int"},
		{Name: "Y", Field: "B", Type: "int"},
		{Name: "合计", Field: "C", Formula: "A2+B2"},
	}
	row := []string{"10", "20", ""}
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 1 {
		t.Fatalf("expected 1 formula error, got %d: %v", len(errs), errs)
	}
	if row[2] != "" {
		t.Errorf("mismatched formula should not be filled, got %q", row[2])
	}
}

func TestValidateAndFillFormulas_NoFillWhenCached(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithData(sheet,
		map[string]interface{}{"A2": 10, "B2": 20, "C2": 99},
		map[string]string{"C2": "A2+B2"},
	)
	defer f.Close()

	headers := []Header{
		{Name: "X", Field: "A", Type: "int"},
		{Name: "Y", Field: "B", Type: "int"},
		{Name: "合计", Field: "C", Formula: "A2+B2"},
	}
	row := []string{"10", "20", "99"}
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if row[2] != "99" {
		t.Errorf("cached value should be preserved, got %q", row[2])
	}
}

func TestValidateAndFillFormulas_FillRoundUp(t *testing.T) {
	const sheet = "Sheet1"
	f := buildExcelWithData(sheet,
		map[string]interface{}{"M2": 7, "Q2": 1750, "P2": 10},
		map[string]string{"C2": "ROUNDUP(M2*1000000000/Q2/3600/P2,0)"},
	)
	defer f.Close()

	headers := []Header{
		{Name: "参数量", Field: "M", Type: "float"},
		{Name: "速度", Field: "Q", Type: "float"},
		{Name: "时长", Field: "P", Type: "float"},
		{Name: "预算卡数", Field: "C", Formula: "ROUNDUP(M2*1000000000/Q2/3600/P2,0)"},
	}
	row := []string{"7", "1750", "10", ""}
	errs := validateAndFillFormulas(f, sheet, 2, row, headers)
	if len(errs) != 0 {
		t.Errorf("expected no errors, got %v", errs)
	}
	if row[3] == "" {
		t.Error("expected ROUNDUP formula to be calculated, got empty")
	}
}
