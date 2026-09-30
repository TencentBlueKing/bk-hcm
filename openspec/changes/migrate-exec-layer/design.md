## Context

> **权威性说明（2026-09 起）**：退出码、label 检查时机、`PENDING` 开关、ID 复用与「已初始化」定义，以 `openspec/changes/adjust-migration-specs/` 及其对应的 `adjustments-2026-09-23.md` 为准。下文 D1 退出码表、D4 末段「连库前 label / PENDING → 退出码 3」、以及「有 4 用 4」类叙述已被取代：现为 0–6 七档；label 与未开开关的 `PENDING` 为退出码 5，在连库后的执行前校验汇总；疑似 ID 复用为退出码 6；缺表（两张都要在）与非法记录为退出码 3。

`migrate/util/`（MetaOrm + 幂等 DDL/DML）已经落地并有表驱动测试，但没有任何调用方：没有注册表、没有版本比较、没有执行记录表、没有可执行入口。本次设计执行层，把「一堆写好的 `up()` 函数」变成「一次可重复、可续跑、有留痕的部署步骤」。

参考实现是 bk-cmdb 的 `admin_server/upgrader`（仓库内 `migrationV2/cmdb-src`）。它每一层单看都很弱——无锁、无事务、无逐条记录、无重试——组合起来五年没出大事故，前提是 MongoDB 的宽容语义、慢发布节奏、migration 单源头。HCM 三条都不满足：MySQL 的加列 / 加索引依赖倒挂是硬报错，发布节奏快，且内外部双源头合入。所以本设计在两处加厚（逐条记录、明确事务策略），其余分层直接借鉴。

权威输入是本地记忆 `.cursor/docs/iwiki/4039435213-hcm-migration/understanding.md`（iWiki 文档 4039435213 的图文理解，含 2026-09-22 补入的特性分支 / 归档 / 默认与补跑，以及同日的执行层纠正）。

约束：

- 迁移文件只 import `migrate/register` 和 `migrate/util`，不 import `migrate/engine`。
- ORM 走裸 `Do()`，不挂 `ModifySQLOpts`，避免租户 SQL 改写把迁移打到错误的表上。
- 迁移 Job 单实例、失败阻断发布、不自动重试。

## Goals / Non-Goals

**Goals:**

- 定义 `hcm-migrate` 的全部子命令、参数、退出码与输出，使出包和 Helm hook 能无歧义地调用，且不需要任何额外的一次性执行机制。
- 定义版本号的解析与**全序**比较，使执行顺序在任意文件组合下确定且可复现。
- 定义注册表的 `init()` 期强校验，把「版本写错、ID 重复、未定版占位」从运行期错误提前成启动期失败。
- 定义 `hcm_migration_record` 的结构、建表时机、写入时机，以及「库当前版本」的计算方式。
- 定义主库与 OBS 两份连接的构造、缺省跳过、失败隔离。
- 定义逐条执行判定（默认 / 补跑、版本上限、断点续跑、`--plan`）与每迁移的事务策略。

**Non-Goals:**

- 不实现任何 Go 代码，本变更只产出规划。
- 不修改 `migrate/util/**` 的行为，也不为它追补 spec。
- 不迁移 118 个历史 SQL 文件，不实现 `new-migrate.sh` / `release-migrate.sh` / `archive.sh` 三个脚本的内部逻辑（本次只约定它们与执行层的接口，即文件名、注册参数、`imports.go`）。
- 不做迁移回滚（`down`）。已经执行的变更无法逆转，靠幂等和人工处理。
- 不做分布式锁（见「无分布式锁」决策）。

## Decisions

### D1 命令与参数

四个子命令：`init` / `up` / `status` / `list`。全局参数沿用仓库既有约定。

全局参数（所有子命令可用）：


| 参数                     | 类型                | 默认  | 说明                                                                                                                     |
| ---------------------- | ----------------- | --- | ---------------------------------------------------------------------------------------------------------------------- |
| `--config-file` / `-c` | string            | 无   | 配置文件路径，沿用 `pkg/cc` 的加载方式                                                                                               |
| `--database` / `-d`    | 字符串列表，可重复，也可用逗号分隔 | 不传  | 选择本次作用的库。不传表示配置里已启用的全部库。传入名字的子集，例如 `--database main,obs` 或 `-d main -d obs`。名字必须对应已有的 Registry，未知名字退出码 2。没有 `all` 这个取值 |


`up`：执行迁移。


| 参数           | 类型     | 默认      | 说明                                       |
| ------------ | ------ | ------- | ---------------------------------------- |
| `--to`       | string | 空       | 版本上限，只处理 `version <= --to` 的文件。空表示不设上限   |
| `--catch-up` | bool   | `false` | 补跑模式。关掉「无 success 且 version ≤ 库当前即失败」这道闸 |
| `--plan`     | bool   | `false` | 只打印判定结果，不写记录表、不执行 `up()`                 |


`init`：建记录表，并在存量库上落一批 `success` 记录把库版本垫起来。幂等，可以每次部署都跑。


| 参数           | 类型          | 默认      | 说明                                                                       |
| ------------ | ----------- | ------- | ------------------------------------------------------------------------ |
| `--mode`     | string      | 无（必填）   | `empty`（空库：只建表）或 `adopt`（存量库：建表 + 插入执行记录）                                |
| `--baseline` | `库名=版本`，可重复 | 无       | `--mode=adopt` 时指定各库基线，例如 `--baseline main=v1.9.3 --baseline obs=v1.9.3` |
| `--plan`     | bool        | `false` | 只打印将建的表和将标记的 ID、版本                                                       |


`status`：打印每个库的当前版本、各状态计数、未执行清单。未执行项里标出哪些 `version ≤ 库当前`（这些在默认模式下会失败）。只读，不建表。`status` 不接收、也不推断模式：模式是 `up` 那一次调用的 `--catch-up`，`status` 看不到下一次 `up` 会不会带这个参数。

`up` 在开始时打印自己这次收到的模式（默认或补跑），因为这个参数就在本次命令上。

`list`：按执行顺序打印全部已注册迁移（ID、版本、时间戳、语义名）。不连库，供开发本地核对顺序。

退出码（已被 `adjust-migration-specs` 扩展为 0–6，下表为原稿；以实现与 delta 为准）：


| 码   | 含义（原稿）                                              | 现行（adjust-migration-specs） |
| --- | ----------------------------------------------- | --- |
| 0   | 成功；或 `--plan` 判定无问题                             | 同左 |
| 1   | 迁移执行失败（`up()` 返回错误、连库失败、记录表写入失败）                | 执行期失败（含 panic→failed、回填/MarkSuccess 失败、断连/取消） |
| 2   | 参数或配置错误（`--to` 解析不了、`--database` 取值非法、配置缺字段）    | 同左 |
| 3   | 前置校验失败（注册表里有 `pending` 占位符、记录表不存在、记录表里有解析不了的版本） | 仅记录表问题：两表任一缺失，或记录内容非法；**不再**含 PENDING / label |
| 4   | 默认模式检出漏执行（无 success 且 version ≤ 库当前）            | 同左（仅漏执行） |
| 5   | （原稿无） | 注册表内容：未开 `--allow-pending` 的 `PENDING`，或两条及以上版本线。空 label 的数字第四段也算一条线，每个非空 label 各算一条线；没有第四段的版本不单独成线 |
| 6   | （原稿无） | 疑似 ID 复用 |


计划阶段多项问题优先级 6 > 5 > 4；退出码 3 在计划之前返回，不与 4/5/6 同时出现。4 独立于 1，是为了让出包工具能区分「这一包需要开补跑」和「真的跑挂了」，`--plan` 同样返回对应退出码，可在流水线里前置体检。

stdout 只打人要看的判定结果与汇总表，便于 `kubectl logs` 直接阅读；`pkg/logs` 负责过程日志（含 rid 与错误详情）。两者不混排。

### D2 保留 `init --mode`，`up` 不自举建表

沿用技术方案文档的 `init --mode` 与 `init && up` 命令链。方案里两个取值叫 `full` / `existing`，这里改成 `empty` / `adopt`：`full` 听起来像把迁移全跑完，实际只建表；`existing` 只说库是旧的，没说要做的事。`empty` 表示空库只建表，`adopt` 表示把已有库接进框架并垫上基线。两条职责切开：`init` 负责记录表的存在与库版本的起点，`up` 负责把增量跑完。

`init` 的幂等判据是**记录表在不在**，一个明确的布尔值：表已存在则整条命令对该库 no-op，两种 mode 都一样。所以它可以留在每次部署的命令链里，不需要任何额外的一次性执行机制——这正是它比「独立命令 + 单独起 Job」更实在的地方。


| mode    | 记录表不存在                                                 | 记录表已存在 |
| ------- | ------------------------------------------------------ | ------ |
| `empty` | 建空表                                                    | no-op  |
| `adopt` | 建表 + 把 `<= 该库 baseline` 的已注册迁移逐条写 `success`，不执行 `up()` | no-op  |


`--mode` 必填无默认。这一步是人对「这个环境是空库还是存量库」的确认，不给默认值就是不让它被猜。

`--baseline` 按库给：`--baseline main=v1.9.3 --baseline obs=v1.9.3`。两个库的 schema 进度可以不同，所以基线是每库一个值，不是全局一个。`--mode=adopt` 却一个 `--baseline` 都没给时以退出码 2 失败，否则等于悄悄退化成 `empty`。被选中的库里没有对应基线条目的，按 `empty` 处理（建空表），这样「主库是存量、OBS 是新库」这种混合情况不用跑两次命令。

`**up` 不自举建表**：记录表或审计表任一不存在时 `up` 以退出码 3 失败，提示先跑 `init`（「已初始化」= 两表都在）。这条比自动建表重要得多。设想一个还没接入框架的存量库，有人单独跑了 `up`：若 `up` 自动建出一张空账本，库当前版本就是空，于是全部历史迁移都判定为「高于库当前」，在一个有数据的生产库上从头重放一遍。让 `up` 在缺表时直接失败，这条路就被堵死了。

曾考虑取消 `init`、让 `up` 自举并把基线折进 `up`（类似 Flyway 的 `baselineOnMigrate`，只在「无记录表且库里已有业务表」时盖章）。否决，两个原因：一是「库里已有业务表」判不准——迁移本身也建表，半截状态、空 OBS 库都会让这个启发式出错，而它守的偏偏是全系统唯一能静默跳过的操作；二是划定基线是人的决策，折进 `up` 等于绕过人工确认隐式执行。`init --mode` 不够优雅，但判据可靠、决策显式。

二进制只打进迁移镜像，**不进业务 Pod**。业务进程里没有这份程序。Helm hook Job 每次部署跑 `init && up`，`--mode` 与 `--baseline` 由该环境的 values 提供。存量环境第一次部署时 `init` 建表并垫好库版本，之后每次部署 `init` 都是 no-op，values 里那几个参数留着也无害。非 Kubernetes 的二进制部署同理，部署脚本里就是 `init && up` 这条链。

仍然保留的一处风险：全新空库的环境如果 values 误配成 `--mode=adopt --baseline main=v1.9.3`，第一次部署就会把 `<= v1.9.3` 的迁移全标成 success 而不执行，得到一个残缺的库。这是首次安装时的人为配置错误，发生在人主动选定基线的时刻，不会随时间静默累积；缓解靠 `--plan` 预演和接入说明。

### D3 目录与文件职责

```text
migrate/
├── migrate.go                 入口：pkg/logs 打 stderr，调用 cli.Run
├── cli/                       子命令装配：init / up / status / list
│   ├── cli.go                 Run、全局参数、退出码接线
│   ├── init.go
│   ├── up.go
│   ├── status.go
│   ├── list.go
│   └── output.go              计划 / status / list / 汇总的 stdout 表格
├── Makefile                   build / test（挂接 scripts/check-migrate.sh）
├── scripts/
│   ├── lib.sh                 共用 version_group / replace_in_file
│   ├── new-migrate.sh         建模板（时间戳 + 目录）并写 imports.go
│   ├── release-migrate.sh     pending 整目录定版
│   ├── check-migrate.sh       出包门禁：目录与空白 import 一致；依赖门禁（禁 import engine|schema|cli）也在此
│   └── archive-feat-migrate.sh 特性分支归档：-label 版本收成主线三位版本
├── register/
│   ├── register.go            Migration / Registry / Regist / All；Main 与 Obs 两个实例
│   ├── version.go             Version / Suffix / Parse / Compare / IsPending
│   ├── register_test.go
│   └── version_test.go
├── schema/                    与 engine 平级：记录表与运行审计表
│   ├── schema.go              两表 DDL、InitTables、MissingTables、BaselineMigrations、CheckInitialized
│   ├── recordstore.go         记录读写 + 库当前版本计算
│   ├── auditstore.go          NewAudit().Begin / End
│   └── *_test.go
├── engine/
│   ├── datasource.go          cc 配置 → 主库与 OBS 的 dao.Set；缺省跳过
│   ├── plan.go                Prepare / BuildPlan / CollectPlanErrors；Plan 含 Records
│   ├── executor.go            Execute / NewBlockedResult
│   └── *_test.go
└── migrations/
    ├── imports.go             空白 import，把版本包拉进编译
    ├── main/                  → register.Main
    └── obs/                   → register.Obs
```

退出码哨兵（`ErrUsage` / `ErrPrecondition` / `ErrMissed` / `ErrRegistry` / `ErrIDReuse`）与 `ExitCode` 映射放在 `pkg/migrate/errors.go`，不在 `engine` 内。

`Regist(id, version, timestamp, description string, up UpFunc)` 五个位置参数，`up` 的签名是 `func(ctx context.Context, o orm.Interface) error`，和 `migrate/util` 里每个幂等辅助函数对齐。校验不过时 `panic`，沿用仓库里 `pkg/async/action` 注册失败的写法；这些调用都在 `init()` 里，panic 即进程在 `main` 之前终止。panic 文本带上全部五个字段，因为这是读日志的人定位到具体文件的唯一线索。

Migration ID 只收小写标准格式的 UUID，不收大写、不带连字符、花括号或 `urn:uuid:` 前缀的写法。执行器是拿这个串和记录表里的值比对来判断「已执行」的，MySQL 比较时大小写不敏感而 Go 敏感，同一个 ID 允许多种写法会让它要么撞唯一键、要么被当成两条重跑一遍。

`engine` 依赖 `register`，`register` 不依赖 `engine`。迁移文件只 import `register` + `util`，这条靠 `scripts/check-migrate.sh` 检查 `migrations/` 子树里是否出现 `migrate/engine`、`migrate/schema`、`migrate/cli` 导入——Go 编译器不会阻止这种导入（没有环），所以必须有门禁。

### D4 版本解析与比较器

语法三种写法，一条正则：

```text
^v(\d+)\.(\d+)\.(\d+)(?:\.(\d+)|-([a-z][a-z0-9]*)\.(\d+))?$
```

额外拒绝前导零（`v01.9.3`、`v1.9.3.01`）。数字第四段为 0 时收成没有第四段：`v1.9.3.0` 与 `v1.9.3` 是同一版本。`v1.9.3-tenant.0` 不收，标签还在，仍然晚于 `v1.9.3`。

文件名用下划线把版本和时间戳切开，不用点：

```text
v1.9.3_20260905160000_add_bk_asset_id.go
v1.9.3.1_20260904090000_ziyan_backfill.go
v1.9.3-tenant.1_20260905160000_foo.go
```

版本号自己含有点号和短横线。分隔符再用点，`v1.9.3.1.20260904090000_foo.go` 就要靠「14 位数字是时间戳」来猜切在哪，而 `v1.9.3.0` 又和 `v1.9.3` 等同，点号更分不清。第一个下划线左边是版本，交给上面的正则；右边是 14 位时间戳和语义名。`pending/` 里的文件本来就是 `20260920153012_add_legacy_asset.go`，定版后只是在前面加上版本和下划线。

```go
// Suffix 是版本号的第四段。Label 为空表示内部特有的纯数字序号。
type Suffix struct {
    Label string
    Seq   int
}

// Version 是一条迁移的版本号。Suffix 为 nil 表示三位版本。
type Version struct {
    P1, P2, P3 int
    Suffix     *Suffix
    Raw        string
}
```

比较器给出**全序**，目的只是拿到确定、可复现的执行顺序，**不用来挑哪条线的文件可以跑**（见 D8 偏离 2）。逐级比较：

1. `P1` → `P2` → `P3`，按数值。
2. 第四段：没有第四段（含收成无第四段的 `.0`）小于任何真正的第四段。同一前缀下三位先跑，`.1` 及以后后跑。
3. 两边都有第四段：先比 `Label`（空串最小，其余按字典序），再比 `Seq` 按数值。
4. 时间戳：14 位定长，按字符串比即等于按时间比。
5. Migration ID：按字符串比。最后一级决胜，保证同版本同时间戳的两个文件顺序也确定。

前 3 级只看版本号，由 `register.Compare(a, b Version) int` 给出。第 4、5 级的时间戳和 Migration ID 是迁移的字段、不是版本的字段，由注册表的 `All` 在 `Compare` 返回 0 时接着比（见 D3 的 `Migration` 结构）。两处同在 `register` 包，合起来就是上面这个全序。

第四段（数字序号或标签）表示这条分支自己的特有提交。一棵正常的树里版本线最多一条：内部主线是空标签的 `.N`，tenant 分支是 `-tenant.N`。同一个注册表里出现两条线，说明别的分支的特有提交混进来了，或者归档没做完。这不是排序能收拾的情况。**现行规则（取代「连库前退出码 3」）**：`up` 在连库之后的执行前校验中发现两条及以上版本线，以退出码 5 失败，列出每条线及其 Migration ID（排序），一条都不执行。空 label 的数字第四段也算一条线，每个非空 label 各算一条线；没有第四段的版本不单独成线。与 PENDING、漏执行、ID 复用同一轮汇总，优先级见 D1。

所以第 3 条不是为了「把两条线排得好看」。跨标签的先后只让 `Compare` 在任何输入上都有确定结果，避免比较函数本身不确定。先比 `Label` 再比 `Seq` 是这个全序里最简单的写法。它不会被执行到：那种输入在排序之前就已经失败了。

正常顺序（一棵树里只有一个标签，从先到后）：

```text
v1.9.3          与 v1.9.3.0 相同，并列时再比时间戳和 ID
v1.9.3.1
v1.9.3.2
v1.9.3-tenant.1
v1.9.3-tenant.2
v1.9.4
v1.9.10
```

边界：


| 输入                                                      | 结果                                                          |
| ------------------------------------------------------- | ----------------------------------------------------------- |
| `v1.9.3` 与 `v1.9.3.0`                                   | 同一版本。两者都在时，顺序看时间戳，再看 Migration ID                           |
| `v1.9.9` 与 `v1.9.10`                                    | 按数值，`v1.9.9 < v1.9.10`                                      |
| `v1.9.3.2` 与 `v1.9.3-tenant.1`                          | 数字第四段整体先于带标签第四段                                             |
| `v1.9.3-tenant.1` 与 `v1.9.3-woa.1`                      | 比较器按标签字典序给出先后；注册表里同时出现这两个标签时，执行前校验以退出码 5 失败，不会按这个顺序执行 |
| `1.9.3` / `v1.9` / `v1.9.3.1.2`                         | 非法                                                          |
| `v1.9.3.tenant.1` / `v1.9.3-Tenant.1` / `v1.9.3-tenant` | 非法                                                          |
| `v01.9.3`                                               | 非法（前导零）                                                     |
| `PENDING`                                               | 非法版本，但是合法的占位符，单独判定                                          |


未定版占位符是 `register` 包里的常量，值为大写的 `PENDING`。注册、`IsPending` 和执行前扫描都用这个常量，不写字符串字面量。目录仍然叫 `pending/`，文件名也不带这个词。`Parse("PENDING")` 返回错误。**现行规则（取代「连库前一律退出码 3」）**：未开 `--allow-pending` 时，执行前校验以退出码 5 失败并列出 ID；打开开关则可执行一次。`list` 例外：照常打印，版本列就是 `PENDING`，退出码 0。

不引入 bk-cmdb 的 `remapVersion` / `wrongVersion` 历史错版号映射表。那是它五年里积累的存量错版号补丁（`x18_10_10_01` → `x18.10.10.01`）；HCM 的版本号是全新的，没有存量，对非法版本一律硬失败即可，不给「错版号也能跑」留口子。

版本上限 `--to` 用同一个比较器：只处理 `Compare(m.Version, ceiling) <= 0` 的文件。`--to v1.9.3` 包含写成 `v1.9.3.0` 的文件，因为两者等同；**不包含** `v1.9.3.1`。要跑到内部四位文件，必须传确切的四位上限（如 `--to v1.9.3.2`）。

### D5 记录表

表名 `hcm_migration_record`，**每库一份**，建在对应库内。不是在主库开一张总表加 DB 字段区分——OBS 库被重建后，主库账本仍显示已执行，缺表就发现不了。

```sql
CREATE TABLE `hcm_migration_record` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增主键',
  `migration_id` VARCHAR(64) NOT NULL COMMENT '迁移 ID',
  `version` VARCHAR(64) NOT NULL COMMENT '最近一次成功时注册的版本',
  `status` VARCHAR(16) NOT NULL COMMENT '执行状态，取值 running/success/failed',
  `message` VARCHAR(1024) NOT NULL DEFAULT '' COMMENT '失败摘要',
  `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uidx_migration_id` (`migration_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin COMMENT='migration 执行记录表';
```

`version` 宽 64，和注册时的长度上限是同一个数：超长的版本在 `init()` 注册时就 panic，不会等到写库时被截断或拒绝。排序规则用 `utf8mb4_bin`，`migration_id` 的唯一索引按字节比较，和 Go 的字符串比较一致。查重和状态更新都走这个唯一索引，不另建索引。

建表时机（先有鸡还是先有蛋）：记录表**不写成一条迁移**，否则「建账本」这件事没法记在账本里。`init` 调用 `util.CreateTableIfNotExists`，把上面这条 `CREATE TABLE` 原样传进去。存在性判断和跳过都在这个工具里，`engine` 不自己拼 `IF NOT EXISTS`，也不直接查 `information_schema`。

`CreateTableIfNotExists` 返回 `(created bool, err error)`。`created` 为 true 表示这次执行了 DDL、表是新建的；表已存在，或任一错误，都是 false。`init` 只在 `created` 且 `--mode=adopt` 时插入基线记录，表已经在就整库 no-op。

`up` / `status` / `list` 都不建表。`up` 在缺表时以退出码 3 失败并提示先跑 `init`（理由见 D2：自动建表会让未接入的存量库被整段重放）；`status` 报「尚未初始化」。

一条迁移在表里**只有一行**，`migration_id` 唯一。查重看 ID，不看 version。写入时机：


| 时机                  | 操作                                                                                  |
| ------------------- | ----------------------------------------------------------------------------------- |
| 执行前                 | `INSERT ... ON DUPLICATE KEY UPDATE status='running', message='', updated_at=NOW()` |
| 成功                  | `UPDATE status='success', message='', version=<此刻注册的 version>`                      |
| 失败                  | `UPDATE status='failed', message=<截断后的错误>`                                          |
| `init --mode=adopt` | 建表后对 `<= 该库 baseline` 的项 `INSERT` 一行 `success`，不执行 `up()`                           |


`version` 存的是**执行当时**的快照。代码里的 version 后来会改（定版、归档），所以按 ID 读到 success 而记录里的 version 与此刻注册的不一致时，打 WARN「坐标被改过」并照常跳过——归档后在特性分支环境上这条 WARN 是预期，不是失败。

**库当前版本**必须在 Go 里算：取本库全部 `success` 记录，逐条 `Parse`，用比较器取最大。不能用 SQL 的 `ORDER BY version DESC LIMIT 1`——版本串的字典序和版本序不一致，`v1.9.10` 会排在 `v1.9.9` 前面。也不是「最后写入的那一条」：补跑写入的成功记录版本偏低，不应该把库当前改小。

读取记录时，不论该行是 `running`、`success` 还是 `failed`，只要 `version` 解析不了就以退出码 3 失败，不是跳过。检查发生在装载记录的时候，和未知 `status` 放在一起。账本里有本工具不认识的行，意味着它被别的东西写过；此时算出来的「库当前」偏小，会让本该报错的漏执行伪装成「> 库当前」而被执行，比直接拒绝危险得多。

`message` 截断到 1024：按字节截断且不能把多字节字符切一半。完整错误进 `pkg/logs`，表里只留摘要。

### D6 数据库连接管理

配置沿用 dataservice 的那份 yaml。Job 与 dataservice Deployment 共用 ConfigMap，挂成一个文件，`--config-file` 指向这个路径。迁移镜像里没有 data-service，也不 import `cmd/data-service`，不访问它的进程。

读配置只有一条路径：`cc.InitService(cc.DataServiceName)`，然后 `cc.LoadSettings` 用 `ReadFile` 读这个文件并 `yaml.Unmarshal`。`loadFromFile` 按进程里的 `ServiceName()` 决定反序列化成哪种结构；只有先登记成 `data-service`，文件才会进 `*DataServiceSetting`。`cc.DataService()` 不读文件，它只是把已经载入内存的配置做类型断言。断言成功后取 `Database`（`cc.DataBase`）和 `OBSDatabase`（`*cc.DataBase`，yaml 里没有这一段就是 `nil`），各调一次 `dao.NewDaoSet`。

`LoadSettings` 会跑 `DataServiceSetting.Validate()`。除了两个库，它还检查 `network`、`service`、`crypto`、`cmdb`。所以挂进去的必须是 dataservice 的完整配置文件，不能是只含数据库段的片段。这些检查是 `pkg/cc` 里的字段校验，随迁移二进制编译进来，不是 data-service 的启动逻辑。

- 迁移函数拿到的是 `dao.Set` 的 `GetOrm()` 裸 `Do()`，不挂 `ModifySQLOpts`。租户 SQL 改写会把语句重写到别的表上，对 DDL 是灾难。
- `OBSDatabase` 为 `nil`（外部版）时整段跳过，打一条显眼的 Warn 日志说明「未配置 OBS 库，跳过」，退出码仍为 0。不能静默跳过。
- 顺序固定先主库、后 OBS。两库之间无顺序依赖，固定顺序只为让日志和排障可预期。
- 失败隔离：主库失败则直接退出，不再连 OBS；OBS 失败则主库已写的 success 记录保留，**不回滚**。跨库没有事务，也不做补偿。任一库失败整个进程非 0 退出，Job 失败，阻断发布。
- 以后加库：多一个配置对象 + 一个 `Registry` 实例 + 一棵目录，执行器按「有连接才跑」循环，不需要改判定逻辑。

DB 就绪：靠 Job 的 initContainer（busybox `nc`）等端口通，程序自身只做一次连接探测，失败就带明确错误退出。不在进程内加重试轮询——重试会把「配置写错」和「DB 还没起来」两种错误都拖成超时，反而更难排查，而部署顺序本来就是部署层该管的事。

### D7 执行语义

注册表是全量文件。遍历顺序固定为版本号从旧到新、同版本比时间戳（再比 ID）。补跑不另排序、不另切一批，只改一条判定：

同一个 Registry 里允许两条迁移使用相同的 Migration ID。这是内外部合入的正常结果：同一段变更内部写成四位、外部写成三位，两份都留着，ID 相同。注册期不拒绝。执行时只认 ID：排序靠前的那条若还没有 success 就执行，另一条读到 success 后跳过，不调用它的 `up()`。记录里的 version 与这条注册 version 不一致时照常 WARN，不失败。

```text
for 每个已排序、且不超过版本上限的 m:
  已有 success            → 跳过（只认 Migration ID，同 ID 的另一份文件走这里）
    记录 version ≠ 注册 version → WARN：坐标被改过
  超过版本上限            → 不执行
  没有成功记录（无记录 / running / failed）:
    默认模式:
      version ≤ 库当前    → 漏执行，退出码 4
      version > 库当前    → 执行
    补跑模式:
      一律执行
```

事务策略：`engine` **不在 `up()` 外面包事务**，由每条迁移自己决定。DDL 走 `util`、不开事务——MySQL 的 DDL 隐式提交，包了也回滚不了，反而给人「失败会回滚」的错觉；DML 由迁移作者自己开小事务分批，避免大表长事务锁表。记录表的写入一律在迁移事务之外，`failed` 记录必须能在迁移自己的事务回滚后仍然留下。

无分布式锁：并发安全由部署形态保证——Job `parallelism: 1`、`backoffLimit: 0`、`restartPolicy: Never`，二进制部署是 `&&` 链里的单进程。全局只有一个写者。风险是有人给 Job 加副本或并行度，这一条必须写进 chart 注释；进程级互斥（MySQL `GET_LOCK`）留作后续增强，本次不做。

无自动重试：失败即中断，后面的文件不执行，业务不启动，等人介入。失败后库当前版本仍停在已成功记录的最大版本；重跑时已 success 的按 ID 跳过，等于从断点续跑。

`--plan` 打印每条的判定结果（`EXECUTE` / `SKIP-SUCCESS` / `ABOVE-MAX-VERSION` / `MISSING`），不写记录表、不跑 `up()`，默认 / 补跑与版本上限同样生效。若默认模式下判出 `MISSING`，`--plan` 也返回退出码 4。

### D8 相对技术方案文档的偏离

集中列在这里，便于评审时逐条否决或接受。

1. `**init --mode` 的取值与语义**：保留方案的 `init --mode` 与 `init && up` 命令链，但取值从 `full` / `existing` 改为 `empty` / `adopt`。`empty` 只建空表，`adopt` 建表加垫库版本。幂等判据是记录表是否存在，`--mode` 必填无默认，`--baseline` 按库给且 `adopt` 缺基线即失败。另外 `up` 在缺表时失败而不是自举建表。理由见 D2。
2. **执行层不按版本线 / 标签过滤候选**。编译进二进制的迁移一律要执行，只受「按 ID 跳过已成功、版本上限、默认 / 补跑判定」三件事约束；主线不跑 `-tenant` 靠这些文件不合进主线（git 目录与分支合入纪律）保证。据此 `migrationV2/docs/migration-version.md` 里「先留下当前线该跑的」一段作废，执行器也没有「当前线」这类配置项。
3. **比较器补全跨标签全序和 ID 末位决胜，但两条版本线共存时直接失败**。方案只定义了同标签之间比 `Seq`。跨标签先后只为让比较函数确定，不作为可执行的顺序；同一注册表出现两条版本线（数字第四段与具名 label，或两个非空 label）视为特有提交混入或归档未完成，**现行为退出码 5**（执行前校验，非连库前 3）。
4. **明确 `--to` 的边界语义**：`--to v1.9.3` 包含 `v1.9.3.0`，不包含 `v1.9.3.1`。方案提了版本上限但没展开这条边界，也把 `.0` 当成了另一个版本。
5. **记录表里出现无法解析的 `version` 时硬失败**（退出码 3），方案未提这种情形。
6. `**--mode` 必填无默认，`adopt` 缺 `--baseline` 即失败**。方案未规定缺省行为，留默认会让「空库还是存量库」这个人工判断被猜。
7. **不引入 `remapVersion` 式历史错版号映射**，非法版本一律硬失败。
8. `**engine` 不包裹事务**，方案只说「每迁移明确事务策略」，这里把责任明确落到迁移自己，并规定记录写入在迁移事务之外。
9. **复用 dataservice 的 `cc` 配置段**，不新增 migrate 专属配置段。
10. **新增退出码 4 / 5 / 6** 区分漏执行、注册表内容问题与疑似 ID 复用（详见文首权威性说明与 `adjust-migration-specs`），使出包工具和 `--plan` 能前置判断。
11. **预演参数叫 `--plan`**，方案文档里的 `dry-run` 指的是同一件事。
12. `**--database` 是库名列表**，不传表示已启用的全部库，不设 `all` / `main` / `obs` 三选一。库数量增加时不用改参数形态。
13. **拒绝版本号前导零**。数字第四段 0 收成无第四段，所以 `v1.9.3` 与 `v1.9.3.0` 等同；前导零仍然拒绝。
14. **文件名里版本和时间戳用下划线分隔**，不用点。版本号自身已有点号，且 `.0` 与三位等同，点号切不开。
15. **Migration ID 限定为小写标准 UUID**。方案只说「合法 UUID」。理由见 D3：多种写法会和记录表的比对打架。
16. **`All` 把 PENDING 的迁移排在最后**，它们之间再按时间戳和 ID。方案没定义占位符在排序里的位置。占位符没有版本可比，当成零值会被排到最前面，正好排在最该先跑的位置上；未开 `--allow-pending` 时执行前校验会因占位符失败（退出码 5），这个顺序也影响 `list` 的显示。
17. **`util.CreateTableIfNotExists` 现在返回 `(created bool, err error)`**。`created` 仅在本次真正执行了 DDL 时为 true；表已存在，或任一错误（非法表名、空 DDL、`HasTable` 失败、`Exec` 失败）都是 false。`init` 靠这个返回值区分「这次新建」和「本来就在」。
18. **基线记录用一条多行 INSERT 写入，全成或全不成**。插入失败时 `InitRecordTable` 删掉刚建的 `hcm_migration_record`，让 init 可以整段重试。否则留下一张空表，下次 init 会当成已初始化而 no-op，随后 `up` 会把基线以内的迁移全部重跑。删除也失败时，错误信息要求操作者手工删掉这张表再重试。
19. **基线选择跳过 `PENDING`，并按 migration ID 去重，保留执行顺序里的第一次出现**。同一个 ID 插两行会被 `migration_id` 唯一键拒绝，整批插入失败。高于基线的那一份不占这个名额，后面一份不高于基线的仍会写入。
20. **记录表里未知的 `status` 是前置条件失败（退出码 3）**。和无法解析的版本同一个理由：这张表是别的东西写的。`running` / `success` / `failed` 以外的值，包括空串和大小写不同的写法，一律拒绝。
21. **相等版本的库当前版本取字典序较小的原文**。例如同时有 `v1.9.3` 和 `v1.9.3.0` 时，`CurrentVersion` 固定返回原文 `v1.9.3`，结果不依赖 map 的遍历顺序。只有 `.0` 那一条时，原文保持 `v1.9.3.0`，不会被改写成三位。
22. **`--database` 的空元素、重复和大小写**。空元素（如 `main,`、单独的空串或纯空白）是退出码 2；重复名字合并成一个；输出顺序始终先 main 后 obs，与输入顺序无关；名字大小写敏感，`MAIN` 不是 `main`。
23. **退出码不在 engine 里映射**。项目里没有进程退出码框架，`errf` 是接口业务错误码（错误文本是 JSON、`NewFromErr` 只留 message、断链），不适用。哨兵错误 `ErrUsage` / `ErrPrecondition` / `ErrMissed` / `ErrRegistry` / `ErrIDReuse` 与 `ExitCode` 放在 `pkg/migrate/errors.go`，产生处用 `%w` 包装。CLI（`migrate/cli`）经 `errors.Is` 映射到退出码 2/3/4/5/6，链上没有哨兵错误时为 1，nil 为 0。
24. **success / failed 写入把影响行数 0 当成错误（记录不存在）**。`MarkRunning` 总是先执行，而且状态一定会变，所以更新到 0 行说明该走的那一行不在。
25. **`Load` 对每一行校验 `version`，不限状态**。未知 `status` 仍先拒绝。通过之后，`running` / `success` / `failed` 的 `version` 都要能解析，失败同样是前置条件错误（退出码 3），并且不返回已经读到的记录。`CurrentVersion` 的行为不变：只在 `success` 行里取最大，解析失败仍拒绝，用来挡住没有经过 `Load` 拼出来的记录。

## 待确认

以下条目写在 `migrate-registry` spec 里，review 确认前不实现，也不改变现有行为。

1. **相同 ID 的动作名必须一致。** 不比较迁移文件正文。内外部都保留的同一段通用变更，文件名里时间戳之后的动作名（即 `Regist` 的 description）必须相同，版本号和时间戳可以不同。`init` / `up` 连库前扫描：动作名一致则放行，不一致则以退出码 3 失败。复制文件忘了换 ID 时，动作名通常不同，于是在执行前失败，而不是被按 ID 跳过。代价是动作名略有出入的两份通用文件也会失败；规范没被遵守时，失败发生在执行前，不会漏执行。

## Risks / Trade-offs

- 补跑开关发完忘记改回默认 → 「漏执行直接失败」对后续每次发布都失效，版本填错的文件会被静默执行掉 → 缓解：`status` 标出「无 success 且 version ≤ 库当前」的清单，让人看得出默认模式会不会失败；`up` 开头打印自己这次收到的模式。`--plan` 与实跑用退出码 4 明确区分需要补跑的场景。chart values 里就近写注释说明它是按包一次性开关。模式不落库，`status` 推断不了下一次 `up` 会不会带 `--catch-up`。
- `imports.go` 漏空白 import → 该迁移永远不跑且不报错，这是全系统唯一的静默失效点 → 缓解：`new-migrate.sh` 自动写入、出包门禁 `check-migrate`、`list` 输出可人工对照目录。
- 单写者假设被破坏（给 Job 加副本或并行度）→ 两个进程同时执行同一条迁移 → 缓解：chart 里固定 `parallelism: 1` 并加注释；后续可加 MySQL `GET_LOCK`。
- 归档进来的老文件打在已经演进过的库上 → 加列靠幂等能过，但删列、按旧列回填会中途失败 → 缓解：Job 失败即阻断发布、业务不启动，人工处理；已执行的变更无法逆转。
- 基线定错，或空库环境误配成 `--mode=adopt` → 该跑的迁移被标成 success 永不执行，得到残缺的库 → 缓解：`--mode` 必填、`adopt` 缺 `--baseline` 直接失败、支持 `--plan` 先看清单、接入说明里写明如何选定基线。这是首次安装时的人为配置错误，`init` 的表存在即 no-op 保证它不会随后续每次部署反复发生。
- `message` 截断吞掉根因 → 缓解：完整错误进 `pkg/logs`，表里只留摘要。
- 复用 dataservice 配置文件意味着 migrate 跟着它的结构走 → `Validate()` 还检查 `network`、`service`、`crypto`、`cmdb`，dataservice 配置重构会连带影响加载 → 缓解：运行时只用 `Database` / `OBSDatabase`；不 import `cmd/data-service`。
- adopt 的基线本身不落库 → 库当前版本只取已写入的 success 记录的最大值。注册表里没有不超过基线的迁移时一行都不写，库当前为空，与 empty 相同；有记录时最大值也可能低于基线。之后带进来版本不超过基线、却高于这些记录的文件，会被默认模式当成新迁移执行，而不是判为漏执行 → 缓解：当前接受此风险（首次接入时基线以内的历史变更应已作为迁移文件注册，此时最大记录贴近基线）；`init --plan` 会打印将标记的 ID 与版本，从中能看出接入后的库当前版本是多少、离基线有多远。

