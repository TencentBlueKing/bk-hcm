---
name: go-excel-export
description: Excel 导出功能实现指南，包括 excelize 库使用、Handler 定义、Batch/Stream 模式选择、表头与数据填充。当实现 Excel 导出接口、处理大数据量导出、或使用 excelize 库时使用。
---

# Excel 导出规范

## 快速参考

| 场景 | 规则文件 |
|-----|---------|
| Handler 定义与命名 | [rules/handler-pattern.md](rules/handler-pattern.md) |
| 导出模式选择（Batch/Stream） | [rules/export-mode.md](rules/export-mode.md) |
| 表头与数据填充 | [rules/data-filling.md](rules/data-filling.md) |
| 并发安全与预加载 | [rules/concurrency.md](rules/concurrency.md) |
| 辅助函数与工具 | [rules/helpers.md](rules/helpers.md) |

## 核心原则

1. **统一接口**：通过泛型接口 `BatchExportHandler` / `StreamExportHandler` 定义导出行为
2. **内存友好**：大数据量（>10万行）使用 Stream 模式流式写入
3. **可组合性**：支持单 Sheet 和多 Sheet 导出模式
4. **并发安全**：Stream 模式使用锁机制保护共享状态

## 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                      BaseExporter[T]                        │
│  ┌─────────────────────────────────────────────────────┐   │
│  │  handler: BatchExportHandler[T]                      │   │
│  │       or StreamExportHandler[T]                      │   │
│  │  mode: ExportModeBatch / ExportModeStream            │   │
│  │  isMultiSheet: bool                                  │   │
│  └─────────────────────────────────────────────────────┘   │
│                           │                                 │
│                  ExportExcel(kt, resp)                      │
│                           │                                 │
│         ┌─────────────────┴─────────────────┐              │
│         ▼                                   ▼              │
│   ┌───────────┐                      ┌───────────┐         │
│   │  Batch    │                      │  Stream   │         │
│   │ GetItems  │                      │ GetItems- │         │
│   │ AddHeader │                      │  Stream   │         │
│   │ AddRows   │                      │ AddStream-│         │
│   │           │                      │  Header   │         │
│   │           │                      │ AddStream-│         │
│   │           │                      │  Rows     │         │
│   └───────────┘                      └───────────┘         │
└─────────────────────────────────────────────────────────────┘
```

## 常见问题

### Q: 什么时候使用 Batch 模式，什么时候使用 Stream 模式？

**A**: 
- **Batch 模式**：数据量较小（<10万行），一次性加载到内存，代码简单
- **Stream 模式**：数据量大或不确定，需分批获取并流式写入，内存友好但代码较复杂

### Q: 为什么数据行要从第 2 行开始？

**A**: 第 1 行预留给表头。在计算实际行号时，应使用 `rowIndex + 2`（0-indexed 循环）。

### Q: 如何支持中文文件名下载？

**A**: 使用 `url.QueryEscape` 对文件名编码，并设置 `filename` 和 `filename*=utf-8''` 两种格式以兼容不同浏览器。

### Q: 如何预加载关联数据避免重复查询？

**A**: 使用 `sync.Once` 机制，确保无论 `GetItemsStream` 被调用多少次，关联数据只会加载一次。

## 相关规范

本规范与以下规范有关联，必要时请一并参考：
- [go-naming-convention](../go-naming-convention/SKILL.md) - Handler 命名规范
- [api-implementation-workflow](../../.cursor/skills/api-implementation-workflow/SKILL.md) - API 接口实现流程
