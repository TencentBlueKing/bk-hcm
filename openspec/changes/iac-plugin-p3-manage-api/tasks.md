## 1. 对接接口与权限

- [x] 1.1 Schema 是否存在先放在 switch handler 的 `isIaCPluginSchemaExists`；对接方交付查询接口前默认返回 `true` 并打 warn：交付前 Schema 本就不存在，返回 false / 错误会让"启用"永久失败，无法联调
- [x] 1.2 `pkg/iam/meta/resource.go` 新增 `IaCPlugin` 资源类型，并加入合法集合（已在 P1 完成）
- [x] 1.3 `cmd/auth-server/service/auth/adaptor.go`、`gen_id.go` 增加 `iac_plugin` 映射（find / update / create → `global_configuration`，已在 P1 完成）

## 2. API 结构体

- [x] 2.1 `pkg/api/cloud-server` 新增 list 响应 `IaCPluginListResult`（请求复用通用 `ListReq`）与 switch 请求 `IaCPluginSwitchReq`（`plugin_name` + `state`），带 `Validate()`
- [x] 2.2 `Validate()`：`state` 仅允许 `enabled` / `disabled`；非法值由 handler 单测覆盖

## 3. 状态规则

- [x] 3.1 `cmd/cloud-server/service/iac-plugin`：`decideSwitch`（无需处理 / 批量改状态 / 拒绝及原因）
- [x] 3.2 表驱动单测覆盖方案表格全部格子：enabled/disabled/failed/init × 目标 disabled/enabled，以及 Schema 缺失
- [x] 3.3 可用性判定函数本阶段不实现，运行态接入请求入口时再加

## 4. cloud-server 服务

- [x] 4.1 新增 `cmd/cloud-server/service/iac-plugin`：`InitService`，注册 list 与 switch 路由，并在 `service.go` 注册
- [x] 4.2 cloud-server 直接复用进程级锁 `lock.Manager.TryLockWithTTL`（已在 `service.NewService` 中基于 `cc.CloudServer().Service.Etcd` 初始化）
- [x] 4.3 list handler：鉴权（`iac_plugin` / `find`）→ 调 data-service 列表
- [x] 4.4 switch handler：鉴权（`update`）→ 按请求 `plugin_name` 执行 `TryLockWithTTL` → 查同名全部记录并取当前最高版本 → （仅 `disabled → enabled` 时）`isIaCPluginSchemaExists` → `decideSwitch` → 批量改状态（DAO 事务内自动写 update 审计，幂等时不调更新故不写审计） → `defer` 释放锁
- [x] 4.5 拿不到锁返回"进行中"错误；锁用 `defer` 释放
- [x] 4.6 handler 单测（`iac_plugin_test.go`）：列表鉴权与成功返回、switch 参数校验 / 无权限；状态流转由 `decideSwitch` 表驱动单测完整覆盖

## 5. 文档

- [x] 5.1 `docs/api-docs/web-server/docs/scr/` 新增 list 与 switch 两个接口文档，版本占位 `v9.9.9`
- [x] 5.2 文档内含请求 / 响应示例与错误码说明（含"进行中"、"先恢复"的提示语）

## 6. 验证

- [ ] 6.1 本地联调：创建若干 `iac_plugin` 记录（含多版本），走通列表、停用、启用、幂等、`failed` 拒绝启用、并发持锁
- [ ] 6.2 确认状态变化后审计记录内容正确
- [x] 6.3 `go build ./...` 与新增包 `go test` 通过
