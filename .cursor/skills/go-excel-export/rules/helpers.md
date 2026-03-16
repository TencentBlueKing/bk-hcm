# 辅助函数与工具

## 常用工具函数

| 函数 | 位置 | 用途 |
|------|------|------|
| `cellRenderPtr[T]` | `helper.go:59` | 处理指针类型字段，nil 返回空字符串 |
| `mergeCell` | `helper.go:70` | 合并单元格 |
| `setStreamRow` | `helper.go:105` | StreamWriter 写入行数据的封装 |
| `getOrgHeader` | `helper.go:89` | 根据视角和聚合级别获取组织表头 |
| `addDownExcelHTTPHeader` | `helper.go:36` | 设置文件下载 HTTP 响应头 |

## 函数详解

### cellRenderPtr - 空指针值处理

```go
func cellRenderPtr[T any](ptr *T) interface{} {
	if ptr == nil {
		return ""
	}
	return *ptr
}
```

### mergeCell - 单元格合并

```go
func mergeCell(kt *kit.Kit, file *excelize.File, sheetName string,
	startCol, startRow, endCol, endRow int,
) error {
	start, err := excelize.CoordinatesToCellName(startCol, startRow)
	if err != nil {
		logs.Errorf("start coordinates to cell failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	end, err := excelize.CoordinatesToCellName(endCol, endRow)
	if err != nil {
		logs.Errorf("end coordinates to cell failed, err: %v, rid: %s", err, kt.Rid)
		return err
	}

	return file.MergeCell(sheetName, start, end)
}
```

### addDownExcelHTTPHeader - HTTP 响应头设置

```go
func addDownExcelHTTPHeader(writer http.ResponseWriter, name string) {
	contentType := ""

	switch {
	case strings.HasSuffix(name, ".xls"):
		contentType = "application/vnd.ms-excel"
	case strings.HasSuffix(name, ".xlsx"):
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	default:
		contentType = "application/octet-stream"
	}

	writer.Header().Set("Content-Type", contentType)

	// 支持中文文件名
	encodedName := url.QueryEscape(name)
	writer.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s\"; filename*=utf-8''%s", encodedName, encodedName))
	writer.Header().Set("Accept-Ranges", "bytes")
	writer.Header().Set("Cache-Control", "must-revalidate, post-check=0, pre-check=0")
	writer.Header().Set("Pragma", "no-cache")
	writer.Header().Set("Expires", "0")
}
```

## 文件名生成

### 规范

- 文件名格式：`{业务标识}_{时间戳}.xlsx`
- 时间戳格式：`20060102150405`

### 简单文件名

```go
func (e *Handler) BuildFileName() (string, error) {
	fileName := fmt.Sprintf("%s_%s.xlsx", 
		orgObsCostTrendWithResFileName, 
		time.Now().Format("20060102150405"))
	return fileName, nil
}
```

### 带日期范围的文件名

```go
func (e *Handler) BuildFileName() (string, error) {
	dateRangeStr := strings.ReplaceAll(e.req.DateRange.String(), " ", "")
	comparisonDateRangeStr := strings.ReplaceAll(e.req.Comparison.DateRange.String(), " ", "")
	fileName := fmt.Sprintf("%s_%svs_%s_%s.xlsx", 
		orgObsCostDistributionFileName, 
		dateRangeStr,
		comparisonDateRangeStr, 
		time.Now().Format("20060102150405"))
	return fileName, nil
}
```

## 样式设置

### 表头样式

```go
func getHeaderStyle(file *excelize.File) (int, error) {
	return file.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"F0F0F0"},
			Pattern: 1,
		},
		Font: &excelize.Font{
			Bold: true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
}
```

### 居中加粗样式

```go
style, err := file.NewStyle(&excelize.Style{
	Alignment: &excelize.Alignment{
		Horizontal: "center", 
		Vertical: "center",
	},
	Font: &excelize.Font{Bold: true},
})
```

## excelize 常用函数

| 函数 | 用途 |
|------|------|
| `excelize.CoordinatesToCellName(col, row)` | 坐标转单元格名称，如 (1,2) → "A2" |
| `file.SetSheetRow(sheet, cell, &data)` | 写入一行数据 |
| `file.SetCellValue(sheet, cell, value)` | 写入单个单元格 |
| `file.MergeCell(sheet, start, end)` | 合并单元格 |
| `file.NewStyle(&excelize.Style{})` | 创建样式 |
| `file.SetCellStyle(sheet, start, end, styleID)` | 应用样式 |
| `writer.SetRow(cell, values)` | StreamWriter 写入行 |

## 常量

```go
const DefaultSheetName = "Sheet1"
```

## 错误处理

统一使用 `errf.NewFromErr` 封装错误：

```go
// 参数验证失败
return errf.NewFromErr(errf.InvalidParameter, err)

// 操作被中止
return errf.NewFromErr(errf.Aborted, err)
```
