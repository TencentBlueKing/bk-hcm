## Why

`openspec/changes/migrate-exec-layer` 里的 spec 仍按 UUID、执行前一律拒绝 `PENDING`、按 ID 静默跳过来写。2026-09-23 起的方案调整已经改了身份、目录、未定版和漏跑判定，2026-09-25 又确定用 `applied_pkg` 发现粘错的 ID，并增加运行审计表。spec 不跟着改，后面实现会对着过期规则写。

## What Changes

- **BREAKING** Migration ID 从 UUID 改为全大写的 `日期-分钟-tag-4 位随机`。`Regist` 不再单独收语义描述。
- 每条迁移一个目录、一个包、一处 `Regist`。`package` 子句统一为 `migration`。`imports.go` 每条迁移一行。`check-migrate` 增加对同 ID、不同迁移后缀的失败项。
- 删除「相同 ID 的动作名一致」这条待确认需求。不拿目录 tag 去对 ID 的 tag 段。
- `Regist` 从传入值的类型取导入路径，写入 `Pkg`。记录表增加 `applied_pkg`。按 ID 跳过时比迁移后缀，对不上以退出码 6 失败。
- `PENDING` 改为全局开关 `--allow-pending`（默认关）。打开时执行一次，记录写 `PENDING`，不计入库当前版本，定版后回填。
- 执行判定从 for 循环里挪到执行前校验：读完记录表、算出库当前版本后一次列全问题，一条都不执行。多个库先全部校验完再执行。
- 新增能力 `migrate-audit`：每库一张 `hcm_migration_audit`。一次进程运行在每个库一行，开头插入、结尾更新。执行层不调用审计器；由 CLI 调用。

本变更以 delta spec 为准。已落地的 `migrate/register/`、`migrate/engine/` 部分实现见 `adjustments-2026-09-23.md` 文末实现进度。

## Capabilities

### New Capabilities

- `migrate-audit`: 运行审计表 `hcm_migration_audit` 的粒度、字段、写入时机，以及与执行层的边界。

### Modified Capabilities

这些能力的现行需求写在 `openspec/changes/migrate-exec-layer/specs/`，尚未同步到 `openspec/specs/`。本变更的 delta 针对那份需求。

- `migrate-registry`: ID 格式与 `Regist` 签名、导入路径校验、一目录一包与 `check-migrate`、删除动作名需求。
- `migrate-record`: `applied_pkg` 列、写入时机、`PENDING` 不计入库当前版本、定版回填、空 `applied_pkg` 的失败。
- `migrate-version`: `PENDING` 不再一律在执行前失败，改为受 `--allow-pending` 控制。
- `migrate-executor`: 执行前校验、疑似 ID 复用（退出码 6）、跳过时比迁移后缀、`PENDING` 的执行与回填。
- `migrate-cli`: `--allow-pending`、退出码 0–6、`--plan` 与实跑使用同一套判定。

## Impact

- 不改动 Access、Service、Resource 层任何在跑服务。影响范围是独立二进制 `hcm-migrate` 的 spec。
- 要改的 spec：`openspec/changes/migrate-exec-layer/specs/` 下的 `migrate-registry`、`migrate-record`、`migrate-version`、`migrate-executor`、`migrate-cli`。本变更新增 `specs/migrate-audit/spec.md`。
- 方案来源：`.cursor/docs/iwiki/4039435213-hcm-migration/adjustments-2026-09-23.md`。
- 后续实现会改 `migrate/register`、`migrate/engine`，并在每个被迁移的库增加 `hcm_migration_audit`。记录表 `applied_pkg` 随尚未合入的建表语句一起加，不另写 ALTER。
