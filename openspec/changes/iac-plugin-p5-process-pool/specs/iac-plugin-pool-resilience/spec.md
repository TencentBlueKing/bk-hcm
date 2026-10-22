## ADDED Requirements

### Requirement: 准备期预约
调度规则 SHALL 增加规则 1b：状态为 `starting` / `configuring`、`binding.key` 与队首一致且 `boundAt` 为空、`len(pending) < maxLeasesPerProcess` 的进程，队首请求 MUST 出队并加入该进程的 `pending`，不触发新的配置或创建。选择优先级位于"直接交付"之后、"空闲进程配置"之前。`pending` 的长度（含触发者）MUST NOT 超过 `maxLeasesPerProcess`。配置成功后，池 MUST 先设置 `boundAt`，再按 `pending` 顺序逐个交付并递增 `inflight`，跳过已取消的请求，最后清空 `pending` 并转 `running`。预约 MUST NOT 构成插队：只有当前队首才可能被预约。

#### Scenario: 同配置突发共享一个进程
- **WHEN** `maxLeasesPerProcess=3`，三个相同 key 的请求先后到达，第一个触发新建进程
- **THEN** 第二、第三个请求被预约到同一进程的 `pending`，配置成功后三个 lease 一起交付，不新建第二个进程

#### Scenario: 预约已满
- **WHEN** 目标进程的 `pending` 已有 `maxLeasesPerProcess` 个请求
- **THEN** 后续同 key 请求不能再预约，按后续规则处理

#### Scenario: 预约不改变顺序
- **WHEN** 队首为 C2，后面有同 key 的 C2' 与别的 key 的 C1
- **THEN** 仅当 C2 出队后 C2' 才能成为队首并预约，C1 不会越过它们

#### Scenario: 已取消的预约在交付时跳过
- **WHEN** `pending` 中某请求已取消，配置成功交付时
- **THEN** 该请求被跳过，不产生 lease，其占用的预约名额在这次准备结束时释放

### Requirement: 失败分流与重试
进程侧失败 SHALL 按下表处理，业务错误 MUST 与传输错误区分：

| 失败 | 进程 | 请求 |
| --- | --- | --- |
| `ConfigureProvider` 返回插件业务错误（缺密钥、鉴权失败等） | 回 `running`，`binding = nil` | 触发者与全部预约一起直接失败，不重试 |
| `Start` 失败或超时 | 确认无残留后删记录 | `pending` 按原相对顺序回队尾，`attempts` 超过 `maxRetries` 则失败 |
| `ConfigureProvider` 超时、传输错误 | 转 `closing` | 同上 |
| 准备期间进程退出 | 删记录 | 同上 |
| 池已关闭 | 按关闭流程 | 返回"池已关闭"错误 |

重试的请求 MUST 放回队尾，与新增请求同等对待，MUST NOT 保留原位置；重试 MUST 仍使用同一份 config，并仍受请求 ctx 与 `maxQueueWait` 约束；已 `done` 的请求 MUST NOT 被重试。进程转入重试路径时 MUST 先取走 `pending` 并置空，防止被重复回收。池 MUST NOT 引入熔断。

#### Scenario: 业务错误整批失败
- **WHEN** 配置因缺少密钥被插件拒绝，`pending` 中有 3 个请求
- **THEN** 3 个请求都收到该业务错误，不重试，进程回到 `running` 且 `binding` 为空，可用于下一个请求

#### Scenario: 传输错误后重试
- **WHEN** 配置超时，`pending` 中有 2 个请求，`maxRetries=2`
- **THEN** 进程被关闭；2 个请求按原相对顺序回队尾，`attempts` 各加 1

#### Scenario: 重试耗尽
- **WHEN** 一个请求的 `attempts` 已超过 `maxRetries`
- **THEN** 不再入队，直接返回最近一次失败原因

#### Scenario: 重复回收不重复入队
- **WHEN** 定时任务先回收了某进程的 `pending`，随后该进程的异步任务才写回
- **THEN** 写回时 `pending` 为空，不会重复入队

#### Scenario: 失败不泄漏凭证
- **WHEN** 任一失败路径写日志或返回错误
- **THEN** 内容不含 config 或凭证，只含 key 前缀与错误原因

### Requirement: 取消的三种位置
请求取消 SHALL 按所处位置处理：在队列中则置 `done`、`queue.Remove`、`queued=false`；在进程 `pending` 中则只置 `done`，不从 `pending` 移除，配置继续进行，交付与回收时跳过；已交付则取出 lease 并立即归还。取消与交付 MUST 在池锁内以 `done` 互斥，二者只有一个生效。取消流程需要归还 lease 时 MUST 先释放池锁，避免重入死锁。

#### Scenario: 预约中的请求取消
- **WHEN** 请求已在某进程 `pending` 中，调用方取消
- **THEN** 配置继续，交付时跳过该请求，其他预约者不受影响

#### Scenario: 取消触发调度
- **WHEN** 队首请求被取消
- **THEN** 取消后立即触发调度，后一个请求成为队首并继续派发

### Requirement: 池内定时巡检
每个池 SHALL 按 `poolCheckInterval` 周期巡检自己的进程，慢操作在锁外执行。

**空闲进程**的口径 MUST 统一为：状态为 `running` 且 `inflight == 0` 的进程，不论是否已配置、配置是否已过期；尚未被任何请求认领的预热进程（`starting` 且 `pending` 为空）MUST 按"将成为空闲"计入。预热与回收 MUST 使用同一个空闲数 `idle` 与同一个下限 `minIdleProcesses`（记为 `min`）：
- 预热：`idle < min` 时补建 `min - idle` 个未配置进程（缺口计入 `starting`，不计 `configuring`）；容量有余才补，失败按退避重试；池满时 MUST NOT 挤占服务中的进程。
- 空闲回收：**仅当 `idle > min` 时才回收**；回收对象 MUST 是空闲时长超过 `idleReclaimTimeout` 的空闲进程，按空闲最久优先，**一轮最多回收 `idle - min` 个**，使回收后空闲数不低于 `min`；`idle <= min` 时 MUST NOT 回收任何进程。
- 退出检查：`Exited()` 为真的进程 MUST 删记录、释放容量、回收其 `pending`，并触发调度。
- 连接检查：对 `running` 状态的进程调用 `CheckHealth()`，报错则转 `closing`，确认退出后删记录；`starting` / `configuring` 的进程 MUST NOT 做健康检查，由任务超时处理。
- 每轮巡检末尾 MUST 无条件触发一次调度，作为事件遗漏的兜底。

进程是否可用 MUST 只依据 `Exited()`、`CheckHealth()` 与 `ConfigureProvider` 失败，MUST NOT 依据业务调用失败判断。

#### Scenario: 预热补足
- **WHEN** `minIdleProcesses=1`，池内没有任何空闲进程且容量有余
- **THEN** 异步新建一个未配置的进程

#### Scenario: 已配置的空闲进程也算空闲
- **WHEN** `minIdleProcesses=1`，池内有一个空闲且已配置的进程，没有未配置进程
- **THEN** `idle == min`，不补建预热进程

#### Scenario: 池满不预热
- **WHEN** 池内进程数已达 `maxProcesses`
- **THEN** 不新建预热进程，也不关闭服务中的进程

#### Scenario: 空闲数不超过下限不回收
- **WHEN** `minIdleProcesses=1`，仅有 1 个空闲进程且已超过 `idleReclaimTimeout`
- **THEN** 不回收

#### Scenario: 只回收超出下限的部分
- **WHEN** `minIdleProcesses=2`，有 5 个空闲进程且全部超过 `idleReclaimTimeout`
- **THEN** 本轮回收其中空闲最久的 3 个，剩余 2 个

#### Scenario: 未超时的不回收
- **WHEN** `minIdleProcesses=1`，有 3 个空闲进程，仅 1 个超过 `idleReclaimTimeout`
- **THEN** 只回收那 1 个

#### Scenario: 回收后被占用由预热补回
- **WHEN** 回收后空闲数等于 `min`，随后一个请求占用了该空闲进程
- **THEN** 下一轮巡检 `idle < min`，补建一个未配置进程

#### Scenario: 进程意外退出
- **WHEN** 某 `running` 进程 `Exited()` 为真
- **THEN** 删除记录、释放容量，已交付 lease 在归还时被忽略，其后续 RPC 返回错误

#### Scenario: 连接不可用
- **WHEN** `CheckHealth()` 对某 `running` 进程返回错误
- **THEN** 该进程转 `closing`，确认退出后才删记录

#### Scenario: 兜底调度
- **WHEN** 某事件触发点遗漏了调度，队首有可用名额但没人唤醒
- **THEN** 下一轮巡检的无条件调度会把名额派给队首
