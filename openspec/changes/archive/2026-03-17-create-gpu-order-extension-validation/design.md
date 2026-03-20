## Context

CreateGpuDemandOrder 接口当前流程：req.Validate()（结构校验）→ 鉴权 → 获取 templateID → 创建主单 → 创建子单。extension 字段为 `types.JsonField`（json.RawMessage），直接存入数据库，未按 schema 做值校验。

Excel import 阶段已实现 `ValidateCellValue` 对单元格做 type/enum/range/length 校验。CreateGpuDemandOrder 需要复用此逻辑校验 extension。

## Goals / Non-Goals

**Goals:**
- 复用 `ValidateCellValue` 校验 extension 中每个值
- 新增 `ValidateExtension` 作为 `[]interface{}` → `ValidateCellValue` 的适配层
- controller 层在创建工单前执行校验，失败则拒绝

**Non-Goals:**
- 不校验 fixed_headers 对应的字段（demand_year/demand_month/gpu_num/qpm_max 已有 validate tag）
- 不修改已有的 req.Validate() 逻辑

## Decisions

### Decision 1: ValidateExtension 放在 `pkg/tools/excel/validate.go`

**选择**：在已有的 validate.go 中新增 `ValidateExtension` 函数。

**理由**：与 `ValidateCellValue` 同文件，职责一致。

### Decision 2: 直接对 interface{} 做类型校验，不转 string

**选择**：新增 `ValidateTypedValue(val interface{}, header Header) []string`，直接根据 Go 值类型和 header.Type 校验，不做 interface{} → string → parse 的转换。

**理由**：JSON 反序列化后值已有类型信息（float64/string/nil），直接校验更准确、更高效。具体策略：
- `nil` / 空字符串 → 按 required 校验
- header.Type == "int" → 检查值是否为 `float64` 且 `v == float64(int64(v))`（整数判定）+ 范围校验
- header.Type == "float" → 检查值是否为 `float64`
- header.Type == "enum" → 直接与 value 列表比较（类型敏感）
- header.Type == "string" → 检查值是否为 `string` + 长度校验

### Decision 3: 校验时机在 controller 层

**选择**：在 `Controller.CreateGpuDemandOrder` 中，获取 schema 后、创建主单前执行校验。

**理由**：复用已有的 `getLatestGpuTplSchema`，schema 获取一次即可。校验失败直接返回错误，避免创建了主单后子单校验失败导致数据不一致。

## Risks / Trade-offs

- [额外一次 schema 查询] CreateGpuDemandOrder 原来只查 templateID，现在还需查 tpl_schema。但 `getLatestGpuTplSchema` 已有，且查询量极小 → 可接受
