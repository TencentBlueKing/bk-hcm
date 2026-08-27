## Context

### 现状：CLB 资源锁的两 flow 结构

腾讯云单个 CLB 不支持并发操作，HCM 通过 `resource_flow_lock` 表串行化。该表主键是 `(res_type, res_id)`，`owner` 字段存持锁的 flowID——**锁的互斥性由 DB 主键保证**，`checkResFlowRel` 只是一次提前失败的预检测。

每次 CLB 异步操作产生两个独立 flow：

```
调用方（cloud-server）
  ├── checkResFlowRel(resID)          预检测，撞锁则 errf.LoadBalancerTaskExecuting
  ├── CreateCustomFlow                主 flow，state=Init
  ├── CreateTemplateFlow              watch flow，state=Pending（返回的 flowID 被丢弃）
  ├── ResFlowLock                     insert resource_flow_lock，owner=主flowID
  └── UpdateCustomFlowState           主 flow Init→Pending

task-server
  主 flow    : Pending →(dispatcher)→ Scheduled →(scheduler)→ Running → Success
  watch flow : Pending →(dispatcher)→ Scheduled →(scheduler)→ Running → 每 500ms 轮询主 flow
                                                                        └→ 终态则 ResFlowUnLock
```

**主 flow 自身不解锁**，解锁完全依赖 watch flow 走完同一条调度链。所以「主 flow 已 Success」与「锁已释放」之间必然存在时间窗。

### 时间窗的四段构成

| 段 | 长度 | 当前可观测性 |
|---|---|---|
| watch flow 等 dispatcher（Pending→Scheduled） | `dispatcher.watchIntervalSec` 一个周期 | 无日志 |
| watch flow 等 scheduler（Scheduled→Running） | ≥ 一个周期，且受打分排序影响 | 仅 `logs.V(3)` |
| watch task 等 worker（initQueue → fast/slow queue → worker） | 取决于 worker 占用 | 无日志 |
| 主 flow 终态 → watch 下次轮询命中 | ≤ 500ms | 无日志 |

第二段有系统性风险：`runScheduledFlow` 每轮只取 `listScheduledFlowLimit = 20` 个 flow，按 `caculateFlowTypeScore` 排序，其中 `norExecTime = 1/(1+execTime)`。watch flow 的实测执行时间约等于被监听主 flow 的时长（它一直轮询到主 flow 终态），execTime 大 → 得分低 → 并发高时 watch flow 被系统性排到后面。叠加 `OperateWatchTimeout = 5 * time.Minute`——单个 watch task 最长占用一个 worker 5 分钟做纯轮询——CLB 操作一多就形成 worker 被 watch task 占满、新 watch flow 更进不来的正反馈。这是「近期频繁报错」而非偶发的一个合理解释，但**当前没有任何日志能验证**。

### 关键约束：rid 无法跨越请求边界

- `exec.kt` 在 `NewExecutor` 时创建一次，**整个 task-server 进程共用一个 rid**。`UpdateTask`、`workerDo` 的失败日志都用它。
- `task.Kit` 由 `listTaskByFlowID` 中的 `kt.NewSubKit()` 派生，而该 `kt` 的根是 scheduler 每轮轮询时 `NewKit()` 新建的。`NewSubKit` 只在父 rid 后追加随机后缀，因此 `task.Kit.Rid` 形如 `<轮询rid>/<flow后缀>/<task后缀>`。
- 结论：**两个 rid 都不是调用方那次 HTTP 请求的 rid**。异步执行阶段的日志无法通过 rid 回溯到标准运维的调用。

## Goals / Non-Goals

**Goals:**

1. 从日志能回答「第二步撞锁时，锁被哪个 flow 持有、该 flow 何时加锁」。
2. 能量化「主 flow 进入终态 → 锁释放完成」的实际延迟。
3. 能把 watch flow 与它监听的主 flow 在日志中关联起来。
4. 能区分「锁还在」与「锁已放但 `resource_flow_rel` 未收尾」，以及识别锁泄漏（解锁删了 0 行）。
5. 能量化 flow 在派发、调度选中、worker 排队三段的等待时长。
6. 所有新增日志带 rid，且带 flowID / resID 作为跨服务关联键。

**Non-Goals:**

- 不改变加锁与解锁的时机、不改变 flow/task 的状态机。
- 不修复该时间窗本身（如把删锁与 flow 终态更新做成原子）。这需要独立评估，留待本变更观测到数据后再提。
- 不调整 `listScheduledFlowLimit`、`OperateWatchTimeout`、flow 优先级配置等调度参数。
- 不新增 Prometheus 指标。已有 `async-flow-task-metrics`、`clb-submit-metrics` 等 spec 覆盖指标侧；本变更只做日志，因为需要的是单次复现的因果链，不是聚合趋势。
- 不改 DB schema、不改对外 API。
- **不改任何函数/方法/接口签名**（含 DAO 层），不改对外错误消息，不做同构实现的合并重构。本变更是纯日志变更，代码结构保持原样。唯一例外见 D1：把丢弃已有返回值的 `_, err :=` 改为接收该返回值用于打日志。
- **不为打日志侵入框架**：不给 `Task`、`Flow`、`InitPayload`、`FlowSlaveOperateWatchOption` 等结构体加字段，不改变数据在框架内的流转方式，不新增 DB 查询。日志只能使用当前作用域已有的变量，详见 D3b。

## Decisions

### D1：撞锁日志复用 `checkResFlowRel` 已有的返回值，不新增查询

`checkResFlowRel` 已经返回 `*corelb.BaseResFlowLock`（含 `Owner`、`Revision.CreatedAt`），但所有调用方都写成 `_, err :=` 丢掉了它。

决定：调用方改为接收该返回值并在撞锁时打印 `owner`、`createdAt`。这是本变更中唯一触及调用点写法的改动——**不改任何签名、不改任何行为**，只是不再丢弃一个已经返回的值。若连这个也不做，`owner` 无从获取，需求「得知道加解锁时的 owner 是哪个 flow」无法满足。

- 备选方案「在 `checkResFlowRel` 内部打日志」：被否。该函数有 cloud-server 的两份同构实现（`logics/load-balancer/res_lock.go` 与 `service/load-balancer/async_target_group_add_rs.go`）以及 CVM 版本，内部打日志会丢失调用方上下文（哪个入口、哪一步），而入口信息正是区分「第一步」与「第二步」的关键。
- 备选方案「新增一次 list 查询」：被否，数据已在手上。
- 同理适用于 `CreateTemplateFlow` 返回的 watch flowID（见 D4）。

### D2：持锁 flowID 只进日志，不改对外错误消息

现状 `errf.Newf(errf.LoadBalancerTaskExecuting, "resID: %s is processing", resID)`，标准运维侧只看到被打码的 resID。

决定：**错误消息保持原样**，持锁 flowID 只出现在日志里。

- 备选方案「错误消息追加 `by flow: %s`」：被否。这是对外行为变更，超出本变更「只补日志」的范围，且可能影响下游对错误串的匹配。
- 代价：提单人在 SOPS 界面仍拿不到 flowID，需要开发查日志。可接受——排查入口是 `resID` + 时间点，用它捞日志即可拿到 owner。
- 如后续确认需要错误消息带 flowID，另开变更评估兼容性。

### D3：锁泄漏的识别放在 watch action 已有的查询结果上，不加查询也不改 DAO 签名

`ResFlowUnLock` 按 `(res_id, res_type, owner=flowID)` 删除。若 owner 不匹配，DAO 的 `dao.Orm.Txn(tx).Delete` 返回 0 行但被 `_` 丢弃 → 无错误、无日志、锁永久泄漏。

关键观察：**这个信息在 watch action 里已经有了**。`processResFlow` 的终态分支第一件事就是 `act.queryResFlowLock(kt, opt)`，其过滤条件正是 `res_id + res_type + owner = opt.FlowID`；查不到时当前直接 `return true, nil` 静默跳过。这个静默跳过恰好等于「主 flow 已终态，但锁不在该 flow 名下」——即 owner 不匹配或已被解锁。

决定：在该静默跳过处补一条 Warn 日志。data-service 侧的 `ResFlowUnLock` 只加常规的成功日志。

- 备选方案「`ResourceFlowLockDao.DeleteWithTx` 签名改为返回 `(int64, error)`」：被否。DAO 接口的结构变更，牵连接口定义与两个调用方。
- 备选方案「`ResFlowUnLock` 事务内删除前先查一次实际 owner」：被否。虽然不改签名，但为打日志新增了 DB 查询，属于侵入业务逻辑；而且这个信息在 watch 侧本来就有，重复查一次没有必要。
- 覆盖面差异：watch action 是解锁的唯一主路径，该处日志覆盖了实际关心的场景。若日后发现有绕过 watch 的解锁调用方也需要观测，再单独评估。
- 为何不把不一致当错误返回：解锁可能被重复调用（watch retry），查不到记录是幂等场景下的正常结果。变成错误会让 watch task 失败重试，反而放大问题。

### D3b：拿不到的字段不打，用两条日志的时间戳差替代

`InitPayload.entryTime` 记录了 task 进入 initQueue 的时刻，但它只在 `watchInitQueue` / `initWorkerTask` 的作用域内可见——`workerDo(task)` 只拿到 `*Task`。要在 `workerDo` 里算出 fast/slow 队列的等待时长，就得把时间戳挂到 `Task` 结构体上。

决定：**不加字段**。改为在 `initWorkerTask` 推入队列前打一条日志（含目标队列），`workerDo` 已有的 `start execute task` 日志作为另一端，两条日志的时间戳差即队列等待时长。initQueue 自身的等待时长在 `watchInitQueue` 里用现成的 `payload.entryTime` 打出。

- 这条决策是本变更的通用原则：**为打日志侵入框架数据结构是被禁止的**。凡是当前作用域拿不到的字段，一律改用「两条日志夹一段时间」或「同一 flowID 的另一条日志」来还原，不惜牺牲单条日志的自解释性。
- 同理，`updateFlowStateAndReason` 遇 CAS 失败时不查库中实际状态，只打期望的源/目标；实际状态由同一 flowID 的其他状态更新日志还原。
- 代价：需要按 flowID / taskID 聚合多条日志才能得到完整链路，无法靠单条日志下结论。可接受——排查本来就是按 flowID 捞全量日志。

### D4：watch flow 与主 flow 的关联在创建侧建立，而非执行侧

`CreateTemplateFlow` 的返回值（watch flowID）在所有 `createFlowTask` 里都被 `_` 丢弃。watch action 内部只知道被监听的主 flowID，不知道自己所属的 watch flowID（`FlowSlaveOperateWatchOption` 里没有这个字段，也拿不到）。

决定：在创建侧打一条 Info 日志记录 `watchFlowID ↔ mainFlowID ↔ resID` 的对应关系。执行侧只打主 flowID 与 resID，靠创建侧那条日志做关联。

- 备选方案「把 watch flowID 透传进 `FlowSlaveOperateWatchOption`」：被否。watch flow 创建时其自身 flowID 尚未生成（参数在创建请求里），做不到；要做需改异步框架给 action 注入所属 flowID，超出本变更范围。

### D5：解锁延迟的耗时口径

在 `processResFlow` 观察到主 flow 处于终态时取当前时间，与 `flowInfo.UpdatedAt`（主 flow 变为终态的时刻）相减，得到「终态 → 被 watch 观察到」的延迟；解锁完成后再打一次「终态 → 解锁完成」的总延迟。

- 这两个数正好覆盖需求怀疑的窗口。分成两段是为了区分「watch 没被调度」与「解锁本身慢」。
- 权衡：`UpdatedAt` 是 DB 的 `on update current_timestamp`，精度为秒，且与应用机器有时钟偏差。作为量级判断（毫秒级 vs 十秒级 vs 分钟级）足够，不用于精确计时。这一点需在日志字段命名上体现（如 `terminalToUnlockSec`），避免误读为精确值。

### D6：双 rid 打印，沿用代码中已有的 `exeRid` / `rid` 约定

`executor.go` 中已存在双打先例：`... exeRid: %s, taskRid: %s`（`workerDo` 的 patch 失败分支）。

决定：统一为 `exeRid: %s`（`exec.kt.Rid`，标识执行器实例）+ `rid: %s`（`task.Kit.Rid`）。按项目日志规范，`rid: %s` 必须位于消息末尾，因此 `exeRid` 放在其前面。

- 为何不只保留 `task.Kit.Rid`：`exeRid` 在多副本部署下用于确认是哪个 task-server 进程处理的，与 `worker` 字段互相印证，对排查「flow 被派发到某节点后卡住」有价值。
- 由于 D 段约束（rid 不跨请求边界），**所有新增日志必须同时带 flowID 与 resID**，这是硬性要求而非可选项，写入 spec。

### D7：日志级别与量级控制

异步框架的轮询路径是高频路径。分级原则：

| 场景 | 级别 | 理由 |
|---|---|---|
| 加锁成功、解锁成功、watch 进入/退出、flow 派发、状态更新成功 | `Info` | 每个 flow 生命周期内次数有限（个位数） |
| 解锁时库中 owner 与请求 flowID 不一致或记录不存在、撞锁、CAS 未更新、watch 空转等待 | `Warn` | 异常但流程继续 |
| watch 的 500ms 轮询每一跳、scheduler 每轮的候选列表 | `logs.V(3)` | 频率过高，仅调试时开启 |
| 撞锁导致 return | `Error` | 流程终止 |

- 「每 flow 个位数条 Info」的量级是可接受的：本身每个 CLB 操作已经在打若干条日志。
- watch 的轮询循环**不得**每跳打 Info，否则单个 watch task 5 分钟能打 600 条。只在状态发生变化或首次进入时打。

## Risks / Trade-offs

- **[日志量激增]** → 严格按 D7 分级；轮询循环内一律 `V(3)`；上线后先观察 task-server 日志量再决定是否收紧。
- **[改动范围失控成重构]** → 这是本变更最主要的风险。红线写在 proposal 的 Impact 与本文 Non-Goals：不改签名、不加结构体字段、不加查询、不改错误消息、不合并同构实现。评审时逐个 diff 核对，出现其中任一即视为超范围。
- **[日志字段不自解释，需聚合多条才能定位]** → D3b 的直接代价。缓解：所有新增日志强制带 flowID 与 resID（见 spec 的公共字段约定），保证按 flowID 一次捞全；tasks 的 9.7 要求实测一次两步流程，人工确认日志确实能拼出完整链路。
- **[`UpdatedAt` 秒级精度与时钟偏差导致耗时读数失真]** → 字段命名体现量级语义；结论只依据量级差异，不依据具体数值。若量级判断也不可靠，后续再考虑在 flow 表或 ShareData 中记录应用侧终态时间戳。
- **[观测到数据后才能定位，本变更本身不修复问题]** → 这是有意的取舍，需求本身要求「补日志到线上看是否复现」。但要接受可能需要等待一次线上复现；因此日志必须一次到位，覆盖四段窗口，避免补了一轮还是缺关键字段。
- **[cloud-server 存在 `lockResFlowStatus` / `checkResFlowRel` 的多份同构实现]**（`logics/load-balancer`、`service/load-balancer`、`logics/cvm`）→ 本变更不做合并重构（团队规则：避免大范围重构），但必须三处都改到，否则会出现「有的入口有日志、有的没有」的排查陷阱。tasks 中逐处列出。

## Migration Plan

无 DB 变更、无 API 变更、无签名变更，纯日志改动，可直接随版本发布。

回滚策略：改动全部是增量日志语句（外加一次事务内只读查询），回滚即回退代码，无数据残留、无兼容性问题。

## Open Questions

1. 是否需要同时在 `resource_flow_lock` 表增加一列记录加锁来源（入口 API / SOPS 任务 ID），以便把锁直接关联到调用方？本变更范围内不做（涉及 schema），但如果单靠 flowID 仍无法回溯到 SOPS 的哪一次调用，可能是后续必要项。
2. 观测到延迟数据后的修复方向：主 flow 末尾内联解锁 / flow 终态更新与删锁同事务 / 给 `FlowSlaveOperateWatch` 单独的 worker 池。需另开变更评估。
3. 是否需要一个巡检任务主动告警「flow 已终态但锁仍存在超过 N 秒」？`cmd/cloud-server/service/task/task_timing.go` 的 `isFlowDone` 已经在查这个组合条件，可复用其逻辑。本变更不做，但值得确认是否应纳入。
