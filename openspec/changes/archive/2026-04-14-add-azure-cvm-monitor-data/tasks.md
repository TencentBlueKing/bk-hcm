## 1. 接口与文档定义

- [x] 1.1 更新 `docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`，补充 `vendor=azure` 支持说明和专有入参（`metric_namespace`、`aggregation`、`auto_adjust_timegrain`、`top`、`orderby`、`filter`、`result_type`）。
- [x] 1.2 在文档中补充 Azure 返回扩展字段约定，明确 `extensions` 至少包含 `unit`、`cost`、`granularity`、`namespace`、`resource_region`。
- [x] 1.3 在文档中补充 Azure 流量语义策略，明确“原生优先、兜底回退”与 `extensions` 语义标识字段（含 `source_metric_name`、`semantic_phase`、`traffic_scope`、`is_fallback`）。

## 2. cloud-server API 层参数扩展

- [x] 2.1 扩展 `pkg/api/cloud-server/cvm/monitor.go` 的请求结构，新增 Azure 专有参数字段并补充注释。
- [x] 2.2 在 `Validate(vendor)` 中实现 Azure 专有参数校验与非 Azure 场景拦截，保证参数组合合法且错误信息可读。
- [x] 2.3 为新增 Azure 参数校验补充单元测试，覆盖合法参数、非法参数和非 Azure 误传场景。

## 3. cloud-server 服务层 Azure 分支接入

- [x] 3.1 在 `cmd/cloud-server/service/cvm/monitor.go` 的 `getMonitorData` 增加 `vendor=azure` 分支处理。
- [x] 3.2 新增 Azure 分组查询处理函数，按 `(account_id, region)` 分组调用 hc-service Azure 监控接口并聚合结果。
- [x] 3.3 在 Azure 分支回填统一 `MonitorDataPointResp`，确保主结构字段与既有厂商一致，厂商差异放入 `extensions`。

## 4. hc-service API 与服务路由扩展

- [x] 4.1 在 `pkg/api/hc-service/cvm` 新增 Azure monitor 请求/响应结构及 `Validate` 逻辑。
- [x] 4.2 在 `cmd/hc-service/service/cvm` 注册 Azure 监控路由并新增 handler，实现请求解析、参数校验和 adaptor 调用。
- [x] 4.3 在 hc-service 层将 Azure 专有参数完整透传到 adaptor option，并保持与 cloud-server 字段命名一致。

## 5. adaptor 类型与 Azure 监控实现

- [x] 5.1 在 `pkg/adaptor/types/cvm` 新增 Azure 监控 option/result 结构，包含专有参数与标准化输出字段。
- [x] 5.2 在 `pkg/adaptor/azure` 新增 `monitor.go`，封装 Azure Metrics List 调用与请求参数映射。
- [x] 5.3 在 Azure adaptor 中实现监控结果标准化转换，输出 `dimensions`、`timestamps`、`values` 与 `extensions`。
- [x] 5.4 在 Azure adaptor 中实现“原生优先、兜底回退”语义处理逻辑，并在 `extensions` 写入回退标识与语义来源字段。

## 6. 验证与回归

- [x] 6.1 增加/更新单元测试，覆盖 cloud-server 与 hc-service 的 Azure 请求校验、参数透传与错误场景。
- [x] 6.2 增加/更新 adaptor 侧测试，覆盖 Azure 查询结果解析、扩展字段填充、四象限语义回退场景。
- [x] 6.3 执行回归测试，验证 `tcloud`、`huawei`、`aws` 监控查询行为未受影响。
