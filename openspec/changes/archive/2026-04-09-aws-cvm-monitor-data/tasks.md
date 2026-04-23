## 1. 接口契约与文档更新

- [x] 1.1 更新 `docs/api-docs/web-server/docs/resource/monitor/list_cvm_monitor_data.md`，补充 `vendor=aws` 支持说明、AWS 专有 UTC 时间参数、AWS Phase 1 语义（`Lan*`/`Wan*` 总流量映射）与扩展字段约定
- [x] 1.2 更新 `pkg/api/cloud-server/cvm/monitor.go` 的 `GetMonitorDataReq`，新增 `vendor=aws` 专有 UTC 时间字段并实现参数校验规则（ISO8601、必填约束、与其他 vendor 互斥）
- [x] 1.3 明确 `MonitorDataPointResp.extensions` 在 AWS 场景的最小字段约束（源指标/语义阶段/流量范围/单位）并在接口类型注释中补充说明

## 2. cloud-server 监控分支改造

- [x] 2.1 在 `cmd/cloud-server/service/cvm/monitor.go` 的 `getMonitorData` 中新增 `enumor.Aws` 分支入口
- [x] 2.2 参照 `getHuaWeiMonitorData` 的分组模式实现 AWS 监控查询逻辑：按 `(account_id, region)` 分组、组装下游请求、聚合回包
- [x] 2.3 完成 AWS 回包到统一 `GetMonitorDataResp` 的映射，确保 `id/ip/region/instance_id/timestamps/values/extensions` 字段兼容现有响应结构
- [x] 2.4 增加关键失败路径日志（缺失实例元数据、分组下游失败、维度映射失败）并保证错误可定位到 account/region

## 3. hc-service AWS monitor 接口实现

- [x] 3.1 在 `cmd/hc-service/service/cvm/aws.go` 注册 AWS CVM monitor 路由（对齐 `/vendors/aws/cvms/monitor/data` 约定）
- [x] 3.2 在 `pkg/api/hc-service/cvm` 新增 AWS monitor 请求/响应协议结构及校验逻辑（包含 UTC 时间参数）
- [x] 3.3 实现 AWS monitor handler：解析请求 -> 初始化 AWS adaptor client -> 调用 adaptor 监控查询 -> 组装标准 `MonitorDataPointResp`
- [x] 3.4 确保 hc-service 错误透传与上下文保留（鉴权失败、区域异常、CloudWatch 查询失败、限流）

## 4. adaptor aws CVM 监控封装

- [x] 4.1 在 `pkg/adaptor/types/cvm` 新增 AWS CVM monitor option/result 类型定义（与现有 monitor data 结构对齐）
- [x] 4.2 在 `pkg/adaptor/aws` 新增 CVM monitor 封装，复用 `pkg/adaptor/aws/cloudwatch.go` 的 `GetMetricData` 实现查询，不重复实现分页归并逻辑
- [x] 4.3 实现 `metric_name` 到 CloudWatch 查询参数映射（Phase 1：`LanOuttraffic`/`WanOuttraffic` -> `NetworkOut`，`LanIntraffic`/`WanIntraffic` -> `NetworkIn`）
- [x] 4.4 完成查询结果映射：保持 AWS 原始值语义和单位，不做 Mbps 转换，并在 `extensions` 中写入语义标识字段

## 5. 客户端接线与集成验证

- [x] 5.1 在 `pkg/client/hc-service` AWS CVM client 中新增 monitor 接口调用方法，接通 cloud-server 到 hc-service 的请求链路
- [x] 5.2 补充/更新单元测试与集成测试：参数校验、vendor aws 分支路由、分组查询、CloudWatch 结果映射、Phase 1 语义扩展字段
