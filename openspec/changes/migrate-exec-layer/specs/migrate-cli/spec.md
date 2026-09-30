## ADDED Requirements

### Requirement: 子命令集合

`hcm-migrate` SHALL 提供四个子命令：`init`（建记录表，存量库另垫库版本）、`up`（执行迁移）、`status`（查看每库进度）、`list`（按执行顺序列出已注册迁移）。

#### Scenario: 无子命令时打印帮助

- **WHEN** 不带子命令执行 `hcm-migrate`
- **THEN** 打印帮助信息并以非 0 退出

#### Scenario: 未知子命令报错

- **WHEN** 执行一个不存在的子命令
- **THEN** 命令以退出码 2 失败并打印帮助

### Requirement: 全局参数

所有子命令 SHALL 支持 `--config-file` / `-c`（配置文件路径）与 `--database` / `-d`。`--database` 是库名列表：可重复，也可用逗号分隔。不传 SHALL 表示配置里已启用的全部库。传入的每个名字 MUST 对应一个已有的 Registry，未知名字以退出码 2 失败。MUST NOT 使用 `all` 作为魔法取值。

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

### Requirement: up 参数

`up` SHALL 支持 `--to`（版本上限，默认空表示不设上限）、`--catch-up`（补跑模式，默认 `false`）、`--plan`（默认 `false`）。`up` SHALL 在开始执行前打印本次命令收到的模式，便于从日志判断是否误开补跑。

#### Scenario: 默认不设上限且为默认模式

- **WHEN** 执行 `up`
- **THEN** 处理全部注册迁移，且按默认模式判定

#### Scenario: 打印本次模式

- **WHEN** 执行 `up --catch-up`
- **THEN** 输出显式标明本次为补跑模式

### Requirement: init 命令

`init` SHALL 负责记录表的建立与库版本的起点。`--mode` MUST 必填且无默认值，取值 `empty`（只建空表）或 `adopt`（建表并把 `<= 该库基线` 的已注册迁移逐条写入 `success` 而不执行它们）。`init` SHALL 幂等，判据为记录表是否存在：表已存在时对该库整条命令 no-op，两种 mode 皆然，因此 `init` SHALL 可以每次部署都执行。`--plan` SHALL 只打印将建的表与将标记的 ID、版本。

#### Scenario: 空库建空表

- **WHEN** 对不存在记录表的空库执行 `init --mode=empty`
- **THEN** 建立 `hcm_migration_record` 且不写入任何记录

#### Scenario: 存量库垫起库版本

- **WHEN** 对不存在记录表的存量库执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 建表后为版本不超过 `v1.9.2` 的迁移各写入一行 `success`，且它们的 `up()` MUST NOT 被调用

#### Scenario: 表已存在则 no-op

- **WHEN** 记录表已存在，再次执行 `init --mode=adopt --baseline main=v1.9.2`
- **THEN** 该库不被修改，既不建表也不写入任何记录

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

### Requirement: up 不建表

`up` MUST NOT 创建记录表或审计表。两表任一不存在时 `up` MUST 以退出码 3 失败并提示先执行 `init`（「已初始化」= 两表都在）。此约束防止未接入框架的存量库因自动建出空账本而被从头重放全部历史迁移。

#### Scenario: 缺表时拒绝执行

- **WHEN** 目标库不存在记录表或审计表，执行 `up`
- **THEN** 命令以退出码 3 失败并提示先执行 `init`，MUST NOT 建表，MUST NOT 执行任何迁移

#### Scenario: 存量库不会被重放

- **WHEN** 一个有数据但尚未接入框架的存量库上单独执行 `up`
- **THEN** 命令失败退出，历史迁移 MUST NOT 被执行

### Requirement: 部署形态

迁移二进制 MUST 只打进迁移镜像，MUST NOT 打进业务 Pod。Helm hook Job 的命令 SHALL 为 `init && up`，`--mode` 与 `--baseline` 由该环境的 values 提供。MUST NOT 依赖任何额外的一次性 Job 或人工登录容器来完成初始化。

#### Scenario: 业务镜像不含迁移程序

- **WHEN** 检查业务 Deployment 的镜像与命令
- **THEN** 其中没有 `hcm-migrate`

#### Scenario: 部署链一次跑完

- **WHEN** 存量环境首次部署
- **THEN** hook Job 执行 `init && up`，`init` 建表并垫库版本，`up` 接着跑增量，无需额外操作

#### Scenario: 后续部署 init 无副作用

- **WHEN** 同一环境再次部署，values 中仍保留 `--mode=adopt` 与 `--baseline`
- **THEN** `init` 因记录表已存在而 no-op，`up` 正常跑增量

### Requirement: status 与 list 输出

`status` SHALL 对每个库打印库当前版本、各状态计数与未执行清单，并标出未执行项里哪些 `version ≤ 库当前`。`status` MUST NOT 接收模式参数，也 MUST NOT 推断下一次 `up` 会使用哪种模式，因为模式只存在于那一次 `up` 的 `--catch-up` 上。`status` 为只读，MUST NOT 建表或写入。`list` SHALL 按执行顺序打印全部已注册迁移的 ID、版本、时间戳与语义名，且 MUST NOT 需要数据库连接。

#### Scenario: status 只读

- **WHEN** 执行 `status`
- **THEN** 输出各库进度，数据库内容不发生任何变化

#### Scenario: status 标出默认模式会失败的项

- **WHEN** 某库有未执行迁移且其版本不高于库当前版本，执行 `status`
- **THEN** 该项被标出，输出不声称本次将使用默认模式或补跑模式

#### Scenario: list 不连库

- **WHEN** 数据库不可达时执行 `list`
- **THEN** 仍能按顺序打印全部已注册迁移

### Requirement: 退出码

命令 SHALL 使用固定退出码：`0` 成功；`1` 执行期失败；`2` 参数或配置错误；`3` 记录表问题（两表任一不存在，或记录内容非法）；`4` 默认模式检出漏执行；`5` 注册表内容问题（未开 `--allow-pending` 的 `PENDING`，或两个及以上不同非空 label）；`6` 疑似 ID 复用。退出码 `4` MUST 与 `1` 区分，使出包工具能判断本包是否需要开补跑。计划阶段多项问题优先级 MUST 为 6 > 5 > 4。权威全文见 `openspec/changes/adjust-migration-specs/specs/migrate-cli/spec.md`。

#### Scenario: 漏执行返回专用退出码

- **WHEN** 默认模式下检出漏执行
- **THEN** 退出码为 4

#### Scenario: 占位符在开关关闭时返回注册表退出码

- **WHEN** 未传 `--allow-pending`，注册表中存在常量 `PENDING`
- **THEN** 退出码为 5

#### Scenario: 疑似 ID 复用返回专用退出码

- **WHEN** 检出迁移后缀不同的同 ID
- **THEN** 退出码为 6

#### Scenario: 迁移执行失败返回 1

- **WHEN** 某条迁移的 `up()` 返回错误
- **THEN** 退出码为 1

### Requirement: 输出分工

stdout SHALL 只承载供人阅读的判定结果与汇总表，过程日志与错误详情 SHALL 走 `pkg/logs`，两者 MUST NOT 混排，使 Job 日志可直接阅读。

#### Scenario: 判定清单可直接阅读

- **WHEN** 查看迁移 Job 的日志
- **THEN** 能看到成块的判定结果与汇总，不被过程日志打断
