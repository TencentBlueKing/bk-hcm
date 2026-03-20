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
	"strconv"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// --- test helpers ---

func newTestExcelFile(sheets map[string][][]string) *excelize.File {
	f := excelize.NewFile()

	first := true
	for name, rows := range sheets {
		if first {
			f.SetSheetName("Sheet1", name)
			first = false
		} else {
			f.NewSheet(name)
		}
		for rowIdx, row := range rows {
			for colIdx, val := range row {
				colName, _ := excelize.ColumnNumberToName(colIdx + 1)
				cell := colName + strconv.Itoa(rowIdx+1)
				f.SetCellValue(name, cell, val)
			}
		}
	}

	buf, _ := f.WriteToBuffer()
	result, _ := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	return result
}

type orderedSheet struct {
	name string
	rows [][]string
}

func newTestExcelFileOrdered(sheets []orderedSheet) *excelize.File {
	f := excelize.NewFile()

	for i, s := range sheets {
		if i == 0 {
			f.SetSheetName("Sheet1", s.name)
		} else {
			f.NewSheet(s.name)
		}
		for rowIdx, row := range s.rows {
			for colIdx, val := range row {
				colName, _ := excelize.ColumnNumberToName(colIdx + 1)
				cell := colName + strconv.Itoa(rowIdx+1)
				f.SetCellValue(s.name, cell, val)
			}
		}
	}

	buf, _ := f.WriteToBuffer()
	result, _ := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
	return result
}

func buildOneSheetExcel(name string, rows [][]string) func() *excelize.File {
	return func() *excelize.File {
		return newTestExcelFileOrdered([]orderedSheet{{name: name, rows: rows}})
	}
}

func buildTwoSheetExcel(
	name1 string, rows1 [][]string,
	name2 string, rows2 [][]string,
) func() *excelize.File {
	return func() *excelize.File {
		return newTestExcelFileOrdered([]orderedSheet{
			{name: name1, rows: rows1},
			{name: name2, rows: rows2},
		})
	}
}

// --- integrityTestCase & runner ---

type integrityTestCase struct {
	name      string
	buildFile func() *excelize.File
	schema    *Schema
	wantErr   bool
	errMsg    string
}

func runIntegrityTests(t *testing.T, tests []integrityTestCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			excelFile := tt.buildFile()
			defer excelFile.Close()

			err := ValidateFileIntegrity(excelFile, tt.schema)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Fatalf("expected error containing %q, got: %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
		})
	}
}

// --- TestValidateFileIntegrity: basic scenarios ---

func TestValidateFileIntegrity_ValidFile(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "valid_file_matches_schema",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"业务名称", "GPU型号", "数量"},
				{"biz1", "A100", "10"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					Headers: []Header{
						{Name: "GPU型号", Type: "string", Field: "B"},
						{Name: "数量", Type: "int", Field: "C"},
					},
				}},
			},
		},
	})
}

func TestValidateFileIntegrity_HeadRowValidation(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "head_row_less_than_1",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"业务名称"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 0, RowStart: 2,
					FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
				}},
			},
			wantErr: true,
			errMsg:  "head_row must be >= 1",
		},
		{
			name: "row_start_3_header_on_row_2",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"标题行"},
				{"业务名称", "GPU型号"},
				{"biz1", "A100"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 2, RowStart: 3,
					FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					Headers:      []Header{{Name: "GPU型号", Type: "string", Field: "B"}},
				}},
			},
		},
	})
}

// --- TestValidateFileIntegrity: sheet match scenarios ---

func TestValidateFileIntegrity_MissingSheet(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "missing_sheet",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"业务名称"},
			}),
			schema: &Schema{
				Sheets: []Sheet{
					{
						Name: "GPU需求", HeadRow: 1, RowStart: 2,
						FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					},
					{
						Name: "汇总", HeadRow: 1, RowStart: 2,
						FixedHeaders: []Header{{Name: "合计", Type: "string", Field: "A"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing sheets",
		},
	})
}

func TestValidateFileIntegrity_ExtraSheet(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "extra_sheet",
			buildFile: buildTwoSheetExcel(
				"GPU需求", [][]string{{"业务名称"}},
				"多余Sheet", [][]string{{"data"}},
			),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
				}},
			},
			wantErr: true,
			errMsg:  "extra sheets",
		},
	})
}

func TestValidateFileIntegrity_MultiSheet_Valid(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "multiple_sheets_all_valid",
			buildFile: buildTwoSheetExcel(
				"GPU需求", [][]string{{"业务名称"}, {"biz1"}},
				"汇总", [][]string{{"合计"}, {"100"}},
			),
			schema: &Schema{
				Sheets: []Sheet{
					{
						Name: "GPU需求", HeadRow: 1, RowStart: 2,
						FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					},
					{
						Name: "汇总", HeadRow: 1, RowStart: 2,
						FixedHeaders: []Header{{Name: "合计", Type: "string", Field: "A"}},
					},
				},
			},
		},
	})
}

func TestValidateFileIntegrity_MultiSheet_HeaderMissing(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "multiple_sheets_second_missing_header",
			buildFile: buildTwoSheetExcel(
				"GPU需求", [][]string{{"业务名称"}, {"biz1"}},
				"汇总", [][]string{{"其他列"}, {"xxx"}},
			),
			schema: &Schema{
				Sheets: []Sheet{
					{
						Name: "GPU需求", HeadRow: 1, RowStart: 2,
						FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					},
					{
						Name: "汇总", HeadRow: 1, RowStart: 2,
						FixedHeaders: []Header{{Name: "合计", Type: "string", Field: "A"}},
					},
				},
			},
			wantErr: true,
			errMsg:  "missing columns: 合计",
		},
	})
}

// --- TestValidateFileIntegrity: fixed header scenarios ---

func TestValidateFileIntegrity_FixedHeader_SkipDash(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "skip_field_dash_headers",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"业务名称", "GPU型号"},
				{"biz1", "A100"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{
						{Name: "业务名称", Type: "string", Field: "A"},
						{Name: "计算列", Type: "string", Field: "-"},
					},
					Headers: []Header{{Name: "GPU型号", Type: "string", Field: "B"}},
				}},
			},
		},
	})
}

func TestValidateFileIntegrity_FixedHeader_Hidden(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "hidden_fixed_header_not_in_excel_passes",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"年份", "月份", "使用场景"},
				{"2026", "9", "文生图"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{
						{Name: "年份", Type: "enum", Field: "A"},
						{Name: "月份", Type: "enum", Field: "B"},
						{Name: "预算卡数", Type: "int", Field: "C", Hidden: true},
					},
					Headers: []Header{{Name: "使用场景", Type: "string", Field: "D"}},
				}},
			},
		},
	})
}

func TestValidateFileIntegrity_FixedHeader_MissingVisible(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "missing_visible_fixed_header_fails",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"月份", "使用场景"},
				{"9", "文生图"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{
						{Name: "年份", Type: "enum", Field: "A"},
						{Name: "月份", Type: "enum", Field: "B"},
					},
					Headers: []Header{{Name: "使用场景", Type: "string", Field: "C"}},
				}},
			},
			wantErr: true,
			errMsg:  "missing columns: 年份",
		},
	})
}

func TestValidateFileIntegrity_FixedHeader_MixedVisibleHidden(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "mixed_visible_hidden_headers",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"年份", "月份", "使用场景", "卡型"},
				{"2026", "9", "文生图", "H20"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{
						{Name: "年份", Type: "enum", Field: "A"},
						{Name: "月份", Type: "enum", Field: "B"},
						{Name: "预算卡数", Type: "int", Field: "C", Hidden: true},
						{Name: "QPM峰值", Type: "int", Field: "-", Hidden: true},
					},
					Headers: []Header{
						{Name: "使用场景", Type: "string", Field: "D"},
						{Name: "卡型", Type: "enum", Field: "E"},
						{Name: "隐藏业务列", Type: "string", Field: "F", Hidden: true},
					},
				}},
			},
		},
	})
}

// --- TestValidateFileIntegrity: ext header scenarios ---

func TestValidateFileIntegrity_ExtHeader_Missing(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "missing_header_column",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"业务名称", "GPU型号"},
				{"biz1", "A100"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					Headers: []Header{
						{Name: "GPU型号", Type: "string", Field: "B"},
						{Name: "数量", Type: "int", Field: "C"},
					},
				}},
			},
			wantErr: true,
			errMsg:  "missing columns: 数量",
		},
	})
}

func TestValidateFileIntegrity_ExtHeader_HiddenSkipped(t *testing.T) {
	runIntegrityTests(t, []integrityTestCase{
		{
			name: "hidden_headers_skipped_in_validation",
			buildFile: buildOneSheetExcel("GPU需求", [][]string{
				{"业务名称"},
				{"biz1"},
			}),
			schema: &Schema{
				Sheets: []Sheet{{
					Name: "GPU需求", HeadRow: 1, RowStart: 2,
					FixedHeaders: []Header{{Name: "业务名称", Type: "string", Field: "A"}},
					Headers:      []Header{{Name: "隐藏列", Type: "string", Field: "B", Hidden: true}},
				}},
			},
		},
	})
}
