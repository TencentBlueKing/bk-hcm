## Context

当前GPU需求提报流程分为两步：
1. 前端上传Excel，调用 `ExcelImportGpuDemand` 接口获取解析后的 details 列表（预览）
2. 用户确认后，需要将 details 持久化为数据库中的主单（`res_plan_demand_gpu_order`）和子单（`res_plan_demand_gpu_suborder`）

data-service 层已有 `BatchCreateResPlanDemandGpuOrder` 和 `BatchCreateResPlanDemandGpuSubOrder` 接口，woa-server 层的 `ResourcePlanClient` 也已封装好对应的客户端方法。当前缺少的是 woa-server 层的业务编排接口，将前端传入的扁平 details 列表转换为主单+子单的两级结构并持久化。

## Goals / Non-Goals

**Goals:**
- 提供 woa-server 业务层接口，完成 GPU 需求提报主单和子单的创建
- 主单关联最新 GPU 模板 ID，初始状态为 `INIT`
- 将前端传入的 details 中每条记录转换为一条子单，extension 字段序列化为 JSON 存储
- 子单通过 `order_id` 关联主单，初始状态为 `INIT`
- 整个操作需要保证数据一致性（主单创建成功后才创建子单，子单创建失败需要合理报错）

**Non-Goals:**
- 不涉及子单的审批流程触发（后续流程由其他接口负责）
- 不涉及 Excel 上传解析逻辑（已由 `ExcelImportGpuDemand` 完成）
- 不修改 data-service 层的现有接口
- 不处理子单数量超过 100 条的分批创建（当前 data-service 限制为单次最多 100 条）

## Decisions

### 1. 请求类型定义位置

**决定**: 在 `pkg/api/woa-server/` 下新增 `res_plan_gpu_demand_order.go` 文件定义请求类型。

**理由**: 与已有的 `res_plan_gpu_excel_import.go` 保持一致的文件组织方式，woa-server 对外接口的请求/响应类型统一放在 `pkg/api/woa-server/` 下。

### 2. details 中 extension 字段的处理

**决定**: 前端传入的 `extension` 为 `[]interface{}` 类型，直接 JSON 序列化后存入子单的 `Extension`（`types.JsonField`）字段。

**理由**: `types.JsonField` 底层是 `string`，通过 `types.NewJsonField(detail.Extension)` 即可完成序列化。与 Excel 导入预览接口返回的 extension 数据格式一致。

### 3. qpm_max 字段类型处理

**决定**: API 接口层接收 `float64` 类型（与接口文档一致），存入子单时转换为 `int64`（子单表字段类型为 `int64`）。

**理由**: 数据库子单表的 `qpm_max` 字段为 `int64`，需要在 logics 层做类型转换。使用 `int64(detail.QpmMax)` 截断小数部分，与现有表结构保持一致。

**备选方案**: 修改子单表字段为 `float64`，但这涉及数据库 migration 且超出本次变更范围。

### 4. 模板 ID 获取方式

**决定**: 复用已有的 `getLatestGpuTplSchema` 方法中获取最新模板的查询逻辑，单独抽取一个 `getLatestGpuTemplateID` 方法，仅获取模板 ID 而无需解析 schema。

**理由**: 创建主单只需要模板 ID，不需要解析 tpl_schema 的完整内容。抽取单独方法避免不必要的 JSON 反序列化开销。

### 5. 调用链路

**决定**: woa-server service handler → logics Controller 方法 → data-service client（先创建主单获取 ID，再创建子单）。

```
前端 → woa-server service handler (解析请求、鉴权)
     → logics.CreateGpuDemandOrder (业务编排)
       → getLatestGpuTemplateID (获取最新模板ID)
       → client.BatchCreateResPlanDemandGpuOrder (创建主单，获取主单ID)
       → 遍历 details 构建子单请求 (转换 extension、关联 order_id)
       → client.BatchCreateResPlanDemandGpuSubOrder (批量创建子单)
     ← 返回主单 ID
```

**理由**: 遵循项目已有的分层架构，service 层只负责请求解析和鉴权，业务逻辑在 logics 层完成，数据操作通过 data-service client 调用。

### 6. 错误处理策略

**决定**: 如果主单创建成功但子单创建失败，直接返回错误，不回滚主单。

**理由**: data-service 层的 `BatchCreateResPlanDemandGpuOrder` 内部已使用事务（`AutoTxn`），主单创建是原子的。子单创建也使用独立事务。如果子单创建失败，主单状态为 `INIT`（未进入审批流程），可由前端重试或管理员清理。避免跨服务分布式事务的复杂性。

## Risks / Trade-offs

- **[子单数量限制]** data-service 的 `BatchCreateResPlanDemandGpuSubOrder` 限制单次最多 100 条。如果 details 超过 100 条需要分批调用。 → 当前 Excel 模板行数预期不会超过 100 条，暂不处理分批逻辑，在 Validate 中限制 details 数量不超过 100。
- **[主单子单非事务]** 主单和子单分两次 RPC 调用创建，非原子操作。 → 主单状态为 `INIT`，子单创建失败不会触发下游流程，影响可控。
- **[qpm_max 精度丢失]** float64 转 int64 会丢失小数部分。 → 当前业务场景 QPM 为整数或接受截断，如需精度可后续修改表结构。
