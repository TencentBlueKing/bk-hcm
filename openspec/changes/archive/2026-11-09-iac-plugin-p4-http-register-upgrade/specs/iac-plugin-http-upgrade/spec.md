## ADDED Requirements

### Requirement: 升级接口
cloud-server SHALL 提供 `POST /api/v1/cloud/iac_plugins/{id}/upgrade`，无请求体，同步执行，完成后才响应，成功响应 `{id, plugin_name, plugin_version, state}`。用于管理员在插件方发版后触发升级检查，或在插件处于 `failed` 时触发恢复。权限为 `iac_plugin` / `update`。cloud-server 只做鉴权，MUST 转发给 hc-service 执行。tf_provider 插件 MUST 返回错误，提示通过新增版本目录并滚动重启 hc-service 升级。

#### Scenario: tf 插件
- **WHEN** 对 `protocol=tf_provider` 的记录调用升级
- **THEN** 返回错误并提示滚动重启方式

#### Scenario: 插件名不存在
- **WHEN** `{id}` 记录不存在
- **THEN** 返回记录不存在错误

### Requirement: 升级判定流程
升级与巡检 SHALL 共用同一流程：按插件名 `TryLock`（拿不到：接口返回"升级进行中"，巡检跳过本轮）→ 读取该插件名下版本最高的记录 L → 经 hc-service 探活取版本 V → 判定：V 高于 L → 插入 V 的 `init` 记录并同步；V 等于 L 且 L 为 `failed` 或 `init` → L 改为 `init` 并同步；V 等于 L 且 L 为 `enabled` 或 `disabled` → 无需处理；V 低于 L → 返回"暂不支持回滚"。与注册、修改共用同一把插件名锁。

#### Scenario: 发现新版本
- **WHEN** L 版本 `1.2.0`，探活 V 为 `1.3.0`
- **THEN** 插入 `1.3.0` 的 `init` 记录并同步，L 保持原状

#### Scenario: 版本相同无需处理
- **WHEN** V 等于 L 且 L 为 `enabled`
- **THEN** 直接成功，不改动，不写审计

#### Scenario: 异常恢复
- **WHEN** V 等于 L 且 L 为 `failed`
- **THEN** L 改为 `init`，重新同步

#### Scenario: 版本回退
- **WHEN** V 低于 L
- **THEN** 返回"暂不支持回滚"，记录不变

#### Scenario: 拿不到锁
- **WHEN** 同名插件注册或升级中
- **THEN** 接口返回"升级进行中"，不探活

### Requirement: 升级期间的请求保护与状态继承
升级期间新版本记录处于 `init`，请求入口读取版本最高的记录，不是 `enabled` 即拒绝新请求；旧版本记录在升级期间 MUST NOT 被修改。同步成功后，新版本记录 SHALL 继承上一版本记录（版本最高的旧记录）的状态：上一版本为 `disabled` 则新版本为 `disabled`，否则为 `enabled`。同步失败时新版本记录为 `failed`，旧版本记录保持原样，错误信息带 `message`。新 Schema 的协议版本不在支持列表内属同步失败（由对接方校验，本流程按失败处理）。

#### Scenario: 继承停用
- **WHEN** 上一版本为 `disabled`，新版本同步成功
- **THEN** 新版本为 `disabled`

#### Scenario: 同步失败旧版本不动
- **WHEN** 新版本同步失败
- **THEN** 新版本为 `failed`，旧版本记录保持原样

#### Scenario: 停用且 Schema 缺失时恢复
- **WHEN** 版本最高记录为 `disabled` 且 Schema 不存在，调用升级接口
- **THEN** 重做同步，成功后仍为 `disabled`

### Requirement: 巡检
hc-service SHALL 在 master 副本（`IsMaster()`）上运行定时巡检（默认间隔 10 秒，可配置），对所有 `protocol=http` 的插件探活并走升级判定流程。巡检 MUST 对 `disabled` 的插件也探活。探活失败的处理：当前版本为 `enabled` → 失败计数加一，连续达到 3 次（计数在内存）则改为 `failed` 并写 `message`；当前版本为 `disabled` 或其他 → 只记告警日志，不改记录。探活成功 MUST 清零该插件的失败计数；`failed` 的插件探活成功后 MUST 走重入（改为 `init` 并重新同步），恢复后按状态继承规则收尾。拿不到锁的插件 MUST 跳过本轮。非 master 副本 MUST NOT 执行巡检。

#### Scenario: 连续三次失败
- **WHEN** `enabled` 插件连续三个周期探活失败
- **THEN** 记录改为 `failed`，`message` 写明探活失败

#### Scenario: 中途成功清零
- **WHEN** 前两次失败，第三次成功
- **THEN** 计数清零，记录保持 `enabled`

#### Scenario: disabled 探活失败
- **WHEN** `disabled` 插件探活失败
- **THEN** 只记告警日志，记录不变

#### Scenario: 异常自动恢复
- **WHEN** `failed` 插件探活成功
- **THEN** 走重入流程，同步成功后恢复为 `enabled`（上一版本为 `disabled` 时为 `disabled`）

#### Scenario: 拿不到锁
- **WHEN** 某插件正被管理员升级
- **THEN** 巡检跳过该插件，下一轮再处理

#### Scenario: 非 master
- **WHEN** 当前副本不是 master
- **THEN** 不执行巡检

### Requirement: 系统触发的审计
升级新增版本记录 MUST 写 create 审计；巡检触发的状态变化（进入或恢复 `failed`、升级新增版本）MUST 写审计，操作人为后端 kit 的 `hcm-backend-admin`（与 tf_provider 注册一致）；无变化（版本相同、告警日志）MUST NOT 写审计。

#### Scenario: 巡检置异常
- **WHEN** 巡检把 `enabled` 改为 `failed`
- **THEN** 产生 update 审计，操作人 `hcm-backend-admin`，`detail.changed` 为 `{"state":"failed"}`

#### Scenario: 版本相同
- **WHEN** 升级或巡检发现版本相同
- **THEN** 不产生审计
