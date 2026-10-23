## ADDED Requirements

### Requirement: 进程状态与容量占用
池 SHALL 在内存中为每个插件进程维护一条记录，记录受池锁保护、不落库；进程退出即删记录，新进程使用新的 ID，ID MUST NOT 复用。进程状态 MUST 为 `starting`（正在启动握手，无执行器）、`configuring`（正在 `ConfigureProvider`，包括重配）、`running`、`closing`（已开始关闭）。所有状态 MUST 占用容量，只有确认进程退出、或确认创建时无残留后才释放容量。`running` 的使用中 / 空闲 / 已满子状态 MUST 由 `binding` 与 `inflight` 推导，不单独存储。`inflight` 任何时候 MUST 满足 `0 <= inflight <= maxLeasesPerProcess`，且在 `starting` / `configuring` 期间 MUST 为 0。

#### Scenario: 创建中的进程占容量
- **WHEN** 池内有一个 `starting` 进程，`maxProcesses` 为 1
- **THEN** 调度器不再新建进程

#### Scenario: 关闭中仍占容量
- **WHEN** 进程进入 `closing`，尚未确认退出
- **THEN** 容量仍被占用；确认退出并删记录后才释放

#### Scenario: 新进程使用新 ID
- **WHEN** 一个进程退出后又新建进程
- **THEN** 新进程 ID 与已退出进程不同

### Requirement: 异步创建与配置，锁外执行并写回
创建（`Start`）与配置（`ConfigureProvider`）、关闭（`Close`）和业务 RPC MUST 在池锁外执行，锁内 MUST 只做内存读写。池 MUST 使用只有一把锁的模型：锁内改状态、锁外做慢操作、锁内写回。每个进程同一时刻 MUST 至多存在一个异步任务；每个异步任务 MUST 写回恰好一次。创建与配置使用池自己的 ctx 加超时（`startTimeout`、`configureTimeout`），MUST NOT 跟随请求 ctx 取消：请求取消后进程仍然创建 / 配置完成，作为空闲进程留给后面的队首。任务 panic MUST 被 recover，并按传输错误处理，仍然写回一次。写回前 MUST 核对进程仍在进程表中；不在则不写进程记录，但仍要处理其请求。

#### Scenario: 配置成功后交付
- **WHEN** `ConfigureProvider` 成功返回
- **THEN** 池在锁内设置 `boundAt`，进程转 `running`，向触发请求交付 lease，`inflight` 加 1，再触发一次调度

#### Scenario: 请求取消后进程保留
- **WHEN** 触发创建的请求在进程启动期间被取消
- **THEN** 进程仍创建完成并作为未配置的空闲进程保留，供后续队首使用，不因请求取消而被关闭；因已没有请求在等，MUST NOT 再执行 `ConfigureProvider`（避免无谓地把凭证写进进程）

#### Scenario: 外部调用不响应 ctx
- **WHEN** `Start` 或 `ConfigureProvider` 超过各自超时仍未返回（实现未响应 ctx）
- **THEN** 池不再等待，按超时失败写回、请求收到可用 `errors.Is(err, context.DeadlineExceeded)` 识别的错误，并关闭该进程，由进程管理的 `Close` 让阻塞的调用返回；迟到的结果被丢弃

#### Scenario: 配置开始即覆盖 binding
- **WHEN** 进程开始重配为新 key
- **THEN** `binding` 立即写为新目标（`boundAt` 为空），旧配置不再被规则 1 命中

#### Scenario: 任务 panic
- **WHEN** 异步任务发生 panic
- **THEN** 被恢复并按传输错误处理，进程被关闭，请求不会被遗落，池仍可继续调度

#### Scenario: 写回时进程已不在
- **WHEN** 任务写回时进程已被移出进程表
- **THEN** 不修改任何进程记录，但其请求按失败流程处理

### Requirement: 失败处理（阶段一）
阶段一 SHALL 采用最简失败语义：`Start` 失败、`ConfigureProvider` 失败或超时的，触发该次任务的请求 MUST 直接返回错误，MUST NOT 重试；对应进程 MUST 转 `closing` 并在锁外关闭，确认退出后才删记录。返回给调用方的错误 MUST NOT 包含 config 或凭证内容。

#### Scenario: 启动失败
- **WHEN** `Start` 返回错误
- **THEN** 确认无残留后删除记录、释放容量，请求收到不含凭证的错误

#### Scenario: 配置失败
- **WHEN** `ConfigureProvider` 返回错误或超时
- **THEN** 进程转 `closing` 并被关闭，请求收到错误

#### Scenario: 无法确认无残留
- **WHEN** `Start` 失败后无法确认进程已退出
- **THEN** 进程转 `closing`，由关闭流程兜底，容量暂不释放

### Requirement: 归还与调度联动
`Release()` SHALL 在池锁内完成：进程不存在则忽略；`inflight` 减 1；`inflight` 归 0 且为 `running` 时记录 `idleSince`；池已关闭时该进程转 `closing`；`closing` 的进程归还后 MUST NOT 回到空闲。归还完成后 MUST 触发一次调度，把空出的名额交给队首。`Release()` MUST NOT 阻塞。

#### Scenario: 归还后唤醒队首
- **WHEN** 队首因容量已满在等待，某 lease 归还
- **THEN** 同一次 `Release` 内触发调度，队首获得名额

#### Scenario: closing 进程归还
- **WHEN** 处于 `closing` 的进程上最后一个 lease 归还
- **THEN** 只递减 `inflight`，进程不会变为 `running` 空闲

### Requirement: 池关闭
`Close(ctx)` SHALL 停止接收新请求，用 `queue.Drain()` 取出所有排队请求并返回"池已关闭"错误；空闲进程 MUST 立即转 `closing` 并关闭；有 lease 的进程 MUST 调用进程对象的 `Close(ctx)`，由其等待在途 RPC 返回、超时后强制退出，池 MUST NOT 自行统计在途 RPC。`starting` / `configuring` 的进程 MUST 等任务写回后再关闭。传入的 ctx 超时后 MUST 强制结束剩余进程。处于 Plan 与 Apply 之间的 lease，其后续调用 MAY 返回错误。

#### Scenario: 排空队列
- **WHEN** 池关闭时队列中有 3 个等待请求
- **THEN** 3 个请求都收到"池已关闭"错误，且不再被调度

#### Scenario: 创建中关闭
- **WHEN** 池关闭时某进程正在 `starting`
- **THEN** 池不等待创建完成才返回，创建完成写回时该进程直接转 `closing` 并关闭

#### Scenario: 关闭超时
- **WHEN** `Close(ctx)` 的 ctx 超时，仍有进程未退出
- **THEN** 强制结束这些进程
