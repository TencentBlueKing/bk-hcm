# 接口设计规范

## 1. 接口命名

接口统一命名为 `Interface`，位于各 DAO 包中：

```go
// Interface holds all the supported operations for the obs bill.
type Interface interface {
    // ...
}
```

## 2. 标准接口示例

```go
// Interface holds all the supported operations for the system config.
type Interface interface {
    List(kt *kit.Kit, opt *types.ListOption) (*types.ListSystemConfigResult, error)
    ListTxn(kt *kit.Kit, tx *sqlx.Tx, opt *types.ListOption) (*types.ListResult[*table.SystemConfig], error)
    Create(kt *kit.Kit, tx *sqlx.Tx, model *table.SystemConfig) (string, error)
    BatchCreate(kt *kit.Kit, tx *sqlx.Tx, models []table.SystemConfig) ([]string, error)
    BatchUpdate(kt *kit.Kit, tx *sqlx.Tx, models []table.SystemConfig) (int64, error)
    GetByTypeAndName(kt *kit.Kit, typ, name string) (*table.SystemConfig, error)

    WithAudit(audit audit.AuditorIf) Interface
}
```

## 3. 复合接口模式

对于大型 DAO，使用接口嵌入来组织：

```go
type Interface interface {
    List(kt *kit.Kit, opt *types.ListOption) (*types.ListObsBillResult, error)
    ListBrief(kt *kit.Kit, opt *types.ListOption) (*types.ListObsBillBriefResult, error)
    // ... 其他方法 ...
    
    AllocInterface  // 嵌入分摊接口
}

// AllocInterface 分摊相关操作
type AllocInterface interface {
    GetProdAllocCostMap(kt *kit.Kit, ...) (map[int64]map[int64]float64, error)
    // ...
}
```

## 4. 方法签名规范

### 4.1 参数顺序

```go
func (d *dao) MethodName(kt *kit.Kit, tx *sqlx.Tx, ...) (Result, error)
```

1. `kt *kit.Kit` - 上下文信息（必须，第一个参数）
2. `tx *sqlx.Tx` - 事务对象（可选，CUD 操作）
3. 业务参数（如 `opt *types.ListOption`、`model *table.Entity`）

### 4.2 返回值规范

| 操作类型 | 返回值 |
|---------|-------|
| 列表查询 | `*types.ListResult[T]` |
| 单条查询 | `*T, error` |
| 单条创建 | `string, error`（返回 ID） |
| 批量创建 | `[]string, error`（返回 ID 列表） |
| 更新/删除 | `int64, error`（返回影响行数） |

## 5. 方法命名规范

### 5.1 查询方法

| 前缀 | 用途 | 示例 |
|-----|------|------|
| `Get*` | 获取单个/聚合结果 | `GetByID`, `GetCostTrend`, `GetStatusCostMap` |
| `List*` | 列表查询 | `List`, `ListBrief`, `ListArchive` |
| `Count*` | 计数查询 | `CountByBizID` |
| `Agg*` | 聚合查询（ES） | `AggUsageByProduct`, `AggUsageTrendByDateRange` |

> **注意**：项目中聚合方法命名存在两种风格：
> - ES DAO 倾向使用 `Agg*`（如 `AggUsageByProduct`）
> - MySQL DAO 倾向使用 `Get*`（如 `GetCostTrend`）

### 5.2 写入方法

| 前缀 | 用途 | 示例 |
|-----|------|------|
| `Create` | 单条创建 | `Create` |
| `BatchCreate` | 批量创建 | `BatchCreate` |
| `BatchUpdate` | 批量更新 | `BatchUpdate` |
| `BatchDelete` | 批量删除 | `BatchDelete` |

### 5.3 特殊后缀

| 后缀 | 用途 | 示例 |
|-----|------|------|
| `*Txn` | 事务内执行 | `ListTxn` |
| `*Brief` | 返回简要信息 | `ListBrief` |

## 6. 通用类型

### 6.1 查询参数

```go
// From: pkg/dal/dao/types/types.go:36-40
type ListOption struct {
    Fields []string           `json:"fields"`
    Filter *filter.Expression `json:"filter"`
    Page   *types.BasePage    `json:"page"`
}
```

### 6.2 返回结果

```go
// From: pkg/dal/dao/types/types.go:30-33
type ListResult[T any] struct {
    Count   uint64 `json:"count"`
    Details []T    `json:"details"`
}
```

## 7. 接口设计检查清单

✅ **接口设计时应确认**：
- [ ] 接口命名为 `Interface`
- [ ] 第一个参数为 `kt *kit.Kit`
- [ ] CUD 操作包含 `tx *sqlx.Tx` 参数
- [ ] 方法命名遵循前缀规范
- [ ] 返回值类型符合规范
- [ ] 大型接口考虑使用嵌入拆分
