> 分两个阶段交付（对应 design.md Migration Plan）：
> - **第一阶段**（1~3）：配置、校验、`InheritState` 公共化、`Prober`。hc-service 暂不接入、不暴露路由，可独立单测、不依赖对接方。
> - **第二阶段**（4~10）：hc-service 编排与内部路由、cloud-server 转发 handler、巡检、接口文档、联调。
>
> 编排放 hc-service（与 `tf-plugin` 同层），cloud-server 只做参数校验 + 鉴权 + 转发，见 design D1。

## 1. 配置（第一阶段）

- [x] 1.1 `pkg/cc`：hc-service 新增顶层 `iacGateway`，类型直接用 `ApiGateway`（与 `cmdb` 同形，不另立类型）；`endpoints` 为空表示不使用 http 插件、跳过校验，配置时走 `ApiGateway.validate()` 且只允许一个地址（它是 `api_path` 安全校验的基准）。`iacPlugin` 新增 `httpProbeTimeoutSeconds`（默认 10）、`httpCheckIntervalSeconds`（默认 60）、`httpCheckFailureThreshold`（默认 3）。cloud-server 不新增配置。探活响应体上限用常量（1MB）
- [x] 1.2 `hc_service.yaml`、helm `values.yaml`（顶层 `iacGateway`、`hcservice.iacPlugin` 三项）与 hc-service configmap 同步

## 2. 校验与公共逻辑（第一阶段）

- [x] 2.1 `pkg/iacplugin.ValidateAPIPath`：以 `/` 开头、不以 `/` 结尾、无 scheme / host / userinfo / `?` / `#` / `\`、无 `.` / `..` 段（含 `%2e%2e` 编码形式）、无空白与控制字符；`JoinPluginURL` 拼接后校验 scheme+host 与前缀一致、路径以前缀路径开头。P1 `HTTPExtension.BaseURL`（`base_url`）改为 `APIPath`（`api_path`），list 接口文档同步
- [x] 2.2 `pkg/iacplugin.ValidateCrypto`：`crypto_type` 仅支持 `RSA`；公钥必须是单个 `PUBLIC KEY`（SPKI）PEM 块、可解析、RSA、≥3072 位
- [x] 2.3 状态继承抽到 `pkg/iacplugin.InheritState`，`tf-plugin/register` 改用公共实现（删除其私有 `inheritedState`），补表驱动单测
- [x] 2.4 校验函数单测：合法后缀、带 host、`//host`、路径穿越、2048 位公钥、EC 公钥、PKCS1 公钥、非 PEM、多个 PEM 块、不支持的加密类型

## 3. 探活（第一阶段）

- [x] 3.1 `cmd/hc-service/logics/iac-plugin/http-plugin`：通过南向 `plugin.Client` 请求 `GET {endpoints[0]}{api_path}/health` 探活取版本；网关 `endpoints` 为空时返回默认版本并打告警，便于本地联调
- [x] 3.2 南向客户端单测（`httptest`）由 `pkg/thirdparty/api-gateway/plugin` 覆盖

## 4. hc-service 接入（第二阶段）

- [x] 4.1 `http-plugin`：按 `api_path` 缓存复用 `plugin.Client`；`schemaGetter(apiPath)` 返回 `integration.SchemaGetter`，请求 `GET {endpoints[0]}{api_path}/schema`
- [x] 4.2 `Service.initIaCPlugin`（`NewService` 中调用）：初始化 `http-plugin.Service`（交付前 `SchemaSyncer` 为 `unintegratedSyncer`）
- [x] 4.3 插件名锁：`Service.iacMutex` 进程内长期持有（`RegisterPlugins` 不再现建现关），tf 启动注册用阻塞 `Lock`，http 接口与巡检用 `TryLockWithTTL`，键 `lock.IaCPluginKey`，租约 `iacPlugin.lockLeaseSeconds`，与 P3 cloud-server 启停互斥
- [x] 4.4 新增 `cmd/hc-service/logics/iac-plugin/record`：`List` / `Get` / `Insert` / `Update` / `MarkFailed` / `Truncate`，`tf-plugin/register` 改用（删除其私有 `truncate` 与重复的写库代码）

## 5. 注册（第二阶段）

- [x] 5.1 `httpplugin.Register`：`TryLock` → 同名记录 protocol 检查 → 探活（含版本号格式校验）→ `Latest` 后直接分支（无记录 / 版本更高插入、相同且 `init` / `failed` 用本次参数覆盖重入、相同且 `enabled` / `disabled` 返回"已注册"、更低"暂不支持回滚"）→ `syncSchemaAndUpdate`（同步成功按 `InheritState` 收尾，失败 `MarkFailed` 并返回错误）
- [x] 5.2 hc-service 内部路由 `POST /api/v1/hc/iac_plugins/http/register`；`pkg/api/hc-service/iac-plugin` 的 `HTTPRegisterReq` / `HTTPPluginResult`（cloud-server 对外接口与内部路由共用同一结构）；`pkg/client/hc-service` 新增 `IaCPlugin` 客户端（前缀 `/api/v1/hc`）
- [x] 5.3 cloud-server handler `POST /iac_plugins/register`：`Validate()`（插件名字符集、`api_path` 形状、加密配置）→ 权限 `iac_plugin` / `create` → 转发
- [x] 5.4 审计：data-service DAO 事务内写，操作人为 cloud-server 透传的用户
- [ ] 5.5 编排不写单测，在 10.1 联调中验证：首次成功、已注册、init / failed 重入、拿不到锁、与 tf 重名、校验失败不写库、探活失败不写库、同步失败置 failed

## 6. 升级（第二阶段）

- [x] 6.1 `httpplugin.Upgrade(id)` → `upgradeTo`（巡检复用）：更高新增（`extension`、`memo` 复制自 L）、相同且 `init` / `failed` 重入、相同且 `enabled` / `disabled` 无操作、更低"暂不支持回滚"
- [x] 6.2 tf_provider 记录拒绝并提示滚动重启
- [x] 6.3 `disabled` 且 Schema 缺失时重做同步、仍为 `disabled`（`isSchemaExists` 交付前默认存在，同 P3）
- [x] 6.4 hc-service 内部路由 `POST /iac_plugins/http/{id}/upgrade` + 客户端 `UpgradeHTTP`；cloud-server handler `POST /iac_plugins/{id}/upgrade`，权限 `iac_plugin` / `update`，响应含 `id`、`plugin_name`、`plugin_version`、`state`
- [x] 6.5 审计：新增版本写 create，状态变化写 update，无变化（版本相同）不写库故不写审计
- [ ] 6.6 编排不写单测，在 10.1 联调中验证：新版本、版本相同、异常恢复、版本回退、继承停用、同步失败旧版本不动、拿不到锁

## 7. 修改（第二阶段）

- [x] 7.1 `httpplugin.Update(id, patch)`：状态限制（仅 `enabled` / `disabled`）、至少一个字段（请求 `Validate`）、无 `plugin_name` 字段
- [x] 7.2 未改 `api_path`：`buildExtension` 合并后（改了加密配置则整体 `ValidateCrypto`）以当前状态条件更新 `extension` / `memo`，不改 `state`、不同步
- [x] 7.3 改 `api_path`（`updateAPIPath`）：探活新地址，更高插入新版本记录（用新地址与合并后的配置）、相同则用新地址同步成功后再更新、更低拒绝
- [x] 7.4 hc-service 内部路由 `POST /iac_plugins/http/{id}/update` + 客户端 `UpdateHTTP`；cloud-server handler `POST /iac_plugins/{id}/update`，权限 `iac_plugin` / `update`；审计由 DAO 记录变更字段
- [ ] 7.5 编排不写单测，在 10.1 联调中验证：仅改备注、改加密配置、改地址（高 / 同 / 低版本）、同步失败不改记录、状态不允许、tf 插件

## 8. 巡检（第二阶段）

- [x] 8.1 `httpplugin.RunCheck`：仅 `sd.IsMaster()` 副本执行，按 `httpCheckIntervalSeconds` 分页列出所有 `protocol=http` 插件名，串行逐个走 `upgradeTo`（含 `disabled`）；`app.Run` 就绪后 `StartHTTPPluginCheck`，`finalizer` 经 `StopPluginTasks` 停止
- [x] 8.2 失败计数（内存 `failCounts` map）：`enabled` 连续 `httpCheckFailureThreshold` 次失败以 `expect_state=enabled` 置 `failed` 并写原因；`disabled` 等其他状态只告警且清零；成功清零；`failed` 探活成功走重入；失去 master 时清空计数
- [x] 8.3 拿不到锁的插件跳过本轮
- [x] 8.4 审计：后端 kit，操作人 `hcm-backend-admin`
- [ ] 8.5 master 判定、连续失败置 failed、加锁、恢复在 10.1 联调验证

## 9. 文档（第二阶段）

- [x] 9.1 `docs/api-docs/web-server/docs/scr/` 新增 `register_iac_plugin.md`、`upgrade_iac_plugin.md`、`update_iac_plugin.md`，版本占位 `v9.9.9`，含错误提示语

## 10. 联调与验证（第二阶段）

- [ ] 10.1 起一个模拟 http 插件（`/health` 返回 `plugin_version`），在 `SchemaSyncer` 占位（返回错误）与临时放行两种情况下走通：注册 → 升级 → 修改 → 停用 → 巡检异常 → 恢复
- [ ] 10.2 并发验证：手动升级、巡检、P3 启停同时触发，只有一方执行
- [ ] 10.3 与对接方联调真实 `SchemaSyncer` 与 http 取 Schema 接口，确认超时与错误形态
- [x] 10.4 `go build ./...` 与新增包 `go test` 通过
