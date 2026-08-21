# async-flow-task-metrics Specification

## Purpose
TBD - created by archiving change add-sops-fine-grained-metrics. Update Purpose after archive.
## Requirements
### Requirement: Flow 终态耗时指标
系统 MUST 在 task-server 中按 Flow 终态上报 `flow_exec_cost_seconds`。指标 label MUST 为 `flow_name,state`，统计对象 MUST 为 Flow 从开始执行或可可靠获取的创建时间到终态的耗时，成功与失败终态均 MUST 记录耗时。

#### Scenario: Flow 成功终态
- **WHEN** Flow 进入成功终态并完成状态更新
- **THEN** 系统按 `flow_name,state=success` 记录 `flow_exec_cost_seconds`

#### Scenario: Flow 失败终态
- **WHEN** Flow 进入失败、取消或超时等非成功终态并完成状态更新
- **THEN** 系统按 `flow_name,state` 记录 `flow_exec_cost_seconds`

### Requirement: Flow 失败计数指标
系统 MUST 在 Flow 进入非成功终态后上报 `flow_fail_total`。指标 label MUST 为 `flow_name,state`。Flow 级别不携带 `err_type`，因为 Flow 终态由「树内某个 task 失败」推导而来，准确归因需遍历 tree 或回查 backend，代价较高且会与 `task_fail_total` 重复；精准错误类型 MUST 通过 `task_fail_total{action_name,state,err_type}` 下钻查询。

#### Scenario: Flow 执行失败
- **WHEN** Flow 因任务失败、超时、取消或框架错误进入非成功终态
- **THEN** 系统按 `flow_name,state` 增加 `flow_fail_total`，错误类型不在该指标的 label 中体现

#### Scenario: 错误类型下钻
- **WHEN** 需要查询 Flow 失败的具体错误类型分布
- **THEN** 用户 MUST 通过 `task_fail_total` 按对应 `action_name` / `err_type` 进行查询，而非直接在 `flow_fail_total` 上聚合

### Requirement: Task action 执行耗时指标
系统 MUST 在 task-server 执行 `act.Run()` 前后上报 `task_exec_cost_seconds`。指标 label MUST 为 `action_name,state`，每次 action 执行尝试 MUST 记录一次耗时，失败重试 MUST 按多次执行样本记录。

#### Scenario: Task action 成功
- **WHEN** `act.Run()` 成功返回并进入成功状态更新流程
- **THEN** 系统按 `action_name,state=success` 记录 `task_exec_cost_seconds`

#### Scenario: Task action 失败重试
- **WHEN** `act.Run()` 返回错误且该 action 后续发生重试
- **THEN** 系统对每次执行尝试分别记录 `task_exec_cost_seconds`

### Requirement: Task action 失败计数指标
系统 MUST 在 Task action 执行失败时上报 `task_fail_total`。指标 label MUST 为 `action_name,state,err_type`，并不得携带 `bk_biz_id`、`vendor` 或 `operation_type` 等业务维度。

#### Scenario: Task action 执行失败
- **WHEN** `act.Run()` 返回错误、上下文取消、超时或任务进入失败终态
- **THEN** 系统按 `action_name,state,err_type` 增加 `task_fail_total`

#### Scenario: 通用异步指标不携带业务标签
- **WHEN** CLB Flow 或非 CLB Flow 上报 Flow/Task 通用指标
- **THEN** 指标标签仅包含规定的框架维度，不包含 `bk_biz_id`、`vendor` 或 `operation_type`

### Requirement: 保持既有异步指标兼容
系统 MUST 保持既有异步指标继续可用，不得重命名或删除 `hcm_async_task_init_queue_size`、`hcm_async_flow_type_running_num`、`hcm_async_flow_type_exec_duration_milliseconds`。新增 Flow/Task 指标 MUST 使用明确的 `_seconds` 后缀表达耗时单位。

#### Scenario: 既有看板继续查询旧指标
- **WHEN** 现有看板或告警继续查询旧异步指标
- **THEN** 系统仍暴露旧指标，新增指标不破坏旧查询

