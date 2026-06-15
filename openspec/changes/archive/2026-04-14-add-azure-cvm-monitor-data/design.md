## Context

当前 CVM 监控链路已经打通 `cloud-server -> hc-service -> adaptor` 的多层调用模型，并已支持 `tcloud`、`huawei`、`aws`。其中，`cloud-server` 负责实例鉴权、按 `(account_id, region)` 分组聚合；`hc-service` 负责厂商路由与请求转换；`adaptor` 负责云厂商 SDK/API 调用与结果标准化。  
本次变更需要在不破坏既有接口兼容性的前提下，将 Azure 接入该链路，并补齐 Azure 专有参数透传与 `extensions` 扩展信息输出。同时，针对“内网/外网流量四象限”能力，需设计“原生优先、兜底回退”的可观测机制。

约束条件：
- 对外接口路径与主结构保持不变：`POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data`。
- 非 Azure 厂商行为不变。
- Azure 专有参数仅在 `vendor=azure` 生效。
- 厂商扩展信息统一承载于 `extensions` 字段。

## Goals / Non-Goals

**Goals:**
- 为 `vendor=azure` 提供可用的 CVM 监控查询能力，支持流量相关指标查询与统一输出。
- 在 `pkg/api/cloud-server/cvm`、`cmd/cloud-server/service/cvm`、`pkg/api/hc-service/cvm`、`cmd/hc-service/service/cvm`、`pkg/adaptor/azure` 建立完整 Azure 监控调用链路。
- 支持 Azure 专有参数：`metric_namespace`、`aggregation`、`auto_adjust_timegrain`、`top`、`orderby`、`filter`、`result_type`。
- 在返回数据中通过 `extensions` 输出 Azure 扩展字段，至少包含 `unit`、`cost`、`granularity`、`namespace`、`resource_region`。
- 明确四象限语义策略：优先使用 Azure 原生可分离能力，无法分离时回退并显式标注语义来源与阶段。

**Non-Goals:**
- 不在本次变更中引入新的对外 API 路径或版本。
- 不在本次变更中重构既有 `tcloud/huawei/aws` 的核心监控实现。
- 不承诺本次即完成 Azure 所有资源类型与所有维度组合的公网/内网精准拆分。
- 不在本次变更中实现跨厂商统一的监控缓存/限流中间层。

## Decisions

### 决策 1：沿用现有三层架构扩展 Azure 监控链路

**方案**：保持 `cloud-server -> hc-service -> adaptor` 分层职责，新增 Azure 对应请求结构、服务路由和 adaptor 实现。  
**原因**：与现有 AWS/Huawei 实现一致，改动边界清晰，测试与回归成本可控。  
**备选方案**：直接在 `cloud-server` 调 Azure SDK，绕过 `hc-service`。  
**不选原因**：破坏服务分层，增加凭证与云厂商访问耦合，不利于后续运维治理。

### 决策 2：Azure 参数采用“通用参数 + vendor 专有参数”模型

**方案**：通用参数继续沿用（`metric_name/period/ids` + 时间窗口），Azure 专有参数作为可选字段新增到请求结构，仅在 Azure 分支校验和使用。  
**原因**：兼容现有接口与调用方；支持 Azure Monitor 查询增强能力。  
**备选方案**：将所有厂商参数强行统一到一个最小公共子集。  
**不选原因**：会损失 Azure 能力，且未来扩展性差。

### 决策 3：四象限流量语义采用“原生优先、兜底回退”

**方案**：
- 优先尝试 Azure 原生可分离能力（如可判定内外网的指标/维度组合）；
- 若当前请求无法获得四象限原生拆分，则回退到可用总量指标并在 `extensions` 标注：
  - `source_metric_name`
  - `semantic_phase`
  - `traffic_scope`
  - `is_fallback`
**原因**：在保证可用性的同时保留语义透明度，避免静默误导调用方。  
**备选方案**：固定采用 AWS Phase1 同构映射。  
**不选原因**：忽略 Azure 潜在原生能力，不符合“原生优先”要求。

### 决策 4：统一 `extensions` 作为多云厂商扩展信息容器

**方案**：所有厂商扩展字段继续放入 `extensions`，Azure 补充 `unit/cost/granularity/namespace/resource_region`，并保留语义标识字段。  
**原因**：无需变更主响应结构，兼容历史消费者；新增字段对调用方是向后兼容。  
**备选方案**：为 Azure 新增顶层字段。  
**不选原因**：会破坏跨厂商一致性并增加前端/调用方分支逻辑。

### 决策 5：按 `(account_id, region)` 分组发起 Azure 监控查询

**方案**：复用现有分组策略，cloud-server 先从 DB 查询实例，再按账户与地域分组调用 hc-service。  
**原因**：与现有厂商一致，便于复用错误处理与聚合流程。  
**备选方案**：按单实例逐个请求。  
**不选原因**：调用次数大、性能差、易触发云侧限流。

## Risks / Trade-offs

- **[Azure 维度能力不一致导致四象限精度波动]** → 通过 `extensions.is_fallback` 与语义字段显式标注，并在文档中说明能力边界。
- **[Azure 查询参数组合复杂，易触发服务端校验失败]** → 在 API 层增加 vendor 级参数组合校验，错误信息保持可读并指向具体参数。
- **[跨层新增结构体较多，存在字段遗漏风险]** → 按“API 定义 -> service 转换 -> adaptor option -> 结果回填”顺序对齐字段，并补充单测覆盖关键参数。
- **[新增 Azure 查询导致调用时延增加]** → 复用分组批量查询与现有并发策略，避免单实例串行调用。
- **[对既有厂商回归风险]** → 修改点限定在 Azure 分支与共用结构新增字段，保持原分支行为不变并做回归测试。

## Migration Plan

1. 更新 OpenAPI/Markdown 文档，明确 Azure 参数与扩展字段。
2. 扩展 `pkg/api/cloud-server/cvm` 请求结构与 `Validate(vendor)` 逻辑，加入 Azure 参数校验规则。
3. 在 `cmd/cloud-server/service/cvm/monitor.go` 新增 Azure 分支，复用分组聚合流程。
4. 新增 `pkg/api/hc-service/cvm` 的 Azure monitor req/resp 定义与校验。
5. 在 `cmd/hc-service/service/cvm` 增加 Azure monitor 路由与处理函数，调用 adaptor。
6. 在 `pkg/adaptor/types/cvm` 增加 Azure monitor option/result。
7. 在 `pkg/adaptor/azure` 新增 `monitor.go`，封装 Azure Metrics List 调用、参数转换与结果标准化。
8. 完成单测与集成验证：参数校验、链路可达、语义字段、回退标记、兼容性回归。

回滚策略：
- 代码回滚到不含 Azure monitor 分支版本即可恢复原行为；
- 由于新增字段均为向后兼容，不涉及数据迁移与存量数据修复。

## Open Questions

- Azure 在当前账户权限与资源模型下，是否能稳定提供可用于内/外网四象限判定的 NIC 维度数据；若部分地域不可用，回退策略是否需要按地域开关。
- `result_type` 是否允许透传全部 Azure 枚举值，或在 HCM 侧限定为 `Data/Metadata` 以减少无效组合。
- `orderby` 与 `top` 的组合是否在监控场景默认启用，还是仅在调用方明确传参时启用。
