> 实现前请先通读 `proposal.md`、`specs/async-flow-cleanup/spec.md`、`design.md`，并遵循 `.cursor/rules/` 下的项目规范：
> `naming-convention-core`（命名）、`import-standard`（三段式导入与包别名）、`comment-standard`（注释）、
> `error-handling`（`errf` 错误码与错误传播）、`logging-standard`（级别与 rid）、`constant-define`（禁止魔数、枚举放 `pkg/criteria/enumor`）。
> 额外约束：单函数不超过 80 行；指针/值转换统一用 `cvt "hcm/pkg/tools/converter"` 的 `PtrToVal` / `ValToPtr`；map 相关逻辑优先用 `pkg/tools/maps`；非必要不写 `else`。

## 1. 枚举与常量

- [x] 1.1 在 `pkg/criteria/enumor/cron_task.go` 新增 `CronTaskAsyncFlowAndTaskCleanup CronTask = "async_flow_and_task_cleanup"`，并按文件既有风格补中文注释
- [x] 1.2 在 `pkg/criteria/constant/` 下新增清理相关常量：默认周期 60（分钟）、默认保留天数 180、默认批间隔 100（毫秒）。不定义批间隔下限常量：下限若等于默认值则配置项只能上调，等于半个假配置项，AC-P02 的 100ms 由默认值保证。单批条数也不在 `constant` 包定义，它不是配置项，由 `asyncflowcleanup` 包内的常量 `flowBatchSize = 100` 直接给出（design D-4 / D-12）
- [x] 1.3 确认复用既有枚举 `enumor.FlowBillMainAccountSummary`（`pkg/criteria/enumor/async_flow_name.go:179`），不新增重复定义；在 `asyncflowcleanup` 包内定义清理白名单 `cleanupFlowNames []enumor.FlowName`（当前只有该枚举一项），过滤条件与起点游标的命中判定共用它，元素个数受 `filter.DefaultMaxInLimit` 约束（design D-14）

## 2. 配置定义与校验（pkg/cc）

- [x] 2.1 在 `pkg/cc/types.go` 新增 `AsyncFlowAndTaskCleanup` 结构体，字段 `Enabled *bool` / `IntervalMin *int` / `RetentionDays *int` / `BatchIntervalMs *int`，yaml tag 分别为 `enabled` / `intervalMin` / `retentionDays` / `batchIntervalMs`；四个字段全用指针的原因（区分「未配置」与「显式配置了非法值」，否则 `trySetDefault` 会把非法值静默改写成默认值、使 `validate` 的下界校验永不可达）写进结构体注释（design D-12）
- [x] 2.2 实现 `func (c *AsyncFlowAndTaskCleanup) trySetDefault()`：只在字段为 nil 时填 1.2 中的默认常量（`Enabled` 填 `cvt.ValToPtr(true)`），显式配置的非法值原样保留交给 validate 拦截
- [x] 2.3 实现 `func (c AsyncFlowAndTaskCleanup) validate() error`：`cvt.PtrToVal(c.Enabled)` 为 false 时直接返回 nil；否则校验 `IntervalMin > 0`、`RetentionDays > 0`、`BatchIntervalMs > 0`（三项均只校验为正、不设上下限），错误信息带上具体字段名与当前值。启动期校验是唯一防线，清理逻辑内不再做运行期收敛（design D-12）
- [x] 2.4 在 `pkg/cc/service.go` 的 `TaskServerSetting` 新增字段 `AsyncFlowAndTaskCleanup AsyncFlowAndTaskCleanup \`yaml:"asyncFlowAndTaskCleanup"\``
- [x] 2.5 在 `TaskServerSetting.trySetDefault()` 中调用 `s.AsyncFlowAndTaskCleanup.trySetDefault()`
- [x] 2.6 在 `TaskServerSetting.Validate()` 中调用 `s.AsyncFlowAndTaskCleanup.validate()` 并向上返回错误（注意 `Validate()` 是值接收者、`trySetDefault()` 是指针接收者）

## 3. 配置落地（yaml + helm）

- [x] 3.1 在 `cmd/task-server/etc/task_server.yaml` 的 `async` 配置段之后新增 `asyncFlowAndTaskCleanup` 段，四个配置项均带中文注释说明含义、默认值与取值范围
- [x] 3.2 在 `docs/support-file/helm/values.yaml` 的 `taskserver` 段下同步新增 `asyncFlowAndTaskCleanup` 配置，内容与 3.1 保持一致
- [x] 3.3 在 `docs/support-file/helm/templates/taskserver/configmap.yaml` 中渲染 `asyncFlowAndTaskCleanup` 配置段，字段引用方式与该文件内 `async` 段的既有写法保持一致
- [x] 3.4 自检 3.1~3.3 三处的字段名、默认值、注释完全一致，避免配置漂移

## 4. 清理主逻辑（`cmd/task-server/logics/asyncflowcleanup`）

- [x] 4.1 新建包 `cmd/task-server/logics/asyncflowcleanup`，定义 `Logics` 结构体（持有 `dao dao.Set`）与构造函数 `NewLogics(dao dao.Set) *Logics`，并写包注释
- [x] 4.2 定义 `Result` 结构体：`DeletedFlowCount int` / `DeletedTaskCount int` / `Duration time.Duration`，供 handler 返回与日志共用
- [x] 4.3 在 `Logics` 上实现进程内防重入：持有 `mu sync.Mutex`，对外入口 `Cleanup(kt *kit.Kit) (*Result, error)` 用 `TryLock()` 抢锁，抢不到时返回 `errf.New(errf.Aborted, "async flow and task cleanup is already running")`；`defer Unlock()`（design D-6，spec: Single-Flight Cleanup Execution）
- [x] 4.4 在 `Cleanup` 开头判定 `cvt.PtrToVal(cc.TaskServer().AsyncFlowAndTaskCleanup.Enabled)`，为 false 时返回明确错误/空结果并输出 info 日志（spec: Cleanup Configuration And Validation，AC-005）
- [x] 4.5 计算本轮 cutoff：`time.Now().AddDate(0, 0, -retentionDays).Format(constant.TimeStdFormat)`，整轮复用；输出「本轮开始，cutoff=xxx」的 info 日志并带 rid（spec: Cleanup Observability Logging）
- [x] 4.6 实现 flow 侧主键游标分页：以 `lastFlowID` 为游标，过滤条件 `tools.ExpressionAnd(tools.RuleIn("name", cleanupFlowNames), tools.RuleLessThan("updated_at", cutoff), tools.RuleIDGreaterThan(lastFlowID))`，`Fields: []string{"id"}`，`Page` 设 `Sort: "id"` / `Order: core.Ascending` / `Limit: flowBatchSize` / `Start: 0`；首批 `lastFlowID` 取自 4.15 的起点游标定位结果（design D-2 / D-13 / D-14，spec: Scheduled Cleanup / Cleanup Scope）
- [x] 4.7 实现 task id 查询：对 flow id 按 `slice.Split(flowIDs, 500)` 分片，每片用 `tools.ContainersExpression("flow_id", batch)` + `Fields: []string{"id"}` 翻页取全量 task id（走 `idx_flow_id`）（design D-4）
- [x] 4.8 实现单批事务删除：`dao.Txn().AutoTxn(kt, func(txn, opt))` 内，先按 `slice.Split(taskIDs, 500)` 分片调 `dao.AsyncFlowTask().DeleteWithTx(kt, txn, tools.ContainersExpression("id", chunk))`，再调 `dao.AsyncFlow().DeleteWithTx(kt, txn, tools.ContainersExpression("id", flowIDs))`；顺序不可颠倒（design D-3，spec: Cascading Deletion，AC-001/AC-S02）
- [x] 4.9 每批删除成功后输出 info 日志（本批 flow 数、task 数、rid），并 `time.Sleep(time.Duration(batchIntervalMs) * time.Millisecond)`（spec: Rate-Limited Batch Deletion，AC-P02/AC-010）
- [x] 4.10 实现主循环：批次返回条数为 0 时正常结束；返回条数小于 `flowBatchSize` 时删完该批后结束；否则推进 `lastFlowID = flowIDs[len-1]` 继续。单轮不设总量上限（AC-008、spec: 单轮不封顶直到清空）
- [x] 4.11 错误处理：任一批的查询或删除失败，按 `logs.Errorf("...err: %v, ...rid: %s", err, kt.Rid)` 输出并**终止本轮**返回错误，不做无限重试；已提交的批次不回滚（幂等，下一轮继续）
- [x] 4.12 结束时输出 info 日志：本轮累计删除 flow 数、task 数、总耗时、rid，并填充 `Result` 返回（AC-010）
- [x] 4.13 拆分函数确保每个函数不超过 80 行（建议 `Cleanup` / `listExpiredFlowIDs` / `listTaskIDsByFlowIDs` / `deleteBatchWithTx` 四个层次）
- [x] 4.14 核实 `core.NewBackendKit()` 的 `SetBackendTenantID()` 对 `orm.NewInjectTenantIDOpt` 的实际影响，确认多租户环境下清理不会被限制在单一租户；若会，记录结论并在本文件补充后续处理项（design D-11 / R-6 / Open Question 5）
  - **核实结论：会被限制**。`async_flow` / `async_flow_task` 均为 `EnableTenant: true`（`pkg/dal/table/table.go:413`）；
    `core.NewBackendKit()` → `SetBackendTenantID()` 在开启多租户时置租户为 `system`，`InjectTenantIDOpt.enabled()`
    返回 true，所有查询与删除都会带上 `tenant_id = 'system'`。而 `bill_main_account_summary` flow 的创建方
    （`cmd/account-server/logics/bill/mainaccount_controller.go:141` 的 `getInternalKit()` + `kt.SetTenant(mac.TenantID)`）
    按主账号的**真实租户**落库，两者口径不一致；人工触发路径拿到的是调用方租户，同样会被限制。
  - **处理：按 design D-11 逐租户清理**——`Cleanup` 先经 data-service 取全部租户 id（`tenantLister`），
    再对每个租户派生 `kt.NewSubKitWithTenant(tenantID)` 调 `cleanupLoop`，沿用项目里跨租户定时任务的统一写法。
    串行执行（不用 errgroup）以保住批间隔限速；租户列表不按 `status = enable` 过滤，避免已禁用租户的数据永久残留。
    租户列表必须走 data-service 而非直连 dao：未开启多租户时 data-service 会合成 `default` 租户兜底，dao 层不会，
    直连会导致单租户部署一个租户都查不到、清理空转。后续需在多租户环境实测确认，见 8.19。
- [x] 4.15 实现起点游标定位 `locateStartCursor`：在 `cleanupLoop` 之前，用 `ScanFlowsAfter` 做有界扫描（过滤条件只有 `tools.RuleIDGreaterThan(cursor)` 一条，`Fields: []string{"id", "name"}`，`Limit: scanWindow = core.DefaultMaxPageLimit`），逐窗调 `pickStartCursor` 找第一条命中白名单的记录，取**它前一条**的 id 作为起点（命中记录是窗口首条时游标保持不变，整窗不命中则推进到窗口末尾继续下一窗，空窗即表尾结束）；定位阶段不删数据、不做批间隔限速；结束后输出含租户 id、游标、扫描窗数的 info 日志（design D-13，spec: Start Cursor Location And Cross-Round Reuse）
- [x] 4.16 起点游标按租户存入 `Logics.startCursors map[string]string` 并跨轮保留，每轮重新定位时从上一轮的游标续扫；删除循环轮内推进的 `lastFlowID` **不得**写回 `startCursors`（会跳过保留期内的命中记录，构成漏清）；游标只存内存不落库，进程重启后回到空串、退化为从表头扫描（design D-13）

## 5. cron 任务与 task-server 接入

- [x] 5.1 新建 `cmd/task-server/task/async_flow_and_task_cleanup.go`，定义 `AsyncFlowAndTaskCleanupTask`（持有 `logics *asyncflowcleanup.Logics` 与 `sd serviced.State`）与构造函数 `NewAsyncFlowAndTaskCleanupTask(...) (croncore.Task, error)`，形态对齐 `cmd/woa-server/task/apply_recommend.go`
- [x] 5.2 实现 `Name() string` 返回 `string(enumor.CronTaskAsyncFlowAndTaskCleanup)`
- [x] 5.3 实现 `Next() (time.Time, error)` 返回 `time.Now().Add(time.Duration(cc.TaskServer().AsyncFlowAndTaskCleanup.IntervalMin) * time.Minute)`，不对周期做兜底收敛（合法性由启动期校验 + 「关闭时不注册调度」共同保证，design D-9）
- [x] 5.4 实现 `GetURL() string` 返回 `"/async_flow_and_task/cleanup"`
- [x] 5.5 实现 `Do(kt *kit.Kit) error`：先判 `t.sd == nil || !t.sd.IsMaster()` 则输出 info 日志后返回 nil（AC-006，spec: Master-Only Execution），再调用 `t.logics.Cleanup(kt)`；被防重入拒绝时降级为 info 日志并返回 nil，避免 cron 框架把「跳过」记为执行失败（AC-007）
- [x] 5.6 在 `cmd/task-server/service/service.go` 的 `Service` 结构体新增 `sd serviced.State`、`cleanupLogics *asyncflowcleanup.Logics`、`tasks map[enumor.CronTask]croncore.Task` 三个字段
- [x] 5.7 在 `NewService` 中给上述字段赋值（`sd` 直接使用入参 `sd serviced.ServiceDiscover`，其内嵌了 `State`，**无需修改 `NewService` 签名与 `cmd/task-server/app/app.go`**，design D-1）
- [x] 5.8 新增 `func (s *Service) initCronTask() error`：调用 `cron.Init(context.Background(), metrics.Register())`，构造清理 Task 放入 `s.tasks`，再经 `registerCleanupCronTask` 注册调度；形态对齐 `cmd/woa-server/service/service.go:702`。注意 `s.tasks` 无条件写入（人工触发路由要取 `GetURL()`），而 `cron.Register` 仅在 `enabled=true` 时执行——否则「`enabled: false` + `intervalMin: 0`」这个校验放过的组合会让 `Next()` 返回非未来时间、调度退化成忙循环（design D-9）
- [x] 5.9 在 `NewService` 返回前调用 `initCronTask()`，失败时记 error 日志并向上返回（启动期快速失败，design R-6）
- [x] 5.10 确认 task-server 的优雅退出路径中已覆盖 cron（如未覆盖，参照 woa-server 在 shutdown notifier 中调用 `cron.Stop()`）

## 6. 人工触发入口

- [x] 6.1 在 `cmd/task-server/service/capability/capability.go` 的 `Capability` 新增字段 `Tasks map[enumor.CronTask]croncore.Task` 与 `CleanupLogics *asyncflowcleanup.Logics`
- [x] 6.2 在 `cmd/task-server/service/service.go` 的 `apiSet()` 中把 `s.tasks` 与 `s.cleanupLogics` 填入 `capability.Capability`
- [x] 6.3 在 `cmd/task-server/service/controller/controller.go` 的 `service` 结构体新增 `cleanupLogics` 字段并在 `Init` 中赋值
- [x] 6.4 实现 handler `CleanupAsyncFlowAndTask(cts *rest.Contexts) (any, error)`：直接调用 `svc.cleanupLogics.Cleanup(cts.Kit)`（**不做 master 判定**，design D-7；**不做 IAM 鉴权**，与 task-server 其余内部接口一致），成功返回 `Result`，被防重入拒绝时把 `errf.Aborted` 错误原样返回（AC-007/AC-009）
- [x] 6.5 在 `controller.Init` 中注册路由：`h.Add("CleanupAsyncFlowAndTask", http.MethodPost, cap.Tasks[enumor.CronTaskAsyncFlowAndTaskCleanup].GetURL(), svc.CleanupAsyncFlowAndTask)`，路径字面量从 `GetURL()` 取以保证与定时任务不分叉（对齐 `cmd/woa-server/service/res-sync/service.go:59` 的既有注册模式）
- [x] 6.6 验证完整路径为 `POST /api/v1/task/async_flow_and_task/cleanup`（`ws.Path("/api/v1/task")` 由 `apiSet()` 提供）

## 7. 单元测试

- [x] 7.1 `pkg/cc` 配置测试：缺省配置经 `trySetDefault` 后为 `enabled=true / 60 / 180 / 100`（spec: 缺省配置取默认值）
- [x] 7.2 `pkg/cc` 配置测试：`enabled=true` 且 `intervalMin <= 0` / `retentionDays <= 0` / `batchIntervalMs <= 0` 时 `validate()` 分别返回带字段名的错误；并覆盖「`batchIntervalMs` 取低于默认 100 的正值仍合法」（不设下限）（AC-011）
- [x] 7.3 `pkg/cc` 配置测试：`enabled=false` 时即使其余项非法也校验通过（spec: 关闭开关时不校验其余项）
- [x] 7.4 `pkg/cc` 配置测试：yaml 中显式 `enabled: false` 与完全不配置该段，`trySetDefault` 后分别得到 false 与 true（验证 `*bool` 的必要性）
- [x] 7.5 清理逻辑测试：过滤表达式构造正确——只含 `name IN cleanupFlowNames`、`updated_at < cutoff`、`id > lastFlowID` 三条规则，不含任何 `state` 条件（`TestBuildExpiredFlowFilter`；AC-002/AC-003/AC-004，spec: Cleanup Scope）
- [x] 7.6 清理逻辑测试：id 分片逻辑——给定 1200 个 task id，`slice.Split` 后为 500/500/200 三片；并断言 `scanWindow <= core.DefaultMaxPageLimit`、`flowBatchSize <= filter.DefaultMaxInLimit`（`TestSplitIDsWithinInLimit` / `TestBatchSizeWithinLimits`；AC-P01，design D-4 / D-13）
- [x] 7.7 清理逻辑测试：`Cleanup` 在 `enabled=false` 时不发起任何 DAO 调用（AC-005）
- [x] 7.8 清理逻辑测试：防重入——两个 goroutine 并发调用 `Cleanup`，第二个立即返回 `errf.Aborted`，不进入清理主体（AC-007）
- [x] 7.9 清理逻辑测试：flow 查询返回空列表时，`Cleanup` 正常返回、`Result` 全 0、不报错、不发起删除调用（AC-008）
- [x] 7.10 cron Task 测试：`sd.IsMaster()` 为 false 时 `Do()` 直接返回 nil 且不触发清理（AC-006）
- [x] 7.11 cron Task 测试：`Name()` / `GetURL()` / `Next()` 返回值符合预期（`Next()` 随 `intervalMin` 变化）
- [x] 7.12 起点游标测试：定位扫描的过滤条件只有主键一条规则，空游标时渲染出的 SQL 不含任何附加条件（`TestBuildScanFilterOnlyHasPrimaryKey` / `TestBuildScanFilterSQLWithEmptyCursor`，design D-13）
- [x] 7.13 起点游标测试：`pickStartCursor` 的四类窗口形态——命中记录在中间取其前一条、命中记录是首条时游标不回退、整窗不命中推进到窗口末尾、空窗不推进（`TestPickStartCursor` / `TestPickStartCursorFirstRowIsTarget` / `TestPickStartCursorNoTargetInWindow` / `TestPickStartCursorEmptyWindow`）
- [x] 7.14 起点游标测试：只要 name 命中就停，不判断是否超期，挡住「跳过未超期命中记录导致漏清」的回归（`TestPickStartCursorStopsAtEveryNameHit`，spec: 起点游标不越过任何命中记录）
- [x] 7.15 起点游标测试：定位逐窗向后推进不重扫，删除循环的第一次查询从定位结果开始（`TestCleanupLocateStartCursor` / `TestCleanupRoundStartsFromLocatedCursor`）
- [x] 7.16 起点游标测试：第二轮从上一轮保留的游标续扫而非回到空串（`TestCleanupLocateResumesFromLastCursor`，spec: 下一轮从上一轮的游标续扫）
- [x] 7.17 起点游标测试：删除循环的进度不写回 `startCursors`（`TestCleanupCursorNeverSkipsUnexpiredHit`）；定位阶段不做限速等待，多窗定位总耗时小于一个 `batchIntervalMs`（`TestCleanupLocateHasNoRateLimit`）
- [x] 7.18 运行 `go build ./...` 与 `go test ./pkg/cc/... ./cmd/task-server/...`，全部通过

## 8. 自测与联调验证

- [ ] 8.1 造数：在测试库插入 `name=bill_main_account_summary` 且 `updated_at` 早于 180 天的 flow 及其名下 task，验证一轮清理后两张表均无残留、无孤儿（AC-001）
- [ ] 8.2 造数：插入保留期内的同名 flow，验证一条不删（AC-002）
- [ ] 8.3 造数：插入 `bill_split_daily`、`obs_sync_bill_item` 等白名单以外 flow_name 的超期记录，验证一条不删（AC-003）
- [ ] 8.4 造数：插入 `state=running` 且超期的同名 flow，验证被删除（AC-004）
- [ ] 8.5 造数：flow 名下无 task 的场景，验证只删 flow 且不报错（spec: 名下无 task 的 flow）
- [ ] 8.6 配置验证：`enabled=false` 时到达周期不执行、表数据无变化（AC-005）
- [ ] 8.7 多副本验证：在 slave 实例上确认定时触发被跳过并有 info 日志、数据无变化（AC-006）
- [ ] 8.8 并发验证：构造长耗时清理，期间再次定时触发与人工触发，确认均被拒绝且无并发（AC-007）
- [ ] 8.9 空表验证：无任何符合条件的记录时，任务正常结束、不报错（AC-008）
- [ ] 8.10 人工触发验证：调用 `POST /api/v1/task/async_flow_and_task/cleanup`，确认执行同一套逻辑并返回本轮结果（AC-009）
- [ ] 8.11 日志验证：检查每轮开始 / 起点游标定位 / 每批 / 每轮结束 / 出错五类日志齐全，均携带 rid，结束日志含 flow 数、task 数、总耗时，定位日志含租户 id、游标与扫描窗数（AC-010）
- [ ] 8.12 启动校验验证：`enabled=true` 且 `intervalMin=0` 时 task-server 启动失败并输出明确错误（AC-011）
- [ ] 8.13 SQL 审计：抓取本轮全部 SQL，确认每条 delete 的 where 条件均为主键 `id IN (...)`，单条 IN 列表不超过 500，且不存在 `set session binlog_format`（AC-P01/AC-S01/AC-S02）
- [ ] 8.14 限速验证：从日志时间戳确认相邻两批间隔 ≥ 100ms，峰值速率 ≤ 5000 行/秒（AC-P02）
- [ ] 8.15 存量压测：在约 500 万存量的环境执行首轮清理，观测 MySQL 主从复制延迟不持续增长，清理结束 10 分钟内回落到清理前水平（AC-P03）
- [ ] 8.16 干扰验证：清理期间观察 task-server 的异步任务调度与执行，确认无积压、无超时（AC-P04）
- [ ] 8.17 表结构验证：上线前后比对 `async_flow` / `async_flow_task` 表结构，确认无任何 DDL 变更（spec: 不涉及表结构变更）
- [ ] 8.18 性能核实：对 flow 侧带条件的查询与定位阶段的有界扫描分别执行 `EXPLAIN`，记录实际扫描行数与单次耗时，并记录冷启动定位所需的扫描窗数（design D-13）。**若实测不可接受**，不要在本变更内加索引，而是记录数据并另立一个 DDL 变更提案（`alter table async_flow add index idx_name_updated_at (name, updated_at)`，SQL 文件按约定用 `SQLVER=9999,HCMVER=v9.9.9`）（design R-1 / Open Question 1）

- [ ] 8.19 多租户验证：在开启多租户的环境造多个租户的超期 `bill_main_account_summary` flow，确认一轮清理后各租户数据均被清理（对应 4.14 的逐租户遍历方案，design D-11 / Open Question 5 的闭环）

- [ ] 8.20 起点游标验证：在存量环境连跑至少两轮，从定位日志确认第二轮的游标不回退到空串、扫描窗数明显少于首轮；重启 task-server 后确认游标回到表头且清理结果与重启前一致（design D-13）

## 9. 规范自检与交付

- [x] 9.1 导入分组自检：标准库 / `hcm/` 内部包 / 第三方包三段式，组间空行；运行 `goimports -w .`
- [x] 9.2 命名自检：文件 snake_case、结构体 `*Task` / `*Logics` / `*Result` 后缀、常量 PascalCase、缩写保持大写
- [x] 9.3 日志自检：错误日志用 `logs.Errorf` 且 `err: %v` 在前、`rid: %s` 在末；continue 场景用 `logs.Warnf`；消息小写字母开头
- [x] 9.4 错误自检：业务错误用 `errf.New/Newf` 带错误码，无被 `_` 忽略的 error 返回
- [x] 9.5 注释自检：导出标识符均有以名称开头的注释；超过一行或含业务词汇的注释用中文
- [x] 9.6 函数长度自检：所有新增函数不超过 80 行
- [x] 9.7 基于 `tencent/bcc/v1.9.x` 新开分支提交 MR，MR 描述中显式说明三处有意识的规范偏离：D-10（task-server 直连 DAO）、D-7（人工触发不校验 master）、D-12（`Enabled` 用 `*bool`）
- [x] 9.8 把 design.md 中 5 个 Open Question 的最终结论回填到 MR 描述或 design.md，避免评审阶段重复讨论

## 10. 验收标准覆盖对照

> 实现完成后逐条核对，确保需求文档中的全部验收标准都有对应任务。

| 验收标准 | 覆盖任务 |
|---------|---------|
| AC-001 关联删除无孤儿 | 4.8、8.1 |
| AC-002 保留期内不删 | 4.6、7.5、8.2 |
| AC-003 其他 flow_name 不删 | 4.6、7.5、8.3 |
| AC-004 超期非终态也删 | 4.6、7.5、8.4 |
| AC-005 `enabled=false` 不执行 | 4.4、7.7、8.6 |
| AC-006 slave 跳过 | 5.5、7.10、8.7 |
| AC-007 防重入 | 4.3、5.5、6.4、7.8、8.8 |
| AC-008 无可删记录正常结束 | 4.10、7.9、8.9 |
| AC-009 人工触发 | 6.4、6.5、6.6、8.10 |
| AC-010 日志含条数/耗时/rid | 4.5、4.9、4.12、8.11 |
| AC-011 非法配置启动失败 | 2.3、2.6、7.2、8.12 |
| AC-P01 单批 ≤ 500 | 4.7、4.8、7.6、8.13 |
| AC-P02 批间隔 ≥ 100ms | 4.9、8.14 |
| AC-P03 主从延迟可控 | 8.15 |
| AC-P04 不影响异步调度 | 8.16 |
| AC-S01 不改 binlog_format / 无需额外授权 | 8.13 |
| AC-S02 全部按主键删 | 4.8、8.13 |
| spec: Cleanup Scope Restricted To Whitelisted Flow Names（白名单） | 1.3、4.6、7.5 |
| spec: Start Cursor Location And Cross-Round Reuse（起点游标） | 4.15、4.16、7.12~7.17、8.20 |
