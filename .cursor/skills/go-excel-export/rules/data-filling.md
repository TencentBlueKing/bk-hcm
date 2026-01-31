# 表头与数据填充

## 表头定义

使用 `[]interface{}` 类型定义表头，写入到第一行。

### 静态表头

```go
func (e *OrgObsCostTrendExporterHandler) AddHeader(kt *kit.Kit, file *excelize.File) error {
	headers := []interface{}{"日期", "成本（元）"}
	if err := file.SetSheetRow(webset.DefaultSheetName, "A1", &headers); err != nil {
		logs.Errorf("set header row failed, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}
	return nil
}
```

### 动态表头（根据时间范围生成）

```go
func (e *OrgObsPlatformCostTrendDistributionExporterHandler) AddHeader(kt *kit.Kit, file *excelize.File) error {
	headers := []interface{}{"成本科目", "合计成本"}

	if e.req.DateRange != nil {
		dates, err := e.req.DateRange.RangedMonths()
		if err != nil {
			logs.Errorf("get ranged months failed, err: %v, rid: %s", err, kt.Rid)
			return errf.NewFromErr(errf.InvalidParameter, err)
		}

		for _, date := range dates {
			headers = append(headers, date.Format(constant.YearMonthLayout))
		}
	}

	if err := file.SetSheetRow(webset.DefaultSheetName, "A1", &headers); err != nil {
		logs.Errorf("set sheet row failed, err: %v, rid: %s", err, kt.Rid)
		return errf.NewFromErr(errf.Aborted, err)
	}
	return nil
}
```

### 条件表头（根据视角动态调整）

```go
func (e *OrgDeviceUsageDetailStreamExporterHandler) AddStreamHeader(kt *kit.Kit,
	writers map[string]*excelize.StreamWriter,
) error {
	var header []interface{}
	if e.req.View == enumor.Fin {
		// 财务视角包含二级管理产品字段
		header = []interface{}{
			"固资号", "IP", "设备类型", "部门名称", "中心名称", "小组",
			"业务ID", "业务名称", "业务Set", "运营产品ID", "运营产品名称",
			"二级管理产品ID", "二级管理产品名称", "环境编码", "环境名称",
			// ...
		}
	} else {
		// 其他视角不包含二级管理产品字段
		header = []interface{}{
			"固资号", "IP", "设备类型", "部门名称", "中心名称", "小组",
			"业务ID", "业务名称", "业务Set", "运营产品ID", "运营产品名称",
			"环境编码", "环境名称",
			// ...
		}
	}
	// ...
}
```

## 数据行填充

### 关键规则

1. **数据行从第 2 行开始**（第 1 行为表头）
2. 使用 `excelize.CoordinatesToCellName` 转换坐标
3. 使用 `SetSheetRow`（Batch）或 `StreamWriter.SetRow`（Stream）写入

### Batch 模式填充

```go
func (e *OrgObsCostTrendExporterHandler) AddRows(kt *kit.Kit, file *excelize.File,
	items []anset.ObsBillCostTrendElement) error {

	if len(items) == 0 {
		return nil
	}

	for rowIndex, item := range items {
		actualRow := rowIndex + 2  // 数据从第2行开始

		rowData := []interface{}{
			item.Date.Format(dateFormat),
			cellRenderPtr(item.Cost),
		}

		rowCell, err := excelize.CoordinatesToCellName(1, actualRow)
		if err != nil {
			logs.Errorf("excel coordinates to cell failed, err: %v, rid: %s", err, kt.Rid)
			return errf.NewFromErr(errf.Aborted, err)
		}

		if err := file.SetSheetRow(webset.DefaultSheetName, rowCell, &rowData); err != nil {
			logs.Errorf("set data row failed, err: %v, rid: %s", err, kt.Rid)
			return errf.NewFromErr(errf.Aborted, err)
		}
	}
	return nil
}
```

### Stream 模式填充

```go
func (e *OrgObsBillStreamExporterHandler) AddStreamRows(kt *kit.Kit,
	writers map[string]*excelize.StreamWriter,
	items []webset.GetOrgObsBillDetailElem, sheetName string,
) error {
	// 获取当前行号（需要锁保护）
	e.sheetRowMutex.Lock()
	currentRow, exists := e.sheetRowMap[sheetName]
	e.sheetRowMutex.Unlock()

	if !exists {
		return errf.NewFromErr(errf.Aborted, fmt.Errorf("sheet %s not found in row map", sheetName))
	}

	for _, item := range items {
		values := []interface{}{
			item.FeeDate,
			item.DeptID,
			// ... 其他字段
		}

		cell, err := excelize.CoordinatesToCellName(1, currentRow)
		if err != nil {
			logs.Errorf("excel coordinates to cell failed, err: %v, rid: %s", err, kt.Rid)
			return errf.NewFromErr(errf.Aborted, err)
		}

		if err := setStreamRow(kt, writers, sheetName, cell, values); err != nil {
			return err
		}
		currentRow++
	}

	// 更新行号（需要锁保护）
	e.sheetRowMutex.Lock()
	e.sheetRowMap[sheetName] = currentRow
	e.sheetRowMutex.Unlock()
	return nil
}
```

## 空指针值处理

使用 `cellRenderPtr` 函数处理可能为 nil 的指针字段。

```go
func cellRenderPtr[T any](ptr *T) interface{} {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// 使用示例
values := []interface{}{
	item.Name,
	item.Cost,
	cellRenderPtr(item.ComparisonCost),  // 可能为 nil
	cellRenderPtr(item.ComparisonRate),  // 可能为 nil
}
```

## 错误示例

```go
// ❌ 使用 []string 而非 []interface{}
headers := []string{"日期", "成本"}

// ❌ 直接硬编码单元格
file.SetCellValue(sheetName, "A1", "日期")
file.SetCellValue(sheetName, "B1", "成本")

// ❌ 行号计算错误
for rowIndex, item := range items {
	actualRow := rowIndex + 1  // 会覆盖表头
}

// ❌ 未使用 CoordinatesToCellName
cell := fmt.Sprintf("A%d", rowIndex)

// ❌ 直接使用指针，可能 panic
*item.ComparisonCost  // 如果为 nil 会 panic
```
