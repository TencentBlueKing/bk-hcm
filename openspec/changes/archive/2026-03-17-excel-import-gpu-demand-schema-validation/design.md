## Context

当前 `ExcelImportGpuDemand` 流程为：获取最新模版schema → 校验文件完整性（sheet匹配、列头匹配） → 逐sheet解析数据行 → `buildDetails` 构建detail（类型转换+组装raw_data）。`buildDetails` 中调用 `convertCellValue` 做值类型转换，但转换失败时仅静默保留原始字符串，不产生任何校验错误。`validate_result` 始终为空数组。

需要在 `buildDetails` 阶段增加基于schema的单元格值校验，将错误写入 `validate_result`。

## Goals / Non-Goals

**Goals:**
- 在 `buildDetails` 构建每行detail时，对每个可见列执行类型和约束校验
- Header 结构体扩展 `Min`/`Max` 字段用于范围/长度约束
- 校验逻辑封装为通用函数，放在 `pkg/tools/excel/` 包中，与schema定义同层

**Non-Goals:**
- 不修改API响应结构（validate_result字段已存在）
- 不做跨行/跨列的业务逻辑校验（如"月份必须连续"等）
- 不做required以外的必填校验（如条件必填）
- 不修改已有的类型转换逻辑 `convertCellValue`

## Decisions

### Decision 1: 校验逻辑放在 `pkg/tools/excel/validate.go`

**选择**：新建 `pkg/tools/excel/validate.go`，提供 `ValidateCellValue(val string, header Header) []string` 函数，返回中文错误描述字符串列表。

**理由**：校验逻辑与schema定义强相关，放在excel工具包中可复用。`buildDetails` 在类型转换同时调用校验函数，将返回的错误追加到 `validate_result`。

**替代方案**：在 `cmd/woa-server/logics/plan/` 中直接写校验逻辑 —— 耦合度高，不利于其他模版复用。

### Decision 1.5: 直接返回中文错误描述

**选择**：`validate_result` 保持 `[]string`，直接返回中文错误描述，格式为 `"{列名}: {中文错误描述}"`。中文模版以 `fmt.Sprintf` 格式定义在 `validate.go` 中作为常量集中管理。

**理由**：最简单直接，无额外依赖和 API 变更。前端无需翻译逻辑，直接展示即可。

### Decision 2: Min/Max 使用 `*int64` 指针类型

**选择**：Header 中 `Min`、`Max` 字段类型为 `*int64`，JSON tag 带 omitempty。

**理由**：需要区分"未设置"和"设置为0"两种语义。使用指针类型，nil 表示未设置（不校验），非nil 表示有约束。string 类型的长度和 int 类型的数值范围都是整数，使用 int64 语义更精确。

### Decision 3: 校验顺序 required → type → range/length

**选择**：先校验 required（空值检查），再校验类型合法性，最后校验范围/长度约束。required 失败时跳过后续校验。

**理由**：空值无法做类型和范围校验，required 校验失败后继续校验无意义。但类型校验和范围校验可同时产生错误（如值不是整数，无需再校验范围）。

### Decision 4: enum 比较策略

**选择**：将单元格字符串值转换后与 value 列表比较。value 列表中元素如果是 float64（JSON反序列化结果），先尝试 ParseInt 比较，失败再尝试 ParseFloat 比较；如果是 string，直接字符串比较。

**理由**：复用现有 `convertEnumValue` 的类型推断逻辑，保持一致性。

## Risks / Trade-offs

- [类型转换与校验分离] `convertCellValue` 和 `ValidateCellValue` 逻辑有重叠（都需要判断类型并尝试解析），但职责不同（一个转换、一个校验），保持分离更清晰 → 接受少量重复
- [Min/Max 语义随type变化] 同一字段不同type语义不同可能增加理解成本 → 通过文档和注释明确
