## Context

当前 HCM 的 CVM 监控统一入口已支持 `tcloud`、`huawei`，调用路径为 `cloud-server -> hc-service -> adaptor`。其中 AWS 在仓库中已有 CloudWatch 通用查询能力（`pkg/adaptor/aws/cloudwatch.go` 的 `GetMetricData` / `ListMetrics`），但未接入 CVM 监控统一入口，导致调用侧无法通过 `vendor=aws` 查询主机流量。

本次变更需要同时覆盖接口文档、cloud-server 分支路由、hc-service AWS monitor handler、adaptor AWS CVM 监控封装，属于跨模块改造。约束如下：

- cloud-server 对外参数形态尽量与现有厂商一致；
- `vendor=aws` 新增 AWS 专有 UTC 时间参数（ISO8601）；
- AWS 指标保持云厂商原始单位与语义，不在 HCM 内做 Mbps 转换；
- Phase 1 不做 AWS 内/公网精确拆分，`Lan*` 与 `Wan*` 基于实例总流量口径返回，并通过扩展字段标识语义。

## Goals / Non-Goals

**Goals:**

- 在统一接口 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 中新增 `vendor=aws` 能力。
- 打通 AWS 监控调用链路：cloud-server 分组调用 -> hc-service AWS 监控接口 -> adaptor CloudWatch 查询。
- 复用现有 `pkg/adaptor/aws/cloudwatch.go` 的 `GetMetricData` 能力，避免重复实现分页与结果归并逻辑。
- 在接口契约中增加 AWS 专有时间参数和必要扩展字段，保证对外兼容且可识别 Phase 1 语义。

**Non-Goals:**

- 不实现 AWS 基于 ENI+EIP 的内外网精确拆分。
- 不新增独立的 AWS 专用对外监控接口，仍沿用统一 monitor/data 接口。
- 不改造 sg-proxy 或调用侧逻辑。
- 不修改其他云厂商的流量计算或历史数据结构。

## Decisions

### 1) 复用 CloudWatch 通用查询能力

- **决策**：AWS CVM monitor 实现复用 `pkg/adaptor/aws/cloudwatch.go` 的 `GetMetricData`，仅在 CVM 领域层封装请求构造和结果映射。
- **原因**：
  - 现有能力已处理 CloudWatch 分页和同 `Id` 跨页归并；
  - 能降低重复代码与维护成本；
  - 可保持 AWS 查询行为在仓库内一致。
- **备选方案**：
  - 在 `pkg/adaptor/aws` 直接重新实现一次 `GetMetricData` 调用与分页处理。  
    **不选原因**：重复实现、容易出现行为不一致。

### 2) 保持统一接口风格并新增 AWS 专有时间参数

- **决策**：cloud-server 对外保持统一入参主体（`metric_name`、`period`、`ids`），同时在 `vendor=aws` 下新增 UTC 时间参数（ISO8601）并校验必填。
- **原因**：
  - 满足“对外参数尽量统一”的要求；
  - AWS 时间参数与 `tcloud/huawei` 语义差异明显，独立字段可避免歧义。
- **备选方案**：
  - 直接复用 `start_time/end_time`（本地格式）并在服务端转换。  
    **不选原因**：时区语义不清晰，跨云调用易出错。

### 3) AWS Phase 1 采用“总流量映射”语义

- **决策**：`LanOuttraffic`/`WanOuttraffic` 对应 `NetworkOut`，`LanIntraffic`/`WanIntraffic` 对应 `NetworkIn`，均为实例总流量口径；通过 `extensions` 暴露 `source_metric`、`semantic_phase`、`traffic_scope` 等标识。
- **原因**：
  - AWS EC2 实例级指标无法直接区分内外网；
  - 当前需求优先保证统一接口可用；
  - 与已确认的 Phase 1 策略一致，后续可平滑升级到精确拆分。
- **备选方案**：
  - `Wan*` 返回 0。  
    **不选原因**：与你确认的方案 B 不一致。
  - 直接报“不支持内外网分离”。  
    **不选原因**：会破坏调用侧统一处理路径。

### 4) 指标单位保持云厂商原始语义

- **决策**：不在 HCM 内做 Mbps 转换，AWS 返回值与 CloudWatch 查询语义保持一致；并在扩展字段中显式返回单位/统计方式信息。
- **原因**：
  - 你已明确要求“跟云厂商接口保持一致”；
  - 避免服务端额外换算带来的语义偏差。
- **备选方案**：
  - 统一转换为 Mbps。  
    **不选原因**：与当前约束冲突。

### 5) 分层改造范围

- **决策**：仅在以下层面最小改造：
  - `docs`：更新 monitor/data 文档，补充 aws 入参/出参与语义说明；
  - `pkg/api/cloud-server/cvm`：扩展 `GetMonitorDataReq` 的 aws 时间字段与 vendor 校验；
  - `cmd/cloud-server/service/cvm/monitor.go`：新增 aws 分支与分组调用；
  - `pkg/api/hc-service/cvm` + `cmd/hc-service/service/cvm/aws.go`：新增 AWS monitor 协议与 handler；
  - `pkg/adaptor/types/cvm` + `pkg/adaptor/aws`：新增 AWS monitor option/result 与查询封装。
- **原因**：与现有 `tcloud/huawei` 模式对齐，便于维护和测试。

## Risks / Trade-offs

- [AWS `Lan*` 与 `Wan*` 值相同可能引发误解] → 在接口文档和 `extensions` 中明确 Phase 1 为总流量映射语义，并给出后续精确拆分计划。
- [新增 aws 时间参数后调用方误传旧字段] → 在 `vendor=aws` 的参数校验中强制 UTC 字段必填，返回明确错误信息。
- [CloudWatch 统计窗口与 period 对齐问题导致结果波动] → 透传并记录查询参数（period/stat），在扩展字段回传，便于排障。
- [跨账号跨地域分组调用的部分失败] → 保持现有分组调用日志和错误链路，确保失败可定位到 account/region 维度。

## Migration Plan

1. 先更新 API 文档与请求/响应模型，明确 `vendor=aws` 参数与返回语义。
2. 在 cloud-server 增加 aws 分支和请求下游协议构造，保持原有 tcloud/huawei 逻辑不变。
3. 在 hc-service 的 AWS CVM 服务中新增 monitor 路由和 handler，调用 adaptor 查询。
4. 在 adaptor aws 新增 CVM 监控封装，复用 `cloudwatch.go` 的 `GetMetricData` 实现查询并做领域映射。
5. 联调验证：
   - vendor=aws 参数校验与时间格式；
   - `Lan*`/`Wan*` 四指标返回结构一致性；
   - 扩展字段语义标识是否完整。
6. 回归验证 tcloud/huawei 不受影响。
7. 若出现问题，按服务层回滚（优先回滚 cloud-server/hc-service aws monitor 路由与分支），不影响已有厂商能力。

## Open Questions

- AWS `GetMetricData` 在本场景固定使用的 `Stat`（如 `Sum`）是否需要对外可配置，或在服务端固定为单值策略。
- `extensions` 字段的最小必备键集合（例如 `unit`、`source_metric`、`traffic_scope`、`semantic_phase`）是否需要在 specs 中强制约束。
- 后续 Phase 2（ENI 维度精确拆分）是否作为本能力的后续 requirement，还是独立 capability 建模。
