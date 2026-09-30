## ADDED Requirements

### Requirement: 记录表结构与放置

系统 SHALL 在每个被迁移的库内各建一张 `hcm_migration_record` 表，字段为 `id`（BIGINT UNSIGNED 自增主键）、`migration_id`（VARCHAR(64)，唯一索引 `uidx_migration_id`）、`version`（VARCHAR(64)，最近一次成功时注册的版本）、`status`（VARCHAR(16)，取值 running/success/failed）、`message`（VARCHAR(1024)）、`created_at`、`updated_at`（DATETIME）。表排序规则 SHALL 为 `utf8mb4_bin`，使唯一索引按字节比较。除主键和这个唯一索引外 MUST NOT 再建索引。MUST NOT 只在主库建一张总表靠字段区分库。一条迁移在表中 SHALL 只有一行，查重 MUST 只看 `migration_id`，不看 `version`。注册的版本字符串 MUST NOT 超过 64 个字符，超长在注册时失败。

#### Scenario: 每库各有一张记录表

- **WHEN** 多个库都参与迁移
- **THEN** 两个库内各有一张自己的 `hcm_migration_record`，互不引用

#### Scenario: 其他库被重建后能被发现

- **WHEN** 某个库被重建导致其记录表为空，而主库记录表完好
- **THEN** 该库的迁移被判定为未执行并重新执行

#### Scenario: 重复执行不新增行

- **WHEN** 一条已有记录的迁移再次被处理
- **THEN** 系统更新原有那一行，MUST NOT 插入新行

### Requirement: 记录表的建立

记录表 SHALL NOT 写成一条迁移。`init` SHALL 调用 `util.CreateTableIfNotExists`，传入不含 `IF NOT EXISTS` 的 `CREATE TABLE` 语句。该函数 MUST 返回本次是否新建了表。存在性判断由该工具完成，`engine` MUST NOT 自己拼 `IF NOT EXISTS`，也 MUST NOT 直接查询 `information_schema`。仅当本次新建且 `--mode=adopt` 时，`init` SHALL 插入基线记录；返回未新建时该库整条命令 no-op。`up`、`status`、`list` MUST NOT 建表。

#### Scenario: init 经幂等工具建表

- **WHEN** 目标库不存在 `hcm_migration_record`，执行 `init --mode=empty`
- **THEN** 通过 `util.CreateTableIfNotExists` 建立记录表，下发的语句是 `CREATE TABLE` 而不是 `CREATE TABLE IF NOT EXISTS`

#### Scenario: 重复执行建表语句无副作用

- **WHEN** 记录表已存在，再次执行 `init`
- **THEN** 表结构与数据不变

#### Scenario: up 不建表

- **WHEN** 目标库不存在记录表，执行 `up`
- **THEN** 命令以退出码 3 失败，MUST NOT 创建表

#### Scenario: 只读命令不建表

- **WHEN** 目标库不存在记录表，执行 `status`
- **THEN** 输出提示该库尚未初始化，MUST NOT 创建表

### Requirement: 状态写入时机

系统 SHALL 在执行每条迁移前写入 `running`（已有行则更新回 `running` 并清空 `message`），成功后更新为 `success` 并把 `version` 刷成此刻注册的版本，失败后更新为 `failed` 并写入错误摘要。`init --mode=adopt` SHALL 为基线以内的迁移各写入一行 `success`。记录写入 MUST 在迁移自身事务之外，保证迁移事务回滚后 `failed` 记录仍然留存。`message` 写入 SHALL 截断到 1024 字节且 MUST NOT 把多字节字符切断，完整错误 SHALL 进日志。

#### Scenario: 执行前落 running

- **WHEN** 一条从未执行过的迁移开始执行
- **THEN** 记录表新增一行 `status=running`

#### Scenario: 失败重跑时状态回到 running

- **WHEN** 一条 `failed` 的迁移被重新执行
- **THEN** 原行的 `status` 更新为 `running` 且 `message` 被清空

#### Scenario: 成功刷新版本快照

- **WHEN** 一条迁移执行成功，而它的注册版本自上次记录以来已变化
- **THEN** 该行 `status=success`，`version` 更新为此刻注册的版本

#### Scenario: 失败记录不被迁移事务回滚

- **WHEN** 某迁移在自己的事务中失败并回滚
- **THEN** 记录表中该行仍为 `failed` 并带错误摘要

#### Scenario: 超长错误安全截断

- **WHEN** 迁移返回的错误信息超过 1024 字节且含中文
- **THEN** `message` 被截断到 1024 字节以内且无残缺字符，完整错误出现在日志中

### Requirement: 库当前版本的计算

库当前版本 SHALL 定义为本库全部 `success` 记录中版本号最大的那一条，MUST 在 Go 中逐条解析后用版本比较器求最大值，MUST NOT 使用 SQL 的 `ORDER BY version` 取最大，也 MUST NOT 取最后写入的那一条。读取记录时，`running`、`success`、`failed` 任一状态的 `version` 无法解析，SHALL 与未知 `status` 一样以退出码 3 失败，MUST NOT 跳过该行。

#### Scenario: 不按字典序取最大

- **WHEN** 成功记录含 `v1.9.9` 与 `v1.9.10`
- **THEN** 库当前版本为 `v1.9.10`

#### Scenario: 补跑写入的低版本不降低库当前

- **WHEN** 库当前为 `v1.9.4`，补跑成功写入一条 `v1.9.3` 的 success 记录
- **THEN** 库当前版本仍为 `v1.9.4`

#### Scenario: 空库当前版本为空

- **WHEN** 记录表中没有任何 `success` 记录
- **THEN** 库当前版本为空，所有迁移都被视为高于库当前

#### Scenario: 无法解析的版本导致失败

- **WHEN** 读取记录时，记录表中某条 `running`、`success` 或 `failed` 记录的 `version` 无法解析
- **THEN** 命令以退出码 3 失败，MUST NOT 跳过该行继续计算库当前版本

#### Scenario: 失败记录的版本无法解析

- **WHEN** 读取记录时，某条 `failed` 记录的 `version` 无法解析
- **THEN** 命令以退出码 3 失败，MUST NOT 跳过该行

### Requirement: 版本漂移告警

按 Migration ID 读到 `success` 而记录中的 `version` 与此刻注册的版本不一致时，系统 SHALL 打印 WARN 说明坐标被改过，并照常跳过该迁移。此情形在归档后的特性分支环境上是预期行为，MUST NOT 判为失败。

#### Scenario: 归档后的漂移只告警

- **WHEN** 某迁移记录的版本为 `v1.9.3-tenant.1`，归档后注册版本变为 `v1.9.3`，且记录状态为 `success`
- **THEN** 打印 WARN 并跳过该迁移，命令继续执行且不因此失败
