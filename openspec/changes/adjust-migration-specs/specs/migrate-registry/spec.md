## MODIFIED Requirements

### Requirement: 注册参数与 init 期强校验

`Regist` SHALL 接收 Migration ID、版本、14 位时间戳，以及实现了 `Migrator`（`Up(ctx, o) error`）的值。MUST NOT 再接收语义描述。校验 SHALL 在 `init()` 阶段完成。ID MUST 匹配 `^\d{8}-\d{4}-[A-Z][A-Z0-9]*(-[A-Z0-9]+)*-[0-9A-F]{4}$`，全大写，日期与 `HHmm` 能解析成真实时间，总长不超过 `constant.MigrationIDMaxLen`（64），tag 不超过 45 个字符。版本 MUST 是合法版本号或 `constant.MigrationPendingVersion`（值为 `PENDING`），长度不超过 `constant.MigrationVersionMaxLen`（64）。时间戳 MUST 是 14 位数字且能按 `yyyyMMddHHmmss` 解析成真实时间。

`Regist` SHALL 用 `reflect` 取该值类型的 `PkgPath()`（指针类型先 `Elem`），去掉前缀 `hcm/migrate/migrations/`（`constant.MigrationPkgPrefix`）后写入 `Pkg`。MUST NOT 使用 `runtime.Caller`。路径 MUST 恰好三段：库、分组、迁移目录。库名 MUST 等于当前注册表。迁移目录 MUST 能拆出可选的版本前缀，以及从 14 位时间戳起到结尾的迁移后缀，且该时间戳 MUST 等于本次注册的时间戳参数。去掉前缀后的路径长度 MUST 不超过 `constant.MigrationPkgMaxLen`（255）。`nil` 或 typed-nil 的 migrator MUST 在注册时拒绝。任一项不合法时进程 MUST 立刻终止，不得进入运行期。

`NewMigration(database, pkgPath, id, version, timestamp, up)` SHALL 由 `register` 导出，用于在不注册进 Registry 的情况下构造并校验一条迁移（测试与包外工具）。进注册表的唯一入口仍是 `Regist`。`hcm/pkg/migrate` 的 `PkgSuffix` SHALL 返回迁移后缀。格式校验与字符串辅助（`ValidateMigrationID`、`ValidateTimestamp`、`ParsePkgPath`、`TruncateUTF8` 等）SHALL 放在 `hcm/pkg/migrate`；表名、长度上限与 `PENDING` 等常量 SHALL 放在 `hcm/pkg/criteria/constant`。

#### Scenario: 版本格式非法拒绝启动

- **WHEN** 某迁移注册的版本为 `v1.9.3-Tenant.1`
- **THEN** 进程在启动阶段终止并打印该迁移的 ID 与非法版本

#### Scenario: 时间戳位数不对拒绝启动

- **WHEN** 某迁移注册的时间戳为 `2026090516`
- **THEN** 进程在启动阶段终止

#### Scenario: 时间戳不是真实时间拒绝启动

- **WHEN** 某迁移注册的时间戳为 `20261301000000`
- **THEN** 进程在启动阶段终止

#### Scenario: ID 不是合法格式拒绝启动

- **WHEN** 某迁移注册的 ID 为 `add-bk-asset-id`
- **THEN** 进程在启动阶段终止

#### Scenario: 可读 ID 允许注册

- **WHEN** 某迁移注册的 ID 为 `20260905-1600-ADD-BK-ASSET-ID-A3F9`，版本为 `v1.9.3`，时间戳为 `20260905160000`
- **THEN** 注册成功

#### Scenario: 占位符版本允许注册

- **WHEN** 某迁移注册的版本为常量 `PENDING`
- **THEN** 注册成功，不在 `init()` 里因占位符终止

#### Scenario: 注册进错误的库拒绝启动

- **WHEN** 导入路径以 `hcm/migrate/migrations/main/` 开头的迁移调用 `register.Obs.Regist`
- **THEN** 进程在启动阶段终止

#### Scenario: nil migrator 拒绝启动

- **WHEN** 调用 `Regist` 时传入 `nil` 或 typed-nil 指针
- **THEN** 进程在启动阶段终止

#### Scenario: 去掉前缀后路径过长拒绝启动

- **WHEN** 去掉 `hcm/migrate/migrations/` 后的包路径超过 255 个字符
- **THEN** 进程在启动阶段终止

#### Scenario: NewMigration 不进入注册表

- **WHEN** 调用 `NewMigration` 构造一条合法迁移
- **THEN** 返回校验后的 `Migration`，且对应 Registry 的执行序列不含它

### Requirement: 重复 ID 允许注册

同一个注册表内 SHALL 允许两条迁移使用相同的 Migration ID。注册期 MUST NOT 因 ID 重复终止进程。迁移后缀相同是内部四位与外部三位、pending 定版、归档的正常形态。迁移后缀不同 MUST NOT 在单次 `Regist` 里失败，由执行前校验以退出码 6 拒绝。不同库的记录表相互独立。

#### Scenario: 同库相同 ID 且后缀相同可以注册

- **WHEN** 同一个 `Registry` 内两条迁移注册了相同的 Migration ID，迁移后缀都是 `20260905160000_add_bk_asset_id`，版本分别是 `v1.9.3.1` 与 `v1.9.3`
- **THEN** 两条都注册成功，都出现在执行序列中

#### Scenario: 同库相同 ID 且后缀不同仍能注册

- **WHEN** 同一个 `Registry` 内两条迁移注册了相同的 Migration ID，迁移后缀分别是 `20260905160000_add_bk_asset_id` 与 `20260925103000_add_host_index`
- **THEN** 两条都注册成功，进程不在启动阶段终止

#### Scenario: 不同库相同 ID 可以注册

- **WHEN** `Main` 与 `Obs` 各有一条迁移使用了相同的 Migration ID
- **THEN** 注册成功，两个库的记录表相互独立，互不影响跳过判定

### Requirement: 排序后的执行序列

`Registry` SHALL 提供按执行顺序排好的迁移序列，排序键为版本、时间戳、Migration ID。路径、导入路径与 `package` 子句 MUST NOT 参与排序。版本为 `PENDING` 的迁移 SHALL 排在全部已定版之后，彼此再按时间戳、ID 排序。返回的切片 MUST 是副本，调用方修改不影响注册表内部状态。

#### Scenario: 返回确定顺序

- **WHEN** 多次获取同一注册表的执行序列
- **THEN** 每次返回的顺序完全一致

#### Scenario: 返回副本

- **WHEN** 调用方修改返回的切片
- **THEN** 注册表内部保存的序列不受影响

#### Scenario: 未定版排在已定版之后

- **WHEN** 注册表含 `v1.9.3` 与 `PENDING`
- **THEN** `v1.9.3` 排在 `PENDING` 之前

### Requirement: 漏导入防护

每条迁移 SHALL 在 `imports.go` 里有一行空白 import，指向该迁移目录。空白 import MUST NOT 依赖 `package` 子句的名字。SHALL 提供 `check-migrate`：比较含 `migrate.go` 的目录与 `imports.go` 里的空白 import，不采用重新生成再 diff。下列情况 MUST 失败：漏 import；多余 import；版本目录或 `pending/` 的直接子项不是目录，或迁移目录里还有子目录；迁移目录没有 `migrate.go`，或全目录里 `Regist(` 不是恰好一次且不在 `migrate.go`；`main/` 下出现 `register.Obs`，或 `obs/` 下出现 `register.Main`；同一个库内多条迁移的 ID 相同且迁移后缀不同。

同一个库内 ID 相同且迁移后缀相同 SHALL 只警告，打印目录和 ID，MUST NOT 因此失败。`main` 与 `obs` MUST 分开看。

#### Scenario: 新建迁移自动写入导入

- **WHEN** 通过 `new-migrate.sh` 创建一条迁移目录
- **THEN** 该目录的空白 import 被写入 `imports.go`，且该目录下文件的 `package` 子句为 `migration`

#### Scenario: 门禁检出漏导入

- **WHEN** `migrations/main/v1.9.4/v1.9.4_20260925103000_add_host_index/` 存在但 `imports.go` 未导入它，执行 `check-migrate`
- **THEN** 校验失败并指出该目录

#### Scenario: 同 ID 不同后缀在门禁失败

- **WHEN** 同一个库下两个迁移目录的 `const ID` 相同，迁移后缀不同，执行 `check-migrate`
- **THEN** 校验失败并打印这两个目录和该 ID

#### Scenario: 同 ID 同后缀只警告

- **WHEN** 内部四位目录与外部三位目录的 `const ID` 相同且迁移后缀相同，执行 `check-migrate`
- **THEN** 校验成功，并打印警告，含这两个目录和该 ID

## REMOVED Requirements

### Requirement: 相同 ID 的动作名一致（待确认，review 前不实现）

**Reason**: 动作名和 ID 写在同一次注册里，复制时会一起被带走，对不上粘错的 ID。目录 tag 对 ID tag 段同样不采用。改为比较导入路径的迁移后缀，见 `migrate-executor` 的疑似 ID 复用。
**Migration**: 删除本条及其三个场景。不要在 `Regist` 或 `check-migrate` 里实现动作名比对。

## ADDED Requirements

### Requirement: 一条迁移一个目录

每条迁移 SHALL 是一个目录、一个包、一处 `Regist`。`package` 子句 MUST 为 `migration`。版本目录和 `pending/` 只分组，自身 MUST NOT 是 Go 包。`migrate.go` SHALL 放 ID 常量以及 `init()` 里的那一次 `Regist`。同目录其他文件 MUST NOT 再注册。迁移包之间 MUST NOT 互相 import。迁移文件 SHALL 只 import `migrate/register`、`migrate/util` 与（注册 `PENDING` 时）`pkg/criteria/constant`，MUST NOT import `migrate/engine`。

包的身份 SHALL 是导入路径。`applied_pkg` 与跳过比对 MUST NOT 使用 `package` 子句。

#### Scenario: 包名相同不影响导入

- **WHEN** 两个迁移目录的 `package` 子句都是 `migration`，且 `imports.go` 以空白 import 引用两者的路径
- **THEN** 两者都能完成注册，注册结果里的 `Pkg` 分别是各自的导入路径去掉模块前缀后的值
