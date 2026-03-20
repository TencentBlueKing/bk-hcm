## Context

当前 GPU 需求提报的创建和覆盖接口中，前端传入的子单数据包含两部分：
- **固定字段**：demand_type、demand_year、demand_month、gpu_num、qpm_max，直接作为 `CreateGpuDemandOrderDetail` 结构体的字段
- **动态字段**：extension JSON 数组，按 schema 的 headers 顺序排列

目前 `validateOrderExtensions` 只校验了 extension（动态字段）部分，而固定字段仅依赖结构体 validate tag 做基础校验，没有依据 schema 的 `fixed_headers` 进行业务规则校验（如 enum 枚举值、范围约束）。

Schema 的 `fixed_headers` 中每个 Header 通过 `db_field` 标识其对应的数据库/JSON 字段名（如 `demand_year`、`demand_month`），与 `CreateGpuDemandOrderDetail` 的 JSON tag 一致。

## Goals / Non-Goals

**Goals:**
- 在创建和覆盖接口中，对每个 detail 的固定字段按 schema fixed_headers 进行校验
- 复用已有的 `ValidateTypedValue` 校验逻辑（类型校验 + 范围/枚举校验）
- 通过 `db_field` 建立 schema header 与请求结构体字段的映射关系

**Non-Goals:**
- 不修改现有 validate tag 校验逻辑，两层校验并存（validate tag 做格式校验，schema 做业务规则校验）
- 不修改 Excel 导入预览接口的校验流程（该接口已通过 ValidateCellValue 对所有列做校验）
- 不修改 fixed_headers 的 schema 定义格式

## Decisions

### 1. 通过 JSON 序列化 + db_field 映射固定字段值

**决策**：将 `CreateGpuDemandOrderDetail` 通过 `json.Marshal` → `json.Unmarshal` 转为 `map[string]interface{}`，然后遍历 `fixed_headers` 按 `db_field` 从 map 中查找对应值进行校验。

**理由**：`CreateGpuDemandOrderDetail` 的 JSON tag（如 `json:"demand_year"`）与 `Header.DBField`（如 `demand_year`）天然一致，序列化后自动建立映射关系，无需反射或手动维护字段映射。校验在非热路径上，序列化开销可以忽略。

**替代方案**：
- 反射自动提取结构体字段 → 过于复杂，维护成本高
- 手写 `FixedFieldMap()` 方法返回 map → 显式但新增字段时需同步更新，容易遗漏

### 2. 在 validate.go 中新增 ValidateFixedFields 函数

**决策**：在 `pkg/tools/excel/validate.go` 中新增 `ValidateFixedFields(values map[string]interface{}, headers []Header) []string` 函数，遍历 headers，按 `db_field` 从 map 中取值后调用 `ValidateTypedValue`。

**理由**：与 `ValidateExtension` 平行，保持校验逻辑集中在 excel 包中，方便复用和测试。

### 3. 扩展 validateOrderExtensions 函数

**决策**：在现有的 `validateOrderExtensions` 函数中同时校验固定字段和动态字段，将函数重命名为 `validateOrderDetails` 以反映更完整的校验范围。

**理由**：创建和覆盖接口共用此函数，只需改一处即可覆盖两个接口。

## Risks / Trade-offs

- **[向后兼容]** 新增的校验可能导致之前能通过的请求被拒绝 → 这是预期行为，应在发布说明中注明
- **[db_field 缺失]** 如果 fixed_header 没有配置 db_field，将跳过该字段校验 → 这是合理的降级策略，不会导致误判
