## MODIFIED Requirements

### Requirement: 全局参数

所有子命令 SHALL 支持 `--config-file` / `-c`（配置文件路径）、`--database` / `-d` 与 `--allow-pending`。`--database` 是库名列表：可重复，也可用逗号分隔。不传 SHALL 表示配置里已启用的全部库。传入的每个名字 MUST 对应一个已有的 Registry，未知名字以退出码 2 失败。MUST NOT 使用 `all` 作为魔法取值。`--allow-pending` 默认 MUST 为关。生产、出包与外部版的 Helm values MUST NOT 打开它。

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
- **THEN** 命令以退出码 3 失败

### Requirement: up 参数

`up` SHALL 支持 `--to`（版本上限，默认空表示不设上限）、`--catch-up`（补跑模式，默认 `false`）、`--plan`（默认 `false`）。`up` SHALL 在开始执行前打印本次命令收到的模式，包括是否补跑、是否允许 `PENDING`、版本上限，便于从日志判断是否误开。

#### Scenario: 默认不设上限且为默认模式

- **WHEN** 执行 `up`
- **THEN** 处理全部已定版注册迁移，按默认模式判定，且不执行 `PENDING`

#### Scenario: 打印本次模式

- **WHEN** 执行 `up --catch-up --allow-pending`
- **THEN** 输出显式标明本次为补跑模式且允许 `PENDING`

### Requirement: init 命令

`init` SHALL 调用 `InitTables` 建立记录表与运行审计表，并在需要时写入库版本起点。`InitTables` SHALL 先创建缺失的审计表（失败则返回），再创建缺失的记录表；仅当本次调用新建了记录表且 `--mode=adopt` 时，把不超过该库基线的已定版迁移逐条写入 `success` 而不执行它们。基线插入失败时 MUST 只删除本次新建的记录表。`--mode` MUST 必填且无默认值，取值 `empty`（只建表，不写基线）或 `adopt`。`PENDING` MUST NOT 写入基线。两张表都已存在时，该库对 `init` 无改动；只缺其中一张时，只创建缺失的那张（仅新建记录表时才可能写基线）。因此 `init` SHALL 可以每次部署都执行。`--plan` SHALL 只打印将建的表与将标记的 ID、版本，MUST NOT 写运行审计。

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

#### Scenario: existing 缺基线即失败

- **WHEN** 执行 `init --mode=adopt` 且未提供任何 `--baseline`
- **THEN** 命令以退出码 2 失败，MUST NOT 退化为 `empty`

#### Scenario: 混合场景按库区分

- **WHEN** 主库为存量库、OBS 为新库，执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 主库建表并垫到 `v1.9.2`，OBS 库只建空表

#### Scenario: init 支持预演

- **WHEN** 执行 `init --mode=adopt --baseline main=v1.9.2 --plan`
- **THEN** 打印将建的表与将标记的 ID、版本清单，数据库不被修改

### Requirement: 退出码

命令 SHALL 使用固定退出码：`0` 成功；`1` 执行失败；`2` 参数或配置错误；`3` 前置校验失败（含开关关闭时的 `PENDING`、记录表中无法解析的版本、未知 `status`、`success` 行的 `applied_pkg` 为空、记录表不存在）；`4` 默认模式检出漏执行，或检出疑似 ID 复用。退出码 `4` MUST 与 `1` 区分，使出包工具能判断本包是否需要开补跑。同一次校验里既有退出码 3 的项又有退出码 4 的项时，MUST 使用 4。

#### Scenario: 漏执行返回专用退出码

- **WHEN** 默认模式下检出漏执行
- **THEN** 退出码为 4

#### Scenario: 疑似 ID 复用返回专用退出码

- **WHEN** 检出迁移后缀不同的同 ID
- **THEN** 退出码为 4

#### Scenario: 占位符在开关关闭时返回前置校验退出码

- **WHEN** 未传 `--allow-pending`，注册表中存在常量 `PENDING`
- **THEN** 退出码为 3

#### Scenario: 迁移执行失败返回 1

- **WHEN** 某条迁移的 `up()` 返回错误
- **THEN** 退出码为 1
