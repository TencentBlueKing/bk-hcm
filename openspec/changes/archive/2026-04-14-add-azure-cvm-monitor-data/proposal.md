## Why

当前 `POST /api/v1/cloud/vendors/{vendor}/cvms/monitor/data` 已支持 `tcloud`、`huawei`、`aws`，但尚未支持 `azure`，导致多云场景下 Azure 主机流量监控无法统一接入。随着多云流量分析与安全评估需求落地，需要在保持现有接口兼容的前提下，补齐 Azure 端到端监控查询能力并支持厂商专有参数。

## What Changes

- 扩展对外接口文档，新增 `vendor=azure` 的请求参数说明，补充 Azure 专有入参（含 `top`、`orderby`、`filter`、`result_type`）与返回扩展字段说明。
- 扩展 cloud-server 的 CVM 监控聚合逻辑，在 `getMonitorData` 增加 Azure 分支，按 `(account_id, region)` 分组调用 hc-service 并回填统一 `data_points` 结构。
- 扩展 hc-service 的 Azure CVM 监控接口，新增 Azure 监控请求/响应结构及路由处理，调用 adaptor 的 Azure 监控能力。
- 在 `pkg/adaptor/azure` 新增监控查询实现（对齐现有 huawei 监控实现风格），封装 Azure Metrics List 调用与结果转换。
- 统一多云厂商扩展信息放入 `extensions` 字段，Azure 至少返回 `unit`、`cost`、`granularity`、`namespace`、`resource_region` 等扩展信息。
- Azure 网络流量语义采用“原生优先、兜底回退”：若可直接获得内/外网流入流出则直接返回；若当前能力无法直接分离，则按既定兜底语义返回并在 `extensions` 明确语义来源。

## Capabilities

### New Capabilities
- `azure-cvm-monitor-data`: 为 Azure 云主机提供统一监控数据查询能力，覆盖接口入参、跨层调用链路、适配器查询与扩展字段输出。

### Modified Capabilities
- 无。

## Impact

- 受影响 API/文档：
  - `docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`
- 受影响服务链路：
  - `cmd/cloud-server/service/cvm/monitor.go`
  - `cmd/hc-service/service/cvm/`（新增 Azure 监控处理）
- 受影响适配器：
  - `pkg/adaptor/azure`（新增监控查询实现）
  - `pkg/adaptor/types/cvm`（新增 Azure 监控 option/result 结构）
- 风险与兼容性：
  - 保持现有接口路径与主结构不变，对既有厂商行为保持兼容；
  - 新增 Azure 专有参数为可选扩展，不影响非 Azure 请求；
  - Azure 内外网分离能力存在云厂商能力差异，需通过 `extensions` 提供可观测语义标识。
