## 1. 注册与版本

- [x] 1.1 按 `specs/migrate-registry/spec.md`：ID 改为可读格式；`Regist(id, version, timestamp, Migrator)`；`PkgPath`（指针先 Elem）写入 `Pkg`；去掉 description；导出 `NewMigration`；`migrate.PkgSuffix`；`constant.MigrationPkgMaxLen` 255；拒绝 nil / typed-nil；常量与格式校验分别落在 `pkg/criteria/constant` 与 `pkg/migrate`
- [x] 1.2 从 `migrate-registry` spec 删除「相同 ID 的动作名一致」整节（delta 已 REMOVED；代码未实现该比对）
- [x] 1.3 按 `specs/migrate-version/spec.md`：`Load` 放行 `PENDING`；`CurrentVersion` 跳过 `PENDING`
- [ ] 1.4 按 `specs/migrate-registry/spec.md` 改 `openspec/changes/migrate-exec-layer/specs/migrate-registry/spec.md`，并同步 `check-migrate`、一目录一包叙述

## 2. 记录表

- [x] 2.1 按 `specs/migrate-record/spec.md`：`applied_pkg` 列；`MarkRunning`（insert / duplicate update）与 adopt 基线写入；`MarkSuccess` / `MarkFailed` 不写；`success` 行空 `applied_pkg` → `ErrPrecondition`（退出码 3）
- [x] 2.2 `migrate/schema/schema.go`：`InitTables`（两表 DDL；先审计表后记录表；仅新建记录表时 adopt；基线失败只删记录表；状态与跳过原因枚举在 `pkg/criteria/enumor`）；`recordstore.go` 负责记录表读写
- [x] 2.3 PENDING 定版回填：只更新 `version`，不改 `applied_pkg`，影响 0 行视为正常（`RecordStore.BackfillPendingVersion`；何时调用由 3.1 执行器决定）
- [ ] 2.4 按 delta 改 `openspec/changes/migrate-exec-layer/specs/migrate-record/spec.md`

## 3. 执行与命令

- [x] 3.1 按 `specs/migrate-executor/spec.md`：执行前校验、疑似 ID 复用（统一 owner / `migrate.PkgSuffix`，退出码 6）、label 与 `--allow-pending`（退出码 5）、漏执行（退出码 4）；`Prepare` / `BuildPlan` / `CollectPlanErrors` / `Execute`；循环只照计划执行
- [x] 3.2 按 `specs/migrate-cli/spec.md`：`--allow-pending`；`init` 调用 `InitTables`；CLI 接线 `NewAudit().Begin` / `End`；退出码 0–6 映射（哨兵 `ErrUsage`/`ErrPrecondition`/`ErrMissed`/`ErrRegistry`/`ErrIDReuse`）
- [x] 3.3 按 delta 改 `openspec/changes/migrate-exec-layer/specs/migrate-executor/spec.md` 与 `migrate-cli/spec.md`（退出码 / ID 复用 / label / PENDING 与本变更对齐）

## 4. 运行审计与旧叙述

- [x] 4.1 `migrate/schema/auditstore.go`：`NewAudit().Begin`、`(*Audit).End`；最新仍为 `running` 的行进 warnings；空 rid / 插入失败只警告并空操作；JSON 空列表写 `[]`；warnings/message 200 项与 1024 字节上限
- [x] 4.2 确认 `specs/migrate-audit/spec.md` 仍只放在本变更下，不复制进 `migrate-exec-layer`
- [ ] 4.3 在 `openspec/changes/migrate-exec-layer/proposal.md` 与 `design.md` 里，把仍写着 UUID、连库前一律拒绝 `PENDING`、动作名必须一致的段落改成指向本变更，避免和 spec 矛盾
