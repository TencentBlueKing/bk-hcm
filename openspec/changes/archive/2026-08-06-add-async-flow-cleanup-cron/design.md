## Context

### 当前状态

`async_flow` / `async_flow_task` 两张表自异步框架上线以来没有任何清理机制。`pkg/dal/dao/async/flow.go:336` 与
`pkg/dal/dao/async/task.go` 各有一个 `DeleteWithTx`，但**全仓库无调用方**，数据只进不出。账单分账的
`bill_main_account_summary` flow 占 `async_flow_task` 全表 70% 以上，运维一次手工清理删掉 5,583,462 行。

### 已核实的代码事实（本设计的全部前提）

| 事实 | 证据 |
|------|------|
| task-server **尚未接入** `pkg/cron` 框架，本次为首次接入 | `cmd/task-server/service/service.go` 无 `cron.Init` / `cron.Register`；全仓库仅 `cmd/woa-server/service/service.go:703`、`cmd/agent-server/service/service.go:133` 调用 |
| `serviced.ServiceDiscover` 已内嵌 `State`，master 判定无需改 app.go | `pkg/serviced/serviced.go:36`（`ServiceDiscover` 内嵌 `Service`）→ `pkg/serviced/service.go:41`（`Service` 内嵌 `State`）；`service.NewService(sd serviced.ServiceDiscover, ...)` 已经拿到 `sd` |
| cron `Task` 接口只有 4 个方法 | `pkg/cron/core/task.go:30`：`Name() / Next() / Do(kt) / GetURL()` |
| 调度器为「串行 for 循环 + `Next()` 计算下次时间」，非 crontab 表达式 | `pkg/cron/core/scheduler.go:115`（`executeTask` 的无限 for）、`:136`（`executeSingleRun` 先 `Next()` 再 sleep 再 `Do()`） |
| 每次 `Do()` 由调度器构造 backend kit，自带 rid | `pkg/cron/core/scheduler.go:162`：`core.NewBackendKit().NewSubKitWithCtx(s.ctx)` |
| task-server 的 WebService 根路径为 `/api/v1/task` | `cmd/task-server/service/service.go:271` |
| task-server 直接持有 `dao.Set`，异步框架本就直连 DB | `cmd/task-server/service/service.go:96`、`:138` |
| `async_flow` **没有** `name` 或 `updated_at` 索引 | 建表 `scripts/sql/0012_20231130_1604.sql:116`（仅 PK）；后续仅加 `idx_tenant_id`（`0034`）、`idx_worker_state_id` / `idx_state_id`（`0029_20241030_async_table_optimize.sql:36`） |
| `async_flow_task` **有** `idx_flow_id` | `scripts/sql/0023_20240531_1648_bill.sql:359` |
| 两张表 `EnableTenant: true` | `pkg/dal/table/table.go:420-421` |
| `LessThan` 算子原生支持时间字符串 | `pkg/runtime/filter/operator.go:438` `isNumericOrTime`、`:453` `judgeAndParseTime` |
| 分批清理先例：List(fields=id) → BatchDelete(id IN) 循环 | `cmd/woa-server/logics/applyrecommend/logics.go:336` `cleanupExpiredUserRecommend` |
| `core.DefaultMaxPageLimit = 500`，构成单批条数的上限 | `pkg/api/core/page.go:51` |

### 约束

- 不做任何 DDL（需求「本期不包含」明确排除表结构治理）。
- 不修改 session `binlog_format`，不依赖需要 DBA 额外授权的权限（AC-S01）。
- 所有 delete 必须以主键为条件（AC-S02）。
- 清理不得影响 task-server 的异步调度主流程（AC-P04）。
- 交付基于 `tencent/bcc/v1.9.x` 新开分支；MR 2872 仅作参考，其
  「全表条数阈值 + 只删 `state=success` + 不区分 flow_name」策略**不采用**。

### 干系人

运维（清理的直接受益者与配置调整者）、DBA（主从延迟的兜底方）、异步框架的其他使用方（账单、CVM、LB 等，
要求本变更对它们零影响）。

## Goals / Non-Goals

**Goals:**

- 让 `async_flow` / `async_flow_task` 的体量长期可控，无需人工 delete 兜底。
- 存量（百万~千万级）由程序限速自动消化干净，无需人工介入。
- 清理速率受控，不引发 MySQL 主从延迟持续增长。
- 在无索引支撑的前提下，让扫描代价可控、不随清理进度劣化成 O(N²)。
- 首次把 `pkg/cron` 接入 task-server，为后续 task-server 的定时任务铺路。

**Non-Goals:**

- 不清理白名单以外的任何 flow_name（白名单是代码内常量，当前只有 `bill_main_account_summary` 一项，见 D-13）。
- 不做归档 / 冷备 / 转储，纯物理删除。
- 不做表分区、分表、加索引等表结构治理。
- 不接入监控指标与告警（本期只做日志）。
- 不新增前端页面，不改动任何对外业务接口。
- 不做跨进程分布式锁（防重入只保证单进程内，理由见 D-6）。

## Decisions

### D-1：cron 接入点放在 `service.NewService` 内的 `initCronTask`，不改 app.go

**选型**：在 `cmd/task-server/service/service.go` 的 `Service` 结构体上增加 `sd serviced.State` 与
`tasks map[enumor.CronTask]croncore.Task` 两个字段，在 `NewService` 末尾调用新增的 `initCronTask()`，
内部依次 `cron.Init(context.Background(), metrics.Register())` → 构造 task → `cron.Register([]croncore.Task{...})`。

**理由**：完全复刻 woa-server 的既有形态（`cmd/woa-server/service/service.go:702` `initCronTask`），
接入成本最低、评审心智负担最小。且 `NewService` 的入参 `sd serviced.ServiceDiscover` 已内嵌 `State`
（`pkg/serviced/service.go:41`），直接赋值即可满足 master 判定，**app.go 一行都不用改**。

**被否决的替代方案**：

- *在 `app.go` 里另起一个 cron 初始化流程，把 `serviced.State` 显式传进 `NewService`*——多改一个文件、
  多一层参数，且与 woa-server / agent-server 两处先例都不一致，没有收益。
- *用 `robfig/cron` 之类第三方库直接写 crontab 表达式*——引入新依赖，违反「不新增依赖」的团队规则，
  且放弃了 `pkg/cron/metric` 已有的执行次数 / 错误数 / 耗时指标。

### D-2：按主键游标（keyset）分页扫描 flow，而不是 OFFSET 分页

**选型**：每批查询

```
SELECT id FROM async_flow
WHERE id > :lastID AND name IN (:cleanupFlowNames) AND updated_at < :deadline
ORDER BY id ASC LIMIT 100
```

`lastID` 的初值是本轮为该租户定位出的起点游标（首轮、进程重启后为空串，见 D-13），
每批取回后更新为本批最大 id；返回行数 `< 100` 时本轮结束。

**理由**：这是本设计**最关键**的一条。`async_flow` 上没有 `name` / `updated_at` 索引
（见上表证据），任何按这两个字段的过滤都要走扫描。三点决定了必须用 keyset：

1. **PK 有序 + 早停**。`ORDER BY id` 直接沿 PK 走，MySQL 攒够 `LIMIT` 行就停，不会扫全表。
2. **避免 O(N²) 劣化**。若像 `cleanupExpiredUserRecommend` 那样每轮固定从 `start=0` 重扫，
   随着清理推进，PK 前缀里堆积的「已检查过但不匹配」的行（其他 flow_name、保留期内的行）越来越多，
   每批都要重新跨过它们；千万级存量下这会退化成平方级扫描。keyset 让每行在**一轮之内**最多被检查一次。
   轮与轮之间的那段前缀不归 keyset 管——每轮的第一批查询仍会从表头出发跨过它，这部分由 D-13 的起点游标接手。
3. **超期行天然聚集在 id 低位**。`id` 由 id_generator 生成的 36 进制 8 位左补零字符串，单调递增，
   与时间强相关（这也正是 MR 2872 敢用 `id < minID` 的依据）。所以游标从头出发能立刻命中匹配行。

**被否决的替代方案**：

- *OFFSET 分页（`Page.Start += limit`）*——删除会让后续行前移，`start` 持续前进会**漏删**；
  且 OFFSET 本身要求 MySQL 丢弃前 N 行，代价随进度线性上升。
- *固定 `start=0` 重扫（先例 `cleanupExpiredUserRecommend` 的做法）*——正确性没问题，但如上第 2 点，
  在本场景的数据规模下会劣化。先例的数据量是万级、且表上有可用索引，场景不可比。
- *先算一个 `minID`（如 MR 2872 那样按条数定位第 100 万条），再 `DELETE WHERE id < minID`*——
  违反 AC-S02（必须按主键 IN 删，不能用范围条件直接 delete），且口径已被需求改成按时间。
- *加联合索引 `(name, updated_at)`*——最优解，但需求明确「不涉及 DDL」。**记入 Open Questions**，
  建议作为后续独立需求评估。

### D-3：flow 与 task 在同一事务内按主键成对删除

**选型**：每批拿到 flow id 列表后，在一个 `dao.Txn().AutoTxn` 事务里：

1. `AsyncFlowTask().List(fields=["id"], filter: flow_id IN (flowIDs))` 取出该批 flow 名下的全部 task id
   （走 `idx_flow_id`，代价可控）；
2. `AsyncFlowTask().DeleteWithTx(tx, id IN (taskIDs))`；
3. `AsyncFlow().DeleteWithTx(tx, id IN (flowIDs))`。

**理由**：

- 同事务保证 R-005 / AC-001「不残留孤儿数据」；任一步失败整批回滚，两张表始终一致。
- 先删 task 再删 flow：万一事务外发生意外，残留的是「flow 在、task 已删」而非孤儿 task，
  前者可被下一轮重新识别并清理，后者则永远无人认领。
- 全部走 `DeleteWithTx(expr)` + `tools.ContainersExpression("id", ids)`，生成
  `DELETE FROM ... WHERE id IN (...)`，同时满足 AC-S02 与「复用已有 DAO」。

**被否决的替代方案**：

- *不开事务，先删 task 再删 flow*——中途失败会留下孤儿，违反 R-005。
- *`DELETE FROM async_flow_task WHERE flow_id IN (...)`*——省一次查询、且能走 `idx_flow_id`，但
  where 条件不是主键，直接违反 AC-S02。
- *一条 flow 一个事务*——事务数量爆炸（100 倍），提交开销与 binlog 条目数都显著变差。

### D-4：单批条数固定为 100 不做成配置项；单批 task 数可能远超它，需要对 task 再做二次切片

**选型**：单批 flow 条数是代码常量 `flowBatchSize = 100`，**不进配置段**。100 条 flow 名下的 task
总数可能是几百上千条，因此第 2 步取回的 task id 列表要用
`slice.Split(taskIDs, int(filter.DefaultMaxInLimit))` 再切成 ≤500 一组，逐组调 `DeleteWithTx`（仍在同一事务内）。
同理，`flow_id IN (...)` 的 List 也要分片。

**取 100 而不贴着上限跑的理由**：本需求最主要的稳定性诉求是不把从库拖垮。单批越小，单个删除事务在
ROW 格式 binlog 下产生的 event 越少（100 行/事务 vs 500 行/事务），从库 apply 的尖峰越低，
这正是 R-3 要压低的那个尖峰。代价是同样总量下往返次数变成 5 倍、单轮耗时相应拉长，
而单轮本就不封顶、跨周期继续，这个代价可以接受（耗时量级见 R-2）。

上限侧有两道不可逾越的硬约束，构成该常量的取值天花板：
`core.DefaultMaxPageLimit = 500`（`pkg/api/core/page.go:51`）是单页查询条数上限，
`filter.DefaultMaxInLimit = 500`（`pkg/runtime/filter/expression.go:41`）是 IN 列表元素数硬上限，
超出任一都会在表达式校验阶段直接报错。

**不做成配置项的理由**：它与 `batchIntervalMs` 共同决定删除速率，两个旋钮作用重叠——要降速率，
调批间隔与调批大小效果等价。对外只暴露 `batchIntervalMs` 一个，运维不必面对两个相互耦合的参数，
也省掉一套上界校验（配成 >500 会让每条 SQL 在表达式校验阶段被拒，属于「配了就崩」）。

**二次切片的理由**：AC-P01 要求「单次 delete 语句涉及的记录数不超过 500 条」，约束的是 SQL 层面而非
flow 层面；不切片会产出上千行的单条 delete，直接违反 AC-P01，也会放大单条 binlog event。

**被否决的替代方案**：
- *让单批条数同时约束 flow 与 task 总数*——需要先查 task 数再倒推 flow 批大小，实现复杂且每批规模不稳定，
  收益仅是少一次切片。
- *单批贴着上限取 500*——往返次数最少、单轮最快，但单事务的 binlog event 数是现在的 5 倍，
  与 R-3 要压低的尖峰直接冲突。
- *保留 batchSize 配置项、取值范围 1~500*——即上一版设计。它需要 `pkg/cc` 的边界校验、
  运行期的 `normalizeBatchSize` 兜底、以及三处 yaml 的同步维护，换来的调速能力与 `batchIntervalMs` 重叠。

### D-5：限速用「批间固定 sleep」，不用令牌桶

**选型**：每批事务提交后 `time.Sleep(batchIntervalMs)`；sleep 期间用 `select` 同时监听 `kt.Ctx.Done()`，
以便服务优雅退出时能立刻中断。

**理由**：需求把限速直接定义成「单批 ≤500 条 + 批间隔 ≥100ms」（F-003、AC-P01/P02），
固定 sleep 是对该定义的**字面实现**，验收时可直接观测，无需解释算法。实现取 100 条/批，
理论峰值 1000 行/秒，在需求给出的 5000 行/秒上限内留出 5 倍余量。

**被否决的替代方案**：

- *令牌桶 / `golang.org/x/time/rate`*——能做到更平滑的速率控制，但验收标准是按「批间隔」写的，
  令牌桶反而不好证明 AC-P02，且多一层概念。
- *根据主从延迟动态调速*——最理想，但需要读 `SHOW SLAVE STATUS`（额外 DB 权限，违反 AC-S01），
  且本期不接监控。记入 Open Questions。

### D-6：防重入用进程内互斥（`sync.Mutex.TryLock()`），不做分布式锁

**选型**：清理逻辑持有一个 `sync.Mutex`，对外入口 `Cleanup()` 先 `TryLock()`（Go 1.18+，非阻塞），
抢不到就跳过（定时触发打 info 日志并返回 nil）或返回明确错误（人工触发返回 `errf.Aborted`）。
`defer Unlock()` 保证异常路径也能释放。等价实现是 `atomic.Bool` + `CompareAndSwap(false, true)`，
选 `TryLock()` 是因为语义更直白、误用面更小。

**理由**：满足 R-008 / AC-007。定时触发已被 D-7 限定为**仅 master 执行**，集群内同一时刻只有一个节点
会跑定时清理，所以跨进程并发的唯一入口是「运维在 slave 上手工触发」——这是低频的人为操作，
且清理本身幂等（D-8），最坏结果是两轮删除有重叠、DB 压力翻倍，不会产生数据错误。
用一个原子变量换掉一套 etcd 分布式锁，符合 OpenSpec「Simplicity First」。

**被否决的替代方案**：

- *etcd 分布式锁*——严格正确，但引入锁续期、超时、脑裂等一整套复杂度，为一个幂等的清理任务上这套
  机制性价比过低。**记入 Open Questions**，若后续观测到多节点手工触发确实造成压力再补。
- *靠数据库唯一记录做锁*——需要新表或复用 global_config，本质是自己造分布式锁，同上。

### D-7：master 判定只拦定时触发，人工触发不拦

**选型**：`Do()` 开头 `if t.sd == nil || !t.sd.IsMaster() { 打 info 日志; return nil }`；
人工触发 handler **不做** master 校验。

**理由**：AC-006 只约束「定时器触发时 slave 跳过」，AC-009 则要求人工触发「执行与定时触发完全一致的清理逻辑」。
运维手工触发的场景往往就是「表体量突增、想立刻跑一轮」，此时强制要求打到 master 节点既不可控
（服务发现是负载均衡的）也无必要。

**这是对 woa-server 既有形态的有意偏离，不是沿袭**（本条论据在阶段 5 评审中被更正）：
`cmd/woa-server/service/res-sync/apply_recommend.go:38` 的人工触发调的是 `s.tasks[...].Do(cts.Kit)`，
而 `ApplyRecommendOfflineTask.Do` 开头即 `if t.sd == nil || !t.sd.IsMaster() { return nil }`
（`cmd/woa-server/task/apply_recommend.go:68`）——那条先例的人工触发**会**被 master 门禁拦掉，
打到 slave 会拿到 200 但什么都没做。本设计绕开 `Do()` 直接调 `Logics.Cleanup`，正是为了避免这个结果：
若沿用该形态，运维在 slave 上触发会得到一个成功响应和零效果，反而不满足 AC-009。

**被否决的替代方案**：*人工触发也校验 master，非 master 返回错误并提示重试*——运维体验差，
且需求未要求。

**已知局限**：人工触发能绕过防重入（D-6 的 `sync.Mutex` 是进程内锁，挡不住「A 节点定时跑 + B 节点人工触发」）。
但这个缺口源于 D-6 不做分布式锁的选择，与是否判 master 无关——服务发现是负载均衡的，
加 master 判定只会把「可能并发」换成「可能无效」，并不能消除并发。后果可控：删除按主键幂等、
两轮加锁顺序一致，最坏是 DB 压力翻倍。运维手册需写明「人工触发前先确认没有其它节点在跑清理」。

### D-8：单批失败即终止本轮，已删批次不回滚，靠幂等自愈

**选型**：任一批的查询或事务失败 → `logs.Errorf` → 直接返回 error，本轮结束。已成功提交的批次保留，
不做补偿。下一个周期（默认 60 分钟后）重新走一遍「定位起点游标 → 分批删除」，起点是进程内为该租户
保留的游标（进程重启则回到空串，见 D-13），而不是失败那一批的位置。

**理由**：清理是幂等操作——「删除超期记录」重复执行的结果收敛到同一状态，重跑没有副作用。
需求 F-001「不做无限重试」、AC-P04「不影响调度主流程」都指向「失败即退、等下一轮」。
注意 `pkg/cron/core/scheduler.go:121` 的调度器在 `Do()` 返回 error 后只 sleep 1 秒就重跑 `executeSingleRun`，
而 `executeSingleRun` 会先调 `Next()`——所以真实的重试间隔仍是一个完整的 `intervalMin`，不会形成快速重试风暴。

**被否决的替代方案**：*批级重试 N 次后再放弃*——DB 抖动时反而加重压力，且下一轮天然就是重试。

### D-9：`Next()` 恒返回「当前时间 + intervalMin」不做兜底；开关关闭时不注册调度，但仍留在 `s.tasks` 里

**选型**：三件事配合：

1. `Next()` 恒定返回「当前时间 + intervalMin 分钟」，**不对非法周期做任何收敛**。
2. `enabled=false` 时 `initCronTask` 跳过 `cron.Register`（`registerCleanupCronTask` 内判定），
   但仍把 task 写入 `s.tasks`。
3. `logics.Cleanup` 开头仍判 `enabled` 并返回 `ErrCleanupDisabled`，人工触发路径靠它拦住。

**理由**：`pkg/cron` 是「相对间隔」模型而非 crontab 模型，且 `scheduler.go:148` 是
`if nextTime.After(now)` 才 sleep——`Next()` 返回非未来时间会**跳过等待直接进入下一轮**，退化成忙循环。
而「`enabled: false` + `intervalMin: 0`」是启动期校验刻意放过的合法组合（`enabled=false` 时按约定不校验
其余项）。所以只要注册了调度，就必须在 `Next()` 里对非法周期兜底；改成不注册之后，`Next()` 只可能在
`enabled=true` 时被调用，此时 `intervalMin > 0` 由启动校验保证，兜底成为死代码，可以彻底删掉。
不注册同时也避免调度器周期性唤醒一个必然空转的任务。

第 2 点保留 `s.tasks` 条目是必需的：人工触发的路由注册依赖 `tasks[...].GetURL()`，
从 map 里去掉会让路由注册取到 nil 直接 panic。也就是说「不注册调度」与「人工触发仍可用」并不冲突，
只要把「任务注册表」和「调度注册」这两件事分开。

**被否决的替代方案**：
- *注册调度 + 在 `Next()` 里对非法周期兜底*——即上一版设计。能工作，但要长期维持一层与启动期校验重复的
  收敛逻辑，而这层逻辑的必要性依赖「`enabled=false` 时跳过其余项校验」这个不显眼的约定；
  后来人很容易把它当成冗余防御删掉，从而重新引入忙循环。把不变量前移到注册阶段更稳。
- *`enabled=false` 时连 `s.tasks` 都不放*——人工触发入口会随之消失，且 handler 里 map 取值会 panic。

### D-10：直接用 `dao.Set` 操作 DB，不绕 data-service

**选型**：清理逻辑放在新包 `cmd/task-server/logics/asyncflowcleanup/`，通过 `cap.Dao` / `s.dao` 直接调
`AsyncFlow()` / `AsyncFlowTask()` 的 DAO 方法。

**理由**：`.cursor/rules/api-principle.mdc` 要求「除 data-service 外其他服务禁止直接操作 db」，
但 task-server 对这两张表是**既定例外**——异步框架的 backend 本就是 task-server 直连 DB 读写
`async_flow` / `async_flow_task`（`cmd/task-server/service/service.go:96` 建 dao、`:150`
`backend.Factory(enumor.BackendMysql, dao)`）。清理是同一组表的同一类操作，沿用既有通路一致性最好，
且避免为两张内部表新增一整套 data-service 接口与 client 封装。

**被否决的替代方案**：*在 data-service 新增 async flow/task 的 List/BatchDelete 接口，task-server 通过
client 调用*——严格符合分层规则，但要新增 API 定义 + handler + client 三层代码，工时翻倍，
而这两张表是异步框架的内部表、不对外暴露，分层收益接近零。**这条偏离已在 Open Questions 中标注，
请评审确认。**

**唯一例外是租户列表**：它经 data-service 查询而非直连 dao，原因见 D-11——dao 层没有「未开启多租户时
合成 default 租户」这层兜底。也就是说本包同时持有 `dao.Set` 与 data-service client，前者只用于
`async_flow` / `async_flow_task`，后者只用于列租户。

### D-11：逐租户清理，沿用项目既有的跨租户定时任务写法

**背景**：两张表 `EnableTenant: true`（`pkg/dal/table/table.go:420`），DAO 内部统一用
`orm.NewInjectTenantIDOpt(kt.TenantID)`。而 cron 调度器给的 kit 来自 `core.NewBackendKit()`，
其 `SetBackendTenantID()`（`pkg/kit/kit.go:220`）在**未开启多租户时置为 `default`**（此时
`InjectTenantIDOpt.enabled()` 返回 false，不注入任何租户条件，清理覆盖全表），
在**开启多租户时置为 `system`**（此时会注入 `tenant_id = 'system'`，只清理运营租户的数据）。
而 `bill_main_account_summary` flow 由 account-server 主账号 controller 按**主账号的真实租户**落库
（`cmd/account-server/logics/bill/mainaccount_controller.go:141` `getInternalKit()` 之后 `SetTenant`），
两者口径不一致——直接用调度器给的 kit 会漏清大部分数据。

**选型**：先列出全部租户，再逐个派生 `kt.NewSubKitWithTenant(tenantID)` 执行清理，
即项目里跨租户定时任务的统一写法（`cmd/cloud-server/service/recycle/recycle_timing.go:110`、
`cmd/cloud-server/service/task/task_timing.go:78`、`cmd/cloud-server/service/bill/bill_timing.go:71`、
`cmd/account-server/logics/bill/manager.go:97` 共 4 处先例）。

**理由**：租户隔离仍然生效（每条 SQL 都带本租户的 `tenant_id`），同时又覆盖到全部租户，
不需要任何绕过机制。未开启多租户时租户列表只有 `default`，`InjectTenantIDOpt.enabled()` 返回 false，
行为与「不做租户处理」完全等价，单租户部署不受影响。

**两处有意偏离先例，均有明确理由**：

1. **串行而非 errgroup 并发**。4 处先例都并发跑租户，但本需求的批间隔限速是为了保护 MySQL 主从同步，
   并发会把实际删除速率放大到租户数倍，限速直接失效。租户数量有限，串行的代价只是单轮变长，
   而单轮本就不封顶、跨周期继续。
2. **租户列表不按 `status = enable` 过滤**。先例用的 `ListAllTenantID` 只取启用租户；清理的对象是历史
   垃圾数据，已禁用租户名下的记录更应该清掉，按 enable 过滤会让这部分数据永久残留，与本需求目标相悖。

**租户列表经 data-service 查询，不直连 dao**（是本设计中 D-10「直连 dao」的唯一例外）：
data-service 的 `ListTenant` 在未开启多租户时会无视入参、直接合成一个 `default` 租户返回
（`cmd/data-service/service/tenant/tenant.go:125`），dao 层没有这层兜底。单租户部署的 tenant 表可能
是空的，直连 dao 会一个租户都查不到、清理彻底空转。

**被否决的替代方案**：
- *硬编码 `DefaultTenantID` 关闭租户注入以覆盖全租户*——即上一版实现。它能「一条不漏」，但在整个
  代码库里没有第二处这么做，是独有写法；且它把跨租户物理删除的能力藏在一个工具函数里，
  评审时很难看出影响面。逐租户遍历既达到同样的覆盖度，又保留了租户隔离这道天然护栏。
- *接受只清 backend kit 所在租户*——多租户部署下等于绝大部分数据永远清不掉，不满足需求。

**已知局限**：`tenant_id` 不在租户表内的历史行（如租户已被物理删除）不会被清理。这类数据只能靠
后续独立的数据治理处理，本期不覆盖。

### D-12：配置默认值与常量位置

- 新增 `pkg/cc` 的 `AsyncFlowAndTaskCleanup` 结构体，挂在 `TaskServerSetting` 上（yaml key
  `asyncFlowAndTaskCleanup`），形态对齐 `pkg/cc/ziyan_types.go:1188` 的 `ApplyRecommend`：
  结构体自带 `trySetDefault()`，在 `TaskServerSetting.trySetDefault()` 里调用；校验逻辑放
  `TaskServerSetting.Validate()`。
- `enabled` 用 `*bool` 而非 `bool`。**理由**：默认值是 `true`，用值类型无法区分「用户显式配了 false」
  和「用户没配」，`trySetDefault` 会把显式的 false 覆盖成 true，AC-005 直接失效。
- 默认值（60 / 180 / 100）按
  `.cursor/rules/constant-define.mdc` 定义为 `pkg/criteria/constant/` 下的常量，不写死在校验分支里。
  单批条数不在此列：它不是配置项，而是 `asyncflowcleanup` 包内的常量 `flowBatchSize = 100`（D-4）。
- 三个数值项只校验「> 0」，**不设下限**。`batchIntervalMs` 尤其不设 100ms 下限：下限若等于默认值，
  配置项就只能往上调，等于半个假配置项；AC-P02 的 100ms 由默认值保证，显式调低是运维的知情选择。
  各项也都不做运行期 clamp：`trySetDefault` 只填 nil、`Validate()` 在启动期拦住所有非法值，
  配置进程序时已保证合法，再加一层收敛属于对同一件事的重复防御，反而会掩盖配置错误。
- `enabled=false` 时 `Validate()` 直接 return nil，不校验其余项（AC 场景「关闭开关时不校验其余项」）。

### D-13：每轮先用「只带主键条件」的有界扫描定位起点游标，游标按租户跨轮保留

**背景**：D-2 的 keyset 只解决了轮内的重复扫描。每轮的第一批查询仍然从 `id > ''` 出发，而 `async_flow`
上没有 `name` / `updated_at` 索引，这条带条件的查询要沿 PK 一路过滤到攒够 100 行才停。清理跑一段时间后，
PK 前缀里全是「name 不命中、永远不会被删」的残留行，这段前缀只增不减，每轮的第一批都要重走一遍，
单批耗时随运行时间持续劣化——正是 D-2 想避免的那种劣化，只是从轮内挪到了轮间。

**选型**：在 `cleanupLoop` 之前插入一个定位阶段 `locateStartCursor`：

1. 用 `ScanFlowsAfter` 做有界扫描，过滤条件**只有** `id > :cursor` 一条，
   `Fields` 只取 `id` / `name`，单窗 `Limit = scanWindow = core.DefaultMaxPageLimit`（500）。
2. `pickStartCursor` 在窗口内找第一条 name 命中白名单的行：命中则取**它前一条**的 id 作为游标
   （命中行本身要留给带条件的查询去取，取它自己会差一位跳过去）；命中行就是窗口第一条时游标保持不变；
   整窗不命中则游标推进到窗口末尾，继续下一窗；扫到空窗（表尾）结束。
3. 定位结果存进 `Logics.startCursors[tenantID]`，**跨轮保留**，下一轮从这里续扫。

**几条不变量及其理由**：

- **命中即停，不判断是否超期**。定位阶段拿不到也不该拿 `updated_at`——游标之前必须全是「永远不会被清理」
  的行，才有资格被永久跳过。若为了少扫几行而跨过「name 命中但尚未超期」的行，等它超期时已落在游标后面，
  再也扫不到，直接构成漏清。单测 `TestPickStartCursorStopsAtEveryNameHit` 钉住这条。
- **删除循环的进度不写回 `startCursors`**。带条件的查询会跳过保留期内的命中行，把它的进度当下一轮起点，
  等价于永久跳过那些行，是同一类漏清。单测 `TestCleanupCursorNeverSkipsUnexpiredHit` 钉住这条。
- **每轮都要重新定位**，不能只在首轮定位一次。上一轮删完之后，原本命中的那段位置只剩不命中的残留行，
  起点不继续往后推，这段新沉淀的前缀又会把带条件的查询拖慢。
- **定位阶段不限速**。`batchIntervalMs` 是为了压低删除在从库上的 apply 尖峰，而定位是只读、不产生 binlog；
  冷启动可能要空扫几十上百个窗口，每窗睡 100ms 会把定位本身拖到小时级。单测
  `TestCleanupLocateHasNoRateLimit` 钉住这条。
- **游标是进程内内存态，不落库**。重启后回到空串、退化成从表头扫，正确性不受影响（只是重启后第一轮
  要多扫一遍前缀）。为一个纯性能优化引入一张状态表或一个 global_config 项，性价比不成立。

**被否决的替代方案**：

- *在带条件的查询上直接依赖索引*——最优解，但需求禁止 DDL（Open Question 1 已表态本期不加索引）。
  起点游标正是在「不能加索引」的约束下把前缀扫描摊薄的手段。
- *把删除循环的 `lastFlowID` 直接当作下一轮起点*——实现最省事，但会跳过保留期内的命中行，如上所述是漏清。
- *定位阶段带上 `name IN (...)` 条件，直接查第一条命中行*——看似更省事，但这条查询本身就是那条走不了索引、
  要沿 PK 全程过滤的慢查询，等于把问题原样搬了个位置；而只带主键条件的有界扫描每次都只读固定 500 行，
  代价恒定、可预期。
- *游标持久化到 DB*——能省掉重启后的一次重扫，代价是新增状态表与一致性维护。收益是一次性的，不划算。

### D-14：清理范围用 flow name 白名单，而不是单个 name 等值匹配

**选型**：包内常量 `cleanupFlowNames []enumor.FlowName`，当前只有 `enumor.FlowBillMainAccountSummary`
一项；过滤条件用 `tools.RuleIn("name", cleanupFlowNames)`；定位阶段的命中判定用
`slice.IsItemInSlice(cleanupFlowNames, one.Name)`。

**理由**：本期的清理对象只有账单分账一种 flow，但这两张表的膨胀不是它独有的问题，后续大概率还要纳入
别的 flow 类型。白名单把「清理哪些 flow」收敛成一个常量，扩展时只改一行、过滤条件与定位判定自动跟着走，
不会出现两处判定漏改一处的偏差。元素个数受 `filter.DefaultMaxInLimit`（500）约束，实际远达不到。

**不做成配置项的理由**：改清理范围是有数据风险的动作（配错一个 name 就会误删一整类 flow 的历史数据），
应当走代码评审而不是改 yaml；且 flow name 是 `enumor` 里的枚举，写成配置就失去了编译期约束。

**被否决的替代方案**：*保留 `RuleEqual` 单值匹配，需要时再改*——扩展时要同时改过滤条件与定位判定两处，
且 `RuleEqual` → `RuleIn` 的改动会连带影响单测断言，不如一次到位。

## Risks / Trade-offs

| ID | 风险 → 缓解措施 |
|----|------------------|
| R-1 | **无 `(name, updated_at)` 索引，首轮扫描代价高** → D-2 的 keyset 保证每行在一轮内最多被检查一次，D-13 的起点游标进一步保证 PK 前缀里的残留行跨轮也不重扫，且超期行聚集在 id 低位；配合 100 条/批 + 100ms 间隔，扫描被摊平到长时间窗口。上线后观测慢查询日志，若单批查询耗时异常，作为独立需求评估加索引 |
| R-2 | **首轮消化 500 万+ 存量，耗时可能远超一个 60 分钟周期** → D-6 的防重入保证不会并发叠加。单批 100 条 + 100ms 间隔的理论速率是 1000 行/秒，500 万行约 1.4 小时纯删除时间，加上扫描开销在数小时量级，符合需求「数小时内清理完毕」，但余量比贴上限跑（500 条/批、约 17 分钟）小得多。首轮跨多个 60 分钟周期是预期行为，靠防重入与幂等保证正确性 |
| R-3 | **删除事务在 ROW 格式 binlog 下按行产生 event，可能拉高从库 apply 压力** → 单批取 100 条而非上限 500，把单事务的 event 数压到五分之一（D-4），叠加 100ms 批间隔进一步摊平；AC-P03 要求在存量环境实测主从延迟，实测不达标时上调 batchIntervalMs（配置项，无需改代码）。单批条数不可调，需要改速率时调批间隔即可，两者作用等价（D-4） |
| R-4 | **单轮不封顶，极端情况下清理线程长时间占用一个 DB 连接** → 清理跑在 cron 独立 goroutine，与异步调度的连接池共享但不互斥；`pkg/cc` 的 `limiter.qps=500` 兜底限流；AC-P04 要求实测调度不受影响 |
| R-5 | **误删风险：过滤条件写错会删掉其他 flow_name 或保留期内数据** → 过滤值取自白名单常量 `cleanupFlowNames`（元素为 `enumor.FlowName` 枚举而非字符串字面量，D-14），且过滤与定位共用同一份白名单；单测必须覆盖 AC-002/003/004（保留期内、其他 flow_name、非终态僵尸）三类边界；首次上线建议先把 `retentionDays` 调大（如 365）跑一轮观察，再降到 180 |
| R-6 | **多租户部署下漏清非 `system` 租户的数据** → 已按 D-11 改为先列全部租户、再逐个用该租户的子 kit 清理，租户隔离仍生效；仍需在开启多租户的环境实测各租户数据均被清理（tasks 8.19） |
| R-7 | **task-server 首次接入 cron，初始化失败会阻断服务启动** → `initCronTask` 的错误直接从 `NewService` 返回，属于 fail-fast，与 woa-server 行为一致；配置校验前移到 `cc.LoadSettings`，非法配置在 cron 初始化之前就报错 |
| R-8 | **绕过 data-service 直连 DB 偏离分层规则** → 见 D-10，已说明是 task-server 对这两张表的既定例外，需评审确认 |

## Migration Plan

**部署步骤**

1. 合入代码；`cmd/task-server/etc/task_server.yaml` 与 `docs/support-file/helm`（`values.yaml` +
   `templates/taskserver/configmap.yaml`）两处配置**必须同步**变更，否则 helm 部署的实例读不到配置段
   （会走默认值，行为仍正确，但运维无法调参）。
2. 建议**首次上线时把 `enabled` 配为 `false`** 发布一次，确认 task-server 正常启动、cron 框架接入无副作用。
3. 在灰度环境把 `retentionDays` 设为一个较大值（如 365）打开开关，跑一轮，核对日志中的删除条数与预期一致、
   且未误删其他 flow_name。
4. 生产开启，全程盯 MySQL 主从延迟（AC-P03）与 task-server 异步任务积压（AC-P04）。
5. 首轮存量消化完成后，将 `retentionDays` 收敛到 180。

**回滚策略**

- 一级回滚（秒级，首选）：把 `enabled` 改为 `false` 并重启 task-server，清理立即停止。
  已删除的数据**不可恢复**，但清理是「删超期数据」，业务无依赖。
- 二级回滚：回退代码版本。因不涉及 DDL、不改任何对外接口，回退无兼容性负担。
- 数据兜底：上线前请 DBA 确认 `async_flow` / `async_flow_task` 的备份策略覆盖本次变更窗口。

## Open Questions

> 以下 5 条已在阶段 5 代码评审中全部表态，结论回填于此（对应 tasks.md 9.8）。

1. **是否为 `async_flow` 补 `(name, updated_at)` 联合索引？**
   **结论：本期不加，维持 keyset 扫描 + 起点游标（D-13）。** 需求明确禁止 DDL，加索引须另立独立提案，
   并由 DBA 评估在千万级行表上加索引的窗口与代价。起点游标是在这个约束下把「PK 前缀残留行每轮重扫」
   这部分代价消掉的手段，但它只能摊薄前缀，命中区间内的过滤仍然走扫描，加索引依然是长期最优解。
   **未闭环项**：tasks.md 8.18 要求对 flow 侧查询实测 `EXPLAIN` 记录扫描行数与单批耗时，
   该项依赖真实环境、尚未执行。若实测不可接受，按 8.18 的约定另立 DDL 提案
   （`alter table async_flow add index idx_name_updated_at (name, updated_at)`），不塞进本变更。

2. **防重入是否需要升级为跨进程分布式锁？**
   **结论：本期不做，维持进程内 `sync.Mutex`（D-6）。** 评审确认后果可控：跨副本并发的后果是
   DB 负载升高而非数据错误——删除按主键幂等，两轮加锁顺序一致（均按 id ASC、均先 task 后 flow），
   InnoDB 死锁风险低。最坏是两轮删除集合重叠、DB 压力翻倍。
   **附带影响**：并发时上报条数会偏大，见下方第 5 条相关的 S-7 说明。
   **要求**：运维手册需写明「人工触发前先确认没有其它节点在跑清理」。

3. **是否需要根据主从延迟动态调速？**
   **结论：本期不做，维持固定速率（D-5）。** 动态调速需读从库状态（额外 DB 权限，与 AC-S01 冲突）
   或接入监控（本期不包含）。`batchIntervalMs` 是配置项，实测不达标时调它即可，无需改代码
   （单批条数已改为不可配，见 D-4）。

4. **绕过 data-service 直连 DB 是否被评审接受？**
   **结论：评审明确接受（D-10）。** 理由：task-server 本就是异步框架的 backend，
   `async_flow` / `async_flow_task` 是它自己的状态表，调度、状态流转、重试全程直连；
   为清理单独去 data-service 开一组内部接口反而扩大了这两张表的**外部可写面**，
   且链路更长、失败面更大。
   **接受前提（后续维护必须保持）**：`daoFlowStore` 已把全部 dao 访问收敛在单一类型内，
   `Logics` 只持有 `flowStore` / `tenantLister` 两个接口、cron task 只持有 `*Logics`，
   都拿不到 `dao.Set`。不要把 `dao.Set` 往上层漏。

5. **多租户部署下的清理租户口径（D-11）**
   **结论：实证发现租户口径不一致，已改为「先列全部租户、再逐个用该租户的子 kit 清理」，
   即项目里跨租户定时任务的统一写法。**
   实证过程：两张表 `EnableTenant: true`，而 `bill_main_account_summary` flow 由 account-server
   主账号 controller 按**主账号的真实租户**落库（`mainaccount_controller.go:141` `getInternalKit()`
   之后 `SetTenant`）；cron 的 backend kit 在多租户下是 `system` 租户、人工触发是调用方租户，
   两者都会被注入 `tenant_id` 条件而漏清大部分数据。
   **本条结论经过一次修正。** 初版实现用 `kt.NewSubKitWithTenant(constant.DefaultTenantID)` 关闭租户注入，
   靠「一次查询覆盖全部租户」达到覆盖度。评审指出该写法与代码库其他逻辑不一致——全仓找不到第二处
   为了跨租户而关闭注入的地方，而跨租户定时任务有 4 处统一先例（见 D-11）。改为逐租户遍历后，
   租户隔离这道护栏得以保留，覆盖度不变。
   当初否决「按租户逐个清理」的理由是「cron backend kit 拿不到可信租户列表」，**这个前提是错的**：
   `ListAllTenantID` 这套 helper 早已存在（cloud-server / account-server 各一份），
   经 data-service 查询，且在未开启多租户时会合成 `default` 租户兜底。
   **已落实的护栏**：① 过滤值用 `enumor.FlowBillMainAccountSummary` 常量而非字面量；
   ② 单测断言过滤条件恰好三条规则且不含 state，能挡住「有人顺手改条件」的回归；
   ③ 单测断言每个租户的查询都携带该租户 id、且调用方自己的 kit 不被改动，
   能挡住「退回单租户口径」与「污染上游 kit」两类回归；
   ④ 每轮起始日志打印租户数、每个租户结束时打印该租户删除条数，清理范围在日志里可核对。
   **未闭环项**：tasks.md 8.19 要求在开启多租户的环境实测各租户数据均被清理、且删除范围不超预期。
   这是本条决策唯一的实测支撑，**必须在上线前闭环**——当前正确性仅有代码推理与单测支撑。
