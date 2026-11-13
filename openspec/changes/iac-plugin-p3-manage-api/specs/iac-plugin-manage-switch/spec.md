## ADDED Requirements

### Requirement: 启用与停用接口
cloud-server SHALL 提供 `POST /api/v1/cloud/iac_plugins/switch`，请求体 `{"plugin_name":"...","state":"enabled"}` 或 `{"plugin_name":"...","state":"disabled"}`，目标状态由参数决定，启用与停用 MUST 使用同一接口。接口只修改注册表中的 `state`，修改后立即返回，MUST NOT 直接操作任何进程；返回成功表示状态已修改。`state` 取其他值 MUST 返回参数错误。

#### Scenario: 停用
- **WHEN** 对 `enabled` 的插件请求 `state=disabled`
- **THEN** 状态改为 `disabled` 并立即返回成功

#### Scenario: 非法目标状态
- **WHEN** 请求体 `state` 为 `init` 或 `failed`
- **THEN** 返回参数错误，不修改任何记录

#### Scenario: 插件不存在
- **WHEN** `plugin_name` 对应记录不存在
- **THEN** 返回记录不存在错误

### Requirement: 按当前版本判定的状态规则
接口 SHALL 以该插件名下版本最高的记录（semver，当前版本）的状态为准，并按下表处理：

| 当前状态 | 目标 disabled | 目标 enabled |
| --- | --- | --- |
| enabled | 改为 disabled | 已是目标状态，直接成功 |
| disabled | 已是目标状态，直接成功 | 改为 enabled；当前版本 Schema 必须已存在，否则拒绝并提示先恢复 |
| failed | 改为 disabled | 拒绝，提示先恢复 |
| init | 拒绝，提示注册或升级未完成 | 拒绝，提示注册或升级未完成 |

"先恢复"在 tf_provider 表示等待重新注册，在 http 表示调用升级接口。请求的 `{id}` 可以是该插件名下任一版本，判定始终以当前版本为准。

#### Scenario: 异常状态下不能启用
- **WHEN** 当前版本为 `failed`，请求 `enabled`
- **THEN** 拒绝，错误信息带该记录的 `message` 并按 `protocol` 提示恢复方式

#### Scenario: 异常状态下可以停用
- **WHEN** 当前版本为 `failed`，请求 `disabled`
- **THEN** 改为 `disabled`

#### Scenario: 注册未完成
- **WHEN** 当前版本为 `init`，请求任一目标状态
- **THEN** 拒绝，提示注册或升级未完成

#### Scenario: 幂等
- **WHEN** 当前版本已是目标状态
- **THEN** 直接成功，不改库，不写审计

#### Scenario: 启用时 Schema 缺失
- **WHEN** 当前版本为 `disabled` 且 Schema 不存在，请求 `enabled`
- **THEN** 拒绝，提示先恢复

#### Scenario: 以任一版本 id 请求
- **WHEN** 用旧版本记录的 `id` 请求停用，当前版本为 `enabled`
- **THEN** 按当前版本判定并生效

### Requirement: 状态变化作用于所有相关版本
状态变化 SHALL 作用于该插件名下所有处于 `enabled` / `failed`（目标 disabled）或 `disabled`（目标 enabled）的版本，`init` 版本 MUST 保持不变。tf_provider 与 http 的处理方式 MUST 相同。该变更 MUST 通过 P1 的批量改状态接口一次完成。

#### Scenario: 滚动发布期间停用
- **WHEN** `tencentcloud` 有 `1.83.30`（enabled）、`1.83.34`（enabled）两个版本，请求停用
- **THEN** 两个版本均变为 `disabled`

#### Scenario: init 版本不动
- **WHEN** 存在一个 `init` 版本，请求停用
- **THEN** 该 `init` 版本保持不变（若当前版本即为它则按规则拒绝）

### Requirement: 与注册、升级串行
接口 SHALL 获取以插件名为键的 etcd 锁后再读取与修改，与注册、升级、修改共用同一把锁。拿不到锁 MUST 返回"该插件正在注册、升级或变更中"类错误，不等待。

#### Scenario: 与升级并发
- **WHEN** 升级流程持有该插件名的锁时请求启停
- **THEN** 返回"进行中"错误，记录不被修改

#### Scenario: 锁在返回前释放
- **WHEN** 请求处理完成（成功或失败）
- **THEN** 锁已释放

### Requirement: 启停权限与审计
启停接口 SHALL 要求资源类型 `iac_plugin`、动作 `update`，映射 IAM `global_configuration`。状态实际发生变化时 MUST 写 update 审计（`res_type=iac_plugin`，`res_id`/`res_name` 为当前版本记录的 `id`/`plugin_name`，`bk_biz_id=-1`，`detail.data` 为变更前记录，`detail.changed` 为 `{"state":"<目标状态>"}`）；已是目标状态的幂等返回 MUST NOT 写审计。

#### Scenario: 变化写审计
- **WHEN** `enabled → disabled` 成功
- **THEN** 产生一条 update 审计，`detail.changed` 为 `{"state":"disabled"}`

#### Scenario: 幂等不写审计
- **WHEN** 已是目标状态
- **THEN** 不产生审计

#### Scenario: 无权限
- **WHEN** 调用者无权限
- **THEN** 返回无权限错误，不加锁、不修改

### Requirement: 插件可用性判定函数
系统 SHALL 提供函数：给定发起请求所使用版本的记录，返回是否可接受新请求；不是 `enabled` 时 MUST 返回按状态区分的错误（`init` 提示稍后重试；`disabled` 提示已停用；`failed` 带记录的 `message`）。该函数仅提供判定能力，不在本 change 接入具体请求路径。

#### Scenario: enabled
- **WHEN** 记录为 `enabled`
- **THEN** 返回可用

#### Scenario: failed
- **WHEN** 记录为 `failed` 且 `message` 为 `schema sync failed`
- **THEN** 返回的错误信息包含 `schema sync failed`

#### Scenario: init
- **WHEN** 记录为 `init`
- **THEN** 返回"稍后重试"类错误
