# clb-res-lock-observability Specification

## Purpose
TBD - created by archiving change clb-res-lock-observability. Update Purpose after archive.
## Requirements
### Requirement: 变更范围限定为纯日志

本能力的实现 MUST 只增加日志语句，MUST NOT 改动代码结构。

具体约束：

- MUST NOT 修改任何函数、方法、接口的签名，包括 DAO 层与 client 层。
- MUST NOT 给框架结构体新增字段（如 `Task`、`Flow`、`InitPayload`、`FlowSlaveOperateWatchOption`），也 MUST NOT 为传递日志字段而改变数据在框架内的流转方式。
- MUST NOT 为打日志新增 DB 查询。日志只能使用当前作用域已有的变量。
- MUST NOT 修改对外错误消息、错误码、返回结构。
- MUST NOT 修改加解锁时机、flow/task 状态机、调度策略与参数。
- MUST NOT 合并 `lockResFlowStatus` / `checkResFlowRel` 的多份同构实现。
- 允许把丢弃已有返回值的 `_, err :=` 改为接收该返回值用于打日志，因为这是获取 `owner`、watch flowID 的唯一途径，且不改签名与行为。

当前作用域拿不到的字段 MUST NOT 通过扩展结构体或新增查询来获取，而 MUST 改为由两条日志的时间戳差、或同一 flowID 的另一条日志来还原。

#### Scenario: 实现不含结构性改动

- **WHEN** 审查本变更的代码 diff
- **THEN** 不存在任何函数、方法或接口签名的修改
- **AND** 不存在结构体字段的新增
- **AND** 不存在为打日志而新增的 DB 查询
- **AND** 不存在对外错误消息的修改

#### Scenario: 跨作用域字段用时间戳差替代

- **WHEN** 某段耗时的两个端点分别位于不同函数、且时间戳无法在不加结构体字段的前提下传递
- **THEN** 在两个端点各打一条日志，由日志时间戳差还原该耗时
- **AND** 不给任何结构体新增时间戳字段

#### Scenario: 允许接收已被丢弃的返回值

- **WHEN** 某调用点原本以 `_, err :=` 丢弃 `checkResFlowRel` 或 `CreateTemplateFlow` 的返回值
- **THEN** 允许改为接收该返回值并用于日志
- **AND** 被调用函数自身的签名与行为保持不变

### Requirement: 资源锁日志的公共字段约定

CLB 资源锁（`resource_flow_lock`）相关的所有日志 MUST 同时包含 `resID`、`resType`、涉及的 flowID 与 rid。

由于异步执行阶段的 rid 无法回溯到调用方的原始请求（`task.Kit` 的根 kit 由 scheduler 每轮轮询新建），rid MUST NOT 作为跨服务关联的唯一依据；`resID` 与 flowID MUST 出现在每条日志中作为关联键。

按项目日志规范，`rid: %s` MUST 位于日志消息末尾，错误信息的 `err: %v` MUST 位于消息前部。

#### Scenario: 任意资源锁日志的字段完整性

- **WHEN** 打印任一条资源锁相关日志（加锁、解锁、撞锁、watch 生命周期）
- **THEN** 该条日志同时含有 `resID`、`resType`、相关 flowID
- **AND** 该条日志以 `rid: <rid>` 结尾

#### Scenario: 错误日志的字段顺序

- **WHEN** 打印含错误的资源锁日志
- **THEN** `err: <error>` 出现在业务上下文字段之前
- **AND** `rid: <rid>` 出现在消息末尾

### Requirement: 加锁成功记录日志

对 CLB 资源加锁成功后 MUST 打印 Info 级别日志，记录持锁归属与任务类型。

该要求同时适用于 cloud-server 侧的加锁封装函数与 data-service 侧的加锁接口。cloud-server 中存在多份同构实现（`logics/load-balancer`、`service/load-balancer`、`logics/cvm`），MUST 全部覆盖，否则部分入口将缺失日志。

#### Scenario: cloud-server 加锁成功

- **WHEN** `lockResFlowStatus` 完成加锁并成功把主 flow 状态由 Init 更新为 Pending
- **THEN** 打印 Info 日志，含 `resID`、`resType`、`owner`（即主 flowID）、`taskType`、rid

#### Scenario: data-service 加锁成功

- **WHEN** `ResFlowLock` 接口成功插入 `resource_flow_lock` 记录并创建 `resource_flow_rel` 关联记录
- **THEN** 打印 Info 日志，含 `resID`、`resType`、`owner`、`taskType`、`status`、rid

### Requirement: 撞锁时记录持锁 flow 的归属

资源锁预检测命中时 MUST 打印日志并给出当前持锁的 flowID 与加锁时刻，使调用方能定位是哪个 flow 未释放锁。

`checkResFlowRel` 已返回含 `Owner` 与 `CreatedAt` 的持锁记录，调用方 MUST 使用该返回值而非丢弃它。

预检测的两类命中 MUST 分别打印且可区分：锁表命中表示锁仍被持有；`resource_flow_rel` 中存在 `executing` 状态记录表示锁已释放但关联记录未收尾。

#### Scenario: 锁表命中

- **WHEN** 预检测在 `resource_flow_lock` 中查到该 `(res_id, res_type)` 的记录
- **THEN** 打印日志，含 `resID`、`resType`、持锁 `owner` flowID、该锁的 `createdAt`、当前调用入口、rid

#### Scenario: 关联记录未收尾

- **WHEN** 预检测在 `resource_flow_lock` 中未查到记录，但在 `resource_flow_rel` 中查到状态为 `executing` 的记录
- **THEN** 打印日志明确指出属于关联记录未收尾而非锁被持有，含 `resID`、`resType`、该 rel 记录的 `flowID`、rid

#### Scenario: 对外错误消息保持不变

- **WHEN** 因预检测命中返回 `errf.LoadBalancerTaskExecuting` 错误
- **THEN** 错误消息与变更前完全一致，MUST NOT 追加持锁 flowID 或其他字段
- **AND** 持锁 flowID 仅出现在日志中

### Requirement: 解锁与锁泄漏可观测

解锁成功 MUST 打印 Info 日志。「主 flow 已终态但锁不在该 flow 名下」的情况 MUST 被记录，使锁泄漏可被发现。

`ResFlowUnLock` 按 `(res_id, res_type, owner=flowID)` 删除，owner 不匹配时删除 0 行且不产生错误。该情况 MUST NOT 通过修改 `ResourceFlowLockDao.DeleteWithTx` 签名或在 `ResFlowUnLock` 内新增查询来暴露，而 MUST 利用 watch action 中现成的查询结果——`processResFlow` 终态分支的 `queryResFlowLock` 已按 `owner = flowID` 过滤，其空结果即该信号，当前被静默丢弃。

该情况 MUST NOT 作为错误返回，因为 watch 重试会导致重复解锁，空结果是幂等场景下的正常结果。

#### Scenario: 解锁成功

- **WHEN** `ResFlowUnLock` 成功删除锁记录并更新 `resource_flow_rel` 状态
- **THEN** 打印 Info 日志，含 `resID`、`resType`、请求的 `owner` flowID、目标 `status`、rid

#### Scenario: 主 flow 已终态但锁不在该 flow 名下

- **WHEN** watch action 观察到主 flow 进入终态，且按 `owner = 主flowID` 查询锁表返回空
- **THEN** 打印 Warn 日志，含主 flowID、`resID`、`resType`、观察到的终态、rid
- **AND** 该分支保持原有的跳过行为，不返回错误
- **AND** 不为此新增任何查询

### Requirement: watch flow 与主 flow 的关联可追溯

创建 watch flow 后 MUST 打印 Info 日志记录 watch flowID 与主 flowID 的对应关系，因为执行侧无法获知自身所属的 watch flowID。

各 executor 的 `createFlowTask` 当前丢弃 `CreateTemplateFlow` 的返回值，MUST 改为接收并记录。

#### Scenario: watch flow 创建成功

- **WHEN** 调用 `CreateTemplateFlow` 成功创建 `FlowSlaveOperateWatch` 类型的 watch flow
- **THEN** 打印 Info 日志，含 watch flowID、被监听的主 flowID、`resID`、`resType`、`taskType`、rid

### Requirement: watch action 的生命周期与解锁延迟可观测

`FlowSlaveOperateWatchAction` MUST 记录进入、观察到主 flow 终态、解锁完成三个节点，并给出解锁延迟的量级。

轮询循环内的每一跳 MUST NOT 打印 Info 级别日志（单个 watch task 最长运行 5 分钟、每 500ms 一跳），MUST 使用 `logs.V(3)` 或仅在状态变化时打印。

#### Scenario: watch action 开始执行

- **WHEN** `FlowSlaveOperateWatchAction.Run` 开始执行
- **THEN** 打印 Info 日志，含被监听的主 flowID、`resID`、`resType`、`taskType`、rid

#### Scenario: 观察到主 flow 进入终态

- **WHEN** watch 轮询到主 flow 状态为 `success`、`failed` 或 `cancel`
- **THEN** 打印 Info 日志，含主 flowID、观察到的终态、从主 flow 终态时刻到本次观察的延迟秒数、`resID`、rid
- **AND** 该延迟字段的命名体现其为量级估算（依据 DB 秒级 `updated_at`，非精确计时）

#### Scenario: 解锁完成

- **WHEN** watch 完成 `ResFlowUnLock` 调用
- **THEN** 打印 Info 日志，含主 flowID、`resID`、目标 `resStatus`、从主 flow 终态时刻到解锁完成的总延迟秒数、rid

#### Scenario: watch 等待锁创建期间不刷日志

- **WHEN** 主 flow 处于 `init` 状态且锁尚未创建，watch 持续等待
- **THEN** 该等待状态在一次 watch 执行内至多打印一条 Warn 日志，后续轮询不重复打印 Info 或 Warn

#### Scenario: watch 等待超时

- **WHEN** watch 达到 `OperateWatchTimeout` 仍未观察到主 flow 终态
- **THEN** 打印 Error 日志，含主 flowID、`resID`、`resType`、已等待时长、rid

