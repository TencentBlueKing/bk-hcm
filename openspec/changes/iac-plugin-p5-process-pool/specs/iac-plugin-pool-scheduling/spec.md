## ADDED Requirements

### Requirement: 池交付 lease，不执行业务请求
每个 tf_provider 插件 SHALL 有且仅有一个进程池。调用方 MUST 通过 `Acquire(kt, config)` 申请 lease，成功返回的 lease 对应一个已按该 config 配置好的插件进程上的一个并发名额。池 MUST NOT 自行发起业务请求；业务方法由调用方在持有期内通过 lease 调用，同一 lease 的多次调用 MUST 落在同一进程上。`Lease` MUST 是池对 `tfruntime.Executor` 的封装，只暴露业务方法（`ValidateResourceTypeConfig`、`ValidateDataSourceConfig`、`UpgradeResourceState`、`ReadResource`、`PlanResourceChange`、`ApplyResourceChange`、`ImportResourceState`、`ReadDataSource`），MUST NOT 暴露 `ConfigureProvider` 与 `GetProviderSchema`。lease 已归还后再调用任一业务方法 MUST 直接返回错误，MUST NOT 转发到进程。归还 MUST 只通过 `Lease.Release()`，不带参数、幂等；进程是否可用 MUST 由池判断，不由调用方决定。

#### Scenario: 获取并归还
- **WHEN** 调用方 `Acquire` 成功，使用后调用 `Release()`
- **THEN** 该进程的 `inflight` 减 1，空出的名额按调度规则交给队首请求

#### Scenario: 重复归还
- **WHEN** 同一个 lease 被 `Release()` 两次
- **THEN** 第二次不产生任何效果，`inflight` 不会被多减

#### Scenario: 进程已退出后归还
- **WHEN** lease 对应的进程已被移出进程表，调用方才 `Release()`
- **THEN** 归还被忽略，不影响后来创建的进程（进程 ID 不复用）

#### Scenario: 归还后调用
- **WHEN** lease 已 `Release()`，调用方再调用 `PlanResourceChange`
- **THEN** 返回"lease 已归还"错误，请求不会发到进程

#### Scenario: 调用方无法改配置
- **WHEN** 检查 `Lease` 的方法集
- **THEN** 不含 `ConfigureProvider` 与 `GetProviderSchema`

#### Scenario: 同一次业务内的 Plan 与 Apply
- **WHEN** 调用方在一个 lease 持有期内先后发起 Plan 与 Apply
- **THEN** 两次调用使用同一个进程

### Requirement: 单条严格 FIFO 等待队列
所有请求 SHALL 共用一条等待队列，不按 config 分组，不区分调用方优先级。请求到达时 MUST 进入队尾；调度器 MUST 只从队首取。请求出队后 MUST 要么拿到 lease，要么进入某个进程的预约列表，队列中 MUST NOT 保留准备中的请求。队列有效请求数达到 `maxQueueLength` 时新请求 MUST 被拒绝；单个请求等待超过 `maxQueueWait` 或其 `kt.Ctx` 结束时 MUST 返回错误。已取消的请求 MUST NOT 从队列中间删除，而是记为已取消，待走到队首时清掉，且 MUST NOT 计入队列长度。

#### Scenario: 按到达顺序派发
- **WHEN** 请求 A、B、C 依次入队，且资源足够逐个派发
- **THEN** 派发顺序为 A、B、C

#### Scenario: 队列已满
- **WHEN** 队列有效请求数已达 `maxQueueLength`
- **THEN** 新的 `Acquire` 立即返回"等待者过多"错误，不入队

#### Scenario: 排队超时
- **WHEN** 请求在队列中等待超过 `maxQueueWait`
- **THEN** `Acquire` 返回超时错误，请求被标记取消

#### Scenario: 已取消请求不占队列长度
- **WHEN** 队列中间的请求被取消
- **THEN** `Len()` 立即减 1，该元素在走到队首时被跳过，不被派发

### Requirement: 调度只看队首
调度循环 SHALL 每次只取队首请求并为其选择进程；选中则出队并派发后继续看下一个队首；选不出进程时 MUST 结束本轮，MUST NOT 越过队首去派发后面的请求。选进程 MUST 按下列优先级：
1. 状态为 `running`，配置已生效且未过期、key 与请求一致、`inflight < maxLeasesPerProcess`：直接交付，`inflight` 加 1；多个候选时 `inflight` 少者优先，平局按进程 ID。
2. 状态为 `running`、`inflight == 0` 且未配置或配置已过期：配置成请求的 config。
3. 状态为 `running`、`inflight == 0` 且已生效的是别的 config：重配成请求的 config。
4. 以上均无且池内进程数小于 `maxProcesses`：新建进程再配置。
5. 以上均无：结束本轮，队首及其后请求继续等待。

选进程 MUST 只遍历进程表，耗时与队列长度无关。一轮调度内判断配置是否过期所用的时间 MUST 只取一次。

#### Scenario: 命中已配好的进程
- **WHEN** 队首 key 为 K，存在 `running`、配置已生效且 key 为 K、有空位的进程
- **THEN** 直接交付 lease，不触发配置

#### Scenario: 复用空闲进程重配
- **WHEN** 队首 key 为 K，无匹配进程，存在 `inflight == 0` 的空闲进程（配置了别的 key）
- **THEN** 该进程被重配为 K，状态转 `configuring`，不新建进程

#### Scenario: 新建进程
- **WHEN** 队首 key 为 K，无可复用进程，且进程数小于 `maxProcesses`
- **THEN** 登记一个 `starting` 进程占用容量，在锁外启动并配置

#### Scenario: 队首派不出去时后面不越队
- **WHEN** 队首请求（C3）无匹配、无空闲进程、容量已满，其后的请求（C1）本可填某进程的空位
- **THEN** 本轮结束，C1 继续排队等待，直到 C3 被派发或离队

#### Scenario: 一轮连续派发
- **WHEN** 队列中有三个请求，第一个触发新建、第二个触发新建、第三个命中已有空位
- **THEN** 同一轮调度内依次派发，慢操作均在锁外异步进行

#### Scenario: 有在途 lease 的进程不被重配
- **WHEN** 某进程 `inflight > 0` 且配置为别的 key
- **THEN** 它不是规则 2、3 的候选

### Requirement: 配置有效期
池 SHALL 使用 `maxConfigAge` 限制一次配置被复用的时间，到期在调度时判断，MUST NOT 依赖定时解绑。配置已生效的判据为：`binding != nil` 且 `boundAt != nil` 且当前时间早于 `boundAt + maxConfigAge`。正在配置（`boundAt == nil`）的进程 MUST NOT 被当作已生效。过期的 `binding` 不清除，调度时按"未配置"对待，下一次配置开始时覆盖。过期 MUST 通过 `ConfigureProvider` 重配解决，MUST NOT 关闭进程。过期只阻止新交付，MUST NOT 中断已持有的 lease。

#### Scenario: 过期后不再直接交付
- **WHEN** 进程配置生效已超过 `maxConfigAge`，且空闲
- **THEN** 规则 1 不再选中它；同 key 的请求会通过规则 2 对其重配

#### Scenario: 过期不影响在途 lease
- **WHEN** 进程的配置在 lease 持有期内过期
- **THEN** 已持有的 lease 照常调用，仅新请求不能直接交付该进程

#### Scenario: 正在配置不算已生效
- **WHEN** 进程处于 `configuring`，`boundAt` 为空
- **THEN** 规则 1 不选它，不向其交付 lease

### Requirement: 容量与取消
池的容量参数 SHALL 为：`maxProcesses` 限制池内全部进程数（含 `starting`、`configuring`、`closing`）；`maxLeasesPerProcess` 限制单进程的 lease 数（阶段二起含预约数）。并发名额 MUST 只由队首取用。在 provider 并发能力确认前，`maxLeasesPerProcess` 默认值 SHALL 为 1。`Acquire` 的 ctx 取消时：请求仍在队列中则标记取消并出队；请求已交付但调用方未取走则 MUST 立即归还该 lease。交付与取消 MUST 在池锁内以同一个 `done` 标志互斥，保证不漏还、不重复还。

#### Scenario: 默认一个进程一个并发
- **WHEN** 未配置 `maxLeasesPerProcess`
- **THEN** 每个进程同一时刻至多交付 1 个 lease

#### Scenario: 排队中取消
- **WHEN** 请求在队列中时调用方 ctx 取消
- **THEN** 请求被标记 `done`，不再派发，`Acquire` 返回 ctx 错误

#### Scenario: 交付与取消竞争
- **WHEN** lease 已写入请求的结果通道，同时调用方 ctx 取消
- **THEN** 取消流程取出该 lease 并归还，`inflight` 恢复，不泄漏名额

#### Scenario: 池已关闭
- **WHEN** 池已关闭后调用 `Acquire`
- **THEN** 立即返回"池已关闭"错误
