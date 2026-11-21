## ADDED Requirements

### Requirement: 注册表与库归属

系统 SHALL 提供 `Registry` 类型并暴露实例 `Main`（主库）。迁移的库归属 SHALL 由注册到哪个实例决定，MUST NOT 在运行期靠解析文件路径字符串分流。`migrations/main/**` 注册到 `Main`。新增库时增加一个实例和一棵目录。

#### Scenario: 按实例区分库归属

- **WHEN** 某迁移在 `init()` 里调用 `register.Main.Regist(...)`
- **THEN** 该迁移只在主库的执行序列中出现

#### Scenario: 外部版靠未注册自然不跑

- **WHEN** 二进制不包含 `*.x/` 目录的空白 import
- **THEN** 这些迁移不在任何注册表中，执行时不会被处理，且 MUST NOT 需要任何范围开关

### Requirement: 注册参数与 init 期强校验

`Regist` SHALL 接收 Migration ID、版本、14 位时间戳、语义描述与 `up` 函数。校验 SHALL 在 `init()` 阶段完成：ID MUST 是合法 UUID，版本 MUST 是合法版本号或常量 `PENDING`，时间戳 MUST 是 14 位数字且能按 `yyyyMMddHHmmss` 解析成真实时间，`up` MUST 非空。任一项不合法时进程 MUST 立即终止，不得进入运行期。

#### Scenario: 版本格式非法拒绝启动

- **WHEN** 某迁移注册的版本为 `v1.9.3-Tenant.1`
- **THEN** 进程在启动阶段终止并打印该迁移的 ID 与非法版本

#### Scenario: 时间戳位数不对拒绝启动

- **WHEN** 某迁移注册的时间戳为 `2026090516`
- **THEN** 进程在启动阶段终止

#### Scenario: 时间戳不是真实时间拒绝启动

- **WHEN** 某迁移注册的时间戳为 `20261301000000`
- **THEN** 进程在启动阶段终止

#### Scenario: ID 不是合法 UUID 拒绝启动

- **WHEN** 某迁移注册的 ID 为 `add-bk-asset-id`
- **THEN** 进程在启动阶段终止

#### Scenario: 占位符版本允许注册

- **WHEN** 某迁移注册的版本为常量 `PENDING`
- **THEN** 注册成功，由 `init` / `up` 的执行前扫描负责拦截

### Requirement: 重复 ID 允许注册

同一个注册表内 SHALL 允许两条迁移使用相同的 Migration ID。这是内外部合入的正常结果：同一段变更可以同时以不同版本号、相同 ID 存在。注册期 MUST NOT 因 ID 重复终止进程。执行时按 ID 跳过已成功的那条，见执行器的判定规则。

#### Scenario: 同库相同 ID 可以注册

- **WHEN** 同一个 `Registry` 内两条迁移注册了相同的 Migration ID、不同的版本号
- **THEN** 两条都注册成功，都出现在执行序列中

#### Scenario: 不同库相同 ID 可以注册

- **WHEN** 两个不同库的注册表各有一条迁移使用了相同的 Migration ID
- **THEN** 注册成功，两个库的记录表相互独立，互不影响跳过判定

### Requirement: 相同 ID 的动作名一致（待确认，review 前不实现）

本条等待其他人 review，确认前 MUST NOT 实现，也 MUST NOT 改变现有行为：注册期仍允许相同 ID，执行前扫描不看动作名。

不比较迁移文件正文。两份文件是否是同一段变更，靠正文比对做不到准确。

拟增加的规范：内外部都保留的同一段通用变更，动作名必须相同。动作名是文件名里时间戳之后的那一段，也是 `Regist` 的 description。例如 `v1.9.3_20250925160000_extract_bk_asset_id.go` 与 `v1.9.3.2_20250925160000_extract_bk_asset_id.go` 的动作名都是 `extract_bk_asset_id`。版本号和时间戳可以不同。

拟增加的检查：`init` 与 `up` 在连库之前，按注册表扫描相同 Migration ID 的条目。动作名一致则放行，仍按现有规则由执行器按 ID 跳过已成功的那条。动作名不一致则以退出码 3 失败，列出该 ID 和各自动作名，一条都不执行。`list` 不因这项失败。

这项能拦住的是复制已有迁移文件后忘了换 ID：两段不同的变更动作名通常不同，于是在执行前失败，而不是被按 ID 跳过、永远不跑。

遵守这项的代价：内容与 ID 都相同、只是动作名略有出入的两份通用文件，也会在执行前失败。开发要保证两份文件的动作名一致。

不遵守这项的代价只落在检查上：重复文件在执行前失败。不会变成漏执行一个迁移文件。

#### Scenario: 确认前不检查动作名

- **WHEN** 同一个注册表里两条迁移 ID 相同、动作名不同，且本条尚未确认
- **THEN** 注册成功，执行前扫描不因动作名失败

#### Scenario: 确认后动作名一致则放行

- **WHEN** 本条已确认，同一个注册表里 `v1.9.3_20250925160000_extract_bk_asset_id.go` 与 `v1.9.3.2_20250925160000_extract_bk_asset_id.go` 使用相同 ID，动作名都是 `extract_bk_asset_id`
- **THEN** 执行前扫描放行，两条都留在执行序列里

#### Scenario: 确认后动作名不一致则拒绝执行

- **WHEN** 本条已确认，同一个注册表里两条迁移 ID 相同，动作名分别是 `extract_bk_asset_id` 与 `add_zone`
- **THEN** `init` 或 `up` 在连库之前以退出码 3 失败，输出该 ID 与两个动作名，两条都不执行

### Requirement: 排序后的执行序列

`Registry` SHALL 提供按执行顺序排好的迁移序列，排序键为版本、时间戳、Migration ID，路径与语义名 MUST NOT 参与排序。返回的切片 MUST 是副本，调用方修改不影响注册表内部状态。

#### Scenario: 返回确定顺序

- **WHEN** 多次获取同一注册表的执行序列
- **THEN** 每次返回的顺序完全一致

#### Scenario: 返回副本

- **WHEN** 调用方修改返回的切片
- **THEN** 注册表内部保存的序列不受影响

### Requirement: 漏导入防护

`imports.go` 漏掉某个版本目录的空白 import SHALL 导致该目录下的迁移永不执行且不报错，这是系统唯一的静默失效点。因此 SHALL 提供两道防护：`new-migrate.sh` 创建迁移时自动写入空白 import；出包门禁 `check-migrate` 校验 `migrations/` 下每个版本目录都已被导入，缺失即失败。

#### Scenario: 新建迁移自动写入导入

- **WHEN** 通过 `new-migrate.sh` 创建一条落在新版本目录的迁移
- **THEN** 该目录的空白 import 被写入 `imports.go`

#### Scenario: 门禁检出漏导入

- **WHEN** `migrations/main/v1.9.4/` 存在但 `imports.go` 未导入它，执行 `check-migrate`
- **THEN** 校验失败并指出缺失的目录

### Requirement: 迁移文件的依赖边界

`migrations/` 子树下的文件 SHALL 只 import `migrate/register` 与 `migrate/util`，MUST NOT import `migrate/engine`。由于 Go 编译器不会阻止这种导入，SHALL 提供 Makefile 门禁检查。

#### Scenario: 门禁检出非法依赖

- **WHEN** 某迁移文件 import 了 `hcm/migrate/engine`
- **THEN** 门禁检查失败并指出该文件
