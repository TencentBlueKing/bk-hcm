## ADDED Requirements

### Requirement: 运行记录的粒度

系统 SHALL 在每个被迁移的库内各建一张 `hcm_migration_audit` 表。一次进程运行 SHALL 在每个被选中的库各写一行，这些行 MUST 共用同一个 `run_id`。`run_id` MUST 取自 `hcm/pkg/tools/uuid.UUID()`，并且 MUST 等于这次进程 `kit` 的 `Rid`。MUST NOT 按发布或按部署合并成一行。`list`、`status`、`up --plan` MUST NOT 写这张表。

#### Scenario: 一次运行在每个库一行

- **WHEN** 一次 `up` 依次处理主库与 aux 库
- **THEN** 两个库各有一行 `hcm_migration_audit`，且两行的 `run_id` 相同

#### Scenario: plan 不写运行记录

- **WHEN** 执行 `up --plan`
- **THEN** `hcm_migration_audit` 不被插入或更新

### Requirement: 运行记录表结构

`hcm_migration_audit` SHALL 含下列列：`id`（BIGINT UNSIGNED 自增主键）、`run_id`（VARCHAR(64)，唯一索引 `uidx_run_id`）、`command`（VARCHAR(16)，取值 `up` 或 `init`）、`args`（VARCHAR(1024)）、`binary_version`（VARCHAR(64)）、`git_hash`（VARCHAR(64)）、`status`（VARCHAR(16)，取值 `running`、`success`、`failed`）、`exit_code`（INT，`running` 时为空）、`version_before`（VARCHAR(64)）、`version_after`（VARCHAR(64)）、`skipped`（JSON）、`warnings`（JSON）、`message`（JSON）、`start_at`（DATETIME）、`end_at`（DATETIME，未结束时为空）。表排序规则 SHALL 为 `utf8mb4_bin`。SHALL 另有索引 `idx_start_at`。MUST NOT 包含执行清单、回填清单或运行者列。

`binary_version` 与 `git_hash` SHALL 分别取 `hcm/pkg/version` 的 `VERSION` 与 `GITHASH`，MUST NOT 从配置文件读取，写入时 MUST 截断到 64 字节。`args` SHALL 是 `os.Args[1:]` 以空格拼接后的字符串，超长 MUST 截断到 1024 字节。

#### Scenario: 未注入版本时写入默认值

- **WHEN** 二进制未经 ldflags 注入版本，执行 `up`
- **THEN** 该行 `binary_version` 为 `debug`，`git_hash` 为 `unknown`

#### Scenario: 参数列不含连接密码

- **WHEN** 数据库密码只写在配置文件里，命令行只传配置路径
- **THEN** `args` 含配置路径，不含密码

### Requirement: 写入时机与失败隔离

对 `up`，审计器 SHALL 在连上库之后、`Prepare` 之前插入一行 `status=running`（所有路径含检查失败都须在结束时更新）。对 `init`，审计行 MUST 在该库 `InitTables` 返回之后再插入：首次 init 时审计表在建表前可能尚不存在。运行结束后 SHALL 更新这一行，写入 `end_at`、`status`、`exit_code`、`version_before`、`version_after`、`skipped`、`warnings`、`message`。检查问题 MUST 全部写入 `message`。

对 `up`，每一行的 `exit_code` MUST 是本次进程的退出码（不是该库自身是否成功的局部码）；一次进程内所有已 `Begin` 的 up 审计行 MUST 以同一进程退出码一并 `End`。对 `init`，每个库的审计行在该库 `InitTables` 返回后 `Begin` 并立即以该库自身结果 `End`，`exit_code` 为该库结果对应的退出码（成功为 0）。`status` 为 `success` 当且仅当 `exit_code` 为 0，否则为 `failed`。

JSON 列 MUST 在结束时一次写入；`running` 期间为空；结束时空列表 MUST 写成 `[]` 而不是 `null`。执行层 MUST NOT import 审计器，也 MUST NOT 在迁移循环里调用它。审计器由 CLI 调用。

插入或更新失败时，审计器 SHALL 只打警告，MUST NOT 改变命令的退出码。`kit` 的 `Rid` 为空时，SHALL 只打警告并跳过整次审计，MUST NOT 插入。插入失败后，结束时的更新 MUST NOT 再写。对 `nil` 审计句柄调用结束更新 MUST 什么都不做。

`version_before` SHALL 是执行开始时算出的库当前版本。`version_after` SHALL 是执行结果上的变量：开始时等于 `version_before`；成功执行或回填得到更高的已定版版本时更新；`PENDING` MUST NOT 抬高。`init` 的审计行 `version_after` MUST 保持为空。命令出错时 MUST 仍写入当时的 `version_after`。

#### Scenario: 校验失败仍留下一行

- **WHEN** `Load` 因未知 status 以退出码 3 失败
- **THEN** 该库仍有一行 `hcm_migration_audit`，`status` 为 `failed`，`exit_code` 为 3

#### Scenario: init 在建表后再写审计

- **WHEN** 对两张表都不存在的空库执行 `init --mode=empty`
- **THEN** 先完成 `InitTables`，再插入 `status=running` 的审计行，并以该库自身结果立即更新（成功时 `exit_code` 为 0）

#### Scenario: 各库 exit_code 均为进程退出码

- **WHEN** 主库迁移成功、aux 迁移失败，进程退出码为 1
- **THEN** 两个库的审计行 `exit_code` 均为 1

#### Scenario: 审计失败不改变退出码

- **WHEN** 开头的插入失败，随后迁移本身成功
- **THEN** 命令退出码为 0，且日志中有审计失败的警告

#### Scenario: 空 rid 跳过审计

- **WHEN** 进程 `kit` 的 `Rid` 为空，执行会写审计的命令
- **THEN** 不插入审计行，日志有警告，命令退出码仍按迁移结果

#### Scenario: 执行中途失败仍写入结束后的库版本

- **WHEN** 本次已有一条更高版本的迁移成功，下一条迁移的 `up()` 返回错误
- **THEN** 该行 `status` 为 `failed`，`version_after` 为那条已成功迁移的版本

### Requirement: skipped、warnings 与 message

`skipped` SHALL 是对象数组。每一项 MUST 含 `id`、`version`、`pkg`、`applied_pkg`、`reason`。`reason` 取值 SHALL 为 `applied`（已有 success，按 ID 跳过）、`above_max_version`（超过版本上限）、`baseline`（`init` adopt 直接记为 success）。`applied` 的 `applied_pkg` MUST 是记录表里的值；另外两个取值的 `applied_pkg` MUST 为空串。`skipped` MUST NOT 设条数上限。

`warnings` SHALL 是字符串数组，只追加运行中的警告。`message` SHALL 是字符串数组，按发生顺序追加校验问题和错误，两项 MUST NOT 互相覆盖。这两列每项 SHALL 截断到 1024 字节且不切断多字节字符，每列最多 200 项；超出部分 MUST 换成最后一项，内容为 `... N more omitted`。

本次执行了哪些迁移 MUST NOT 写入本表。排查时 SHALL 用 `hcm_migration_record` 中 `updated_at` 落在本行 `start_at` 与 `end_at` 之间的行来对。

#### Scenario: 按 ID 跳过被逐条记下

- **WHEN** 一条已成功的迁移因迁移后缀相同被跳过
- **THEN** `skipped` 中有一项，`reason` 为 `applied`，`id` 为该迁移 ID，`applied_pkg` 为记录表中的值

#### Scenario: 校验问题与错误都留在 message

- **WHEN** 同一次运行既检出漏执行，又在别的库上出现执行错误
- **THEN** 该库 `message` 数组里同时有漏执行条目和错误条目

#### Scenario: 空列表写成空数组

- **WHEN** 某次运行无跳过、无警告、无 message，正常结束
- **THEN** 该行 `skipped`、`warnings`、`message` 均为 `[]`

### Requirement: 建表与残留的 running 行

`InitTables` SHALL 创建缺失的 `hcm_migration_audit` 与 `hcm_migration_record`：先审计表、后记录表；建表经 `util.CreateTableIfNotExists`，下发的语句 MUST NOT 自带 `IF NOT EXISTS`。库在两张表都存在时视为已初始化；已初始化的库上再次调用 MUST 无改动。只缺一张时 MUST 只创建缺失的那张。基线 adopt 仅在本次新建了记录表时发生。`up` MUST NOT 创建这两张表。审计表不存在或插入失败时，命令 MUST 继续按迁移结果退出。

进程被杀导致一行停在 `running` 且 `end_at` 为空时，当次的 `skipped`、`warnings`、`message` 允许缺失。下一次运行开始时，审计器 SHALL 按 `id` 降序取最新一条仍为 `running` 的行，并把该 `run_id` 追加进本次的 `warnings`。本表 MUST NOT 被当作锁。本期 MUST NOT 清理历史行。

#### Scenario: init 同时建出两张表

- **WHEN** 目标库两张表都不存在，执行 `init --mode=empty`
- **THEN** `hcm_migration_record` 与 `hcm_migration_audit` 都被创建，且运行记录表没有业务迁移行

#### Scenario: 只缺审计表时只建审计表

- **WHEN** 记录表已存在、审计表不存在，执行 `init --mode=empty`
- **THEN** 只创建 `hcm_migration_audit`，记录表不被改动

#### Scenario: 上次 running 被记进这次的警告

- **WHEN** 表中已有一行 `status=running` 且 `end_at` 为空，再次执行 `up` 并正常结束
- **THEN** 新行的 `warnings` 含那次的 `run_id`
