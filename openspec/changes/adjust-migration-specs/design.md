## Context

`migrate-exec-layer` 的 spec 已经写完，代码已落地 `util/`、`register/`、`engine/` 的 `InitTables` / 记录表 / 审计读写（`schema.go`、`recordstore.go`、`auditstore.go`），执行器与 CLI 接线还没写。方案调整记在 `.cursor/docs/iwiki/4039435213-hcm-migration/adjustments-2026-09-23.md`，这些能力尚未同步到 `openspec/specs/`，现行需求就在那个变更的 `specs/` 里。

本变更写 delta spec，并把已实现与未实现对齐到同一份文档。

## Goals / Non-Goals

**Goals:**

- 让 `migrate-registry`、`migrate-record`、`migrate-version`、`migrate-executor`、`migrate-cli` 的需求与方案调整一致。
- 新增 `migrate-audit`，把一次运行的排查信息落在库里。
- 删掉「相同 ID 的动作名一致」和「目录 tag 对 ID tag」这两条没有采用的做法。

**Non-Goals:**

- 不在本变更里写执行器、CLI、`--allow-pending` 接线（见实现进度）。
- 不改 `migrate/util/` 的幂等工具。
- 不引入「迁移版本不能高于发布版本」这道闸。方案里没有定。
- 不规定跳过和疑似 ID 复用的具体排版，只规定必须出现的字段。

## Decisions

### 1. ID 用可读格式，不再用 UUID

格式 `20260923-1945-ADD-BK-ASSET-ID-A3F9`，全大写，总长不超过 64。tag 在 ID 里，`Regist` 不再收 description。14 位时间戳仍是单独参数，ID 里的分钟不代替它。排序仍是版本、时间戳、ID。

目录名保持现有规范，不改成这串 ID。

### 2. 用导入路径的迁移后缀发现粘错的 ID

包的身份是导入路径，`package` 子句统一写 `migration`。`Regist` 对传入值做 `reflect`，取 `PkgPath()`，去掉 `hcm/migrate/migrations/` 前缀后存进 `Pkg`。去掉前缀后长度不超过 `constant.MigrationPkgMaxLen`（255）。不用 `runtime.Caller`：它的路径形态随 `-trimpath` 变化。`NewMigration` 导出给测试与包外工具；进注册表仍只走 `Regist`。迁移后缀用 `hcm/pkg/migrate` 的 `PkgSuffix`。

比对只看迁移后缀，即迁移目录名里从 14 位时间戳起到结尾。版本前缀和分组目录（`v1.9.3/`、`v1.9.3.x/`、`pending/`）不参与。内部提到外部、pending 定版、归档都只改版本前缀，后缀相同，按 ID 跳过。

对不上以退出码 4 失败，和漏执行列在同一份报告里，一条都不执行。修正方式是给复制出来的那条换新 ID，不改记录表来凑后缀。

同一注册表里同 ID、后缀不同，也在执行前校验失败，不在单次 `Regist` 里做，因为另一条可能还没注册。后缀相同只警告。

曾考虑拿目录 tag 对 ID 的 tag 段，以及另注册一个动作名。两者都是手写字符串，复制注册代码时会一起被带走，不采用。

### 3. 执行前校验在 Load 和库当前版本之后

每个库：连库、Load、算一次库当前版本、校验、产出计划、循环只照计划执行。只看注册表的检查也放在这一步，不单独在连库前做。多个库先全部校验完再执行。`--plan` 用同一套判定，不写库，退出码与实跑一致。

多项同时失败时，有退出码 4 就用 4，否则用 3。

### 4. PENDING 用开关，不在注册期拒绝

`--allow-pending` 默认关，由 Helm values 按环境传入。关着时执行前校验以退出码 3 失败。开着时执行一次，`version` 写 `PENDING`，不计入库当前版本。定版后按 ID 跳过，并把记录里的 `PENDING` 回填成此刻注册的版本。改了已执行的 `up()` 不自动重跑，由开发者恢复结构并删记录。

`Parse("PENDING")` 仍然报错，`IsPending` 单独判定。`Load` 放行 `PENDING`，与开关无关。给了版本上限时不执行 PENDING。adopt 不写入 PENDING。

### 5. 运行审计与执行层分开

每个库一张 `hcm_migration_audit`。一次进程运行在每个库一行，`run_id` 用 `uuid.UUID()`，也就是这次进程的 `Rid`，与 agent 的 `runId` 同一套。

`InitTables` 一次处理两张表：先建缺失的审计表，再建缺失的记录表；仅当本次新建了记录表且带基线时写入 adopt；基线插入失败只删记录表。两张表都在则无改动；只缺一张则只建那张。`NewAudit().Begin` 在连库后插入 `running`，结束时 `End` 更新一次。执行层不 import、不调用审计器；由 CLI 接线。交给审计器的结果里带两个版本变量：开始时赋成库当前版本；每条 `MarkSuccess` 之后，版本更高且不是 `PENDING` 就更新 `version_after`。出错也写入当时的值。

`NewAudit().Begin` 按 `id` 降序看最新一条仍为 `running` 的行，把那次 `run_id` 放进本次 warnings。`Rid` 为空或插入失败只打警告并返回空操作。

`skipped` 逐条记录。`message` 合并校验问题和错误，`warnings` 只放警告，都只追加。不记 `executed`、`backfill`、`runner`。本次执行了哪些，用记录表 `updated_at` 落在本次时间窗内的行来对。

审计器读写失败只打警告，不改变退出码。进程被杀时当次的 `skipped`、`warnings`、`message` 会丢，接受。`list`、`status`、`up --plan` 不写这张表。

`binary_version` 与 `git_hash` 读 `hcm/pkg/version` 的变量，由编译时 ldflags 注入，不读配置。

## Risks / Trade-offs

- [两处 spec 同时存在] 本变更是 delta，`migrate-exec-layer` 里仍是旧全文。→ 实现任务把 delta 合进那份全文，读 spec 的人只看合入后的那一份。
- [粘错 ID 但目录名和 ID 都没改、只改了正文] 后缀相同，拦不住。→ 靠 Code Review，spec 不把正文比对加回来。
- [运行审计插入失败] 当次排查信息缺失，迁移仍继续。→ 这是故意的：审计器不能反过来改变部署结果。
- [进程被杀] `running` 行没有 `skipped` 和 `message`。→ 接受。下次运行在自己的 `warnings` 里记下那个 `run_id`。

## Migration Plan

1. 按本变更的 delta 改 `openspec/changes/migrate-exec-layer/specs/` 里对应需求的全文。
2. `migrate-audit` 只放在本变更的 `specs/migrate-audit/spec.md`，不塞进 `migrate-exec-layer`。
3. 不发布、不改库。回退就是还原那些 spec 文件。

## Open Questions

- 同 ID 跳过和疑似 ID 复用的具体排版。字段已经定了：完整 ID、已成功记录的版本、记录里的 `applied_pkg`、当前 `Pkg`。排版实现时与执行器其它输出对齐。
