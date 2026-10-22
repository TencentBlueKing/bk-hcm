## ADDED Requirements

### Requirement: 注册 http 插件
cloud-server SHALL 提供 `POST /api/v1/cloud/iac_plugins/register`，同步执行，注册完成后才响应。cloud-server 只做参数校验与鉴权，MUST 经 `pkg/client/hc-service` 转发给 hc-service，由 hc-service 完成加锁、探活、写记录、同步与收尾。请求体字段：`plugin_name`（必填）、`api_path`（必填，路径后缀）、`crypto_type`（必填）、`crypto_config`（必填）、`memo`（可选，不超过 255 个字符）。成功响应 `{id, plugin_name, plugin_version, state}`。插件版本 MUST 取自探活返回的 `plugin_version`，MUST NOT 由请求传入。

#### Scenario: 注册成功
- **WHEN** 参数合法、探活通过、Schema 同步成功
- **THEN** 插入记录并最终为 `enabled`，响应含 `plugin_version` 与 `state=enabled`

#### Scenario: 缺少必填
- **WHEN** 缺少 `api_path` 或 `crypto_config`
- **THEN** 返回参数错误

#### Scenario: memo 超长
- **WHEN** `memo` 超过 255 个字符
- **THEN** 返回参数错误

### Requirement: 注册流程与前置校验顺序
注册 SHALL 按以下顺序执行，任一步失败即停止：权限（`iac_plugin` / `create`）→ 按插件名 `TryLock`（拿不到立即返回"注册进行中"）→ 同名记录 `protocol` 检查（插件名在 tf_provider 与 http 之间 MUST 不重复）→ 地址与加密配置校验 → 探活 → 读取该插件名下版本最高的记录。地址 / 加密配置校验失败或探活失败时 MUST NOT 写注册表。

#### Scenario: 拿不到锁
- **WHEN** 同名插件正在注册
- **THEN** 立即返回"注册进行中"，不探活

#### Scenario: 与 tf 插件重名
- **WHEN** 已存在同名且 `protocol=tf_provider` 的记录
- **THEN** 拒绝，指出插件名已被 tf_provider 插件占用

#### Scenario: 校验或探活失败不写库
- **WHEN** 地址校验失败或探活失败
- **THEN** 注册表无任何变化

### Requirement: 按已有记录处理
探活成功后，注册 SHALL 按该插件名下版本最高的记录处理：不存在 → 插入 `init` 记录；为 `init` 或 `failed` 且版本与探活一致 → 该记录重入，用本次请求的参数覆盖其 `extension` / `memo` 并改为 `init`（不设超时判断）；为 `init` 或 `failed` 但探活版本更高 → 插入探活版本的 `init` 记录，旧记录保持原状（记录版本不可修改）；为 `init` 或 `failed` 但探活版本更低 → 返回"暂不支持回滚"；为 `enabled` 或 `disabled` → 返回"已注册"，不做任何改动。记录 `extension` MUST 写入 `api_path`、`crypto_type`、`crypto_config`，`protocol=http`，版本取探活返回值。

#### Scenario: 失败记录且探活版本更高
- **WHEN** 版本最高记录为 `failed` 的 `1.0.0`，探活返回 `1.1.0`
- **THEN** 插入 `1.1.0` 的 `init` 记录并同步，`1.0.0` 记录不变

#### Scenario: 已注册
- **WHEN** 版本最高记录为 `enabled`
- **THEN** 返回"已注册"，记录不变

#### Scenario: 上次失败后重试
- **WHEN** 版本最高记录为 `failed`
- **THEN** 用本次参数覆盖并改为 `init`，继续同步

#### Scenario: 上次中断
- **WHEN** 版本最高记录为 `init`
- **THEN** 同样覆盖重入

### Requirement: 同步 Schema 与收尾
记录为 `init` 后，hc-service SHALL 在进程内调用 `integration.SchemaSyncer.Sync`（与 tf_provider 共用同一个实例），getter 为经网关向插件取 Schema 的 http 实现，并按结果收尾：成功 → 条件更新（`expect_state=init`）为 `enabled`；失败 → 条件更新为 `failed` 并写入 `message`，返回错误。注册 MUST NOT 创建连接池或进程。各步骤超时之和 MUST 小于上游网关超时（探活 10 秒，同步超时由对接接口定义）。

#### Scenario: 同步成功
- **WHEN** 同步返回成功
- **THEN** 记录改为 `enabled`

#### Scenario: 同步失败
- **WHEN** 同步返回失败，原因为 `schema_hash mismatch`
- **THEN** 记录改为 `failed`，`message` 含该原因，接口返回错误

#### Scenario: 条件更新冲突
- **WHEN** 收尾时记录已不是 `init`（被其他流程改动）
- **THEN** 返回冲突错误，不覆盖

### Requirement: 修改 http 插件
cloud-server SHALL 提供 `POST /api/v1/cloud/iac_plugins/{id}/update`，同步执行，参数校验与鉴权后转发给 hc-service 执行，针对该插件名下版本最高的记录。可选字段：`api_path`、`crypto_type`、`crypto_config`、`memo`，至少传一个，未传的保持不变；`plugin_name` MUST NOT 可修改。权限为 `iac_plugin` / `update`；与注册、升级、启停共用插件名锁（`TryLock`）。仅 `enabled`、`disabled` 的插件可修改，`init`、`failed` MUST 拒绝。tf_provider 插件 MUST 返回错误并提示其配置由目录和全局配置决定。

#### Scenario: 仅改备注与加密配置
- **WHEN** 传入 `crypto_config` 与 `memo`
- **THEN** 校验加密配置后直接条件更新 `extension` 与 `memo`，不改 `state`、版本，不触发 Schema 同步

#### Scenario: 一个字段都不传
- **WHEN** 请求体全部字段缺省
- **THEN** 返回参数错误

#### Scenario: 状态不允许
- **WHEN** 当前记录为 `failed`
- **THEN** 拒绝修改

#### Scenario: tf 插件
- **WHEN** 目标记录 `protocol=tf_provider`
- **THEN** 返回错误并提示配置来源

### Requirement: 修改 api_path
修改 `api_path` 时，系统 SHALL 先校验并探活新地址：探活版本高于当前记录 → 按升级流程处理（新记录使用新地址）；版本不变 → 重新同步 Schema，成功后更新 `extension`，失败则 MUST NOT 修改记录并返回错误；版本更低 → 返回"暂不支持回滚"。

#### Scenario: 新地址版本更高
- **WHEN** 新地址探活返回的版本高于当前
- **THEN** 走升级流程，新版本记录使用新地址

#### Scenario: 新地址版本相同
- **WHEN** 探活版本与当前一致且同步成功
- **THEN** 更新 `extension.api_path`

#### Scenario: 新地址同步失败
- **WHEN** 版本相同但同步失败
- **THEN** 记录保持原样，返回错误

### Requirement: 注册与修改审计
注册成功（新增记录）MUST 写 create 审计；修改成功 MUST 写 update 审计，`detail.data` 为修改前记录，`detail.changed` 为本次修改的字段；`crypto_config` 在审计中 MUST NOT 记录公钥以外的敏感内容（公钥可记录）。

#### Scenario: 注册审计
- **WHEN** 注册成功
- **THEN** 产生 create 审计，`res_type=iac_plugin`，`bk_biz_id=-1`

#### Scenario: 修改审计
- **WHEN** 修改 `memo` 成功
- **THEN** 产生 update 审计，`detail.changed` 含 `memo`
