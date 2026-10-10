## ADDED Requirements

> 本文件中退出码、执行前校验、ID 复用、计划动作以 `openspec/changes/adjust-migration-specs/specs/migrate-executor/spec.md` 为准（退出码 4=漏执行、5=PENDING/多 label、6=ID 复用；计划动作含 `SKIP-BACKFILL` / `ID-REUSE` / `PENDING-DENIED`）。以下保留原稿结构，已过时处已就地改正。

### Requirement: 遍历顺序

执行器 SHALL 遍历注册表中的全量迁移，顺序为版本号从旧到新、同版本比时间戳、再比 Migration ID。补跑 MUST NOT 另行排序，也 MUST NOT 另切一批文件。执行器 MUST NOT 按版本线或标签筛选候选，编译进二进制的迁移一律进入判定。

#### Scenario: 固定顺序遍历

- **WHEN** 注册表含 `v1.9.4`、`v1.9.3`、`v1.9.3.1`
- **THEN** 判定顺序为 `v1.9.3`、`v1.9.3.1`、`v1.9.4`

#### Scenario: 带标签的文件同样进入判定

- **WHEN** 注册表中含 `v1.9.3-tenant.1`
- **THEN** 该迁移进入判定循环，MUST NOT 因带标签被跳过

#### Scenario: 补跑不改变顺序

- **WHEN** 同一批迁移分别以默认模式和补跑模式判定
- **THEN** 两次遍历的顺序完全相同

### Requirement: 逐条执行判定

对每条迁移，执行器 SHALL 先产出计划再照计划执行：已有 `success` 或本次已有同 ID owner 则按 ID 跳过（后缀相同为 `SKIP-SUCCESS` / `SKIP-BACKFILL`，不同为 `ID-REUSE` 退出码 6）；无成功记录时，默认模式下 `version <= 库当前` 判为漏执行并以退出码 4 失败，`version > 库当前` 则执行；补跑模式下一律执行。超过 `--to` 的为 `ABOVE-MAX-VERSION`，不占 ID。无记录、`running`、`failed` 三种情形 MUST 同等视为「无成功记录」。

#### Scenario: 已成功的按 ID 跳过

- **WHEN** 某迁移在记录表中已是 `success`，且迁移后缀与 `applied_pkg` 相同
- **THEN** 跳过该迁移，MUST NOT 重新执行

#### Scenario: 相同 ID 的另一份文件跳过

- **WHEN** 同一个库的注册表里有两条相同 Migration ID、相同迁移后缀的迁移，排序靠前的那条执行成功
- **THEN** 后一条按 ID 跳过，MUST NOT 调用它的 `up()`

#### Scenario: 默认模式执行更新的迁移

- **WHEN** 库当前为 `v1.9.3`，某无记录的迁移版本为 `v1.9.4`
- **THEN** 执行该迁移

#### Scenario: 默认模式检出漏执行

- **WHEN** 库当前为 `v1.9.4`，某无记录的迁移版本为 `v1.9.3`
- **THEN** 命令以退出码 4 失败，MUST NOT 静默跳过

#### Scenario: 补跑模式执行低版本迁移

- **WHEN** 库当前为 `v1.9.4`，某无记录的迁移版本为 `v1.9.3`，且开启 `--catch-up`
- **THEN** 执行该迁移

#### Scenario: 中断的迁移重跑

- **WHEN** 某迁移记录为 `running`（上次执行中途中断）
- **THEN** 按「无成功记录」判定，在允许的模式下重新执行

#### Scenario: 归档文件与新文件同一轮跑完

- **WHEN** 本次注册表既含版本低于库当前的归档文件，也含高于库当前的新文件，且开启 `--catch-up`
- **THEN** 同一轮从旧到新的遍历中，归档文件与新文件依次执行完毕

### Requirement: 事务策略

执行器 MUST NOT 在 `up()` 外面包裹事务，事务策略由每条迁移自行决定。DDL SHALL 不使用事务；大表 DML SHALL 由迁移作者自行分批提交小事务。记录表写入 MUST 在迁移事务之外。

#### Scenario: 执行器不包裹事务

- **WHEN** 某迁移内部同时包含 DDL 与 DML
- **THEN** 执行器不为其开启外层事务，由迁移自己控制提交边界

#### Scenario: 迁移事务回滚不影响记录写入

- **WHEN** 某迁移在自己的事务中失败并回滚
- **THEN** 该迁移的 `failed` 记录仍被写入

### Requirement: 并发与重试

执行器 MUST NOT 使用分布式锁，并发安全由部署形态保证：Job 的 `parallelism` 为 1、`backoffLimit` 为 0、`restartPolicy` 为 `Never`，二进制部署为单进程。执行器 MUST NOT 自动重试，失败即中断且后续迁移不执行。

#### Scenario: 失败即中断

- **WHEN** 某条迁移执行失败
- **THEN** 该库后续迁移不再执行，命令以非 0 退出

#### Scenario: 不自动重试

- **WHEN** 某条迁移执行失败
- **THEN** 执行器 MUST NOT 重新调用该迁移的 `up()`

### Requirement: 断点续跑

执行失败后库当前版本 SHALL 仍停在已成功记录的最大版本。重新执行时已 `success` 的按 ID 跳过，等效于从断点继续。

#### Scenario: 重跑从断点继续

- **WHEN** 上次执行在第三条迁移失败，修复后重新执行
- **THEN** 前两条按 ID 跳过，从第三条开始执行

#### Scenario: 失败不推高库当前版本

- **WHEN** 某条高版本迁移执行失败
- **THEN** 库当前版本不变，仍为已成功记录中的最大版本

### Requirement: plan

`--plan` SHALL 打印每条迁移的判定结果（`EXECUTE` / `SKIP-SUCCESS` / `SKIP-BACKFILL` / `ABOVE-MAX-VERSION` / `MISSING` / `ID-REUSE` / `PENDING-DENIED`），MUST NOT 写记录表、MUST NOT 调用 `up()`。默认 / 补跑模式与版本上限在 `--plan` 中同样生效。校验失败时 `--plan` 的退出码 MUST 与实跑相同。

#### Scenario: plan 不改库

- **WHEN** 执行 `up --plan`
- **THEN** 打印判定清单，记录表内容与业务表结构均不变

#### Scenario: plan 反映模式差异

- **WHEN** 同一批迁移分别执行 `up --plan` 与 `up --catch-up --plan`
- **THEN** 前者把低于库当前的无记录迁移标为 `MISSING`，后者标为 `EXECUTE`

#### Scenario: plan 预检漏执行

- **WHEN** 默认模式下存在漏执行，执行 `up --plan`
- **THEN** 退出码为 4，供流水线前置判断是否需要开补跑
