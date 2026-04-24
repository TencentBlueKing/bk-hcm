## Context

当前资源视角接口 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 的调用链路已在腾讯云打通：`cloud-server` 完成鉴权和实例分组，调用 `hc-service`，再由 `pkg/adaptor/tcloud` 对接云厂商监控 API。  
但在 `vendor=huawei` 场景下，该链路尚未实现，导致华为云实例无法走统一入口获取监控数据。

本次变更是跨模块扩展，涉及：

- `cloud-server`：入口分发与聚合逻辑扩展；
- `hc-service`：华为云监控查询内部接口新增；
- `pkg/adaptor/huawei`：华为云 CES `BatchListMetricData` 调用实现；
- `docs/api-docs`：接口文档补充 vendor 支持范围。

关键约束如下：

- 对外接口路径、请求结构与响应主结构保持不变，仅扩展 `vendor` 支持范围；
- 继续遵循“按 `(account_id, region)` 分组调用云厂商 API”的现有模式；
- 保持 `tcloud` 现有行为不受影响；
- 华为云监控查询需处理其特有语义：毫秒时间戳、`period` 字符串、`filter` 聚合方式。

## Goals / Non-Goals

**Goals:**

- 在统一接口中新增 `vendor=huawei` 的监控查询能力，返回结构与现有接口一致。
- 在 `cloud-server` 复用现有权限校验、实例信息查询、分组聚合框架，新增华为云分支。
- 在 `hc-service` 新增华为云监控查询接口，调用 `pkg/adaptor/huawei` 完成实际云上请求。
- 在 `pkg/adaptor/huawei` 新增独立监控查询实现，并完成华为云请求/响应到内部协议的转换。
- 补充接口文档中的 vendor 能力说明与华为云约束说明，确保前后端和调用方认知一致。

**Non-Goals:**

- 不改造现有腾讯云监控查询逻辑、协议和行为。
- 不在本次变更中新增跨云统一指标映射层（仅支持调用方传入华为云原生可用指标名）。
- 不扩展到 AWS/GCP/Azure 的监控查询能力。
- 不修改数据库结构（沿用现有 CVM 基础信息进行分组和回填）。

## Decisions

### D1: 复用现有三段式调用链路，按云厂商增加分支

**选择：** 保持 `cloud-server -> hc-service -> adaptor` 的既有分层，仅在各层补齐华为云分支，不引入新中间层。

**理由：**

- 与现有腾讯云实现保持结构一致，降低认知和维护成本。
- 满足“非 data-service 服务不直接操作 DB”的分层规则。
- 便于后续其它云厂商按同模式扩展。

**备选方案：**

- 在 `cloud-server` 直接调用华为云 SDK/HTTP。  
  不选原因：会破坏分层边界，增加 `cloud-server` 外部依赖。

### D2: `cloud-server` 继续按 `(account_id, region)` 分组聚合调用

**选择：** 在 `getMonitorData` 中新增 `vendor=huawei` 分支，并复用与腾讯云一致的分组策略（同账号同地域合并一次调用），最终聚合成统一响应 `data_points`。

**理由：**

- 现有数据模型已通过实例信息提供 `account_id`、`region`、`cloud_id`；
- 华为云接口支持批量 `metrics` 查询，分组后能减少请求次数；
- 可复用现有“cloud_id -> 内部实例ID/IP/region”回填逻辑，保持返回一致性。

**备选方案：**

- 每个实例单独调用华为云监控 API。  
  不选原因：请求量放大，性能与配额风险更高。

### D3: 在 `hc-service` 新增华为云监控接口，按厂商参数语义透传

**选择：** 参考 `GetTCloudMonitorData` 新增 `GetHuaWeiMonitorData`，但不强制与腾讯云保持参数格式一致；由外部按华为云要求直接传入合法参数（包括毫秒时间戳），`cloud-server` 原样下发到 `hc-service` 与 adaptor。

**理由：**

- 满足“不同云参数语义不同”的实际场景，避免为统一而统一导致信息损失或额外转换误差；
- `cloud-server` 不做厂商参数二次转换，减少链路复杂度；
- 将校验重点放在“参数是否满足该 vendor 规则”，而非跨云统一格式。

**备选方案：**

- 强制所有云统一时间格式后再在内部转换。  
  不选原因：与多云参数本身差异不匹配，增加不必要转换逻辑。

### D4: `pkg/adaptor/huawei/monitor.go` 独立实现 CES 调用与响应映射

**选择：** 在 `pkg/adaptor/huawei` 新增 `monitor.go`，封装 `BatchListMetricData` 调用：

- 将内部实例列表转为 `metrics[]`（`namespace=SYS.ECS`，`dimensions=[{name: instance_id, value: <cloud_id>}]`）；
- 透传指标名为 `metric_name`；
- `period` 按华为云要求透传，纳入 `period=1` 实时场景；
- `filter` 使用默认 `average`（与现有接口“未指定统计方式按默认”原则对齐）；
- 保留华为云原生时间语义（毫秒时间戳）并按协议约定返回。

**理由：**

- 适配层职责是“云厂商协议适配”，保持单一职责；
- 与 `pkg/adaptor/tcloud/monitor.go` 结构对齐，便于后续维护；
- 在 adaptor 内统一处理厂商字段差异，避免上层感知复杂度。

**备选方案：**

- 在 `hc-service` 直接拼装华为云请求体并发 HTTP。  
  不选原因：会削弱 adaptor 统一适配层价值，形成重复代码。

### D5: 保持统一基础结构，同时允许扩展厂商特有字段

**选择：** 保持统一基础响应结构（`id/ip/region/instance_id/timestamps/values`），并允许在数据点中增加厂商特有扩展字段（如华为云流量相关维度/字段），用于承载统一字段无法完整表达的监控信息。

**理由：**

- 统一基础字段可保障现有调用逻辑兼容；
- 扩展字段可避免华为云特有监控能力（如内外网流量场景）在适配中被丢失；
- 采用“基础字段稳定 + 扩展字段按 vendor 可选”的方式平衡兼容性与可扩展性。

**备选方案：**

- 严格禁止响应扩展字段。  
  不选原因：会限制多云特性表达，影响华为云监控数据可用性。

## Risks / Trade-offs

- **[风险] 华为云指标名与腾讯云不同，调用方若沿用腾讯云指标将查不到数据**  
  → 在接口文档补充华为云常用指标与差异说明，明确指标名需按 vendor 选择。

- **[风险] 各云时间与 period 语义不同，若误做统一转换会导致数据偏差**  
  → 按 vendor 透传与校验参数，不做跨云格式强制统一；补充分 vendor 的参数校验与测试用例。

- **[风险] 华为云 `period` 与时间跨度存在组合限制，超限可能返回空数据或被服务端调整 `from`**  
  → 保持参数透传，同时在文档中标注限制；后续如有需要再加前置校验。

- **[权衡] 本次固定 `filter=average`，暂不开放客户端选择 min/max/sum/variance**  
  → 换取接口稳定与实现简洁；后续可在不破坏兼容的前提下扩展可选聚合方式。

## Migration Plan

1. 新增 `openspec` 工件完成后，按设计实现 `cloud-server`、`hc-service`、`adaptor` 与 API 协议扩展。
2. 补充 `docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`，更新 `vendor` 支持与华为云说明。
3. 在联调环境验证四类场景：
   - 单账号单地域；
   - 单账号多地域；
   - 多账号多地域；
   - 指标无数据/部分实例无数据。
4. 与现网腾讯云场景回归，确认 `vendor=tcloud` 行为无回归。

**回滚策略：**

- 若上线后出现问题，可先关闭或回退 `vendor=huawei` 分支代码路径；
- 因为不变更 DB 和既有腾讯云逻辑，回滚仅影响新增华为能力，不影响存量能力。

## Open Questions

- 无。
