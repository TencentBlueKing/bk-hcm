# 导出模式选择

## 模式对比

| 特性 | Batch 模式 | Stream 模式 |
|-----|-----------|------------|
| 适用数据量 | <10万行 | 无限制 |
| 内存占用 | 一次性加载 | 分批加载 |
| 代码复杂度 | 简单 | 较复杂 |
| 接口实现 | `BatchExportHandler` | `StreamExportHandler` |

## 接口定义

### Batch 模式接口

```go
type BatchExportHandler[Item any] interface {
	BaseExportHandler[Item]
	GetItems(kt *kit.Kit) ([]Item, error)
	AddHeader(kt *kit.Kit, file *excelize.File) error
	AddRows(kt *kit.Kit, file *excelize.File, items []Item) error
}

// 多 Sheet 版本
type BatchExportHandlerMultiSheet[Item any] interface {
	BaseExportHandler[Item]
	GetItems(kt *kit.Kit) (map[string][]Item, error)  // sheetName -> items
	AddHeader(kt *kit.Kit, file *excelize.File, sheetName string) error
	AddRows(kt *kit.Kit, file *excelize.File, items []Item, sheetName string) error
}
```

### Stream 模式接口

```go
type StreamExportHandler[Item any] interface {
	BaseExportHandler[Item]
	GetItemsStream(kt *kit.Kit) ([]Item, bool, error)  // items, hasMore, error
	AddStreamHeader(kt *kit.Kit, writers map[string]*excelize.StreamWriter) error
	AddStreamRows(kt *kit.Kit, writers map[string]*excelize.StreamWriter, items []Item) error
}

// 多 Sheet 版本
type StreamExportHandlerMultiSheet[Item any] interface {
	BaseExportHandler[Item]
	GetItemsStream(kt *kit.Kit) ([]Item, bool, string, error)  // items, hasMore, sheetName, error
	AddStreamHeader(kt *kit.Kit, writers map[string]*excelize.StreamWriter) error
	AddStreamRows(kt *kit.Kit, writers map[string]*excelize.StreamWriter, items []Item, sheetName string) error
}
```

## 使用示例

### Batch 模式（默认）

```go
return BaseExporter[anset.ObsBillCostTrendElement]{
	handler: &OrgObsCostTrendExporterHandler{
		cs:  cs,
		req: req,
	},
	// mode 默认为 ExportModeBatch
}
```

### Stream 模式

```go
return BaseExporter[webset.GetOrgObsBillDetailElem]{
	handler: &OrgObsBillStreamExporterHandler{
		cs:            cs,
		req:           req,
		currentOffset: 0,
		batchSize:     constant.ObsListQueryLimit,
		sheetRowMap:   make(map[string]int),
	},
	mode:         ExportModeStream,  // 显式设置 Stream 模式
	isMultiSheet: true,              // 多 Sheet 导出
}
```

## 多 Sheet 导出

当需要按维度拆分到多个 Sheet 时，设置 `isMultiSheet: true`。

### GetItems 返回格式差异

**单 Sheet**：
```go
func (e *Handler) GetItems(kt *kit.Kit) ([]Item, error)
```

**多 Sheet**：
```go
func (e *Handler) GetItems(kt *kit.Kit) (map[string][]Item, error) {
	result := make(map[string][]Item)
	for _, dimension := range dimensions {
		result[dimension.Name()] = items
	}
	return result, nil
}
```

## 错误示例

```go
// ❌ 大数据量使用 Batch 模式会导致内存问题
return BaseExporter[LargeDataElement]{
	handler: &LargeDataHandler{},
	// 应该设置 mode: ExportModeStream
}

// ❌ Stream 模式忘记初始化 sheetRowMap
return BaseExporter[Item]{
	handler: &StreamHandler{
		// 缺少 sheetRowMap: make(map[string]int)
	},
	mode: ExportModeStream,
}
```
