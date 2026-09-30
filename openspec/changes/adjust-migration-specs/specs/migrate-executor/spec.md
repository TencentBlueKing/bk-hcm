## MODIFIED Requirements

### Requirement: 逐条执行判定

执行前校验通过后，执行器 SHALL 只照计划执行，MUST NOT 在循环里再做漏执行或疑似 ID 复用判定。计划中每条迁移的动作 SHALL 是下列之一：`EXECUTE`、`SKIP-SUCCESS`、`SKIP-BACKFILL`、`ABOVE-MAX-VERSION`、`MISSING`、`ID-REUSE`、`PENDING-DENIED`。无记录、`running`、`failed` MUST 同等视为无成功记录。默认模式下版本高于库当前的无成功记录 SHALL 执行；补跑模式下无成功记录 SHALL 执行。超过 `--to` 的迁移 MUST NOT 执行、MUST NOT 参与 ID 比较、也 MUST NOT 占有 ID。版本漂移 SHALL 由计划阶段统一产生 WARN 并跳过；记录为 `PENDING` 而注册已定版时走回填，不算漂移。`VersionAfter` SHALL 只在成功执行或回填得到更高的已定版版本时上升，`PENDING` MUST NOT 抬高它。

每个迁移 ID SHALL 有一个 owner：优先是 `success` 记录的 `applied_pkg`；否则是本次按执行顺序第一个计划执行（`EXECUTE`）或判为漏执行（`MISSING`）的副本。之后遇到同 ID 的迁移才是「按 ID 跳过」；此时比较 owner 包与当前包的迁移后缀，不同则 `ID-REUSE`，相同则 `SKIP-SUCCESS`，若 owner 记录版本为 `PENDING` 而当前已定版则 `SKIP-BACKFILL`。回填后同 ID 后续副本若版本不同 SHALL 给出版本漂移 WARN。只在真正发生「按 ID 跳过」时检查后缀，因此注册表是记录表的超集、记录表是注册表已执行子集，两种来源用同一规则。

#### Scenario: 已成功且后缀相同则跳过

- **WHEN** 某迁移在记录表中已是 `success`，且迁移后缀与 `applied_pkg` 相同
- **THEN** 计划为 `SKIP-SUCCESS`，MUST NOT 重新执行

#### Scenario: 相同 ID 且后缀相同的另一份文件跳过

- **WHEN** 同一个库的注册表里有两条相同 Migration ID、相同迁移后缀的迁移，排序靠前的那条计划为执行或漏执行
- **THEN** 后一条按 ID 跳过（`SKIP-SUCCESS` 或 `SKIP-BACKFILL`），MUST NOT 调用它的 `up()`

#### Scenario: 默认模式执行更新的迁移

- **WHEN** 库当前为 `v1.9.3`，某无记录的迁移版本为 `v1.9.4`，且执行前校验通过
- **THEN** 计划为 `EXECUTE`，并调用它的 `up()`

#### Scenario: 补跑模式执行低版本迁移

- **WHEN** 库当前为 `v1.9.4`，某无记录的迁移版本为 `v1.9.3`，且开启 `--catch-up`，执行前校验通过
- **THEN** 计划为 `EXECUTE`，并调用它的 `up()`

#### Scenario: 中断的迁移重跑

- **WHEN** 某迁移记录为 `running`，且执行前校验通过
- **THEN** 按无成功记录进入计划，在允许的模式下重新执行

#### Scenario: 归档文件与新文件同一轮跑完

- **WHEN** 本次注册表既含版本低于库当前的归档文件，也含高于库当前的新文件，且开启 `--catch-up`，执行前校验通过
- **THEN** 同一轮从旧到新的遍历中，归档文件与新文件依次执行完毕

#### Scenario: 超过上限不占 ID

- **WHEN** 某迁移版本高于 `--to`，且同 ID 另有一份在上限内
- **THEN** 超限那份动作为 `ABOVE-MAX-VERSION`，MUST NOT 成为该 ID 的 owner，也不参与后缀比较

### Requirement: plan

`--plan` SHALL 使用与实跑相同的执行前校验和同一份计划，打印每条迁移的判定结果（`EXECUTE` / `SKIP-SUCCESS` / `SKIP-BACKFILL` / `ABOVE-MAX-VERSION` / `MISSING` / `ID-REUSE` / `PENDING-DENIED`）。MUST NOT 写记录表，MUST NOT 写运行审计表，MUST NOT 调用 `up()`。默认模式、补跑模式、版本上限与 `--allow-pending` 在 `--plan` 中同样生效。校验失败时 `--plan` 的退出码 MUST 与实跑相同。

#### Scenario: plan 不改库

- **WHEN** 执行 `up --plan`
- **THEN** 打印判定清单，记录表、运行审计表与业务表结构均不变

#### Scenario: plan 反映模式差异

- **WHEN** 同一批迁移分别执行 `up --plan` 与 `up --catch-up --plan`
- **THEN** 前者把低于库当前的无记录迁移标为 `MISSING`，后者标为 `EXECUTE`

#### Scenario: plan 预检漏执行

- **WHEN** 默认模式下存在漏执行，执行 `up --plan`
- **THEN** 退出码为 4，且清单中列出全部漏执行项

#### Scenario: plan 预检疑似 ID 复用

- **WHEN** 存在迁移后缀不一致的同 ID，执行 `up --plan`
- **THEN** 退出码为 6，且输出含该 ID、owner 包与当前 `Pkg`

## ADDED Requirements

### Requirement: 执行前校验

每个库 SHALL 在连库之后、经 `CheckInitialized`（两张表都存在才算已初始化）、`Load` 记录表、算出库当前版本之后做执行前校验，再产出计划（`Prepare` / `BuildPlan`）。库当前版本本次运行 MUST 只算这一次。校验 SHALL 一次列出全部问题。任一项失败时，该库一条迁移都 MUST NOT 执行。多个库时 MUST 先对所有选中的库完成 `Prepare` 与检查（`CollectPlanErrors`），全部通过后再开始任何一个库的执行。只看注册表的检查 MUST 放在这一步，MUST NOT 单独在连库前做。

`hcm_migration_record` 或 `hcm_migration_audit` 任一不存在，或记录内容非法（未知 `status`、版本不可解析、`success` 行 `applied_pkg` 为空），SHALL 在 Load / Prepare 阶段以退出码 3 返回，发生在生成计划之前，MUST NOT 与退出码 4 / 5 / 6 同时出现。

计划阶段的校验项与退出码 SHALL 为：未开 `--allow-pending` 却注册了 `PENDING`（不论是否指定 `--to`），或注册表出现两个及以上不同的非空版本 label，退出码 5；默认模式下已定版、无 success、版本小于等于库当前，退出码 4；按 ID 跳过时迁移后缀与 owner 不同，退出码 6。多项同时失败时，优先级 MUST 为 6 > 5 > 4（只返回最高优先级对应的退出码，但错误信息列出所有库的全部问题）。补跑模式 MUST 关掉漏执行这道闸，MUST NOT 关掉疑似 ID 复用与 label 检查。

失败输出 MUST 含完整 ID。疑似 ID 复用的输出还 MUST 含 owner 的版本（若来自记录）、owner 包和当前 `Pkg`。label 问题 MUST 列出每个 label 及其迁移 ID（排序）。跳过日志 MUST 打出两边的 pkg。

#### Scenario: 漏执行一次列全且不执行

- **WHEN** 默认模式下有两条已定版迁移无 success 且版本小于等于库当前
- **THEN** 退出码为 4，输出列出这两条的 ID，两条的 `up()` 都不被调用

#### Scenario: 开关关闭时拒绝 PENDING

- **WHEN** 未传 `--allow-pending`，注册表中有版本为 `PENDING` 的迁移，执行 `up`
- **THEN** 退出码为 5，输出列出这些 ID，一条迁移都不执行

#### Scenario: 两个非空 label

- **WHEN** 注册表中同时存在 `v1.9.3-tenant.1` 与 `v1.9.3-woa.1`
- **THEN** 退出码为 5，输出列出每个 label 及其迁移 ID，一条迁移都不执行

#### Scenario: 主库已通过时 OBS 的校验仍先于执行

- **WHEN** 主库校验通过，OBS 库检出漏执行
- **THEN** 主库的迁移也不执行，命令以退出码 4 结束

#### Scenario: 漏执行与 PENDING 同时存在

- **WHEN** 默认模式、未开 `--allow-pending`，同时存在漏执行和 `PENDING`
- **THEN** 两类问题都被列出，退出码为 5，一条迁移都不执行

#### Scenario: ID 复用优先于漏执行

- **WHEN** 同一次运行既有漏执行又有疑似 ID 复用
- **THEN** 两类问题都被列出，退出码为 6

#### Scenario: 缺表在计划之前失败

- **WHEN** 记录表或审计表任一不存在，执行 `up`
- **THEN** 退出码为 3，提示先执行 `init`，MUST NOT 产出可执行计划

### Requirement: 疑似 ID 复用

系统 SHALL 把「按 ID 跳过」时迁移后缀不同视为疑似 ID 复用（退出码 6）。迁移后缀是迁移目录名里从 14 位时间戳起到结尾的部分。版本前缀不同或分组目录不同 MUST NOT 判为复用。内部执行过、外部三位再次注册，以及 pending 执行过再定版，迁移后缀相同，SHALL 按 ID 跳过。复制后忘了换 ID 导致后缀不同时，MUST 以退出码 6 失败。修正方式 SHALL 是给复制出来的那条迁移换一个新 ID。MUST NOT 靠改记录表的 `applied_pkg` 来通过检查。

目录名和 ID 都没改、只改了正文时，后缀相同，系统 MUST NOT 因此失败。

#### Scenario: 内部与外部后缀相同则跳过

- **WHEN** 记录的 `applied_pkg` 为 `main/v1.9.3.x/v1.9.3.1_20260905160000_add_bk_asset_id`，当前 `Pkg` 为 `main/v1.9.3/v1.9.3_20260905160000_add_bk_asset_id`，且已有 success
- **THEN** 跳过该迁移，日志打出这两个 pkg

#### Scenario: 复制后 ID 未改则失败

- **WHEN** 记录的 `applied_pkg` 为 `main/v1.9.3/v1.9.3_20260905160000_add_bk_asset_id`，当前 `Pkg` 为 `main/v1.9.4/v1.9.4_20260925103000_add_host_index`，两者 ID 相同且已有 success
- **THEN** 退出码为 6，不调用当前这份的 `up()`

#### Scenario: 注册表内后缀不同则失败

- **WHEN** 同一注册表内同一 ID 出现两次，迁移后缀不同，且记录表中还没有该 ID
- **THEN** 退出码为 6，靠前的副本成为 owner，后一条为 `ID-REUSE`，两条都不执行

### Requirement: 执行循环

通过 `CollectPlanErrors` 的计划 SHALL 逐条处理：`EXECUTE` 为 `MarkRunning` → `Up` → `MarkSuccess` / `MarkFailed`；`SKIP-BACKFILL` 只回填 `version`；其余可执行的跳过动作只记结果。MUST NOT 有外层事务，MUST NOT 自动重试，遇错即停。`Up` 的 panic SHALL 转成错误并记 `failed`，进程以退出码 1 结束。执行中记录表读写失败、`MarkSuccess` 影响 0 行、连接断开或 context 取消，同样以退出码 1 结束。`running` / `failed` 记录的迁移下次重跑（断点续跑）。

#### Scenario: 遇错即停

- **WHEN** 某条 `EXECUTE` 的 `up()` 返回错误
- **THEN** 该库后续迁移不再执行，记录为 `failed`，退出码为 1

#### Scenario: panic 记为失败

- **WHEN** 某条迁移的 `up()` panic
- **THEN** 记录为 `failed`，退出码为 1，MUST NOT 停在 `running`

### Requirement: PENDING 的执行

`--allow-pending` 打开时，无 success 的 `PENDING` SHALL 在已定版之后执行，彼此按时间戳排序。已有 success 的 SHALL 跳过并按记录表规则回填。`PENDING` MUST NOT 做漏执行判断。给出版本上限时 `PENDING` MUST NOT 执行。`init --mode=adopt` MUST NOT 把 `PENDING` 写入基线。已执行过的 `PENDING` 其 `up()` 被修改后，系统 MUST NOT 自动重跑。

#### Scenario: 开关打开时执行未定版

- **WHEN** 执行 `up --allow-pending`，某 `PENDING` 迁移无 success，且没有版本上限
- **THEN** 计划为执行，成功后记录的 `version` 为 `PENDING`

#### Scenario: 版本上限排除未定版

- **WHEN** 执行 `up --allow-pending --to v1.9.4`，注册表中有 `PENDING`
- **THEN** 该 `PENDING` 迁移不执行

#### Scenario: 基线不包含未定版

- **WHEN** 执行 `init --mode=adopt --baseline main=v1.9.2`，注册表中有版本不超过基线的迁移，也有 `PENDING`
- **THEN** 基线以内的已定版被写成 success，`PENDING` 没有记录行
