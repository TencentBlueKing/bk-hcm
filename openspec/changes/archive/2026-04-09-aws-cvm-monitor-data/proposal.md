## Why

当前 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 在 HCM 仅支持 `tcloud` 与 `huawei`，导致 AWS 主机无法复用现有统一监控入口获取流量数据。业务侧已按统一四指标（`LanOuttraffic`、`LanIntraffic`、`WanOuttraffic`、`WanIntraffic`）消费该接口，需尽快补齐 `vendor=aws` 以减少跨云差异和调用侧改造成本。

## What Changes

- 为 `vendor=aws` 扩展 CVM 监控接口能力，保持 cloud-server 对外参数风格与现有厂商一致。
- 在 `vendor=aws` 下新增 AWS 专有时间参数（UTC ISO8601）并补充对应出参扩展字段，返回结构尽量与当前保持一致。
- 在 cloud-server 的 `getMonitorData` 增加 AWS 分支，按 `account_id + region` 分组调用 hc-service。
- 在 hc-service 新增 AWS CVM 监控路由与处理逻辑，调用 adaptor 层 CloudWatch 能力拉取网络流量指标。
- 在 `pkg/adaptor/aws` 新增 CVM 监控封装，优先复用现有 `pkg/adaptor/aws/cloudwatch.go` 的 `GetMetricData` 能力查询 `NetworkIn/NetworkOut`。
- 按 Phase 1 口径实现 AWS 流量语义：`Lan*` 与 `Wan*` 均基于实例总流量（不做内外网精确拆分），通过扩展字段明确该语义。

## Capabilities

### New Capabilities

- `aws-cvm-monitor-data`: 为统一 CVM 监控入口新增 AWS 厂商支持，覆盖接口契约、cloud-server/hc-service 调用链路和 adaptor 查询封装，保持 AWS 指标单位与语义不变。

### Modified Capabilities

- 无

## Impact

- 影响接口文档：`docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`
- 影响 cloud-server：`cmd/cloud-server/service/cvm/monitor.go` 及相关 API 请求/响应结构
- 影响 hc-service：新增 AWS CVM monitor 路由、请求响应协议与 handler 逻辑
- 影响 adaptor：`pkg/adaptor/aws` 新增 CVM monitor 查询实现，复用/衔接现有 CloudWatch 查询能力
- 指标处理策略：AWS 流量指标保持云厂商原始单位和语义，不在 HCM 内做 Mbps 转换
- 影响 API 类型与 client：`pkg/api/cloud-server/cvm`、`pkg/api/hc-service/cvm`、`pkg/client/hc-service` 等 AWS 监控链路相关定义
