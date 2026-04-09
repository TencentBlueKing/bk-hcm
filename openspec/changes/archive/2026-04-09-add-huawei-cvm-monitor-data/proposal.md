## Why

当前“获取云主机监控数据”接口仅支持 `vendor=tcloud`，导致华为云主机无法通过统一入口查询监控指标，前端与调用方需要做厂商分叉处理。随着多云运营场景扩大，需要在保持现有接口契约不变的前提下补齐华为云能力，降低接入复杂度并保证资源视角监控能力一致。

## What Changes

- 扩展资源视角接口 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data`，新增 `vendor=huawei` 的受理与返回支持。
- 在 `cloud-server` 的 `getMonitorData` 路由分发逻辑中新增华为云分支，沿用现有鉴权、实例校验、按 `(account_id, region)` 分组与结果聚合模式。
- 在 `hc-service` 新增华为云监控查询入口（与腾讯云 `GetTCloudMonitorData` 对齐），用于承接 `cloud-server` 内部调用。
- 在 `pkg/adaptor/huawei` 新增监控数据查询实现，基于华为云 CES `BatchListMetricData` 完成请求组装、时间与指标参数映射、响应转换。
- 补充/更新接口文档，明确 `vendor` 枚举新增 `huawei`，并说明与华为云指标、时间粒度、聚合方式相关的约束和兼容行为。

## Capabilities

### New Capabilities

- `huawei-cvm-monitor-data`: 支持通过统一云主机监控接口查询华为云 CVM（ECS）监控数据，并返回与现有结构一致的数据点集合。

### Modified Capabilities

- 无

## Impact

- 影响服务层：
  - `cmd/cloud-server/service/cvm/monitor.go`（新增华为云分支与聚合调用逻辑）
  - `cmd/hc-service/service/cvm/huawei.go`（新增华为云监控查询对外接口）
- 影响适配层：
  - `pkg/adaptor/huawei/monitor.go`（新增华为云监控查询实现）
- 影响 API 与文档：
  - `docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`（vendor 支持范围与说明更新）
- 影响内部协议：
  - 可能新增或扩展 `pkg/api/hc-service/cvm` 与 `pkg/adaptor/types/cvm` 的华为云监控请求/响应结构。
- 兼容性：
  - 对既有 `tcloud` 行为保持兼容；新增能力不改变现有请求路径与响应主结构，属于向后兼容扩展。
