## ADDED Requirements

### Requirement: 创建后等待云上可见

系统 SHALL 提供公共等待函数。入参为只包含 `ListLoadBalancer(kt, *TCloudListOption)` 方法的接口、地域、本次新建的云 ID 列表，以及可选的等待参数（为空时使用默认常量）。

函数 SHALL 立即查询一次；未全部查到时按 `retry.RetryPolicy` 的退避间隔重试（默认参数下依次约为 0、500ms、1s、1.5s、2s，之后在 1–2.5 秒间随机），默认总上限 30 秒，上限与间隔 SHALL 用常量定义。每轮 SHALL 按 `constant.TCLBDescribeMax`（20）分批查询。全部查到即返回空列表；到达上限时 SHALL 记 Warn 日志（含仍查不到的云 ID 和 rid），返回仍查不到的云 ID，不返回错误。云上查询报错 SHALL 按未查到处理并继续重试；到达上限时最后一轮仍报错，则返回全部目标云 ID。

#### Scenario: 首轮全部可见
- **WHEN** 第 1 轮查询即返回全部目标云 ID
- **THEN** 函数 SHALL 只查询 1 轮，立即返回空列表

#### Scenario: 中途全部可见时提前结束
- **WHEN** 第 1 轮查询不返回目标云 ID，第 2 轮返回全部目标云 ID
- **THEN** 函数 SHALL 在第 2 轮后返回空列表，不等满上限

#### Scenario: 到达上限仍查不到
- **WHEN** 整个等待期内查询都不返回某个目标云 ID
- **THEN** 函数 SHALL 在上限加一轮最大间隔和一次查询的时间内返回，返回值包含该云 ID，不返回错误，并记 Warn 日志

#### Scenario: 超过 20 个 ID 时分批
- **WHEN** 目标云 ID 有 25 个
- **THEN** 每轮 SHALL 发 2 次查询，每次不超过 20 个 ID

#### Scenario: 云上查询一直报错
- **WHEN** 查询在整个等待期内都返回错误
- **THEN** 函数 SHALL 不返回错误，返回全部目标云 ID，并记 Warn 日志

### Requirement: 创建入口先等待再同步建记录

公有云创建入口 `BatchCreateTCloudClb` 在创建任务成功后，SHALL 依次调用等待函数、调用同步、补业务。创建入口 SHALL 不在同步前写入本地记录，本地记录 SHALL 只由同步按云上数据新建。同步 SHALL 使用全部 `SuccessCloudIDs`。

#### Scenario: 云上已可见时交付
- **WHEN** 创建任务成功，等待期内云上返回全部新建 CLB
- **THEN** 交付完成时本地记录存在，名称、VIP、状态等字段来自云上数据，业务为请求中的业务

#### Scenario: 同步前不写本地记录
- **WHEN** 创建任务成功，等待尚未结束
- **THEN** 本地 SHALL 不存在本次新建云 ID 的记录（除非周期同步已先行拉回）

#### Scenario: 批量创建
- **WHEN** 一次创建 3 个 CLB，等待期内云上返回全部 3 个
- **THEN** 交付完成时 3 个本地记录都存在且业务正确，等待阶段每轮只发 1 次云上查询

### Requirement: 同步后补业务

同步完成后，若请求中的业务 ID 大于 0，系统 SHALL 查询本次云 ID 对应的本地记录，把业务为"未分配"（`constant.UnassignedBiz`）的记录改为请求中的业务。业务已是其它值的记录 SHALL 不覆盖，并记 Warn 日志。失败时 SHALL 不重试，直接返回错误，交付失败，错误信息 SHALL 包含本次云 ID。

#### Scenario: 同步新建的记录补业务
- **WHEN** 同步按云上数据新建了业务为"未分配"的记录，请求业务为 100
- **THEN** 该记录的业务 SHALL 被改为 100

#### Scenario: 周期同步抢先建了记录
- **WHEN** 在创建入口同步之前，周期同步已为该云 ID 建了业务为"未分配"的记录
- **THEN** 补业务 SHALL 同样把该记录的业务改为请求中的业务

#### Scenario: 记录业务已是其它值
- **WHEN** 本次云 ID 的本地记录业务已是 200，请求业务为 100
- **THEN** 系统 SHALL 不修改该记录，并记 Warn 日志

#### Scenario: 请求未指定业务
- **WHEN** 请求中的业务 ID 不大于 0
- **THEN** 系统 SHALL 跳过补业务

#### Scenario: 补业务失败
- **WHEN** 查询本地记录或更新业务的调用报错
- **THEN** 创建入口 SHALL 不重试，直接返回错误，错误信息包含本次云 ID

### Requirement: 同步按云上数据写备可用区

`res-sync/tcloud` 新建、更新本地记录时 SHALL 按云上 `BackupZoneSet` 中非空的可用区写入备可用区，保持云上顺序。云上备可用区非空且与本地不同时，同步 SHALL 判定记录有变化并更新。云上未返回备可用区时 SHALL 不修改本地已有值。创建入口 SHALL 不再按请求补备可用区。data-service 批量更新接口 SHALL 支持可选的备可用区字段，未传时 SHALL 不修改该字段。

#### Scenario: 新建记录写入备可用区
- **WHEN** 云上 CLB 的 `BackupZoneSet` 为 [B]，本地无记录
- **THEN** 同步新建的本地记录备可用区 SHALL 为 [B]

#### Scenario: 存量记录补齐备可用区
- **WHEN** 云上 CLB 的 `BackupZoneSet` 为 [B]，本地记录备可用区为空或为 [C]
- **THEN** 同步 SHALL 把本地记录备可用区更新为 [B]

#### Scenario: 云上未返回备可用区
- **WHEN** 云上 CLB 的 `BackupZoneSet` 为 null 或为空，本地记录备可用区为 [B]
- **THEN** 同步 SHALL 不修改本地记录的备可用区

#### Scenario: 不传备可用区的调用方不受影响
- **WHEN** 调用方调用批量更新接口且不传备可用区
- **THEN** 本地记录的备可用区 SHALL 保持不变

### Requirement: 等待超时时交付失败

等待函数返回仍查不到的云 ID 时，创建入口 SHALL 仍先完成同步和补业务（处理已可见的部分），然后返回错误，交付失败。错误信息 SHALL 包含仍查不到的云 ID 和创建成功的全部云 ID。公有云 SHALL 不做异步重试。

#### Scenario: 部分云 ID 一直查不到
- **WHEN** 一次创建 3 个 CLB，等待到达上限时仍有 1 个查不到
- **THEN** 已可见的 2 个 SHALL 有本地记录且业务正确，接口 SHALL 返回错误，错误信息包含查不到的云 ID

#### Scenario: 全部云 ID 一直查不到
- **WHEN** 等待到达上限时全部云 ID 都查不到
- **THEN** 接口 SHALL 返回错误，错误信息包含全部云 ID，本地不新增记录

### Requirement: 创建任务轮询结果判空

`pkg/adaptor/tcloud` 的 `CreateLoadBalancer` 在读取轮询结果前 SHALL 判空。轮询超时且结果为空时，SHALL 返回包含 TencentCloudSDK RequestId 的错误，不发生空指针异常。

#### Scenario: 轮询超时时任务仍在运行
- **WHEN** 创建任务在轮询上限内一直处于运行中，`PollUntilDone` 返回空结果且无错误
- **THEN** `CreateLoadBalancer` SHALL 返回含 RequestId 的错误，不 panic
