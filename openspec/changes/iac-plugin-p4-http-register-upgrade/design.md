## Context

方案（iWiki 4047929736 第四章、5.4~5.6）：http 插件经蓝鲸 API 网关访问，网关环境地址前缀与 HCM 应用鉴权配置在 hc-service；Schema 同步入口 `integration.SchemaSyncer` 也在 hc-service（P2 已注入）。因此"访问插件 + 同步 Schema + 记录收尾"整条链路都放 hc-service，cloud-server 只做参数校验、IAM 鉴权与转发；data-service 负责记录读写与审计。注册 / 升级 / 修改 / 启停共用同一把插件名锁 `lock.IaCPluginKey`。

仓库现状：
- hc-service 已有 tf_provider 的注册（`tfplugin/register`）与运行期检查（`tfplugin/watch`），`saveInitRecord` / `markFailed`（含 message 截断）/ 状态继承的写法可直接复用；
- `pkg/serviced` 的 `IsMaster()` hc-service 同样可用；
- 访问蓝鲸网关的第三方客户端有先例 `pkg/thirdparty/api-gateway/cmdb`（配置直接用 `cc.ApiGateway`）；
- P3 的启停在 cloud-server 用 `lock.Manager.TryLockWithTTL(lock.IaCPluginKey(...))`，与 hc-service 连同一个 etcd，跨服务互斥成立。

## Goals / Non-Goals

**Goals:**
- 注册、升级、修改三条路径共用一套"探活 → 判定 → 同步 → 收尾"步骤，且与 tf_provider 共用同步入口与收尾逻辑。
- 安全边界收紧：`api_path` 无法把请求导向网关之外。
- 巡检与手动升级互不干扰（同锁）。

**Non-Goals:**
- `SchemaSyncer` 的实现。
- 回滚、删除接口。
- 插件运行态调用（launch_config 二次加密、Apply / Plan 等）。
- 巡检高可用（master 切换时失败计数丢失可接受）。

## Decisions

### D1 编排放 hc-service `logics/iac-plugin/httpplugin`，cloud-server 只转发
与 `tfplugin` 同层。对外三个方法 `Register`、`Upgrade(id)`、`Update(id, patch)`，巡检复用 `Upgrade` 的内部步骤。直接持有 `*dataservice.Client`、`*lock.EtcdMutex`、`*Prober`、`integration.SchemaSyncer`（与 tfplugin 共用同一个实例），不抽接口、不做假实现。cloud-server 的 handler 只做参数校验 + IAM 鉴权，经 `pkg/client/hc-service` 调 hc-service 内部路由。

放 hc-service 而不是 cloud-server 的理由：依赖（网关配置、探活、`SchemaSyncer`）全在 hc-service，放 cloud-server 一次注册要"探活 → 回写 init → 同步 → 回写状态"来回跨服务两次，还要为此暴露探活与同步两条内部路由；放 hc-service 只需一次转发。

### D2 判定直接写在流程里，不另立枚举
注册与升级的判定是同一组分支：`latest == nil` 或探活版本更高 → 插入新记录并同步；版本相同且 `init` / `failed` → 重入；版本相同且 `enabled` / `disabled` → 注册返回"已注册"、升级无操作；版本更低 → "暂不支持回滚"。在 `Register` / `Upgrade` 内按 `iacplugin.CompareVersion` 直接 `switch`，该报错的直接返回业务 `errf`，不定义动作枚举再二次分发。状态继承用 `pkg/iacplugin.InheritState`（tf_provider 同一实现）。分支在 9.1 联调中验证。

### D3 地址校验在 cloud-server 与 hc-service 各做一次
cloud-server 做参数形状校验（`ValidateAPIPath`）以便快速失败；hc-service 发请求前做最终校验（`JoinPluginURL`：拼接后 scheme+host 与网关前缀一致），因为网关前缀配置只在 hc-service。两处复用 `pkg/iacplugin` 同一实现。

### D4 探活与取 Schema 是 hc-service 进程内能力，不暴露路由
`Prober.Probe(kt, apiPath)` 只取版本；取 Schema 由 httpplugin 实现 `integration.SchemaGetter`（同一个网关 http 客户端）交给 `SchemaSyncer.Sync`，与 tf_provider 传 `executor.GetProviderSchema` 完全对称。

### D5 网关访问不跟随重定向、固定超时、响应体限长
`http.Client` 设置 `CheckRedirect` 返回 `http.ErrUseLastResponse`，超时默认 10 秒可配置；响应体上限 1MB（常量）。网关配置直接用 `cc.ApiGateway`（顶层 `iacGateway`，与 `cmdb` 同形）；`endpoints` 为空表示不使用 http 插件，配置时只允许一个地址（它是地址校验的基准）。

### D6 失败计数只在巡检里
方案流程图中"连续失败达 3 次"属于巡检行为；手动升级接口探活失败直接返回错误，不计数。计数是巡检 goroutine 内的 `map[pluginName]int`，探活成功清零；仅 master 巡检，切换 master 时计数丢失，最多多等 3 个周期，可接受。

### D7 巡检用 master 判定，而不是每个副本抢锁
hc-service 用 `IsMaster()` 保证只有一个副本巡检，避免多副本重复探活放大对插件后端的压力；锁仍然保留，防止与手动升级、启停并发。巡检配置（间隔、阈值）放 hc-service `iacPlugin`，cloud-server 不新增配置。

### D8 新版本记录的 `extension` 沿用旧记录并按需覆盖
升级新增记录时，`extension`（`api_path`、`crypto_*`）默认复制最高版本记录；`Update` 触发的升级分支用新 `api_path` 覆盖。这样升级不要求管理员重新提交加密配置。

### D9 `crypto_config` 校验与审计
公钥解析用标准库（`encoding/pem` + `x509.ParsePKIXPublicKey`），断言 `*rsa.PublicKey` 且 `N.BitLen() >= 3072`。公钥非机密，审计可记录。审计由 P1 DAO 在事务内写：手动操作的操作人是 cloud-server 透传的用户，巡检用后端 kit 的 `hcm-backend-admin`（与 P2 一致）。

### D10 错误语义
- 拿不到锁：`errf.Aborted` 类，信息"注册进行中 / 升级进行中 / 变更进行中"；
- 已注册、暂不支持回滚：`errf.InvalidParameter` 类业务错误，信息明确；
- 地址校验失败：`errf.InvalidParameter`，不发请求；探活失败（含超时）：`errf.Aborted`，信息带插件 `message`；
- 同步失败：记录置 `failed`，接口返回包含 `message` 的错误。

## Risks / Trade-offs

- [`api_path` 校验不严导致 SSRF] → hc-service 最终校验 scheme+host 与配置前缀一致、禁止 `..`、不跟随重定向；方案 6.1 已列为待确认的安全边界，上线前评审。
- [巡检探活对大量 http 插件造成压力] → 仅 master、串行、单次超时 10 秒；插件数量预计很少。
- [cloud-server → hc-service 的同步调用超过网关超时] → 一次转发内含探活（10 秒）+ 同步，需小于上游网关超时；同步超时在 `SchemaSyncer` 交付时确认。
- [判定分支没有纯函数单测] → 分支少且集中在两个方法里，9.1 联调逐条覆盖。
- [升级期间旧版本继续被使用，新增资源类型的请求会校验失败] → 方案已知限制。

## Migration Plan

1. 合入配置（网关、探活超时、巡检间隔与阈值）、校验函数、`Prober`、`InheritState` 公共化；hc-service 暂不接入、不暴露路由。
2. 合入 hc-service 编排、内部路由、巡检与 cloud-server 转发 handler；巡检默认开启，无 http 插件时为空操作。
3. 回滚：回退版本即可，注册表记录保留。

## Open Questions

- `api_path` 构成规则与校验严格程度（方案 6.1，需协议负责人确认）。
- http 插件取 Schema 的接口路径与响应格式（需协议负责人确认）。
- 手动升级接口的探活失败是否也应计入失败次数：方案图示意为否，本 change 按"否"实现。
