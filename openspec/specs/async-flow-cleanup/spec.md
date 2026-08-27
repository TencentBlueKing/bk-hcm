# Spec: async-flow-cleanup

## Purpose

task-server 上的异步任务历史数据定时清理能力。覆盖清理范围界定、超期判定、flow 与 task 关联删除、分批限速、master 单点执行、防重入、配置化、人工触发入口与运行日志。

`async_flow` / `async_flow_task` 两张表自异步框架上线以来没有任何清理机制，数据只进不出。其中账单分账的
`bill_main_account_summary` flow 持续不断地产生新记录，占 `async_flow_task` 全表 70% 以上。本能力按保留期
自动清理该 flow 的历史数据，替代此前依赖 DBA 手工大批量 delete 的兜底方式。

本能力不涉及任何 DDL，不改动对外业务接口。清理范围由代码内维护的 flow name 白名单界定，白名单当前只有
`bill_main_account_summary` 一项，白名单以外的 flow_name 一律不清理。

## Requirements

### Requirement: Scheduled Cleanup Of Expired Async Flow Records

系统 SHALL 在 task-server 上运行一个名为 `async_flow_and_task_cleanup` 的定时任务，按配置周期（默认 60 分钟）触发一轮清理。一轮清理 MUST 以 `async_flow.updated_at` 早于「当前时间 − `retentionDays`（默认 180 天）」作为超期判定的唯一依据，并把超期记录物理删除。保留期内的记录 MUST 一条不删。

#### Scenario: 超期记录被清理

- **GIVEN** `async_flow` 中存在 `name = bill_main_account_summary` 且 `updated_at` 早于 180 天前的记录
- **WHEN** 清理任务执行一轮
- **THEN** 这些 flow 记录被物理删除，且其 `flow_id` 对应的 `async_flow_task` 记录同时被删除，两张表中不残留孤儿数据

#### Scenario: 保留期内的记录不被清理

- **GIVEN** `async_flow` 中存在 `name = bill_main_account_summary` 且 `updated_at` 在 180 天以内的记录
- **WHEN** 清理任务执行一轮
- **THEN** 这些记录全部保留，一条不删

#### Scenario: 无可清理记录时正常结束

- **GIVEN** 表中不存在任何满足清理条件的记录
- **WHEN** 清理任务执行一轮
- **THEN** 任务正常结束、不报错、两张表数据无变化，并输出「本轮删除 0 条」的日志

#### Scenario: 保留天数可通过配置调整

- **GIVEN** 配置项 `retentionDays` 被设置为 `30`
- **WHEN** 清理任务执行一轮
- **THEN** 超期判定的截止时间为「当前时间 − 30 天」，`updated_at` 早于该时间的目标 flow 被删除

### Requirement: Cleanup Scope Restricted To Whitelisted Flow Names

清理对象 MUST 限定在代码内维护的 flow name 白名单之内，白名单当前只有 `bill_main_account_summary`（`enumor.FlowBillMainAccountSummary`）一项。白名单 MUST 以 `name IN (...)` 条件下推到 flow 查询，元素个数 MUST NOT 超过 `filter.DefaultMaxInLimit`（500）。新增待清理的 flow 类型 MUST 通过在白名单常量中追加实现，MUST NOT 做成配置项。白名单以外的 flow_name 的记录 MUST NOT 被清理，无论多旧。清理 MUST NOT 按 `state` 过滤：`success` / `failed` / `canceled` 与 `init` / `pending` / `scheduled` / `running` 共用同一套保留期。

#### Scenario: 白名单以外的 flow_name 的超期记录被保留

- **GIVEN** 表中存在 `name` 不在白名单内的记录（如 `bill_split_daily`、`obs_sync_bill_item`），且其 `updated_at` 早于 180 天前
- **WHEN** 清理任务执行一轮
- **THEN** 这些记录全部保留，一条不删

#### Scenario: 白名单扩容后新类型一并清理

- **GIVEN** 白名单在代码中追加了一种新的 flow name
- **WHEN** 清理任务执行一轮
- **THEN** flow 查询的过滤条件为 `name IN`（含新追加项），该类型的超期记录与原有类型按同一套保留期一并被清理，无需调整任何配置

#### Scenario: 超期的非终态僵尸任务被清理

- **GIVEN** 存在 `name = bill_main_account_summary`、`state = running` 且 `updated_at` 早于 180 天前的记录
- **WHEN** 清理任务执行一轮
- **THEN** 该 flow 及其名下 task 被删除（超过保留期仍未进入终态的记录视为僵尸任务）

#### Scenario: 不设数据量阈值前置条件

- **GIVEN** 两张表的总行数远低于任何经验阈值（例如全表仅数千行），但其中存在超期的目标 flow
- **WHEN** 清理任务执行一轮
- **THEN** 超期记录照常被删除，系统 MUST NOT 因表体量小而跳过清理

### Requirement: Cascading Deletion Of Flow And Its Tasks

删除一条 flow 时，系统 MUST 连带删除该 flow `flow_id` 名下的全部 `async_flow_task` 记录。task 的超期 MUST NOT 单独判断，一律以其所属 flow 的 `updated_at` 为准。flow 与其 task 的删除 MUST 在同一个数据库事务内成对完成，避免中途失败留下孤儿数据。

#### Scenario: flow 与其名下 task 同事务删除

- **GIVEN** 一批超期的目标 flow，每条 flow 名下均有若干 task
- **WHEN** 清理任务删除这批 flow
- **THEN** 对应的 task 在同一个事务内一并删除，事务提交后两张表中既不存在「flow 已删但 task 残留」，也不存在「task 已删但 flow 残留」

#### Scenario: 名下无 task 的 flow

- **GIVEN** 一条超期的目标 flow 名下没有任何 task 记录
- **WHEN** 清理任务处理该 flow
- **THEN** 只删除 flow 记录，不视为异常，流程继续

#### Scenario: 事务失败时不产生孤儿数据

- **GIVEN** 某一批的删除事务在执行过程中失败
- **WHEN** 事务回滚
- **THEN** 该批的 flow 与 task 均未被删除，两张表保持一致，本轮清理终止并输出 error 日志，等待下一周期重试

### Requirement: Rate-Limited Batch Deletion By Primary Key

所有删除 MUST 以主键 `id` 作为条件执行（`WHERE id IN (...)`），MUST NOT 使用时间范围或 flow_name 直接 delete。单批删除的 flow 记录数 MUST 固定为 100，MUST NOT 做成配置项，且 MUST NOT 超过 `core.DefaultMaxPageLimit` 与 `filter.DefaultMaxInLimit`（两者均为 500）。删除循环 MUST 以主键为游标推进，每批取回后游标推进到本批最后一条 flow 的 id；循环的起点 MUST 取自该租户当轮定位出的起点游标（见 Start Cursor Location And Cross-Round Reuse）。相邻两批之间 MUST 间隔不少于 `batchIntervalMs`（默认 100 毫秒）。单轮 MUST NOT 设置总量上限，持续循环直到没有满足条件的记录为止。

#### Scenario: 单批规模受限

- **GIVEN** 存在 5000 条超期的目标 flow
- **WHEN** 清理任务执行一轮
- **THEN** 每条 delete 语句涉及的 flow 记录数不超过 100 条，全部 5000 条在多批中删完

#### Scenario: 批间隔生效

- **GIVEN** `batchIntervalMs` 为 100
- **WHEN** 清理任务连续执行多批删除
- **THEN** 相邻两批删除之间的实际间隔不小于 100 毫秒，峰值删除速率不超过 5000 行/秒

#### Scenario: 删除语句一律按主键

- **GIVEN** 清理任务正在执行删除
- **WHEN** 审计本轮产生的所有 delete 语句
- **THEN** 每条语句的 where 条件均为主键 `id` 的 IN 列表，不出现按 `updated_at` 或 `name` / `flow_name` 直接 delete 的语句

#### Scenario: 单轮不封顶直到清空

- **GIVEN** 存量约 500 万条超期记录
- **WHEN** 清理任务执行一轮且未发生错误
- **THEN** 本轮持续分批删除直到查不到满足条件的记录才结束，不因累计条数达到某个上限而提前退出

### Requirement: Start Cursor Location And Cross-Round Reuse

每个租户的每一轮清理 MUST 先定位「起点游标」，再以该游标为起点进入删除循环。定位 MUST 使用只带主键游标条件（`id > cursor`）的有界扫描，单窗读取行数固定为 `core.DefaultMaxPageLimit`（500），且只读取 `id` 与 `name` 两个字段；定位阶段 MUST NOT 删除任何数据，也 MUST NOT 施加批间隔限速（限速只约束删除批次）。

扫描 MUST 停在窗口内第一条 name 命中白名单的记录之前，游标取其前一条记录的 id；命中记录本身是否超期 MUST NOT 参与判断，任何命中记录都是停止点。整窗都不命中时游标 MUST 推进到窗口末尾并继续下一窗，已空扫过的记录 MUST NOT 被重扫。

定位出的游标 MUST 按租户在进程内保留，下一轮从该位置续扫，MUST NOT 每轮都回到表头重扫。删除循环轮内推进的游标 MUST NOT 写回该保留位置——删除循环会跨过保留期内的命中记录，用它作为下一轮起点会把这些记录永久跳过。游标是进程内内存态，进程重启后 MUST 重置为空并退化为从表头扫描，这 MUST NOT 影响清理结果的正确性。

#### Scenario: 表头不命中的记录被跳过

- **GIVEN** 某租户 `async_flow` 表主键前缀堆积了大量白名单以外的 flow 记录，之后才出现第一条命中记录
- **WHEN** 清理任务为该租户执行一轮
- **THEN** 定位阶段逐窗向后扫描并跳过这些不命中的记录，删除循环的第一次查询从命中记录的前一条开始，而不是从表头开始

#### Scenario: 起点游标不越过任何命中记录

- **GIVEN** 定位扫描的窗口内出现一条 name 命中白名单但尚未超期的记录
- **WHEN** 定位阶段处理该窗口
- **THEN** 游标停在该记录之前，MUST NOT 越过它；该记录在其超期后的某一轮仍能被扫到并清理

#### Scenario: 下一轮从上一轮的游标续扫

- **GIVEN** 上一轮已为某租户定位出起点游标，且进程未重启
- **WHEN** 下一轮清理为该租户重新定位起点
- **THEN** 定位从上一轮保留的游标继续向后扫描，不回到表头，上一轮已空扫过的前缀不被重复扫描

#### Scenario: 删除进度不污染起点游标

- **GIVEN** 一轮清理定位出起点游标后，删除循环在轮内把游标推进到了更大的主键位置
- **WHEN** 本轮结束
- **THEN** 该租户保留的起点游标仍是定位阶段的结果，未被删除进度顶到更后面

#### Scenario: 扫描到表尾仍无命中记录

- **GIVEN** 某租户表内不存在任何白名单内的 flow 记录
- **WHEN** 清理任务为该租户执行一轮
- **THEN** 定位扫描到表尾结束，游标停在表尾，本轮该租户删除 0 条，任务正常结束且不报错

#### Scenario: 进程重启后游标重置

- **GIVEN** task-server 重启，进程内保留的游标全部丢失
- **WHEN** 重启后的第一轮清理执行
- **THEN** 定位从表头重新开始，清理结果与重启前一致，不出现漏清或重复删除

### Requirement: Master-Only Execution

定时触发的清理 MUST 仅在 master 节点执行。非 master 节点上的定时触发 MUST 直接跳过本轮并输出 info 日志，MUST NOT 产生任何数据变更。

#### Scenario: slave 节点跳过定时清理

- **GIVEN** 当前 task-server 实例不是 master
- **WHEN** 定时器触发清理
- **THEN** 本轮被跳过并输出 info 日志，两张表数据无变化

#### Scenario: master 节点执行定时清理

- **GIVEN** 当前 task-server 实例是 master 且 `enabled` 为 `true`
- **WHEN** 定时器触发清理
- **THEN** 清理逻辑正常执行

### Requirement: Single-Flight Cleanup Execution

同一进程内同一时刻 MUST 只允许一轮清理在执行。上一轮尚未结束时，无论是定时触发还是人工触发，新的触发 MUST 被拒绝并明确说明原因，MUST NOT 并发执行两轮清理。

#### Scenario: 定时触发遇到进行中的清理

- **GIVEN** 一轮清理正在执行中（单轮耗时超过了一个定时周期）
- **WHEN** 定时器再次触发
- **THEN** 新的触发被跳过并输出日志，不出现两轮清理并发执行

#### Scenario: 人工触发遇到进行中的清理

- **GIVEN** 一轮清理正在执行中
- **WHEN** 运维通过 HTTP 入口手工触发
- **THEN** 本次触发被拒绝，接口返回明确的「清理正在进行中」原因

### Requirement: Cleanup Configuration And Validation

清理行为 MUST 通过 task-server 配置文件的 `asyncFlowAndTaskCleanup` 配置段控制，包含 `enabled`（默认 `true`）、`intervalMin`（默认 60）、`retentionDays`（默认 180）、`batchIntervalMs`（默认 100）四项。单批条数固定为 100，MUST NOT 出现在配置段中。缺省项 MUST 按默认值填充。服务启动时 MUST 校验配置：`enabled` 为 `true` 时，`intervalMin` > 0、`retentionDays` > 0、`batchIntervalMs` > 0；校验不通过 MUST 使服务启动失败并输出明确的错误信息。`batchIntervalMs` MUST NOT 设下限——AC-P02 的 100ms 限速由默认值保证，显式调低视为运维自行承担主从延迟风险。`enabled` 为 `false` 时 MUST NOT 校验其余项。配置项 MUST 同步落地到 `cmd/task-server/etc/task_server.yaml` 与 `docs/support-file/helm`（`values.yaml` 与 taskserver `configmap.yaml`）。

#### Scenario: 关闭开关后不执行清理

- **GIVEN** `enabled` 配置为 `false`
- **WHEN** task-server 启动并持续运行
- **THEN** 清理任务不被注册进 cron 调度器，清理逻辑不执行，两张表数据无变化，且调度器不产生该任务的任何空转执行

#### Scenario: 非法周期导致启动失败

- **GIVEN** `enabled` 为 `true` 且 `intervalMin` 配置为 0 或负数
- **WHEN** task-server 启动
- **THEN** 启动失败并输出明确的配置校验错误信息

#### Scenario: 缺省配置取默认值

- **GIVEN** 配置文件中未出现 `asyncFlowAndTaskCleanup` 配置段
- **WHEN** task-server 启动
- **THEN** 启动成功，清理任务以 `enabled=true` / `intervalMin=60` / `retentionDays=180` / `batchIntervalMs=100` 运行

#### Scenario: 关闭开关时不校验其余项

- **GIVEN** `enabled` 为 `false` 且 `intervalMin` 配置为 0（非法值）
- **WHEN** task-server 启动
- **THEN** 启动成功，不因其余配置项非法而报错；且 MUST NOT 因该非法周期使调度退化成忙循环（该配置下清理任务不注册进调度器，`Next()` 不会被调用）

### Requirement: Manual Cleanup Trigger Endpoint

系统 MUST 通过 cron 框架约定的 HTTP 入口暴露人工触发能力，路径为 `POST /api/v1/task/async_flow_and_task/cleanup`。人工触发 MUST 执行与定时触发完全一致的清理逻辑，区别仅在于触发来源。该入口 MUST NOT 新增权限模型，也 MUST NOT 做 IAM 鉴权，沿用 task-server 内部接口的既有约定（task-server 不对外暴露，访问控制由部署网络边界保证）。

#### Scenario: 人工触发不做 IAM 鉴权

- **GIVEN** 调用方通过内部网络访问 task-server
- **WHEN** 调用 `POST /api/v1/task/async_flow_and_task/cleanup`
- **THEN** 系统不执行任何 IAM 权限校验，直接进入清理逻辑，与 task-server 其余内部接口保持一致

#### Scenario: 人工触发执行一轮清理

- **GIVEN** 当前无清理在执行且 `enabled` 为 `true`
- **WHEN** 运维调用 `POST /api/v1/task/async_flow_and_task/cleanup`
- **THEN** 系统执行与定时触发完全一致的清理逻辑，并返回本轮清理结果

#### Scenario: 关闭开关时人工触发

- **GIVEN** `enabled` 配置为 `false`
- **WHEN** 运维调用人工触发入口
- **THEN** 不执行清理，接口明确返回清理功能已关闭

### Requirement: Cleanup Observability Logging

清理过程 MUST 按项目日志规范输出运行日志，所有日志 MUST 携带 `rid`。本期 MUST NOT 接入监控指标与告警。

#### Scenario: 每轮输出开始与结束日志

- **GIVEN** 清理任务执行完一轮
- **WHEN** 查看 task-server 日志
- **THEN** 能看到本轮开始时的保留期截止时间，以及本轮累计删除的 flow 数、task 数与总耗时，且日志携带 `rid`

#### Scenario: 每批输出删除条数

- **GIVEN** 清理任务完成一批删除
- **WHEN** 查看 task-server 日志
- **THEN** 能看到本批删除的 flow 数与 task 数，以及本批结束时推进到的主键游标

#### Scenario: 起点游标定位结果可见

- **GIVEN** 清理任务完成某个租户的起点游标定位
- **WHEN** 查看 task-server 日志
- **THEN** 能看到该租户 id、定位出的游标值与本次定位扫描的窗口数，且日志携带 `rid`

#### Scenario: 出错时输出 error 日志

- **GIVEN** 查询或删除过程中发生错误
- **WHEN** 本轮终止
- **THEN** 按 `logs.Errorf` 输出含错误详情与 `rid` 的日志，等待下一周期重试，不做无限重试

### Requirement: Database Safety Constraints

清理过程 MUST NOT 执行 `set session binlog_format` 或任何需要 DBA 额外授权的语句，MUST 仅使用现有 hcm 数据库账号即可完成全部操作。本能力 MUST NOT 涉及任何 DDL，两张表结构保持不变。

#### Scenario: 不修改 binlog 格式

- **GIVEN** 清理任务正在执行
- **WHEN** 审计本轮产生的所有 SQL
- **THEN** 不出现 `set session binlog_format` 或其他需要额外 DB 授权的语句，全部操作使用现有 hcm 账号完成

#### Scenario: 不涉及表结构变更

- **GIVEN** 本次变更完整上线
- **WHEN** 比对 `async_flow` 与 `async_flow_task` 的表结构
- **THEN** 两张表结构与上线前完全一致，不存在任何 DDL 变更

### Requirement: Non-Interference With Async Scheduling

清理任务 MUST 幂等，重复执行 MUST NOT 产生副作用。单轮失败 MUST NOT 影响后续周期，下一轮从当前状态继续。清理任务的异常 MUST NOT 影响 task-server 的异步任务调度与执行主流程。

#### Scenario: 清理期间调度不受影响

- **GIVEN** 清理任务正在存量环境上执行首轮清理
- **WHEN** 观察 task-server 的异步任务调度与执行
- **THEN** 调度与执行不受影响，无任务积压或超时

#### Scenario: 主从延迟可控

- **GIVEN** 在存量约 500 万条的环境上执行首轮清理
- **WHEN** 观测 MySQL 主从复制延迟
- **THEN** 清理期间延迟不出现持续增长，清理结束后 10 分钟内回落到清理开始前的水平

#### Scenario: 单轮失败后自愈

- **GIVEN** 某一轮清理因 DB 抖动中途失败，已删除的批次不回滚
- **WHEN** 下一个定时周期到达
- **THEN** 新一轮清理从当前状态继续（起点为该租户保留的游标，进程重启后为表头），重复执行不产生副作用，最终仍能把超期数据清理干净
