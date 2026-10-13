# task-progress-metrics Specification

## Purpose
TBD - created by archiving change add-sops-fine-grained-metrics. Update Purpose after archive.
## Requirements
### Requirement: task_manage 终态耗时指标
系统 MUST 在 `task_manage` 主任务进入终态并完成状态更新后上报 `async_task_manage_exec_cost_seconds`。指标 label MUST 为 `bk_biz_id,vendor,operation,state`，耗时口径 MUST 为主任务创建时间到终态更新时间。

#### Scenario: 主任务成功终态
- **WHEN** `task_manage` 从 running 刷新为 success 或 deliver_partial 等终态
- **THEN** 系统按 `bk_biz_id,vendor,operation,state` 记录主任务端到端耗时

#### Scenario: 主任务失败终态
- **WHEN** `task_manage` 从 running 刷新为 failed、cancel 或 timeout 等终态
- **THEN** 系统按 `bk_biz_id,vendor,operation,state` 记录主任务端到端耗时

### Requirement: task_manage 失败计数指标
系统 MUST 在 `task_manage` 主任务进入失败、取消、超时或非预期终态后上报 `async_task_manage_fail_total`。指标 label MUST 为 `bk_biz_id,vendor,operation,state,err_type`。

#### Scenario: 主任务失败计数
- **WHEN** `task_manage` 进入 failed、cancel、timeout 或其它非成功终态
- **THEN** 系统按业务、云厂商、操作、终态和错误类型增加失败计数

### Requirement: task_detail 终态耗时指标
系统 MUST 在 `task_detail` 明细任务进入终态并完成状态更新后上报 `async_task_detail_exec_cost_seconds`。指标 label MUST 为 `bk_biz_id,vendor,operation,state`，首期耗时口径 MUST 使用 `updated_at - created_at`。`task_action_id` 是按任务实例分配的 UUID，会破坏「禁止高基数标签」要求，因此 MUST NOT 作为 label，由 `operation` 提供类型级分组。

#### Scenario: 明细任务成功终态
- **WHEN** `task_detail` 进入 success 终态并完成状态更新
- **THEN** 系统按 `bk_biz_id,vendor,operation,state` 记录 `updated_at - created_at` 耗时

#### Scenario: 明细任务失败终态
- **WHEN** `task_detail` 进入 failed、cancel 或 timeout 等终态并完成状态更新
- **THEN** 系统按 `bk_biz_id,vendor,operation,state` 记录 `updated_at - created_at` 耗时

### Requirement: task_detail 失败计数指标
系统 MUST 在 `task_detail` 明细任务进入失败、取消、超时或非预期终态后上报 `async_task_detail_fail_total`。指标 label MUST 为 `bk_biz_id,vendor,operation,state,err_type`。`task_action_id` 同样 MUST NOT 作为 label。

#### Scenario: 明细任务失败计数
- **WHEN** `task_detail` 进入 failed、cancel、timeout 或其它非成功终态
- **THEN** 系统按业务、云厂商、操作、终态和错误类型增加失败计数

### Requirement: 首期不实现任务状态存量 Gauge
系统 MUST 在首期只实现 `task_manage/task_detail` 的终态耗时和失败计数指标，不得把状态存量 Gauge 与周期聚合作为首期必需交付项。状态存量 Gauge MUST 作为后续迭代能力单独设计。

#### Scenario: 首期任务指标交付
- **WHEN** 首期标准运维细粒度监控上线
- **THEN** 系统提供终态耗时和失败计数指标，不要求提供 pending/running/success/failed/cancel 状态存量 Gauge

#### Scenario: 后续状态聚合扩展
- **WHEN** 后续迭代需要按状态查看任务存量趋势
- **THEN** 系统可新增周期聚合 Gauge，但不得改变首期已定义的终态耗时和失败计数指标语义

