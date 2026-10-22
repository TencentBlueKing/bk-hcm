## ADDED Requirements

### Requirement: 池参数与校验
池 SHALL 通过 `PoolConfig` 结构体接收参数，构造池时传入，池包 MUST NOT 依赖 `pkg/cc` 或任何配置文件；`pkg/cc` 与 yaml 的映射由阶段四（池接入能力）负责。参数及默认值：`maxProcesses`、`maxLeasesPerProcess`（默认 1）、`minIdleProcesses`（默认 1）、`startTimeout`（默认 30 秒）、`configureTimeout`（默认 30 秒）、`maxRetries`（默认 2）、`maxQueueWait`、`maxQueueLength`、`idleReclaimTimeout`（默认 10 分钟）、`maxConfigAge`（默认 30 分钟）、`poolCheckInterval`（默认 30 秒）、`leaseHoldWarn`。校验 MUST 拒绝：`maxProcesses < 1`、`maxLeasesPerProcess < 1`、`minIdleProcesses > maxProcesses`、任何时长为 0 或负数。校验失败时构造池 MUST 返回错误，MUST NOT 返回可用的池。

#### Scenario: 缺省值
- **WHEN** 仅设置 `maxProcesses` 与 `maxQueueLength`，其余参数未设置
- **THEN** 其余参数取默认值，校验通过

#### Scenario: 非法组合
- **WHEN** `minIdleProcesses` 大于 `maxProcesses`
- **THEN** 构造池返回错误

#### Scenario: 非法时长
- **WHEN** `maxConfigAge` 为 0
- **THEN** 构造池返回错误

### Requirement: lease 泄漏告警
池 SHALL 记录每个 lease 的获取时间。持有时长超过 `leaseHoldWarn` 时 MUST 输出告警日志并增加告警指标，日志 MUST 包含插件名、进程 ID 与持有时长。第一版 MUST NOT 强制回收 lease，也 MUST NOT 因告警改变进程状态。同一个 lease 的告警 MUST 只输出一次，MUST NOT 每个巡检周期重复输出。lease 归还后 MUST 不再告警。

#### Scenario: lease 持有过久
- **WHEN** 某 lease 持有时长超过 `leaseHoldWarn`
- **THEN** 输出一条告警日志并增加告警指标，lease 与进程状态不变

#### Scenario: 不重复告警
- **WHEN** 同一个超时 lease 经过多个巡检周期仍未归还
- **THEN** 只告警一次

#### Scenario: 归还后不再告警
- **WHEN** 超时的 lease 随后被归还
- **THEN** 之后不再产生该 lease 的告警

### Requirement: 池指标
池 SHALL 输出以下指标：当前排队数、请求排队时长、各状态（`starting / configuring / running / closing`）进程数、空闲进程数、`inflight` 总数、重试次数、失败次数（按业务错误 / 传输错误 / 启动失败分类）、被队列满拒绝的次数、排队超时的次数。指标 MUST 以插件名区分，名称与注册方式沿用仓库既有的 `pkg/metrics` 习惯。

#### Scenario: 排队时长
- **WHEN** 请求在队列中等待后获得 lease
- **THEN** 排队时长被记录到指标

#### Scenario: 队列满拒绝
- **WHEN** 请求因队列已满被拒绝
- **THEN** 拒绝次数指标加 1

#### Scenario: 失败分类
- **WHEN** 配置因业务错误失败
- **THEN** 业务错误失败次数加 1，传输错误与启动失败次数不变

### Requirement: 日志与指标脱敏
池的所有日志、错误信息与指标标签 MUST NOT 包含 config 或其中的凭证内容。需要标识配置时 MUST 只输出 key（config 规范化后哈希）的前缀。

#### Scenario: 失败日志
- **WHEN** 配置失败时池写日志
- **THEN** 日志包含插件名、进程 ID、key 前缀与错误原因，不包含 config 内容

#### Scenario: 指标标签
- **WHEN** 检查所有池指标的标签
- **THEN** 标签只含插件名、状态或失败分类，不含 key 或 config
