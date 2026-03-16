# Handler 定义与命名

## 命名规范

### Handler 结构体命名

| 类型 | 格式 | 示例 |
|-----|------|------|
| 公开 Handler | `{Domain}{Feature}ExporterHandler` | `OrgObsCostTrendExporterHandler` |
| 私有 Handler | `{domain}{Feature}ExporterHandler` | `svcHCDeviceAmountTrendExporterHandler` |

### 构造函数命名

格式：`New{Domain}{Feature}Exporter`

返回 `BaseExporter[T]` 类型。

> **注意**：历史代码中存在少量使用 `*Exporter` 后缀而非 `*ExporterHandler` 的情况（如 `OrgHmMetricDistExporter`），新代码应统一使用 `*ExporterHandler` 后缀。

## Handler 结构

Handler 结构体必须包含：
- `cs clientset.ClientSet` - 服务客户端集合
- `req *{RequestType}` - 请求参数

Stream 模式 Handler 还应包含：
- `currentOffset int` - 当前分页偏移
- `batchSize int` - 每批次大小
- `sheetRowMap map[string]int` - Sheet 当前行号映射
- `sheetRowMutex sync.Mutex` - 并发保护锁
- `xxxMap map[string]string` - 关联数据映射（可选）
- `mapOnce sync.Once` - 关联数据预加载控制（可选）

## 正确示例

### Batch 模式 Handler

```go
// OrgObsCostTrendExporterHandler is the handler for org OBS cost trend.
type OrgObsCostTrendExporterHandler struct {
	cs  clientset.ClientSet
	req *anset.ExportOrgCostTrendReq
	kt  *kit.Kit
}

// NewOrgObsCostTrendExporter creates a new org OBS cost trend exporter.
func NewOrgObsCostTrendExporter(cs clientset.ClientSet, req *anset.ExportOrgCostTrendReq,
) BaseExporter[anset.ObsBillCostTrendElement] {

	return BaseExporter[anset.ObsBillCostTrendElement]{
		handler: &OrgObsCostTrendExporterHandler{
			cs:  cs,
			req: req,
		},
	}
}
```

### Stream 模式 Handler

```go
type OrgObsBillStreamExporterHandler struct {
	cs            clientset.ClientSet
	req           *webset.ExportOrgObsBillDetailReq
	currentOffset int
	batchSize     int

	sheetRowMap   map[string]int
	sheetRowMutex sync.Mutex

	opProductMap map[string]string
	mapOnce      sync.Once
}

func NewOrgObsBillStreamExporter(cs clientset.ClientSet, req *webset.ExportOrgObsBillDetailReq,
) BaseExporter[webset.GetOrgObsBillDetailElem] {

	return BaseExporter[webset.GetOrgObsBillDetailElem]{
		handler: &OrgObsBillStreamExporterHandler{
			cs:            cs,
			req:           req,
			currentOffset: 0,
			batchSize:     constant.ObsListQueryLimit,
			sheetRowMap:   make(map[string]int),
		},
		mode:         ExportModeStream,
		isMultiSheet: true,
	}
}
```

## 错误示例

```go
// ❌ 命名不规范
type ObsCostExporter struct {  // 缺少 Handler 后缀、缺少领域前缀
	client clientset.ClientSet
}

// ❌ 构造函数命名不规范
func CreateExporter(cs clientset.ClientSet) BaseExporter[SomeType] {
	// ...
}

// ❌ 缺少必要字段
type BadHandler struct {
	req *SomeReq  // 缺少 cs
}
```
