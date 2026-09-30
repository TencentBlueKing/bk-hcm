## MODIFIED Requirements

### Requirement: 命令与用法

CLI SHALL 提供四个子命令：`init` / `up` / `status` / `list`。`help`、`-h`、`--help` SHALL 打印用法并以退出码 0 结束。未给子命令 SHALL 打印用法并以退出码 2 结束。未知子命令 SHALL 以退出码 2 结束。任一子命令出现多余位置参数 SHALL 以退出码 2 结束。

#### Scenario: 无子命令

- **WHEN** 执行 `hcm-migrate` 且无子命令
- **THEN** 打印用法，退出码为 2

#### Scenario: 未知子命令

- **WHEN** 执行 `hcm-migrate foo`
- **THEN** 退出码为 2

#### Scenario: help 成功退出

- **WHEN** 执行 `hcm-migrate --help` 或 `hcm-migrate help`
- **THEN** 打印用法，退出码为 0

### Requirement: 全局参数

所有子命令 SHALL 支持 `--config-file` / `-c`（配置文件路径）、`--database` / `-d` 与 `--allow-pending`。`--database` 是库名列表：可重复，也可用逗号分隔。不传 SHALL 表示配置里已启用的全部库。传入的每个名字 MUST 对应一个已有的 Registry，未知名字以退出码 2 失败。MUST NOT 使用 `all` 作为魔法取值。`--allow-pending` 默认 MUST 为关；仅 `up` 实际使用它。生产、出包与外部版的 Helm values MUST NOT 打开它。

`init` / `up` / `status` MUST 要求 `--config-file`；缺失时以退出码 2 失败。`list` MUST NOT 要求配置文件，也 MUST NOT 连接数据库。

#### Scenario: 默认作用于全部已启用的库

- **WHEN** 执行 `up` 且未指定 `--database`，配置里启用了主库与 OBS 库
- **THEN** 主库与 OBS 库依次处理

#### Scenario: 只执行其中两个库

- **WHEN** 配置里启用了三个库，执行 `up --database main,obs`
- **THEN** 只处理主库与 OBS 库，第三个库不被连接

#### Scenario: 重复传参与逗号等价

- **WHEN** 执行 `up -d main -d obs`
- **THEN** 处理的库与 `up --database main,obs` 相同

#### Scenario: 非法库名报错

- **WHEN** 执行 `up --database foo`
- **THEN** 命令以退出码 2 失败

#### Scenario: 默认拒绝未定版

- **WHEN** 执行 `up` 且未传 `--allow-pending`，注册表中有 `PENDING`
- **THEN** 命令以退出码 5 失败

#### Scenario: 缺配置文件

- **WHEN** 执行 `up` 且未传 `--config-file`
- **THEN** 命令以退出码 2 失败

#### Scenario: list 无需配置

- **WHEN** 执行 `list` 且未传 `--config-file`
- **THEN** 按执行顺序打印所选注册表，退出码为 0

### Requirement: 输出分流

stdout SHALL 承载模式行、计划、汇总、`status` 与 `list` 的表格。过程日志与错误详情 SHALL 经 `pkg/logs` 打到 stderr，使 Job 日志在 pod 销毁后仍可查。同一错误 MUST NOT 既打 stdout 又打一份等价详情到日志以外的通道；失败时 CLI 经 `pkg/logs` 记一条含退出码的错误即可。

#### Scenario: 计划在 stdout

- **WHEN** 执行 `up --plan` 且判定通过
- **THEN** 模式行与计划表出现在 stdout，过程日志在 stderr

### Requirement: up 参数

`up` SHALL 支持 `--to`（版本上限，默认空表示不设上限）、`--catch-up`（补跑模式，默认 `false`）、`--plan`（默认 `false`）。`up` SHALL 在开始执行前先向 stdout 打印本次命令收到的模式，包括是否补跑、是否允许 `PENDING`、版本上限、是否仅计划，便于从日志判断是否误开。`--to` 为 `PENDING` 或非法版本时 MUST 以退出码 2 失败。

#### Scenario: 默认不设上限且为默认模式

- **WHEN** 执行 `up`
- **THEN** 处理全部已定版注册迁移，按默认模式判定，且不执行 `PENDING`

#### Scenario: 打印本次模式

- **WHEN** 执行 `up --catch-up --allow-pending`
- **THEN** 输出显式标明本次为补跑模式且允许 `PENDING`

#### Scenario: --to 非法

- **WHEN** 执行 `up --to PENDING` 或 `up --to not-a-version`
- **THEN** 命令以退出码 2 失败

### Requirement: up 执行顺序

`up` SHALL 按主库再 OBS 的顺序连接每个被选中的库；无配置的库 SHALL 跳过（仅 DataSource.Open 打警告日志，MUST NOT 向 stdout 打印跳过说明）。非 `--plan` 时，审计行 MUST 在连上库之后、`Prepare` 之前 `Begin`。`Prepare` 失败时 MUST 停止后续库的连接与准备，已成功准备的库标记为未执行。全部准备完成后 SHALL 打印每个库的计划，再 `CollectPlanErrors`；任一计划有问题 MUST NOT 执行任何库。执行阶段 SHALL 按序执行并在第一个失败的库处停止，其后库未执行。每个已 `Begin` 的审计行在所有路径上 MUST 以进程退出码 `End`。

#### Scenario: 先全部准备再执行

- **WHEN** 主库计划通过、OBS 计划检出漏执行
- **THEN** 两个库都不执行任何迁移，退出码为 4

#### Scenario: 执行中途停在失败库

- **WHEN** 主库与 OBS 计划均通过，主库执行失败
- **THEN** OBS 不被执行，主库审计 `exit_code` 为 1

#### Scenario: 已完成库仍以进程退出码结束审计

- **WHEN** 主库执行成功、OBS 执行失败
- **THEN** 主库与 OBS 审计行 `exit_code` 均为进程退出码

#### Scenario: plan 不写审计

- **WHEN** 执行 `up --plan`
- **THEN** 不插入审计行，仍打印模式行与计划

### Requirement: init 命令

`init` SHALL 调用 `InitTables` 建立记录表与运行审计表，并在需要时写入库版本起点。`InitTables` SHALL 先创建缺失的审计表（失败则返回），再创建缺失的记录表；仅当本次调用新建了记录表且 `--mode=adopt` 时，把不超过该库基线的已定版迁移逐条写入 `success` 而不执行它们。基线插入失败时 MUST 只删除本次新建的记录表。`--mode` MUST 必填且无默认值，取值 `empty`（只建表，不写基线）或 `adopt`。`PENDING` MUST NOT 写入基线。两张表都已存在时，该库对 `init` 无改动；只缺其中一张时，只创建缺失的那张（仅新建记录表时才可能写基线）。因此 `init` SHALL 可以每次部署都执行。

`--mode=empty` 却带了任一 `--baseline` MUST 以退出码 2 失败。`--mode=adopt` 却未提供任何 `--baseline` MUST 以退出码 2 失败，MUST NOT 退化为 `empty`。`--baseline` 格式 MUST 为 `库名=版本`，可重复；同一库多次给出时后写覆盖先写；格式错误、未知库名、版本为 `PENDING` 或非法 MUST 以退出码 2 失败。已知但未选中的库上的 `--baseline` SHALL 被忽略。被选中却无对应基线条目的库，在 `adopt` 下 SHALL 只建空表。无配置的库 SHALL 跳过（仅 DataSource.Open 打警告日志，MUST NOT 向 stdout 打印跳过说明）。某库 init 失败时 MUST 停止，其后库不被处理。

每个库的 init 审计行在该库 `InitTables` 返回后 `Begin` 并立即以该库自身结果 `End`：审计表在首次 init 前可能尚不存在；`exit_code` 为该库 `InitTables` 错误对应的退出码（成功为 0）。adopt 写入的基线迁移 SHALL 记入审计 `skipped`，`reason` 为 `baseline`；init 审计的 `version_after` MUST 保持为空。stdout SHALL 打印该库结果行（新建时含是否创建审计表/记录表与基线条数，已初始化时为无改动说明），MUST NOT 另打成功日志。

`--plan` SHALL 只打印将建的表；仅当记录表缺失且该库有基线时，再打印将标记为 success 的迁移清单。MUST NOT 写库，MUST NOT 写运行审计。

#### Scenario: 空库建空表

- **WHEN** 对两张表都不存在的空库执行 `init --mode=empty`
- **THEN** 建立 `hcm_migration_record` 与 `hcm_migration_audit`，且不写入任何迁移记录

#### Scenario: 存量库垫起库版本

- **WHEN** 对两张表都不存在的存量库执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 建表后为版本不超过 `v1.9.2` 的已定版迁移各写入一行 `success`，且它们的 `up()` MUST NOT 被调用

#### Scenario: 两张表都已存在则无改动

- **WHEN** 记录表与审计表都已存在，再次执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 该库不被修改，既不建表也不写入任何记录

#### Scenario: 只缺其中一张表

- **WHEN** 记录表已存在、审计表不存在，执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 只创建 `hcm_migration_audit`，MUST NOT 写入基线记录

#### Scenario: 缺少 mode 即失败

- **WHEN** 执行 `init`
- **THEN** 命令以退出码 2 失败并提示 `--mode` 必填

#### Scenario: adopt 缺基线即失败

- **WHEN** 执行 `init --mode=adopt` 且未提供任何 `--baseline`
- **THEN** 命令以退出码 2 失败，MUST NOT 退化为 `empty`

#### Scenario: empty 带基线即失败

- **WHEN** 执行 `init --mode=empty --baseline main=v1.9.2`
- **THEN** 命令以退出码 2 失败

#### Scenario: 基线非法

- **WHEN** 执行 `init --mode=adopt --baseline main=PENDING`，或基线格式不是 `库名=版本`
- **THEN** 命令以退出码 2 失败

#### Scenario: 未选中库的基线被忽略

- **WHEN** 执行 `init --mode=adopt -d main --baseline main=v1.9.2 --baseline obs=v1.9.2`，OBS 未被选中
- **THEN** 主库按基线处理，OBS 不被连接，命令成功

#### Scenario: 混合场景按库区分

- **WHEN** 主库为存量库、OBS 为新库，执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 主库建表并垫到 `v1.9.2`，OBS 库只建空表

#### Scenario: init 支持预演

- **WHEN** 执行 `init --mode=adopt --baseline main=v1.9.2 --plan`
- **THEN** 打印将建的表与将标记的 ID、版本清单，数据库不被修改，且不写审计

#### Scenario: init 审计在建表之后

- **WHEN** 对两张表都不存在的空库执行 `init --mode=empty`
- **THEN** 先完成 `InitTables`，再插入该库审计行并以该库自身结果立即 `End`，`exit_code` 为 0，`version_after` 为空

### Requirement: status 命令

`status` SHALL 只读：不建表、不写记录、不写审计。它 SHALL 用与 `up` 默认模式相同的计划，并打开 `PENDING` 允许，仅用于标出版本不超过库当前且未成功的迁移。每个库 SHALL 打印 `database X, current version: V`、`records: success N, failed N, running N`，以及 `not executed:` 下列出未成功执行的迁移（列 `ID`、`VERSION`、`PKG`）；对版本不超过库当前的项追加 `<= current version`。MUST NOT 打印 `RECORD` 列、`pending` 附注、`not executed: none`，也 MUST NOT 打印 issues / warnings 段。无配置的库 SHALL 跳过（仅警告日志）。某库未初始化或记录非法时 SHALL 打印该库错误并继续展示其余库，命令最终以退出码 3 结束。

#### Scenario: 未初始化库不影响其余展示

- **WHEN** 主库未初始化、OBS 已初始化，执行 `status`
- **THEN** 主库打印缺表提示，OBS 正常打印进度，退出码为 3

### Requirement: list 命令

`list` SHALL 按执行顺序打印每个被选中的注册表（列：`#`、ID、version、timestamp、pkg），`PENDING` 照常列出，退出码为 0。

#### Scenario: 列出 PENDING

- **WHEN** 注册表含 `PENDING`，执行 `list`
- **THEN** 版本列打印 `PENDING`，退出码为 0

### Requirement: 退出码

命令 SHALL 使用固定退出码：

| 码 | 含义 |
| --- | --- |
| 0 | 成功 |
| 1 | 执行期失败：迁移 `Up` 返回错误或 panic（panic 转成错误并记 `failed`）、执行中记录表读写失败（running/success/failed/回填）、`MarkSuccess` 影响 0 行、连接断开或 context 取消 |
| 2 | 参数或配置错误 |
| 3 | 记录表问题：`hcm_migration_record` 或 `hcm_migration_audit` 任一不存在（提示先 `init`），或记录内容非法（未知 `status`、版本不可解析、`success` 行 `applied_pkg` 为空） |
| 4 | 漏执行：默认模式下存在低于等于库当前版本且未成功的迁移；加 `--catch-up` 重跑可修复 |
| 5 | 注册表内容问题：未开 `--allow-pending` 却注册了 `PENDING`（不论是否指定 `--to`），或注册表出现两条及以上版本线。空 label 的数字第四段也算一条线，每个非空 label 各算一条线；没有第四段的版本不单独成线 |
| 6 | 疑似 ID 复用 |

退出码 `4` MUST 与 `1` 区分，使出包工具能判断本包是否需要开补跑。「已初始化」SHALL 定义为两张表都存在；`init` 创建两张表；`up`（含 `--plan`）执行前用同一套 `CheckInitialized`，任一缺失即退出码 3。审计写入失败 SHALL 只 WARN，MUST NOT 影响退出码。

多个库同时存在计划阶段问题时，优先级 MUST 为 6 > 5 > 4（只返回最高优先级对应的退出码，但错误信息列出所有库的全部问题）。退出码 3 在 Load / Prepare 阶段返回，在生成计划之前，MUST NOT 与 4 / 5 / 6 同时出现。`errors.Is` 到退出码的映射由 CLI 完成：`ErrPrecondition`→3、`ErrMissed`→4、`ErrRegistry`→5、`ErrIDReuse`→6、`ErrUsage`→2；其余非 nil 为 1。

#### Scenario: 漏执行返回专用退出码

- **WHEN** 默认模式下检出漏执行
- **THEN** 退出码为 4

#### Scenario: 疑似 ID 复用返回专用退出码

- **WHEN** 检出迁移后缀不同的同 ID
- **THEN** 退出码为 6

#### Scenario: 占位符在开关关闭时返回注册表退出码

- **WHEN** 未传 `--allow-pending`，注册表中存在常量 `PENDING`
- **THEN** 退出码为 5

#### Scenario: 两个非空 label 返回注册表退出码

- **WHEN** 注册表出现两条及以上版本线，包括数字第四段与一个具名 label
- **THEN** 退出码为 5

#### Scenario: 缺表返回记录表问题退出码

- **WHEN** `hcm_migration_record` 或 `hcm_migration_audit` 任一不存在
- **THEN** 退出码为 3，提示先执行 `init`

#### Scenario: 迁移执行失败返回 1

- **WHEN** 某条迁移的 `up()` 返回错误
- **THEN** 退出码为 1
