## Why

HCM 的 DB 变更目前靠人在 DB 平台手工跑 `scripts/sql` / `scripts/obssql`，`hcm_version` 只记一个最终版本号：无法区分「没跑」和「跑到一半失败」，历史 SQL 大量非幂等（重复加列直接报错），内外部两条源头合入时同一变更会出现两个语义相近的文件，空库重建跑不通。

`migrate/util/`（MetaOrm + 幂等 DDL/DML）已经落地，但没有任何东西去调用它：没有注册表、没有版本比较、没有执行记录表、没有可执行入口。本变更规划 `hcm-migrate` 的**执行层**，让迁移文件能被注册、排序、判定、执行并逐条留痕，最终由容器 Job 在业务拉起前自动跑完。

## What Changes

- 新增独立二进制 `hcm-migrate`（入口 `migrate/migrate.go`），子命令 `init` / `up` / `status` / `list`。`--database` 是库名列表（不传表示已启用的全部库），`init` 支持 `--mode` 与 `--baseline`，`up` 支持 `--to` 与 `--catch-up`，两者都支持 `--plan`。
- 新增 `migrate/register/`：`Registry` 类型 + `Main` / `Obs` 两个实例，`Regist` 在 `init()` 阶段做强校验（UUID、版本语法、14 位时间戳、常量 `PENDING`），不合法直接终止进程。目录名仍是小写 `pending/`。
- 新增 `migrate/register/version.go`：`vX.Y.Z` / `vX.Y.Z.N` / `vX.Y.Z-label.N` 三种写法的解析、`Version` / `Suffix` 结构与全序比较器，以及版本上限比较语义。
- 新增 `migrate/engine/`：`recordstore.go`（`hcm_migration_record` 读写）、`schema.go`（记录表 DDL 与建表，只被 `init` 调用）、`executor.go`（按库循环 + 逐条判定 + 默认/补跑模式 + `--plan`；相同 Migration ID 执行时跳过）、`datasource.go`（按 `cc` 配置建连接，按库名列表选择本次作用的库）。
- 新增 `migrate/migrations/` 目录骨架与 `imports.go` 生成/校验机制，替代 bk-cmdb 那种 125 行手工空导入。
- 新增执行记录表 `hcm_migration_record`，**每库一份**，`migration_id` 唯一，逐条记 `running` / `success` / `failed`。
- 新增 `migrate/Makefile` 目标与 `cmd/Makefile` 挂接，产出独立迁移镜像供 Helm hook Job 使用。
- 沿用技术方案的 `init --mode` 与 `init && up` 命令链。取值从方案的 `full` / `existing` 改为 `empty` / `adopt`，并收紧语义：`--mode` 必填无默认、`--baseline` 按库给、`adopt` 缺基线即失败、`up` 在缺记录表时失败而不自举建表。理由见 `design.md` 的 D2 与「偏离技术方案之处」。

本变更**只规划、不实现 Go 代码**；`migrate/util/**` 已完成，不在本次改动范围内。

## Capabilities

### New Capabilities

- `migrate-version`: 迁移版本号的语法、解析、`Version` / `Suffix` 结构、全序比较规则、常量 `PENDING` 与版本上限比较语义。
- `migrate-registry`: 进程内注册表 `Registry`（`Main` / `Obs`）、`Regist` 签名与 `init()` 期强校验、相同 Migration ID 允许注册、`imports.go` 漏导入的防护。
- `migrate-record`: 执行记录表 `hcm_migration_record` 的 DDL、每库放置、建表时机（只由 `init` 建）、写入时机、库当前版本的计算方式与版本漂移告警。
- `migrate-datasource`: 从 `cc` 配置构造主库 / OBS 两份连接、裸 `Do()` 约束、OBS 缺省跳过、库间失败隔离与执行顺序、DB 就绪等待。
- `migrate-executor`: 逐条执行判定循环、默认 / 补跑模式、版本上限、相同 ID 执行时跳过、事务策略、无分布式锁与无自动重试、断点续跑、`--plan` 输出。
- `migrate-cli`: `hcm-migrate` 的子命令、参数（长短名 / 类型 / 默认值）、`init` 的幂等与 mode 语义、`up` 不建表、退出码、部署形态、stdout 与日志分工。

### Modified Capabilities

无。`migrate/util/` 的幂等能力尚未建 spec，本次不追补，也不修改其行为。

## Impact

- 新增目录：`migrate/register/`、`migrate/engine/`、`migrate/migrations/`、`migrate/scripts/`、`migrate/migrate.go`、`migrate/Makefile`。
- 复用不改动：`pkg/dal/dao`（`NewDaoSet`）、`pkg/dal/dao/orm`（`Interface` / `DoOrm`，只用裸 `Do()`）、`pkg/cc`（`DataBase` / `OBSDatabase`）、`pkg/logs`、`pkg/criteria/errf`、`pkg/kit`。
- 复用不改动：`migrate/util/`（MetaOrm + 幂等 DDL/DML），执行层只把 `orm.Interface` 递给迁移函数。
- 构建链：`cmd/Makefile` 与根 `Makefile` 的 docker 目标需要多一个迁移镜像；Helm chart 需要一个带三个 hook 注解的 Job。
- 数据库：主库与 OBS 库各新增一张 `hcm_migration_record` 表。
- 不影响任何在跑的微服务：`hcm-migrate` 是独立进程，业务 Pod 里没有这份镜像。
