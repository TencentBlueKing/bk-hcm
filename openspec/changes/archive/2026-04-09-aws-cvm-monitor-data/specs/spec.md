# aws-cvm-monitor-data

## ADDED Requirements

### Requirement: Unified CVM monitor endpoint SHALL support vendor aws

系统 SHALL 在统一接口 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 上支持 `vendor=aws`，并复用现有接口风格：`metric_name`、`period`、`ids` 作为通用参数。请求在 `vendor=aws` 时 SHALL 额外支持 AWS 专有 UTC 时间参数（ISO8601 格式），响应结构 SHALL 与现有 `data_points` 结构保持兼容。

#### Scenario: vendor aws request is accepted with unified parameters
- **WHEN** 调用方以 `vendor=aws` 调用统一 monitor/data 接口，并提供通用参数与 AWS 专有 UTC 时间参数
- **THEN** 系统 SHALL 返回与现有厂商一致的 `data.data_points[]` 结构（包含 `id`、`ip`、`region`、`instance_id`、`timestamps`、`values`、`extensions`）

#### Scenario: vendor aws request is rejected when aws time parameters are missing or invalid
- **WHEN** 调用方以 `vendor=aws` 调用接口但未提供 AWS 专有 UTC 时间参数，或时间格式不符合 ISO8601
- **THEN** 系统 SHALL 返回参数校验错误，并明确提示 AWS 时间参数缺失或格式非法

### Requirement: Cloud-server SHALL route aws monitor requests by account and region

cloud-server 的 CVM 监控查询逻辑 SHALL 在 `vendor=aws` 时按 `(account_id, region)` 进行实例分组，并对每个分组调用 hc-service AWS CVM monitor 接口，最终聚合返回统一结构结果。

#### Scenario: grouped querying for multiple aws accounts and regions
- **WHEN** 请求中的 `ids` 覆盖多个 AWS 账号或多个地域的实例
- **THEN** 系统 SHALL 按 `(account_id, region)` 分组分别调用下游，并将每组结果合并到统一响应中

#### Scenario: instance metadata missing in grouped aws query
- **WHEN** 某些请求实例在数据库中缺失或无法匹配分组所需元数据
- **THEN** 系统 SHALL 返回可定位的错误信息，且不得返回错误映射的监控数据

### Requirement: Hc-service SHALL expose aws cvm monitor endpoint and query CloudWatch

hc-service SHALL 提供 AWS CVM monitor 路由与处理逻辑，接收 cloud-server 下发的 AWS 监控查询请求，调用 adaptor AWS 查询 CloudWatch 指标并返回标准化 `data_points` 数据。

#### Scenario: hc-service returns data points from cloudwatch query results
- **WHEN** hc-service 收到合法 AWS monitor 请求并成功查询 CloudWatch
- **THEN** hc-service SHALL 返回每个查询对象对应的 `dimensions`、`timestamps`、`values` 与 `extensions` 数据

#### Scenario: hc-service returns downstream errors transparently
- **WHEN** CloudWatch 查询失败（如鉴权失败、区域不可用、限流）
- **THEN** hc-service SHALL 返回错误并保留可排障上下文，不得伪造成功响应

### Requirement: Adaptor aws cvm monitor SHALL reuse cloudwatch getmetricdata capability

`pkg/adaptor/aws` 的 CVM 监控实现 SHALL 复用现有 `pkg/adaptor/aws/cloudwatch.go` 的 `GetMetricData` 能力构建查询，不得重复实现分页归并核心逻辑。CVM 监控层仅负责查询参数组装与结果映射。

#### Scenario: cvm monitor uses existing getmetricdata path
- **WHEN** adaptor 执行 AWS CVM 监控查询
- **THEN** 系统 SHALL 通过既有 `GetMetricData` 能力执行 CloudWatch 查询并获得分页归并后的时序结果

#### Scenario: cloudwatch result mapping preserves original value semantics
- **WHEN** adaptor 将 CloudWatch `MetricDataResults` 映射为 CVM monitor 数据
- **THEN** 系统 SHALL 保持 AWS 原始值语义与单位，不做 Mbps 转换

### Requirement: Aws phase1 traffic semantics SHALL map lan and wan to total traffic

在 Phase 1，系统 SHALL 采用 AWS 总流量映射策略：`LanOuttraffic` 与 `WanOuttraffic` 统一映射 `NetworkOut`，`LanIntraffic` 与 `WanIntraffic` 统一映射 `NetworkIn`。系统 SHALL 通过 `extensions` 明确标识该语义属于 Phase 1 总流量映射。

#### Scenario: lan and wan metrics return the same source values in phase1
- **WHEN** 调用方分别查询 `LanOuttraffic` 与 `WanOuttraffic`（或 `LanIntraffic` 与 `WanIntraffic`）
- **THEN** 系统 SHALL 返回来自同一 AWS 源指标（`NetworkOut` 或 `NetworkIn`）的结果值

#### Scenario: response extension marks phase1 total-traffic semantics
- **WHEN** 系统返回 `vendor=aws` 的监控数据点
- **THEN** `extensions` SHALL 包含可识别的语义标记（例如源指标、语义阶段、流量范围说明），用于表明当前并非内外网精确拆分
