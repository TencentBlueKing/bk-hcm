## Context

标准运维调用 HCM 执行 CLB 批量操作时，请求链路会跨越 `api-server`、`cloud-server`、`task-server`、`hc-service`、`data-service`、`adaptor` 和云厂商 API。当前仓库已有统一 Prometheus 注册能力，核心入口在 `pkg/metrics`，各服务通过 `/metrics` 暴露；REST 层已有 `pkg/rest` 的通用耗时和错误指标，异步框架已有 `pkg/async/consumer/metrics.go` 中的初始化队列长度、运行中 Flow 数和 Flow 类型耗时指标。

现有指标不足以支撑标准运维 CLB 慢任务排查：缺少 submit 阶段按 `operation_type` 隔离的请求量、失败量和耗时，缺少 Task(action) 级执行耗时与失败量，缺少 `task_manage/task_detail` 的业务视角进度与终态指标，也缺少跨服务统一请求指标口径。TAPD 需求和 iWiki 技术方案均强调不改变异步任务框架核心逻辑，优先通过关键节点打点和 `ShareData` 传递上下文完成监控补齐。

相关现有代码落点：

- CLB submit 入口：`cmd/cloud-server/service/load-balancer/import_excel_submit.go`
- CLB ImportExecutor 与任务创建：`cmd/cloud-server/logics/load-balancer`
- 异步框架指标与执行：`pkg/async/consumer/metrics.go`、`executor.go`、`scheduler.go`
- `task_manage/task_detail` 状态刷新：`cmd/cloud-server/service/task/task_timing.go`
- 任务表结构：`pkg/dal/table/task/management.go`、`pkg/dal/table/task/detail.go`
- 统一 metrics 注册与暴露：`pkg/metrics`、`pkg/handler`

## Goals / Non-Goals

**Goals:**

- 为标准运维 CLB 调用链路建立分层监控：入口层、异步框架层、执行层、业务任务层。
- 使用固定 label set 定义 `http_request_*`，覆盖 `api-server`、`hc-service`、`data-service` 的所有服务接口请求和 adaptor 云 API 调用，避免同名指标出现不同标签集合。
- 在 CLB submit 阶段按 `bk_biz_id`、`vendor`、`operation_type` 隔离统计请求量、失败量和耗时。
- 在 Flow 创建上下文中写入 `bk_biz_id`、`vendor`、`operation_type`，支持后续日志和任务关联排查。
- 为 task-server 通用 Flow/Task 指标提供低基数标签设计，避免业务维度污染异步框架指标。
- 为 `task_manage/task_detail` 提供业务视角的终态执行耗时和失败量指标；状态存量 Gauge 放到后续迭代。

**Non-Goals:**

- 不改变现有异步任务调度、重试、回滚、取消和 WatchDog 的核心逻辑。
- 不新增 Flow/Task 数据表字段，不做数据库迁移。
- 不改变现有业务 API 的请求体、响应体和错误码语义。
- 不把主机重装、主机重启、安全组绑定/解绑 CLB 纳入首期业务任务指标；这些接口仅通过通用请求指标获得基础观测能力。
- 不把 rid、任务 ID、Flow ID、实例 ID、监听器 ID、错误信息原文等高基数字段放入 metrics label。

## Decisions

### 0. 指标命名约定（Naming Convention）

所有新增指标遵循全局 `hcm_{subsystem}_{name}` 命名约定，与仓库现有 `hcm_restful_*`、`hcm_cloudapi_*`、`hcm_async_*`、`hcm_version_*` 等指标保持一致。`hcm` 前缀来自 `pkg/metrics/metric.go` 中的 `Namespace = "hcm"` 全局常量，由 Prometheus 客户端按 `Namespace_Subsystem_Name` 自动拼接，不需要在指标定义里显式书写。

因此本 spec 与下文 Decisions / Risks 中出现的 `http_request_*`、`clb_submit_*`、`flow_exec_*`、`task_exec_*`、`async_task_manage_*`、`async_task_detail_*` 均为**逻辑短名**，对应在 `/metrics` 端点上暴露的实际指标名为：

- `http_request_*` → `hcm_http_request_cost_seconds` / `hcm_http_request_total` / `hcm_http_request_fail_total`
- `clb_submit_*` → `hcm_clb_submit_cost_seconds` / `hcm_clb_submit_total` / `hcm_clb_submit_fail_total`
- `flow_exec_*` / `flow_fail_total` → `hcm_async_flow_exec_cost_seconds` / `hcm_async_flow_fail_total`
- `task_exec_*` / `task_fail_total` → `hcm_async_task_exec_cost_seconds` / `hcm_async_task_fail_total`
- `async_task_manage_*` → `hcm_async_task_manage_exec_cost_seconds` / `hcm_async_task_manage_fail_total`
- `async_task_detail_*` → `hcm_async_task_detail_exec_cost_seconds` / `hcm_async_task_detail_fail_total`

PromQL 查询、Grafana 看板、告警表达式 MUST 使用上述完整 `hcm_*` 名称（参见 `tasks.md` PromQL Cookbook）。

### 1. 指标按四层分组，而不是把所有维度放进一组指标

指标分为四层：

1. 统一请求指标：`http_request_cost_seconds`、`http_request_total`、`http_request_fail_total`
2. CLB submit 指标：`clb_submit_cost_seconds`、`clb_submit_total`、`clb_submit_fail_total`
3. 异步框架指标：`flow_exec_cost_seconds`、`flow_fail_total`、`task_exec_cost_seconds`、`task_fail_total`
4. 业务任务指标：`async_task_manage_*`、`async_task_detail_*`

选择分层指标的原因是不同层的排障问题不同：请求指标定位入口和下游依赖，submit 指标定位任务创建阶段，异步框架指标定位调度和 action 执行，业务任务指标定位某个业务和操作类型的端到端影响。把所有标签合并到一组大指标会导致高基数、跨业务污染和查询口径混乱。

备选方案是所有指标都携带 `bk_biz_id/vendor/operation_type/action_name`。该方案排查时看似方便，但非 CLB Flow 无法稳定提供这些字段，会产生大量 `unknown`，并显著增加 Prometheus series 数量，因此不采用。

### 2. `http_request_*` 使用固定 label set

统一请求指标采用以下 label：

- `http_request_cost_seconds`：`component,endpoint,method,vendor`
- `http_request_total`：`component,endpoint,method,vendor`
- `http_request_fail_total`：`component,endpoint,method,vendor,err_type`

`api-server`、`hc-service`、`data-service` 的 `vendor` 固定为 `none`，`endpoint` 使用路由模板，`method` 使用 HTTP 方法。`adaptor` 的 `component` 固定为 `adaptor`，`endpoint` 使用稳定云 API 名称，`method` 固定为 `SDK` 或 `CALL`，`vendor` 使用真实云厂商。

选择固定 label set 的原因是 Prometheus 同名指标必须保持一致标签集合；如果 adaptor 另带 `vendor` 而 HTTP 服务不带，会造成同名指标不兼容。备选方案是 adaptor 使用独立的 `cloud_api_request_*` 指标；当前技术方案要求共用 `http_request_*`，因此通过 `vendor=none` 保持一致。

首期实现要求 `http_request_*` 覆盖 `api-server`、`hc-service`、`data-service` 的所有服务接口，而不是只覆盖标准运维 CLB 链路涉及的核心路由。这样可以保证跨服务请求观测口径一致，避免后续排查其它标准运维调用场景时再次补齐基础指标。

### 3. CLB submit 在入口统一 defer 打点

`ImportSubmit` 进入时记录开始时间，函数返回前统一打 `clb_submit_total` 和 `clb_submit_cost_seconds`；当解码、校验、鉴权、创建 executor 或执行 executor 返回错误时打 `clb_submit_fail_total`，并根据错误归一化 `err_type`。

选择入口统一打点的原因是 submit 指标关心从 cloud-server 接收请求到返回任务创建结果的完整耗时，包括参数处理、鉴权、创建任务管理、创建 Flow、更新任务详情等操作。备选方案是在每个 ImportExecutor 内部分散打点，但会造成多个 executor 重复实现，且难以覆盖入口校验和鉴权错误。

### 4. `bk_biz_id/vendor/operation_type` 通过 `ShareData` 传递

CLB ImportExecutor 创建 Flow 时，将以下字段写入 Flow `ShareData`：

- `bk_biz_id`
- `vendor`
- `operation_type`

这些字段用于日志、必要时通过 Flow 关联业务任务，并为后续扩展保留上下文。task-server 的通用 Flow/Task 指标默认不读取这些字段作为标签。

选择 `ShareData` 的原因是它已经随 Flow 持久化，并在 action 执行时可通过 `ExecuteKit.ShareData()` 读取，不需要改表结构或异步框架模型。备选方案是在 Flow/Task 表新增字段，查询更直接，但涉及模型、存储、迁移和更多调用方改造，本期不采用。

### 5. 异步框架指标保持通用，不携带业务维度

Flow 指标只携带 `flow_name`；Task 指标携带 `flow_name,action_name`；失败指标额外携带 `err_type`。打点位置为：

- `scheduler.executeNext` 中 Flow 进入终态后打 `flow_exec_cost_seconds` 和 `flow_fail_total`
- `executor.runTaskOnce` 中 `act.Run()` 前后打 `task_exec_cost_seconds`，失败时打 `task_fail_total`

选择该方式是因为 `flow_name/action_name` 是异步框架天然稳定字段，能够定位慢 Flow 和慢 action，同时避免 CLB 业务维度进入通用框架指标。备选方案是在所有异步指标读取 `ShareData` 并携带业务标签，但非 CLB 任务不具备统一上下文，会放大标签基数和复杂度。

### 6. `task_manage/task_detail` 承载业务排障维度

CLB 业务维度优先放在 `task_manage/task_detail` 指标中：

- 主任务：按 `bk_biz_id,operation,vendor` 统计端到端耗时，按 `bk_biz_id,operation,vendor,state,err_type` 统计失败。
- 明细任务：按 `bk_biz_id,operation,vendor,task_action_id` 统计执行耗时，按 `bk_biz_id,operation,vendor,task_action_id,state,err_type` 统计失败；首期执行耗时接受使用 `updated_at - created_at` 作为口径。
- 状态分布：首期不实现状态存量 Gauge，后续迭代再通过定时聚合刷新 Gauge，统计 `pending/running/success/failed/cancel` 等状态存量。

选择业务任务层承载 `bk_biz_id` 的原因是 `task_management` 和 `task_detail` 本身已有业务、操作、状态和 Flow 关联信息，符合业务排障视角。备选方案是在 data-service 所有更新接口中统一打业务指标，但 data-service 只能看到写入请求，缺少完整业务语义，且不同调用方的更新含义不完全一致。

### 7. 错误类型归一化为有限枚举

失败计数使用 `err_type`，取值限定为有限集合，例如：

- `timeout`
- `network`
- `cloud_error`
- `hcm_error`
- `invalid_param`
- `auth`
- `cancel`
- `unknown`

`cancel` 是一类独立的"非业务失败"终态（用户主动取消、上游 context 取消），单独成桶可以避免和真正不可分类的 `unknown` 混在一起，便于在告警里把"取消"排除出失败率分子。

选择有限枚举是为了支持聚合查询和告警，避免把错误文本、云厂商错误码或资源 ID 直接放入标签。详细错误仍通过日志和 rid/task_id 关联排查。

### 8. 不复用或重命名已有异步指标

当前已有 `hcm_async_flow_type_exec_duration_milliseconds`，实际打点值为秒，命名和单位存在历史不一致。本次新增 Flow/Task 指标使用明确的 `_seconds` 后缀，不修改旧指标名称和语义，避免影响现有看板或告警。

备选方案是直接修正旧指标名称或单位，但这会造成监控兼容性问题，需要迁移看板和告警，本期不采用。

## Risks / Trade-offs

- [Risk] `bk_biz_id` 作为 label 可能带来较高基数。→ Mitigation：仅在 CLB submit 和 `task_manage/task_detail` 业务指标中使用，通用请求指标和通用异步指标不携带 `bk_biz_id`。
- [Risk] 同名 `http_request_*` 在不同组件打点时 label 不一致。→ Mitigation：定义统一 helper 或统一注册函数，强制 `cost/total` 和 `fail` 使用固定 label set。
- [Risk] `endpoint` 使用原始 URL 会导致高基数。→ Mitigation：HTTP 服务仅使用路由模板，adaptor 仅使用稳定云 API 名称。
- [Risk] 首期不提供 `task_manage/task_detail` 状态存量 Gauge，无法直接通过 metrics 查询 pending/running 存量趋势。→ Mitigation：首期优先保障终态耗时和失败计数；状态存量聚合作为后续迭代单独设计和实现。
- [Risk] 进程重启期间内存中的 Flow 开始时间丢失。→ Mitigation：Flow 耗时优先使用持久化时间字段；如果只使用内存时间，需在指标说明中明确不支持重启后的离线回算。
- [Risk] 分散在多个服务和包中打点，容易出现重复注册或初始化顺序问题。→ Mitigation：所有指标注册走 `pkg/metrics.Register()`，并保持服务 `InitMetrics` 先于路由和异步组件初始化。
- [Risk] task detail 单条耗时缺少独立开始/结束字段，`updated_at - created_at` 可能包含排队或等待时间。→ Mitigation：首期明确接受 `updated_at - created_at` 作为明细任务耗时口径；若后续要求区分真实执行时间和等待时间，再单独评估数据模型扩展。

## Migration Plan

1. 新增 metrics 定义与 helper，先保证指标注册、label set 和错误类型枚举稳定。
2. 在 CLB submit 入口补 `clb_submit_*`，并在 Flow 创建请求中写入 `ShareData` 业务上下文。
3. 在 task-server consumer 层补 Flow/Task 执行耗时和失败指标，保持原有调度逻辑不变。
4. 在 `task_manage/task_detail` 终态更新路径补业务任务耗时和失败指标；状态存量 Gauge 暂不纳入首期。
5. 在 `api-server/hc-service/data-service/adaptor` 接入 `http_request_*`，其中服务端指标覆盖所有服务接口，adaptor 覆盖云 API 调用。
6. 发布后先以看板观察指标是否有异常 series 增长，再启用失败率、P95/P99 慢任务等告警。

回滚策略：

- 所有变更为 metrics 打点和上下文写入，不改变业务接口和任务状态机；如出现异常，可通过回滚代码移除新增打点。
- `ShareData` 新增 key 为向后兼容字段，旧 Flow 不包含这些 key 时按缺失处理，不影响任务执行。
- 看板和告警需在指标稳定后再上线强告警，避免新指标初始化阶段误报。

## Resolved Clarifications

- `task_detail` 执行耗时首期使用 `updated_at - created_at` 作为口径，不新增更精确的开始/结束时间字段。
- `task_manage/task_detail` 首期先实现终态耗时和失败计数，状态存量 Gauge 与周期聚合放到后续迭代。
- `http_request_*` 首期必须覆盖所有服务接口，不限于标准运维 CLB 链路涉及的核心路由。
