## Why

`async_flow` / `async_flow_task` 两张表自异步框架上线以来**从未有过任何清理机制**——代码里只有未被调用的 DAO 层 `DeleteWithTx`，没有定时任务、没有归档、没有保留期配置，数据只进不出。其中账单分账的 `bill_main_account_summary` flow 因主账号 controller 每 10 分钟轮询、且一条 summary flow 结束后立即再创建一条重算 flow，持续不断地产生新记录，占 `async_flow_task` 全表 70% 以上，运维一次手工清理就删掉了 5,583,462 行。

再不做，两张表会持续膨胀，慢查询与 DDL 风险上升，最终只能靠 DBA 手工大批量 delete 兜底，而手工 delete 本身又有把 MySQL slave 拖挂的风险，属于用运维风险换存储空间的不可持续状态。

## What Changes

- **新增 task-server 定时清理任务** `async_flow_and_task_cleanup`：按保留期清理白名单内 flow name 的历史 flow 及其名下 task，白名单当前只有 `bill_main_account_summary` 一项。
- **首次将 `pkg/cron` 框架接入 task-server**：该框架当前只有 woa-server 与 agent-server 在用，本次需要在 `cmd/task-server/service/service.go` 中初始化 cron 调度器，并把已有的 `sd`（`serviced.ServiceDiscover` 已内嵌 `serviced.State`）保存进 `Service` 结构体以支持 master 判定，`cmd/task-server/app/app.go` 无需改动。
- **超期判定口径**：以 `async_flow.updated_at` 早于「当前时间 − retentionDays」为唯一依据，**不按 state 过滤**（`success`/`failed`/`canceled` 与 `init`/`pending`/`scheduled`/`running` 同一套保留期，超期非终态记录视为僵尸任务一并删除）。
- **关联删除**：删除 flow 的同时删除其 `flow_id` 名下的全部 task，同批次成对完成，不留孤儿数据；task 不单独判断超期。
- **清理范围以白名单界定**：flow name 过滤用包内常量 `cleanupFlowNames` 做 `IN` 匹配（当前只有 `bill_main_account_summary`），后续纳入新的 flow 类型只需在白名单里追加一项，过滤条件与起点游标的命中判定共用同一份白名单。
- **分批限速**：一律按主键 `id` 批量删除，单批 100 条（需求上限 500，取更小值以压低单事务的 binlog event 数），批间隔 ≥ 100ms，单轮不设总量上限，跑到无可删数据为止。
- **起点游标定位与跨轮续扫**：每个租户在进入删除循环前，先用「只带主键条件」的有界扫描（单窗 500 行、只取 `id` / `name`、不限速）向后定位到第一条命中白名单的记录之前；游标按租户在进程内保留、跨轮续扫，避免每轮都从表头重扫 PK 前缀里越积越多的不命中记录。游标只由定位阶段决定，删除循环的进度不写回，以免跳过保留期内的命中记录。
- **配置化**：`pkg/cc` 新增 `TaskServerSetting.AsyncFlowAndTaskCleanup` 配置段（`enabled` / `intervalMin` / `retentionDays` / `batchIntervalMs`），同步落地 `cmd/task-server/etc/task_server.yaml` 与 `docs/support-file/helm`。单批条数固定为 100，不做成配置项。
- **人工触发**：暴露 cron 框架约定的 HTTP 入口 `POST /api/v1/task/async_flow_and_task/cleanup`，执行与定时触发完全一致的逻辑。
- **防重入**：同一时刻只允许一轮清理执行，定时或人工触发遇到进行中的清理直接跳过并说明原因。
- **可观测性**：本期只做日志（每轮开始 / 每批 / 每轮结束 / 出错），不接入监控指标与告警。
- **不采用** MR 2872 的「按全表条数阈值保留最新 100 万条、只删 `state=success`、不区分 flow_name」策略，由本方案的按时间保留策略取代。
- 非破坏性变更：不涉及 DDL、不改动任何对外业务接口、不涉及前端。

## Capabilities

### New Capabilities

- `async-flow-cleanup`: task-server 上的异步任务历史数据定时清理能力，覆盖清理范围界定、超期判定、flow/task 关联删除、分批限速、master 单点执行、防重入、配置化、人工触发入口与运行日志。

### Modified Capabilities

无。现有 spec 中 `async-flow-task-metrics` 只覆盖异步任务的指标上报，`apply-recommend` 覆盖申领推荐离线统计，两者的既有 Requirement 均不因本次变更而改变行为，因此不产生 delta。

## Impact

**受影响代码**（预计 11 个文件，全部为新增或增量修改，无删除）：

| 层次 | 文件 | 改动性质 |
|------|------|---------|
| 枚举 | `pkg/criteria/enumor/cron_task.go` | 新增 `CronTaskAsyncFlowAndTaskCleanup` |
| 常量 | `pkg/criteria/constant/`（清理相关默认值/上限常量） | 新增 |
| 配置 | `pkg/cc/service.go` | `TaskServerSetting` 新增配置段 + `trySetDefault` + `Validate` |
| 配置落地 | `cmd/task-server/etc/task_server.yaml` | 新增 `asyncFlowAndTaskCleanup` 配置段 |
| 配置落地 | `docs/support-file/helm/values.yaml`、`docs/support-file/helm/templates/taskserver/configmap.yaml` | 同步新增配置段 |
| cron 接入 | `cmd/task-server/service/service.go` | 新增 `initCronTask`、`Service` 结构体持有 `tasks` 与 `sd` |
| cron 任务 | `cmd/task-server/task/async_flow_and_task_cleanup.go` | 新增，实现 `croncore.Task` |
| 清理逻辑 | `cmd/task-server/logics/asyncflowcleanup/`（新包） | 新增清理主逻辑 |
| 路由 | `cmd/task-server/service/capability/capability.go` | Capability 增加 `Tasks` 字段 |
| 路由 | `cmd/task-server/service/controller/controller.go` | 注册人工触发路由与 handler |

**受影响数据**：`async_flow`、`async_flow_task` 两张既有表的存量与增量数据会被物理删除，**表结构不变（无 DDL）**。首轮上线后会有一次数百万行级别的存量消化过程。

**受影响系统**：MySQL（hcm 库）——使用现有 DB 账号，不需要新增授权、不修改 session `binlog_format`。

**不受影响**：对外业务接口、前端、白名单以外的 flow_name 的记录、异步任务调度主流程。

**交付路径**：基于 `tencent/bcc/v1.9.x` 新开分支实现。MR 2872 目标分支 `bcc/v1.8.x` 已停更且状态为 `cannot_be_merged`，仅作实现参考。
