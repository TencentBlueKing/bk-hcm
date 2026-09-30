## ADDED Requirements

### Requirement: 版本号解析

系统 SHALL 只接受三种写法的迁移版本号：`vX.Y.Z`、`vX.Y.Z.N`、`vX.Y.Z-<label>.N`，其中前三段为十进制数字，`label` 只允许小写字母开头、由小写字母和数字组成，`N` 为十进制数字。解析结果 MUST 为 `Version` 结构（`P1` / `P2` / `P3` / `Suffix` / `Raw`），其中 `Suffix` 为 `nil` 表示三位版本。解析 MUST 拒绝前导零，使版本串与解析结果一一对应。

#### Scenario: 解析三位版本

- **WHEN** 解析 `v1.9.3`
- **THEN** 得到 `P1=1` / `P2=9` / `P3=3`，`Suffix` 为 `nil`，`Raw` 为 `v1.9.3`

#### Scenario: 解析数字第四段

- **WHEN** 解析 `v1.9.3.1`
- **THEN** 得到 `P1=1` / `P2=9` / `P3=3`，`Suffix.Label` 为空串，`Suffix.Seq=1`

#### Scenario: 解析带标签第四段

- **WHEN** 解析 `v1.9.3-tenant.1`
- **THEN** 得到 `P1=1` / `P2=9` / `P3=3`，`Suffix.Label="tenant"`，`Suffix.Seq=1`

#### Scenario: 数字第四段为 0 与三位等同

- **WHEN** 分别解析 `v1.9.3` 和 `v1.9.3.0` 并比较
- **THEN** 两者相等。`v1.9.3.0` 的第四段收成不存在

#### Scenario: 带标签的序号 0 不等于三位

- **WHEN** 比较 `v1.9.3` 与 `v1.9.3-tenant.0`
- **THEN** `v1.9.3` 排在前面，两者不相等

#### Scenario: 拒绝缺少前缀 v

- **WHEN** 解析 `1.9.3`
- **THEN** 返回错误

#### Scenario: 拒绝标签写法错误

- **WHEN** 解析 `v1.9.3.tenant.1`、`v1.9.3-Tenant.1` 或 `v1.9.3-tenant`
- **THEN** 逐个返回错误

#### Scenario: 拒绝段数不合法

- **WHEN** 解析 `v1.9` 或 `v1.9.3.1.2`
- **THEN** 返回错误

#### Scenario: 拒绝前导零

- **WHEN** 解析 `v01.9.3` 或 `v1.9.3.01`
- **THEN** 返回错误

### Requirement: 版本全序比较

系统 SHALL 提供一个对全部合法版本号成立的全序比较器，用于得到确定且可复现的执行顺序。比较 SHALL 依次按：`P1`、`P2`、`P3` 按数值；第四段为 `nil` 的小于任何非 `nil`；两边都有第四段时先比 `Label`（空串最小，其余按字典序）再比 `Seq` 按数值；然后比 14 位时间戳字符串；最后比 Migration ID 字符串。比较器 MUST NOT 用于筛选「哪条版本线的文件可以执行」。跨标签先后只保证比较函数有确定结果，同一注册表出现两个非空标签时不按这个顺序执行。

#### Scenario: 前三段按数值比较

- **WHEN** 比较 `v1.9.9` 与 `v1.9.10`
- **THEN** `v1.9.9` 排在 `v1.9.10` 之前

#### Scenario: 同前缀下三位先于第四段

- **WHEN** 比较 `v1.9.3` 与 `v1.9.3.1`
- **THEN** `v1.9.3` 排在 `v1.9.3.1` 之前

#### Scenario: 数字第四段先于带标签第四段

- **WHEN** 比较 `v1.9.3.2` 与 `v1.9.3-tenant.1`
- **THEN** `v1.9.3.2` 排在 `v1.9.3-tenant.1` 之前

#### Scenario: 同标签按序号数值比较

- **WHEN** 比较 `v1.9.3-tenant.2` 与 `v1.9.3-tenant.10`
- **THEN** `v1.9.3-tenant.2` 排在 `v1.9.3-tenant.10` 之前

#### Scenario: 跨标签按字典序比较

- **WHEN** 比较 `v1.9.3-tenant.1` 与 `v1.9.3-woa.1`
- **THEN** `v1.9.3-tenant.1` 排在 `v1.9.3-woa.1` 之前

#### Scenario: 同版本按时间戳比较

- **WHEN** 两条迁移版本同为 `v1.9.3`，时间戳分别为 `20260905160000` 与 `20260906120000`
- **THEN** 时间戳小的排在前面

#### Scenario: 同版本同时间戳按 ID 决胜

- **WHEN** 两条迁移版本与时间戳完全相同
- **THEN** 按 Migration ID 字符串比较，排序结果 MUST 唯一且可复现

#### Scenario: 排序不筛选版本线

- **WHEN** 注册表中同时存在 `v1.9.3`、`v1.9.3.1` 与 `v1.9.3-tenant.1`
- **THEN** 三条都出现在排序结果中，MUST NOT 因标签不同被剔除

### Requirement: 同一注册表只允许一个非空标签

第四段表示该分支自己的特有提交。`init` 与 `up` MUST 在连接数据库之前扫描注册表中的非空 `Label`。出现两个及以上不同的非空标签时，命令 MUST 以退出码 3 失败，列出这些标签和对应的 Migration ID，且 MUST NOT 执行任何迁移。只有一个非空标签、或全是空标签的数字第四段，SHALL 照常执行。此检查 MUST NOT 按标签挑掉其中一部分文件。

#### Scenario: 两个标签拒绝执行

- **WHEN** 同一注册表中同时有 `v1.9.3-tenant.1` 与 `v1.9.3-woa.1`，执行 `up`
- **THEN** 命令在连库之前以退出码 3 失败，输出两个标签及对应 Migration ID，两条迁移都不执行

#### Scenario: 单个标签照常执行

- **WHEN** 注册表中只有标签 `tenant`，以及没有第四段和数字第四段的迁移，执行 `up`
- **THEN** 这些迁移都进入判定，MUST NOT 因带有 `tenant` 被拒绝

#### Scenario: 纯数字第四段照常执行

- **WHEN** 注册表中的第四段全部是空标签的数字序号，执行 `up`
- **THEN** 命令不因标签检查失败

### Requirement: PENDING 占位符

未定版的迁移 SHALL 以 `register` 包中的常量作为注册版本，常量值为 `PENDING`。调用方 MUST NOT 写这个字符串字面量。目录名 SHALL 保持小写 `pending/`。`Parse` MUST 对 `PENDING` 返回错误，另由 `IsPending` 判定。`init` 与 `up` MUST 在连接数据库之前扫描全部注册项，发现占位符即以退出码 3 失败并列出对应 Migration ID。

#### Scenario: 占位符不是合法版本

- **WHEN** 解析常量值 `PENDING`
- **THEN** 返回错误，且 `IsPending` 判定为真

#### Scenario: 小写不是占位符

- **WHEN** 注册版本为 `pending`
- **THEN** 视为非法版本，进程在启动阶段终止，MUST NOT 当成未定版占位符放行

#### Scenario: 执行前检出占位符

- **WHEN** 注册表中存在版本为 `PENDING` 的迁移，执行 `up`
- **THEN** 命令在连库之前失败，退出码为 3，输出列出占位项的 Migration ID

#### Scenario: list 允许展示占位项

- **WHEN** 注册表中存在版本为 `PENDING` 的迁移，执行 `list`
- **THEN** 版本列打印 `PENDING`，命令退出码为 0

### Requirement: 版本上限比较语义

`--to` 指定的版本上限 SHALL 使用同一个比较器判定，只处理 `version <= 上限` 的迁移。上限本身 MUST 是合法版本号，否则以退出码 2 失败。三位上限包含与之等同的 `.0`，MUST NOT 包含同前缀下序号大于 0 的第四段。

#### Scenario: 上限过滤超出的文件

- **WHEN** 注册表含 `v1.9.3`、`v1.9.4`，执行 `up --to v1.9.3`
- **THEN** 只处理 `v1.9.3`，`v1.9.4` 不执行

#### Scenario: 三位上限包含等同的 .0

- **WHEN** 注册表含 `v1.9.3.0`，执行 `up --to v1.9.3`
- **THEN** `v1.9.3.0` 被处理

#### Scenario: 三位上限不含同前缀的更大第四段

- **WHEN** 注册表含 `v1.9.3` 与 `v1.9.3.1`，执行 `up --to v1.9.3`
- **THEN** 只处理 `v1.9.3`，`v1.9.3.1` 因超出上限不执行

#### Scenario: 上限非法即失败

- **WHEN** 执行 `up --to v1.9`
- **THEN** 命令失败，退出码为 2

### Requirement: 不引入历史错版号映射

系统 MUST NOT 提供 bk-cmdb 式的错版号重映射表（`remapVersion` / `wrongVersion`）。所有非法版本号一律硬失败。

#### Scenario: 非法版本不被容错修正

- **WHEN** 某迁移注册的版本为 `x18_10_10_01` 这类非法串
- **THEN** 系统 MUST 失败，MUST NOT 将其修正为某个合法版本后继续
