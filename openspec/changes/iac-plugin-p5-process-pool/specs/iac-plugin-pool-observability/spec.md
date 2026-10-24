## ADDED Requirements

### Requirement: 池参数与校验
池 SHALL 通过 `PoolConfig` 结构体接收参数，构造池时传入，池包 MUST NOT 依赖 `pkg/cc` 或任何配置文件；`pkg/cc` 与 yaml 的映射由阶段四（池接入能力）负责。参数及默认值：`maxProcesses`、`maxLeasesPerProcess`（默认 4）、`minIdleProcesses`（默认 2）、`startTimeout`（默认 30 秒）、`configureTimeout`（默认 30 秒）、`closeTimeout`（默认 30 秒）、`maxRetries`（默认 2）、`maxQueueWait`、`maxQueueLength`、`idleReclaimTimeout`（默认 10 分钟）、`maxConfigAge`（默认 30 分钟）、`poolCheckInterval`（默认 30 秒）。校验 MUST 拒绝：`maxProcesses < 1`、`maxLeasesPerProcess < 1`、`minIdleProcesses > maxProcesses`、任何时长为 0 或负数。校验失败时构造池 MUST 返回错误，MUST NOT 返回可用的池。

#### Scenario: 缺省值
- **WHEN** 仅设置 `maxProcesses` 与 `maxQueueLength`，其余参数未设置
- **THEN** 其余参数取默认值，校验通过

#### Scenario: 非法组合
- **WHEN** `minIdleProcesses` 大于 `maxProcesses`
- **THEN** 构造池返回错误

#### Scenario: 非法时长
- **WHEN** `maxConfigAge` 为 0
- **THEN** 构造池返回错误

### Requirement: 池指标
池 SHALL 输出以下 11 个原子指标（前缀 `hcm_plugin_pool_`），平均值、比率、成功率等派生值 MUST NOT 直接导出，由查询侧从原子指标计算：

| 指标 | 类型 | 标签 | 记录点 |
| --- | --- | --- | --- |
| `acquire_seconds` | Histogram | plugin, result | `Acquire` 返回时 |
| `lease_hold_seconds` | Histogram | plugin | lease 归还时 |
| `start_seconds` | Histogram | plugin, result | 拉起进程结束时（含预热） |
| `configure_seconds` | Histogram | plugin, result | `ConfigureProvider` 结束时 |
| `retries_total` | Counter | plugin | 请求的重试次数 `attempts` 加 1 时 |
| `queue_length` | Gauge | plugin | 抓取时 |
| `queue_oldest_wait_seconds` | Gauge | plugin | 抓取时 |
| `processes` | Gauge | plugin, state | 抓取时 |
| `inflight_leases` | Gauge | plugin | 抓取时 |
| `max_processes` | Gauge | plugin | 抓取时 |
| `max_leases_per_process` | Gauge | plugin | 抓取时 |

`acquire_seconds` 的 result 取值为 success、business_error、process_error、timeout、canceled、queue_full、closed、invalid_config；`start_seconds` 为 success、failure；`configure_seconds` 为 success、business_error、transport_error；`processes` 的 state 为 starting、configuring、running、closing。

每个事件 MUST 只有一个记录点：`acquire_seconds` MUST 覆盖 `Acquire` 的全部出口，包括超时、取消与拒绝。指标 MUST 定义在 `pkg/metrics`，使用项目的 `metrics.Register()` 与 `metrics.Namespace`，进程池只调用其导出函数。Gauge MUST NOT 由业务代码主动设置：每个池创建时以插件名为常量标签注册自己的状态采集，抓取时持池锁读取一次快照，同一个池的全部 gauge 来自同一时刻；锁内只做 O(进程数) 的内存读取。同一插件同时只能注册一个池，重复注册时记录错误日志，池照常工作但缺少状态类指标。指标层 MUST NOT 维护插件到池的索引，该索引属于阶段四的池注册表。

#### Scenario: 请求数与成功率
- **WHEN** 查询请求总数、成功数、失败数
- **THEN** 分别为 `acquire_seconds_count` 全部 result 之和、`result="success"`、其余 result，三者天然一致

#### Scenario: 超时请求也被统计
- **WHEN** 请求排队超时或调用方取消
- **THEN** `acquire_seconds` 以 result=timeout 或 canceled 记录一次，耗时为实际等待时长

#### Scenario: 时间平均并发
- **WHEN** 查询一段时间内的平均在途 lease 数
- **THEN** 等于该时间窗内 `lease_hold_seconds_sum` 的增长速率，不受抓取间隔影响

#### Scenario: 抓取时状态一致
- **WHEN** Prometheus 抓取
- **THEN** 同一个池的 `queue_length`、`processes`、`inflight_leases` 等来自同一次持锁读取

### Requirement: 日志与指标脱敏
池的所有日志、错误信息与指标标签 MUST NOT 包含 config 或其中的凭证内容。需要标识配置时 MUST 只输出 key（config 规范化后哈希）的前缀。

#### Scenario: 失败日志
- **WHEN** 配置失败时池写日志
- **THEN** 日志包含插件名、进程 ID、key 前缀与错误原因，不包含 config 内容

#### Scenario: 指标标签
- **WHEN** 检查所有池指标的标签
- **THEN** 标签只含插件名、状态或失败分类，不含 key 或 config
