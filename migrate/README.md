# hcm-migrate

## 快速上手

给 HCM 的 MySQL 做表结构变更。变更写在 `migrate/migrations/` 里，随二进制一起编译。每个库自己记哪些已经跑过，主库和 OBS 库各算各的。写迁移不用先看后面的目录说明。

在仓库根目录执行脚本。脚本会改 `migrate/migrations/`，结束时自己跑一遍检查。

### 开发

一条迁移就是一个目录、一个 `migrate.go`。不要手建目录，也不要写版本号。用脚本新建，文件一律进 `pending/`。进哪个发布版本、是主线还是特性分支，由出包的人定版，不在开发这一步决定。

```bash
bash migrate/scripts/new-migrate.sh --database main --desc add_bk_asset_id
bash migrate/scripts/new-migrate.sh --database obs --desc create_obs_bill
```

`--database` 只填 `main`（主库）或 `obs`（OBS 库）。`--desc` 用小写单词，中间下划线，例如 `add_bk_asset_id`。不要加 `--version`。脚本会写好 ID、时间戳和 `imports.go` 里的引用。主库和 OBS 库各做一个目录，不要在一个文件里注册两个库。

脚本生成的 `Up` 里是一段建表、加列、加索引的示例，用注释框住，换成实际变更。建表、加列、加索引、改列名、删列、往 `id_generator` 插一行，都用 `hcm/migrate/util` 里的函数，不要手写「先查再改」。这些函数发现已经做过就跳过，所以一条迁移中途失败后，改完再跑，前面成功的步骤可以再执行一遍。

写完执行：

```bash
bash migrate/scripts/check-migrate.sh
```

它会逐项打印结果，通过打勾，失败打叉并写明原因。它不编译、不连数据库，查这六件事：目录是不是规定的样子、目录名和版本是否匹配、每个目录是不是只在 `migrate.go` 里注册一次、`imports.go` 有没有漏或多、同一个库里同一个 ID 是不是被接到了两段不同的「时间戳_名称」上、迁移有没有去依赖不该依赖的包。

开发时遵守这些约定：

- 合入之后不要改 ID，也不要改目录名里从 14 位时间戳起到结尾的那一段（例如 `20260927160000_add_bk_asset_id`）。程序靠这个判断「是不是重复迁移文件」，对不上会拒绝执行。
- 迁移文件只能引用 `hcm/migrate/register`、`hcm/migrate/util` 和 `hcm/pkg/criteria/constant`。版本保持脚本写上的 `constant.MigrationPendingVersion`，不要改成版本字符串。
- 不要引用 `engine`、`schema`、`cli`，也不要引用另一条迁移。
- `package` 固定写 `migration`。一个目录只注册一次。
- 外部版是对外发布的那一套代码，不含只在内部用的库和变更。两边都要的迁移，把整个目录（含 `migrate.go` 和同目录里的其他文件）拷到外部版的 `pending/`，不要重新生成。只在内部用的（例如只改 OBS 库）不要拷过去。拷目录不会改外部版的 `imports.go`，这一行很容易漏。漏了编译不会报错，这条迁移也不会进二进制，外部版上永远不跑。拷完必须在外部版仓库跑 `check-migrate.sh`。
- 内部提到外部的同一条迁移，ID 原样带过去。目录名也保持原样，从 14 位时间戳起到结尾的那一段必须一致，例如内部是 `pending/20260927160000_add_bk_asset_id/`，外部也是这个目录名。这段或 ID 有一边改了，程序会当成两条迁移复用了同一个 ID，拒绝执行。
- 拷过去之后如果改了 `Up` 里的逻辑，内部和外部两份一起改。只改一边，另一边再跑会对不上。

合 MR 时：

- 文件还在 `pending/`，就让它留在 `pending/`。不要在合入时改版本、挪目录。定版是出包的人做的事。
- 不进这一包的 MR，不要合进出包分支。出包分支上只留这一包要发的变更。

### 出包

每次发布只做一件事：把这个库的 `pending/` 整目录定成一个版本号。不要手改版本字符串，不要按文件挑着定。OBS 库单独定一次，不和主库混在同一条命令里。

① 合外部版代码进内部

外部版已经定过版的迁移，直接合进内部，不要为了「看起来重复」删掉其中一份。两边如果是同一条（ID 相同，目录名里从 14 位时间戳起到结尾也相同），两份都留。执行时按 ID 跳过，只会跑先排到的那一份。

② 确认 `pending/`

定版前看一遍这个库的 `pending/`。里面的目录必须都属于这一版。不该发的，先从出包分支拿掉，再定版。

定完之后这个库的 `pending/` 应该是空的。`up` 默认会扫描还停在 `PENDING` 的迁移，发现了就报错，一条都不执行。

③ 定版

一条命令定掉所选库里全部 `pending/`。脚本把占位版本换成这个版本号，并按版本挪目录。ID 和「时间戳_名称」不动。已经在环境里跑过的，定版后再跑会跳过，并把记录里的版本改成现在的版本，不会把 `Up` 再执行一遍。

| 这一包定成 | 目录会进 | 谁用 |
| --- | --- | --- |
| `v1.9.3` | `v1.9.3/` | 外部版，三位主线版本 |
| `v1.9.3.1` | `v1.9.3.x/` | 内部版。`.1`、`.2` 是内部在这个主线版本上的第几包 |
| `v1.9.3-tenant.1` | `v1.9.3.x/` | 特性分支。同一次发布不要再混进另一个标签，例如 `-woa` |

内部主库：

```bash
bash migrate/scripts/release-migrate.sh --database main --version v1.9.3.1
```

外部版主库用三位版本，目录进 `v1.9.3/`，不要定成内部的 `.1`：

```bash
bash migrate/scripts/release-migrate.sh --database main --version v1.9.3
```

特性分支单独定成自己的版本，目录进 `v1.9.3.x/`，不要定成主线的 `v1.9.3` 或内部的 `v1.9.3.1`。`.1`、`.2` 是这个分支上的第几包：

```bash
bash migrate/scripts/release-migrate.sh --database main --version v1.9.3-tenant.1
```

OBS 库单独再跑一次，版本号用这一包自己的版本。内部包用 `v1.9.3.1`，特性分支包用 `v1.9.3-tenant.1`：

```bash
bash migrate/scripts/release-migrate.sh --database obs --version v1.9.3.1
bash migrate/scripts/release-migrate.sh --database obs --version v1.9.3-tenant.1
```

一次只定一个版本号。下一包再从那时的 `pending/` 定为 `v1.9.3.2`，不要把两包收进同一个版本。

④ 检查 `imports.go`

定版脚本会改 `imports.go`。从内部拷到外部版的目录不会改这一行，这是最容易漏的地方。提交前再跑一次 `check-migrate.sh`，确认这一版的 `v1.9.3/`、`v1.9.3.x/` 和 `obs/` 下每个迁移目录都有一行引用。漏了这行，编译不会报错，这条迁移也不会进二进制，环境上永远不跑。同一个 ID、「时间戳_名称」也相同、只是版本前缀不同，检查会警告但不会失败，这是内外两份都留时的正常结果。

特性分支合回主线时，再用 `archive-feat-migrate` 把 `v1.9.3-tenant.1` 这类目录收到对应的主线版本下。这不是每次定版都要做的步骤。

```bash
bash migrate/scripts/archive-feat-migrate.sh --label tenant --dry-run
bash migrate/scripts/archive-feat-migrate.sh --label tenant
```

先用 `--dry-run` 看报告：每条的旧路径、新路径、版本变化，以及会删掉的空分组。目标目录已存在且 ID 不同、两条落到同一目录、或者归档后执行顺序变了，脚本都会停下，一个文件都不改。

它改的是目录上的版本前缀和注册版本，收成主线版本（例如 `v1.9.3`）。ID 和「时间戳_名称」仍然不动。特性环境上已经执行过的，归档后再跑会跳过；日志里如果出现「记录里的版本和现在注册的版本不一样」，这是预期的警告，不是失败。

## 目录和程序


```text
migrate/
├── migrate.go                 命令行入口
├── scripts/
│   ├── new-migrate.sh         开发：新建一条迁移
│   ├── release-migrate.sh     出包：把 pending 定到某个版本
│   ├── archive-feat-migrate.sh 出包：把特性分支收进主线版本
│   ├── check-migrate.sh       提交前检查目录和注册
│   └── lib.sh                 上面几个脚本共用的函数，不要单独执行
├── migrations/
│   ├── imports.go             每条迁移一行引用，把它们编进二进制
│   ├── main/                  主库
│   └── obs/                   OBS 库
├── util/                      建表、加列、加索引这些可重复执行的函数
├── register/                  启动时收集迁移，并按版本排序
├── engine/                    决定执行、跳过还是失败，然后真正跑 Up
├── schema/                    每个库里的执行记录表、运行审计表
└── cli/                       init / up / status / list
```

`migrations/main` 和 `migrations/obs` 下面只有三种分组：`pending/`、`v1.9.3/` 这种三位版本、`v1.9.3.x/` 这种带第四段的版本。分组目录本身不是 Go 包，真正的代码在它下一层的迁移目录里。

二进制编进当前这棵树里的全部迁移。环境上怎么跑，只看这个库自己的记录，不看别的库，也不看代码里的版本号大小本身：

| 命令 | 干什么 |
| --- | --- |
| `init --mode=empty` | 只建记录表和审计表。新库用这个 |
| `init --mode=adopt --baseline main=v1.9.2` | 建表，并把该版本及更早的迁移记成已成功，不执行它们。已经在用的库接入时用，避免把老变更再跑一遍 |
| `up` | 执行还没成功、且版本不低于当前库版本的迁移 |
| `up --catch-up` | 把漏在当前版本下面的也补上。主线先上去、特性分支后合进来时用这个 |
| `up --plan` | 只打印会执行什么，不改库 |
| `status` | 看每个库跑到哪、哪些还没成功 |
| `list` | 按执行顺序列出编进二进制的迁移，不连数据库 |

`up` 遇到一条失败就停。这条在记录表里是 `failed`，错误摘要在 `message` 里（过长会截断，完整内容在 stderr）。前面已经成功的不会重跑；把这次失败改好再执行，会从这条接着做。进程在记下 `running` 之后被杀掉，也按没成功处理，下次重跑这一条。

同一个库里如果两条迁移用了同一个 ID，但「时间戳_名称」不同，程序当成 ID 被复制后没换，拒绝执行。修正办法是给新的那条换一个新 ID，不要去改记录表。

## 记录表和审计表

`init` 在每个库里各建这两张表。主库和 OBS 库各有一份，互不代替。查问题先分清库。

`hcm_migration_record` 回答「这条迁移在这个库上跑过没有」。一个迁移 ID 一行。

| 列 | 看什么 |
| --- | --- |
| `migration_id` | 迁移 ID。同一个 ID 不会有第二行 |
| `status` | `running` 正在跑或上次没写完，`success` 已成功，`failed` 失败了。后两种下次都会再跑这条 |
| `version` | 最近一次成功时，代码里注册的版本。定版、归档之后代码里的版本会变，和这里不一致只打警告，仍按已成功跳过 |
| `applied_pkg` | 实际执行它的目录，例如 `main/pending/20260927160000_add_bk_asset_id`。内外两份靠这个和 ID 认是不是同一条 |
| `message` | 失败摘要，最长约 1024 字节。里面带有 `Up` 返回的错误；用了 `util` 时，MySQL 的 `Error 编号` 也在这串字里。完整内容在 stderr |

`hcm_migration_audit` 回答「这一次命令在这个库上发生了什么」。一次 `up` 或 `init` 在每个库写一行。这些行的 `run_id` 相同，也是日志里的 `rid`。`list`、`status`、`up --plan` 不写这张表。审计写失败只打警告，不改变命令的退出码。

| 列 | 看什么 |
| --- | --- |
| `command`、`args` | 跑的是 `up` 还是 `init`，以及命令行参数。密码在配置文件里，不会出现在 `args` |
| `status`、`exit_code` | 这次命令的结果。`running` 且 `end_at` 为空，多半是进程被杀掉了 |
| `version_before`、`version_after` | 这次开始和结束时，这个库的当前版本 |
| `binary_version`、`git_hash` | 跑的是哪一个二进制 |
| `skipped` | 跳过了哪些迁移，以及原因 |
| `warnings` | 版本不一致这类警告 |
| `message` | 这次的校验问题和错误。过长会截断，全文在 stderr |

这次到底执行了哪些，审计表不逐条记。用记录表里 `updated_at` 落在该行 `start_at` 和 `end_at` 之间的那些行来对。
