## MODIFIED Requirements

### Requirement: PENDING 占位符

未定版的迁移 SHALL 以 `constant.MigrationPendingVersion` 作为注册版本，常量值为 `PENDING`。调用方 MUST NOT 写这个字符串字面量。目录名 SHALL 保持小写，且位于 `pending/` 下，目录名本身没有版本前缀。`Parse` MUST 对 `PENDING` 返回错误，另由 `IsPending` 判定。`Regist` MUST 接受该常量，MUST NOT 在 `init()` 里因它终止进程。是否允许执行由 `--allow-pending` 在执行前校验决定，见 `migrate-executor`。`list` MUST NOT 因占位符失败。

#### Scenario: 占位符不是合法版本

- **WHEN** 解析常量值 `PENDING`
- **THEN** 返回错误，且 `IsPending` 判定为真

#### Scenario: 小写不是占位符

- **WHEN** 注册版本为 `pending`
- **THEN** 视为非法版本，进程在启动阶段终止，MUST NOT 当成未定版占位符放行

#### Scenario: 注册期接受占位符

- **WHEN** 某迁移以常量 `PENDING` 调用 `Regist`
- **THEN** 注册成功，进程进入运行期

#### Scenario: list 允许展示占位项

- **WHEN** 注册表中存在版本为 `PENDING` 的迁移，执行 `list`
- **THEN** 版本列打印 `PENDING`，命令退出码为 0
