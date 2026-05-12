## Why

标准运维平台调用 HCM 执行 CLB 相关批量操作时，请求会经过 `api-server`、`cloud-server`、`task-server`、`hc-service`、`adaptor` 和数据服务等多个环节，目前缺少按链路层次拆解的请求量、失败量、耗时和任务进度监控。业务反馈任务执行慢或失败时，现有指标难以在短时间内判断问题位于提交入口、异步调度、具体 action、云厂商 API、数据层还是业务任务明细。

本次变更基于 TAPD 需求 `1069995598133689383` 和 iWiki 技术方案《标准运维调用链路-细粒度监控技术方案》，补齐标准运维 CLB 调用链路的细粒度 metrics，为告警、看板和慢任务排查 SOP 提供统一数据口径。

## What Changes

- 新增统一请求指标，覆盖 `api-server`、`hc-service`、`data-service` 的 HTTP 请求和 `adaptor` 云厂商 API 调用，统一输出请求耗时、请求总量和失败总量。
- 新增 CLB submit 入口指标，按 `bk_biz_id`、`vendor`、`operation_type` 统计提交阶段的请求量、失败量和耗时，确保不同 CLB 操作类型隔离统计。
- 在 CLB Flow 创建阶段写入 `ShareData["bk_biz_id"]`、`ShareData["vendor"]`、`ShareData["operation_type"]`，用于日志、任务上下文和必要时的关联排查。
- 扩展 task-server 通用异步框架指标，按 `flow_name` 和 `action_name` 统计 Flow/Task 执行耗时与失败量；通用异步指标不携带 CLB 业务标签，避免非 CLB Flow 产生大量 `unknown` 维度。
- 新增 `task_manage` 与 `task_detail` 业务任务指标，按 `bk_biz_id`、`operation`、`vendor`、`state`、`task_action_id` 等业务维度统计端到端执行耗时、明细执行耗时、失败量和状态分布。
- 明确 metrics 标签约束：禁止使用原始 URL、rid、任务 ID、实例 ID、错误信息原文等高基数字段作为标签；同名指标必须使用固定 label set。
- 本期不改变异步任务框架核心调度模型，不新增 Flow/Task 持久化字段，不把主机重装、主机重启、安全组绑定/解绑 CLB 纳入首期业务指标范围；这些接口只受益于通用请求指标。

## Capabilities

### New Capabilities

- `sops-request-metrics`: 标准运维调用链路的统一请求指标能力，覆盖服务入口请求与云厂商 API 调用，并约束固定指标名、标签集合、错误类型和高基数标签禁用规则。
- `clb-submit-metrics`: CLB submit 阶段的业务入口监控能力，按 `bk_biz_id`、`vendor`、`operation_type` 隔离统计提交请求量、失败量和耗时，并在 Flow 创建上下文中传递业务维度。
- `async-flow-task-metrics`: task-server 通用 Flow/Task 执行监控能力，按 `flow_name`、`action_name` 统计执行耗时和失败量，辅助定位异步调度与 action 执行瓶颈。
- `task-progress-metrics`: `task_manage` 与 `task_detail` 的业务进度监控能力，支持按业务、操作类型、云厂商和任务明细动作观测端到端耗时、明细耗时、失败量和状态分布。

### Modified Capabilities

无。本次变更不修改现有 `openspec/specs/` 下的正式能力契约。

## Impact

- 影响服务与模块：
  - `cmd/api-server`、`cmd/hc-service`、`cmd/data-service` 的统一 HTTP 入口或 middleware/filter。
  - `pkg/adaptor` 或各云厂商 SDK 包装层的云 API 调用打点。
  - `cmd/cloud-server/service/load-balancer/import_excel_submit.go` 的 CLB submit 入口。
  - `cmd/cloud-server/logics/load-balancer` 下 CLB ImportExecutor 创建 Flow、任务管理和任务详情的流程。
  - `pkg/async/consumer/metrics.go`、`executor.go`、`scheduler.go` 的异步框架指标定义与打点。
  - `cmd/cloud-server/service/task/task_timing.go` 以及 `task_manage/task_detail` 状态更新路径。
- 影响 API 行为：不改变现有业务接口请求/响应结构，不改变异步任务调度语义。
- 影响监控系统：新增 Prometheus 指标及看板/告警查询口径，需保证同名指标 label set 固定且控制标签基数。
- 影响排障流程：支持按照提交入口、异步框架、执行层、业务任务主单与明细四层定位 CLB 标准运维慢任务和失败任务。
