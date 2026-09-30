## MODIFIED Requirements

### Requirement: 记录表结构与放置

系统 SHALL 在每个被迁移的库内各建一张 `hcm_migration_record` 表，字段为 `id`（BIGINT UNSIGNED 自增主键）、`migration_id`（VARCHAR(64)，唯一索引 `uidx_migration_id`）、`version`（VARCHAR(64)，最近一次成功时注册的版本）、`applied_pkg`（VARCHAR(255)，执行该迁移 ID 的 Go 包，存导入路径去掉 `hcm/migrate/migrations/` 前缀后的部分）、`status`（VARCHAR(16)，取值 running/success/failed）、`message`（VARCHAR(1024)）、`created_at`、`updated_at`（DATETIME）。表排序规则 SHALL 为 `utf8mb4_bin`，使唯一索引按字节比较。除主键和这个唯一索引外 MUST NOT 再建索引。MUST NOT 只在主库建一张总表靠字段区分库。一条迁移在表中 SHALL 只有一行，查重 MUST 只看 `migration_id`，不看 `version`。注册的版本字符串 MUST NOT 超过 64 个字符，超长在注册时失败。

记录表尚未合入时 SHALL 直接把 `applied_pkg` 写进建表语句，MUST NOT 另写 ALTER。

#### Scenario: 每库各有一张记录表

- **WHEN** 主库与 OBS 库都参与迁移
- **THEN** 两个库内各有一张自己的 `hcm_migration_record`，互不引用

#### Scenario: OBS 库被重建后能被发现

- **WHEN** OBS 库被重建导致其记录表为空，而主库记录表完好
- **THEN** OBS 库的迁移被判定为未执行并重新执行

#### Scenario: 重复执行不新增行

- **WHEN** 一条已有记录的迁移再次被处理
- **THEN** 系统更新原有那一行，MUST NOT 插入新行

### Requirement: 状态写入时机

系统 SHALL 在执行每条迁移前写入 `running`（已有行则更新回 `running` 并清空 `message`），并把 `applied_pkg` 写成这次执行的 `Pkg`。成功后更新为 `success` 并把 `version` 刷成此刻注册的版本，MUST NOT 改 `applied_pkg`。失败后更新为 `failed` 并写入错误摘要，MUST NOT 改 `applied_pkg`。`init --mode=adopt` SHALL 为基线以内的迁移各写入一行 `success`，`applied_pkg` 为被记入基线的那条迁移的 `Pkg`。按 ID 跳过时 MUST NOT 改该行。记录写入 MUST 在迁移自身事务之外，保证迁移事务回滚后 `failed` 记录仍然留存。`message` 写入 SHALL 截断到 1024 字节且 MUST NOT 把多字节字符切断，完整错误 SHALL 进日志。

#### Scenario: 执行前落 running

- **WHEN** 一条从未执行过的迁移开始执行，其 `Pkg` 为 `main/v1.9.3/v1.9.3_20260905160000_add_bk_asset_id`
- **THEN** 记录表新增一行 `status=running`，`applied_pkg` 为该 `Pkg`

#### Scenario: 失败重跑时状态回到 running

- **WHEN** 一条 `failed` 的迁移被另一个包重新执行
- **THEN** 原行的 `status` 更新为 `running`，`message` 被清空，`applied_pkg` 更新为这次执行的 `Pkg`

#### Scenario: 成功刷新版本快照

- **WHEN** 一条迁移执行成功，而它的注册版本自上次记录以来已变化
- **THEN** 该行 `status=success`，`version` 更新为此刻注册的版本，`applied_pkg` 保持 `MarkRunning` 写入的值

#### Scenario: 失败记录不被迁移事务回滚

- **WHEN** 某迁移在自己的事务中失败并回滚
- **THEN** 记录表中该行仍为 `failed` 并带错误摘要

#### Scenario: 超长错误安全截断

- **WHEN** 迁移返回的错误信息超过 1024 字节且含中文
- **THEN** `message` 被截断到 1024 字节以内且无残缺字符，完整错误出现在日志中

#### Scenario: 跳过不改 applied_pkg

- **WHEN** 一条已成功的迁移因迁移后缀相同被跳过
- **THEN** 该行的 `applied_pkg` 与 `version` 保持跳过前的值

### Requirement: 库当前版本的计算

库当前版本 SHALL 定义为本库全部 `success` 记录中、版本不是 `PENDING` 的那些里版本号最大的那一条。MUST 在 Go 中逐条解析后用版本比较器求最大值，MUST NOT 使用 SQL 的 `ORDER BY version` 取最大，也 MUST NOT 取最后写入的那一条。`Load` SHALL 放行 `PENDING`，与 `--allow-pending` 无关。读取记录时，`running`、`success`、`failed` 任一状态的 `version` 无法解析且不是 `PENDING`，或 `status` 未知，或 `success` 行的 `applied_pkg` 为空，SHALL 以退出码 3 失败，MUST NOT 跳过该行。

#### Scenario: 不按字典序取最大

- **WHEN** 成功记录含 `v1.9.9` 与 `v1.9.10`
- **THEN** 库当前版本为 `v1.9.10`

#### Scenario: 补跑写入的低版本不降低库当前

- **WHEN** 库当前为 `v1.9.4`，补跑成功写入一条 `v1.9.3` 的 success 记录
- **THEN** 库当前版本仍为 `v1.9.4`

#### Scenario: 空库当前版本为空

- **WHEN** 记录表中没有任何已定版的 `success` 记录
- **THEN** 库当前版本为空，所有已定版迁移都被视为高于库当前

#### Scenario: PENDING 不抬高库当前版本

- **WHEN** 成功记录含 `v1.9.2` 与 `PENDING`
- **THEN** 库当前版本为 `v1.9.2`

#### Scenario: 无法解析的版本导致失败

- **WHEN** 读取记录时，记录表中某条 `running`、`success` 或 `failed` 记录的 `version` 无法解析且不是 `PENDING`
- **THEN** 命令以退出码 3 失败，MUST NOT 跳过该行继续计算库当前版本

#### Scenario: 成功行缺少 applied_pkg

- **WHEN** 读取记录时，某条 `success` 记录的 `applied_pkg` 为空
- **THEN** 命令以退出码 3 失败

## ADDED Requirements

### Requirement: PENDING 定版回填

已有 `success` 且记录的 `version` 为 `PENDING` 时，系统 SHALL 按 ID 跳过该迁移，并把该行的 `version` 更新为此刻注册的版本。更新语句 MUST 带条件 `status = success` 且 `version = PENDING`，影响 0 行视为正常。回填 MUST NOT 修改 `applied_pkg`。回填本身 MUST NOT 当成版本漂移。同 ID 有两份时，执行顺序靠前的那份完成回填后，后一份看到的已不是 `PENDING`，走版本漂移警告。

#### Scenario: 定版后跳过并回填

- **WHEN** 记录为 `success` 且 `version` 为 `PENDING`，此刻注册版本为 `v1.9.3`，迁移后缀与 `applied_pkg` 相同
- **THEN** 不调用 `up()`，该行 `version` 变为 `v1.9.3`，`applied_pkg` 不变

#### Scenario: 重复回填无副作用

- **WHEN** 记录的 `version` 已经是 `v1.9.3`，再次执行会走到回填语句
- **THEN** 该行不被改写，命令不因此失败
