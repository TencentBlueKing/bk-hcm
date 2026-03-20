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

// Package excel defines some tools like parsor validator for excel file operation.
package excel

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"hcm/pkg/criteria/constant"
	"hcm/pkg/logs"

	"github.com/xuri/excelize/v2"
)

// ValidateFileIntegrity 校验Excel文件与Schema的完整性，依次执行sheet匹配校验和列头匹配校验。
// 列头行取每个sheet的head_row字段指定的行。
func ValidateFileIntegrity(excelFile *excelize.File, schema *Schema) error {
	if err := validateSheets(excelFile, schema); err != nil {
		return err
	}

	if err := validateHeaders(excelFile, schema); err != nil {
		return err
	}

	return nil
}

// validateSheets 校验Excel文件的sheet名称和数量是否与Schema一致
func validateSheets(excelFile *excelize.File, schema *Schema) error {
	excelSheets := excelFile.GetSheetList()
	excelSheetSet := make(map[string]struct{}, len(excelSheets))
	for _, name := range excelSheets {
		excelSheetSet[name] = struct{}{}
	}

	var missing, extra []string
	for _, s := range schema.Sheets {
		if _, ok := excelSheetSet[s.Name]; !ok {
			missing = append(missing, s.Name)
		}
	}
	for _, name := range excelSheets {
		visible, err := excelFile.GetSheetVisible(name)
		if err != nil {
			return err
		}
		if !visible {
			logs.Warnf("sheet[%s] is hidden, skip validation", name)
			continue
		}

		if schema.FindSheet(name) == nil {
			extra = append(extra, name)
		}
	}

	if len(missing) > 0 || len(extra) > 0 {
		var parts []string
		if len(missing) > 0 {
			parts = append(parts, fmt.Sprintf("missing sheets: %s", strings.Join(missing, ", ")))
		}
		if len(extra) > 0 {
			parts = append(parts, fmt.Sprintf("extra sheets: %s", strings.Join(extra, ", ")))
		}
		return fmt.Errorf("excel file sheet mismatch, %s", strings.Join(parts, "; "))
	}

	return nil
}

// validateHeaders 校验每个sheet的列头行是否与Schema中定义的fixed_headers和headers匹配。
// 仅校验可见列头（field != "-" 且 hidden != true）。
func validateHeaders(excelFile *excelize.File, schema *Schema) error {
	for _, sheet := range schema.Sheets {
		if sheet.HeadRow < 1 {
			return fmt.Errorf("sheet[%s] head_row must be >= 1, current value: %d",
				sheet.Name, sheet.HeadRow)
		}

		headerRow, err := readHeaderRow(excelFile, sheet.Name, sheet.HeadRow)
		if err != nil {
			return err
		}

		headerSet := make(map[string]struct{}, len(headerRow))
		for _, h := range headerRow {
			trimmed := strings.TrimSpace(h)
			if trimmed != "" {
				headerSet[trimmed] = struct{}{}
			}
		}

		var missingCols []string
		for _, h := range sheet.AllVisibleHeaders() {
			if _, ok := headerSet[h.Name]; !ok {
				missingCols = append(missingCols, h.Name)
			}
		}

		if len(missingCols) > 0 {
			return fmt.Errorf("sheet[%s] header mismatch, missing columns: %s",
				sheet.Name, strings.Join(missingCols, ", "))
		}
	}

	return nil
}

// readHeaderRow 读取指定sheet指定行号的列头数据
func readHeaderRow(excelFile *excelize.File, sheetName string, headerRowIdx int) ([]string, error) {
	rows, err := excelFile.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read sheet[%s]: %w", sheetName, err)
	}
	defer rows.Close()

	var headerRow []string
	rowNum := 0
	for rows.Next() {
		rowNum++
		if rowNum == headerRowIdx {
			headerRow, err = rows.Columns()
			if err != nil {
				return nil, fmt.Errorf("failed to read sheet[%s] row %d: %w",
					sheetName, headerRowIdx, err)
			}
			break
		}
	}

	if headerRow == nil {
		return nil, fmt.Errorf("sheet[%s] header row not found (row %d)", sheetName, headerRowIdx)
	}

	return headerRow, nil
}

// buildFieldIndices 将header的field列名（如A、B、C）转换为0-based列索引数组
func buildFieldIndices(sheetName string, headers []Header) ([]int, error) {
	indices := make([]int, 0, len(headers))
	for _, h := range headers {
		colNum, err := excelize.ColumnNameToNumber(h.Field)
		if err != nil {
			return nil, fmt.Errorf("sheet[%s] invalid field column name %q: %w", sheetName, h.Field, err)
		}
		indices = append(indices, colNum-1)
	}

	return indices, nil
}

// extractFieldValues 按列索引从原始行数据中提取指定列的值
func extractFieldValues(columns []string, fieldIndices []int) []string {
	row := make([]string, len(fieldIndices))
	for i, idx := range fieldIndices {
		if idx < len(columns) {
			row[i] = columns[idx]
		}
	}

	return row
}

// isEmptyRow 判断一行是否为空行。满足以下任一条件即视为空行：
//   - 所有单元格值均为空
//   - 所有非空单元格均为公式单元格（无用户输入数据，包括公式产生的错误值如 #DIV/0!）
func isEmptyRow(excelFile *excelize.File, sheetName string, rowNum int, columns []string) bool {
	for colIdx, val := range columns {
		if strings.TrimSpace(val) == "" {
			continue
		}
		colName, err := excelize.ColumnNumberToName(colIdx + 1)
		if err != nil {
			return false
		}

		cellRef := colName + strconv.Itoa(rowNum)
		formula, _ := excelFile.GetCellFormula(sheetName, cellRef)
		if formula == "" {
			return false
		}
	}

	return true
}

// cellRefRe 匹配 Excel 单元格引用（列1-3字母 + 可选$ + 行号），限制列为1-3个字母以避免
// 误匹配函数名或 Sheet 名（如 ROUNDUP、Sheet1）。group: (非字母前缀)(列名)($?)(行号)
var cellRefRe = regexp.MustCompile(`(?i)((?:^|[^A-Za-z$]))(\$?[A-Z]{1,3})(\$?)(\d{1,7})`)

// replaceFormulaRowRefs 将公式中所有相对行引用的行号替换为 toRow。
// 绝对行引用（行号前有 $）不做调整。用于 schema 公式中行号不固定的场景，
// 无需知道原始行号即可将所有单元格引用指向目标行。
func replaceFormulaRowRefs(formula string, toRow int) string {
	toStr := strconv.Itoa(toRow)

	return cellRefRe.ReplaceAllStringFunc(formula, func(match string) string {
		sub := cellRefRe.FindStringSubmatch(match)
		prefix, col, dollar := sub[1], sub[2], sub[3]
		if dollar == "$" {
			return match
		}
		return prefix + col + toStr
	})
}

// normalizeFormula 去除前导 '=' 并转大写，用于公式比较。
func normalizeFormula(f string) string {
	return strings.TrimPrefix(strings.ToUpper(strings.TrimSpace(f)), "=")
}

// validateAndFillFormulas 校验公式列正确性，并对通过校验但缓存值为空的单元格通过 CalcCellValue 补算。
// row 与 headers 按索引一一对应，补算结果直接写入 row，返回公式校验错误列表。
func validateAndFillFormulas(excelFile *excelize.File, sheetName string, rowNum int, row []string, headers []Header,
) []string {
	var errs []string
	for i, h := range headers {
		if h.Formula == "" {
			continue
		}

		cellRef := h.Field + strconv.Itoa(rowNum)
		rawActual, _ := excelFile.GetCellFormula(sheetName, cellRef)
		actual := normalizeFormula(rawActual)
		expected := replaceFormulaRowRefs(normalizeFormula(h.Formula), rowNum)

		if actual != expected {
			errs = append(errs,
				fmt.Sprintf(constant.ExcelValidateFormulaMismatch, h.Name, expected, actual))
			continue
		}

		// 如果缓存值为空，则通过 CalcCellValue 补算，补算失败时不做特殊处理，值仍为空
		if i < len(row) && strings.TrimSpace(row[i]) == "" {
			val, calcErr := excelFile.CalcCellValue(sheetName, cellRef)
			if calcErr != nil {
				logs.Warnf("failed to calc cell value: %v, sheet: %s, cell: %s, err: %v", calcErr, sheetName,
					cellRef, calcErr)
				continue
			}

			// 如果公式计算出现错误，则不进行补算，值仍为空 #为excel单元格格式错误标识
			if val != "" && !strings.HasPrefix(val, "#") {
				row[i] = val
			}
		}
	}

	return errs
}

// ParseSheetRowsAndFormulas 在单次行遍历中完成数据解析、公式校验和缺失公式值补算。
// 返回的 rows 与 formulaErrs 按索引一一对应，每个元素分别为该行的字段值和公式校验错误。
func ParseSheetRowsAndFormulas(excelFile *excelize.File, sheet *Sheet) (
	rows [][]string, formulaErrs [][]string, err error,
) {
	if sheet.RowStart < 2 {
		return nil, nil, fmt.Errorf("sheet[%s] row_start must be >= 2, current value: %d",
			sheet.Name, sheet.RowStart)
	}

	excelHeaders := sheet.AllExcelHeaders()
	fieldIndices, err := buildFieldIndices(sheet.Name, excelHeaders)
	if err != nil {
		return nil, nil, err
	}

	sheetRows, err := excelFile.Rows(sheet.Name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read sheet[%s]: %w", sheet.Name, err)
	}
	defer sheetRows.Close()

	rowNum := 0
	for sheetRows.Next() {
		rowNum++
		if rowNum < sheet.RowStart {
			continue
		}

		columns, colErr := sheetRows.Columns()
		if colErr != nil {
			return nil, nil, fmt.Errorf("failed to read sheet[%s] row %d: %w", sheet.Name, rowNum, colErr)
		}

		if isEmptyRow(excelFile, sheet.Name, rowNum, columns) {
			continue
		}

		row := extractFieldValues(columns, fieldIndices)
		fErrs := validateAndFillFormulas(excelFile, sheet.Name, rowNum, row, excelHeaders)
		rows = append(rows, row)
		formulaErrs = append(formulaErrs, fErrs)
	}

	return rows, formulaErrs, nil
}
